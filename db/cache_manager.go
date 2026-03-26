/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	cache "github.com/Comcast/goburrow-cache"
	"github.com/go-akka/configuration"
	"github.com/gocql/gocql"
	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/util"
	log "github.com/sirupsen/logrus"
)

const (
	DEFAULT_TENANT_ID   = "COMCAST"
	CACHE_KEY_DELIMITER = "::"
)

var (
	Conf          *configuration.Config
	syncCacheLock sync.Mutex
)

// ConfigInjection - dependency injection
func ConfigInjection(conf *configuration.Config) {
	Conf = conf
}

// CacheChangeNotifier is an interface for notifications on cache changed events
type CacheChangeNotifier interface {
	Notify(tenantId string, tableName string, changedKey string, operation OperationType)
}

// CacheRefreshTask background task to refresh cache
type CacheRefreshTask struct {
	lastRefreshedTimestamp time.Time
	refreshAttemptsLeft    int32
	stopped                chan bool
	ticker                 *time.Ticker
}

func (t *CacheRefreshTask) doSyncChanges() {
	syncCacheLock.Lock()
	defer syncCacheLock.Unlock()
	now := time.Now().UTC()
	log.Debugf("doSyncChanges. starting cache update [%v - %s]", t.lastRefreshedTimestamp.Format(time.RFC3339), now.Format(time.RFC3339))
	if t.refreshAttemptsLeft == 0 {
		log.Debug("attempting full cache refresh")
		// load all data
		t.lastRefreshedTimestamp = now
		failedTables := GetCacheManager().RefreshAll(DEFAULT_TENANT_ID)
		if len(failedTables) == 0 {
			t.refreshAttemptsLeft++
		} else {
			log.Errorf("failed to refresh cache for table(s): %v", failedTables)
		}
	} else {
		// load only changed data
		_, err := GetCacheManager().SyncChanges(t.lastRefreshedTimestamp, now, true)
		if err == nil {
			t.lastRefreshedTimestamp = now
			t.refreshAttemptsLeft = cacheManager.settings.retryCountUntilFullRefresh
			//	Ws.UpdateCacheSyncMetrics(true)
		} else {
			log.Errorf("failed to sync cache changes: %v", err)
			t.refreshAttemptsLeft--
			//Ws.UpdateCacheSyncMetrics(false)
		}
	}
}

func (t *CacheRefreshTask) Run() {
	go func() {
		for {
			select {
			case <-t.stopped:
				log.Debug("stopping cache refresh task")
				return
			case <-t.ticker.C:
				now := time.Now().UTC()
				log.Debugf("starting cache update [%v - %s]", t.lastRefreshedTimestamp.Format(time.RFC3339), now.Format(time.RFC3339))
				if t.refreshAttemptsLeft == 0 {
					log.Debug("attempting full cache refresh")
					// load all data
					t.lastRefreshedTimestamp = now
					failedTables := GetCacheManager().RefreshAll(DEFAULT_TENANT_ID)
					if len(failedTables) == 0 {
						t.refreshAttemptsLeft++
					} else {
						log.Errorf("failed to refresh cache for table(s): %v", failedTables)
					}
				} else {
					// load only changed data
					_, err := GetCacheManager().SyncChanges(t.lastRefreshedTimestamp, now, true)
					if err == nil {
						t.lastRefreshedTimestamp = now
						t.refreshAttemptsLeft = cacheManager.settings.retryCountUntilFullRefresh
					} else {
						log.Errorf("failed to sync cache changes: %v", err)
						t.refreshAttemptsLeft--
					}
				}
			}
		}
	}()
}

func (t *CacheRefreshTask) Stop() {
	t.ticker.Stop()
	t.stopped <- true
}

// CacheSettings settings for the cache
type CacheSettings struct {
	// Duration of tick for which we check for changed keys in cassandra
	tickDuration int32

	// Changed keys retry load count until a full refresh is attempted
	retryCountUntilFullRefresh int32

	changedKeysTimeWindowSize int32

	// Indicates whether or not cache keys are elapsing
	reloadCacheEntries bool

	// Timeout for cache keys to elapse
	reloadCacheEntriesTimeout int64

	// Timeunit for cache keys elapsing timeout (NANOSECONDS, MICROSECONDS, MILLISECONDS, SECONDS, MINUTES, HOURS, DAYS)
	reloadCacheEntriesTimeUnit string

	// Number of entries that until exceeded will be processed by single thread (namely rules)
	// if exceeded will be processed by Runtime.getAvailableProcessors() threads and though splited into chunks
	numberOfEntriesToProcessSequentially int32

	// Keys chunk size that is used to load keys during initial cache load
	keysetChunkSizeForMassCacheLoad int32

	// Indicates whether or not to copy or clone the cached data
	// since the operation is causing performance issue
	cloneDataEnabled bool

	applicationCacheEnabled bool

	groupServiceExpireAfterAccess int64

	groupServiceRefreshAfterWrite int64
}

// CacheStats statistics for the cache.LoadingCache
type CacheStats struct {
	DaoRefreshTime time.Time     `json:"daoRefreshTime"`
	CacheSize      int           `json:"cacheSize"`
	RequestCount   uint64        `json:"requestCount"`
	NonAbsentCount int           `json:"nonAbsentCount"`
	EvictionCount  uint64        `json:"evictionCount"`
	HitRate        float64       `json:"hitRate"`
	MissRate       float64       `json:"missRate"`
	TotalLoadTime  time.Duration `json:"totalLoadTime"`
}

// Statistics cache statistics of all tables for a tenant
type Statistics struct {
	TableStats map[string]CacheStats `json:"Statistics"`
}

type CacheInfo struct {
	cache          cache.LoadingCache
	tenantId       string
	DaoRefreshTime time.Time
}

// Cache of all tables for a tenant. Key is table name, value is CacheInfo
type TableCacheInfo map[string]CacheInfo

// CacheManager a cache manager
type CacheManager struct {
	settings                  CacheSettings
	cacheChangeNotifier       atomic.Value
	refreshCacheTask          CacheRefreshTask
	applicationCacheEnabled   bool
	tableCaches               map[string]TableCacheInfo // key is tenantId, e.g. "COMCAST"
	applicationCaches         map[string]cache.Cache    // key is tenantId
	groupServiceFeatureCaches map[string]cache.Cache    // key is tenantId
}

// CacheManager a cache manager
var cacheManager CacheManager
var initOnce sync.Once
var refreshCacheMutex sync.Mutex
var grpCacheLoadFunc cache.LoaderFunc
var tenants []string = []string{DEFAULT_TENANT_ID}

// GetTenants returns list of tenants. TODO: Extend to support multiple tenants in the future
func GetTenants() []string {
	return tenants
}

// GetCacheManager Initializes a CacheManager
func GetCacheManager() *CacheManager {
	initOnce.Do(func() {
		cacheManager = CacheManager{}
		cacheManager.tableCaches = make(map[string]TableCacheInfo)
		cacheManager.applicationCaches = make(map[string]cache.Cache)
		cacheManager.groupServiceFeatureCaches = make(map[string]cache.Cache)

		if Conf == nil {
			// Handle ServerConfig not initialized yet
			cacheManager.settings.tickDuration = 60000
			cacheManager.settings.retryCountUntilFullRefresh = 10
			cacheManager.settings.changedKeysTimeWindowSize = 900000
			cacheManager.settings.reloadCacheEntries = false
			cacheManager.settings.reloadCacheEntriesTimeout = 1
			cacheManager.settings.reloadCacheEntriesTimeUnit = "DAYS"
			cacheManager.settings.numberOfEntriesToProcessSequentially = 10000
			cacheManager.settings.keysetChunkSizeForMassCacheLoad = 500
			cacheManager.settings.cloneDataEnabled = false
			cacheManager.settings.applicationCacheEnabled = false
			cacheManager.settings.groupServiceExpireAfterAccess = 240
			cacheManager.settings.groupServiceRefreshAfterWrite = 240
		} else {
			cacheManager.settings.tickDuration = Conf.GetInt32("xconfwebconfig.xconf.cache_tickDuration", 60000)
			cacheManager.settings.retryCountUntilFullRefresh = Conf.GetInt32("xconfwebconfig.xconf.cache_retryCountUntilFullRefresh", 10)
			cacheManager.settings.changedKeysTimeWindowSize = Conf.GetInt32("xconfwebconfig.xconf.cache_changedKeysTimeWindowSize", 900000)
			cacheManager.settings.reloadCacheEntries = Conf.GetBoolean("xconfwebconfig.xconf.cache_reloadCacheEntries", false)
			cacheManager.settings.reloadCacheEntriesTimeout = Conf.GetInt64("xconfwebconfig.xconf.cache_reloadCacheEntriesTimeout", 1)
			cacheManager.settings.reloadCacheEntriesTimeUnit = Conf.GetString("xconfwebconfig.xconf.cache_reloadCacheEntriesTimeUnit", "DAYS")
			cacheManager.settings.numberOfEntriesToProcessSequentially = Conf.GetInt32("xconfwebconfig.xconf.cache_numberOfEntriesToProcessSequentially", 10000)
			cacheManager.settings.keysetChunkSizeForMassCacheLoad = Conf.GetInt32("xconfwebconfig.xconf.cache_keysetChunkSizeForMassCacheLoad", 500)
			cacheManager.settings.cloneDataEnabled = Conf.GetBoolean("xconfwebconfig.xconf.cache_clone_data_enabled", false)
			cacheManager.settings.applicationCacheEnabled = Conf.GetBoolean("xconfwebconfig.xconf.application_cache_enabled", false)
			cacheManager.settings.groupServiceExpireAfterAccess = Conf.GetInt64(fmt.Sprintf("xconfwebconfig.%v.cache_expire_after_access_in_mins", Conf.GetString("xconfwebconfig.xconf.group_service_name")))
			cacheManager.settings.groupServiceRefreshAfterWrite = Conf.GetInt64(fmt.Sprintf("xconfwebconfig.%v.cache_refresh_after_write_in_mins", Conf.GetString("xconfwebconfig.xconf.group_service_name")))
		}
		cacheManager.applicationCacheEnabled = cacheManager.settings.applicationCacheEnabled

		tenants := GetTenants()
		for _, tenantId := range tenants {
			// Initialize cache for each table for the tenant
			tableCaches := make(TableCacheInfo)
			cacheManager.tableCaches[tenantId] = tableCaches

			for tableName, tableInfo := range tableConfig {
				if !tableInfo.CacheData {
					continue // No caching for this table
				}

				// Generate a load function for the table
				loadFn := generateLoadFunction(tenantId, tableName)

				// Create a LoadingCache for each table
				var loadingCache cache.LoadingCache
				if cacheManager.settings.reloadCacheEntries {
					freshAfterWrite := getDuration(
						cacheManager.settings.reloadCacheEntriesTimeout, cacheManager.settings.reloadCacheEntriesTimeUnit)
					loadingCache = cache.NewLoadingCache(
						loadFn,
						cache.WithMaximumSize(0), // Unlimited number of entries in the cache.
						cache.WithRefreshAfterWrite(freshAfterWrite), // Expire entries after specified duration since last created.
					)
				} else {
					loadingCache = cache.NewLoadingCache(
						loadFn,
						cache.WithMaximumSize(0), // Unlimited number of entries in the cache.
					)
				}
				tableCaches[tableName] = CacheInfo{tenantId: tenantId, cache: loadingCache}
			}

			// Initialize application cache for each tenant
			cacheManager.applicationCaches[tenantId] = cache.New(cache.WithMaximumSize(0))

			// Initialize group service feature tags cache for each tenant
			cacheManager.groupServiceFeatureCaches[tenantId] = cache.NewLoadingCache(
				GetGrpCacheLoadFunc(),
				cache.WithMaximumSize(100),
				cache.WithExpireAfterAccess(time.Duration(cacheManager.settings.groupServiceExpireAfterAccess)*time.Minute),
				cache.WithRefreshAfterWrite(time.Duration(cacheManager.settings.groupServiceRefreshAfterWrite)*time.Minute),
			)
		}

		cc, ok := GetDatabaseClient().(*CassandraClient)
		if ok && !cc.IsTestOnly() {
			// Initiate precaching
			cacheManager.initiatePrecaching()

			// Start background task to refresh cache
			cacheManager.refreshCacheTask = CacheRefreshTask{
				lastRefreshedTimestamp: time.Now().UTC(),
				refreshAttemptsLeft:    cacheManager.settings.retryCountUntilFullRefresh,
				stopped:                make(chan bool),
				ticker:                 time.NewTicker(time.Duration(cacheManager.settings.tickDuration) * time.Millisecond),
			}
			cacheManager.refreshCacheTask.Run()
		}
	})

	return &cacheManager
}

// SetCacheChangeNotifier sets a notifier to be called on cache changed events
func (cm *CacheManager) SetCacheChangeNotifier(notifier CacheChangeNotifier) {
	if notifier == nil {
		panic("SetCacheChangeNotifier: notifier cannot be nil")
	}
	cm.cacheChangeNotifier.Store(notifier)
}

// GetCacheStats returns cache statistics for the specified tenant and table
func (cm CacheManager) GetCacheStats(tenantId string, tableName string) (*CacheStats, error) {
	cacheInfo, err := cm.getCacheInfo(tenantId, tableName)
	if err != nil {
		return nil, err
	}

	return cacheInfo.getStats(), nil
}

func (cacheInfo *CacheInfo) getStats() *CacheStats {
	// Get current cache stats
	stats := cache.Stats{}
	cacheInfo.cache.Stats(&stats)

	cacheStats := CacheStats{
		DaoRefreshTime: cacheInfo.DaoRefreshTime,
		CacheSize:      cacheInfo.cache.Size(),
		RequestCount:   stats.RequestCount(),
		EvictionCount:  stats.EvictionCount,
		HitRate:        stats.HitRate(),
		MissRate:       stats.MissRate(),
		TotalLoadTime:  stats.TotalLoadTime,
	}

	// NonAbsentCount is effectively CacheSize as the cache does not store nil values.
	// Avoiding GetAllValues() which is O(N) and expensive.
	cacheStats.NonAbsentCount = cacheStats.CacheSize

	return &cacheStats
}

// GetStatistics returns cache statistics of all tables for the specified tenant
func (cm CacheManager) GetStatistics(tenantId string) *Statistics {
	statistics := Statistics{TableStats: make(map[string]CacheStats)}

	tableCaches, ok := cm.tableCaches[tenantId]
	if !ok {
		log.WithFields(log.Fields{"tenantId": tenantId}).Warnf("GetStatistics called for an unknown tenant")
		return &statistics
	}

	for tableName, cacheInfo := range tableCaches {
		// Get current cache stats
		cacheStats := cacheInfo.getStats()
		statistics.TableStats[tableName] = *cacheStats
	}

	return &statistics
}

// getCacheInfo returns CacheInfo for the specified tenant and table
func (cm CacheManager) getCacheInfo(tenantId string, tableName string) (*CacheInfo, error) {
	tableCaches, ok := cm.tableCaches[tenantId]
	if !ok {
		err := fmt.Errorf("cache not found or configured for tenant %s", tenantId)
		return nil, err
	}

	cacheInfo := tableCaches[tableName]
	if cacheInfo.cache == nil {
		err := fmt.Errorf("cache not found or configured for tenant %s table %s", tenantId, tableName)
		return nil, err
	}
	return &cacheInfo, nil
}

// getCache returns a LoadingCache for the specified tenant and table
func (cm CacheManager) getCache(tenantId string, tableName string) (cache.LoadingCache, error) {
	cacheInfo, err := cm.getCacheInfo(tenantId, tableName)
	if err != nil {
		return nil, err
	}
	if cacheInfo.tenantId != tenantId {
		err := fmt.Errorf("cache tenantId mismatch for tenant %s table %s: actual tenantId is %s", tenantId, tableName, cacheInfo.tenantId)
		return nil, err
	}
	return cacheInfo.cache, nil
}

func (cm CacheManager) ForceSyncChanges() {
	cm.refreshCacheTask.doSyncChanges()
}

// Preloading caches
func (cm CacheManager) initiatePrecaching() {
	// Just a debugging convenience, load the tables in the same order every time
	tableList := make([]string, 0, len(tableConfig))
	for tableName := range tableConfig {
		tableList = append(tableList, tableName)
	}
	sort.Strings(tableList)

	var wg sync.WaitGroup

	precachingStart := time.Now()

	for _, tenantId := range tenants {
		fields := log.Fields{"tenantId": tenantId}
		log.WithFields(fields).Debug("initializing cache...")

		for _, tableName := range tableList {
			wg.Add(1)

			go func(tableName string) {
				defer wg.Done()

				tableInfo := tableConfig[tableName]
				if tableInfo.CacheData {
					log.WithFields(fields).Debugf("precaching for table %s...", tableName)
				} else {
					log.WithFields(fields).Debugf("skipped precaching table %s", tableName)
					return
				}

				start := time.Now()
				var entries map[string]interface{}
				var err error
				if tableInfo.IsCompressAndSplit() {
					entries, err = GetCompressingDataDao().GetAllAsMap(tenantId, tableName, true)
				} else {
					entries, err = GetSimpleDao().GetAllAsMap(tenantId, tableName, 0)
				}

				if err != nil {
					log.WithFields(fields).Fatalf("failed to preload cache for table %s: %v", tableName, err)
					panic(err)
				}

				cache, err := cm.getCache(tenantId, tableName)
				if err != nil {
					log.WithFields(fields).Fatalf("failed to preload cache for table %s: %v", tableName, err)
					panic(err)
				}

				for k, v := range entries {
					cache.Put(k, v)
				}

				log.WithFields(fields).Debugf("table %s precached %v entries in %v", tableName, len(entries), time.Since(start))
			}(tableName)
		}
	}

	wg.Wait()

	log.WithFields(log.Fields{"duration": time.Since(precachingStart).String()}).Info("precache duration")
}

func (cm CacheManager) GetChangedKeysTimeWindowSize() int32 {
	return cm.settings.changedKeysTimeWindowSize
}

// RefreshAll Refresh all caches and return list of table names which were not refreshed
func (cm CacheManager) RefreshAll(tenantId string) []string {
	refreshCacheMutex.Lock()
	defer refreshCacheMutex.Unlock()

	log.WithFields(log.Fields{"tenantId": tenantId}).Debug("starting cache refresh...")

	var failedToRefreshTables []string
	for _, tableInfo := range tableConfig {
		var err error
		var start = time.Now()
		if tableInfo.CacheData {
			err = GetCachedSimpleDao().RefreshAll(tenantId, tableInfo.TableName)
		} else {
			continue // Skip since data is not cache
		}

		duration := time.Since(start)

		if err == nil {
			cache, _ := cm.getCache(tenantId, tableInfo.TableName)
			log.WithFields(log.Fields{"tenantId": tenantId}).Debugf("cache refreshed for table %s: precached %v entries in %v", tableInfo.TableName, cache.Size(), duration)
		} else {
			log.WithFields(log.Fields{"tenantId": tenantId}).Errorf("failed to refresh cache for table %s: %v", tableInfo.TableName, err)
			failedToRefreshTables = append(failedToRefreshTables, tableInfo.TableName)
		}
	}

	return failedToRefreshTables
}

// Refresh Refresh cache for the specified table
func (cm CacheManager) Refresh(tenantId string, tableName string) error {
	refreshCacheMutex.Lock()
	defer refreshCacheMutex.Unlock()

	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		log.WithFields(log.Fields{"tenantId": tenantId}).Errorf("failed to refresh cache table %s: %v", tableName, err)
		return err
	}

	if tableInfo.CacheData {
		err = GetCachedSimpleDao().RefreshAll(tenantId, tableInfo.TableName)
		if err != nil {
			log.WithFields(log.Fields{"tenantId": tenantId}).Errorf("failed to refresh cache for table %s: %v", tableInfo.TableName, err)
			return err
		}
	} else {
		log.WithFields(log.Fields{"tenantId": tenantId}).Debugf("unable to refresh cache table %s, data is not cached", tableInfo.TableName)
	}

	return nil
}

// Invalidate Evict an entry from cache
func (cm CacheManager) Invalidate(tenantId string, tableName string, key string) error {
	cache, err := cm.getCache(tenantId, tableName)
	if err != nil {
		return err
	}
	if util.IsBlank(key) {
		return fmt.Errorf("failed to invalidate cache for tenant %s table %s: key is blank", tenantId, tableName)
	}

	cache.Invalidate(key)
	cm.ApplicationCacheDeleteAll(tenantId, tableName) // Invalidate application cache entries for this table

	return nil
}

// InvalidateAll Evict all entries from cache
func (cm CacheManager) InvalidateAll(tenantId string, tableName string) error {
	cache, err := cm.getCache(tenantId, tableName)
	if err != nil {
		return err
	}

	cache.InvalidateAll()
	cm.ApplicationCacheDeleteAll(tenantId, tableName) // Invalidate application cache entries for this table
	return nil
}

// SyncChanges Updates changes for given time-window defined by
// start (lower bound, inclusive) and end (upper bound, exclusive)
func (cm CacheManager) SyncChanges(startTime time.Time, endTime time.Time, apply bool) (changedList []interface{}, err error) {
	startTS := util.GetTimestamp(startTime)
	currentRowKey := startTS - (startTS % int64(cm.settings.changedKeysTimeWindowSize))

	endTS := util.GetTimestamp(endTime)
	endRowKey := endTS - (endTS % int64(cm.settings.changedKeysTimeWindowSize))

	startUuid, err := util.UUIDFromTime(startTS, 0, 0)
	if err != nil {
		return nil, err
	}

	ranges := make(map[int64]*RangeInfo)
	ranges[currentRowKey] = &RangeInfo{StartValue: startUuid}
	currentRowKey += int64(cm.settings.changedKeysTimeWindowSize)
	for currentRowKey <= endRowKey {
		ranges[currentRowKey] = nil
		currentRowKey += int64(cm.settings.changedKeysTimeWindowSize)
	}

	// Load all changes from DB
	log.Debugf("sync cache, getting changed keys [%v - %v]: %v", startTS, endTS, buildLogForRanges(ranges))
	changedList, err = cm.loadChanges(ranges)
	if err != nil {
		return nil, err
	}

	// Apply changes to cache
	if apply {
		log.Debugf("sync cache, getting changed keys [%v - %v]: %v", startTS, endTS, buildLogForRanges(ranges))
		err = cm.applyChanges(changedList)
	}

	return changedList, err
}

func (cm CacheManager) loadChanges(ranges map[int64]*RangeInfo) ([]interface{}, error) {
	var result []interface{}

	for rowKey, rangeInfo := range ranges {
		// TODO check if GetRange support tenantId is empty for logs table
		list, err := GetListingDao().GetRange("", TABLE_CHANGE_EVENTS, rowKey, rangeInfo)
		if err != nil {
			return nil, err
		}
		result = append(result, list...)
	}

	return result, nil
}

// Apply changed data to keep cache in sync
func (cm CacheManager) applyChanges(changedDataList []interface{}) error {
	changedTables := map[string]util.Set{}   // Keep track of tables which were changed; tenantId -> Set of tableNames
	tablesToRefresh := map[string]util.Set{} // Keep track of tables which need to be refreshed; tenantId -> Set of tableNames

	// Ensure whatever's already in the list will get refresh
	defer func() {
		for tenantId, tables := range tablesToRefresh {
			refreshTables(tenantId, tables.ToSlice())
		}
	}()

	for _, obj := range changedDataList {
		data := *obj.(*ChangedData)
		fields := log.Fields{"tenantId": data.TenantId}

		// Fixed issue w/ Xconf AS wrote changedKey field with double quotation marks ("")
		l := len(data.ChangedKey)
		if l > 0 && data.ChangedKey[0] == '"' && data.ChangedKey[l-1] == '"' {
			data.ChangedKey = data.ChangedKey[1 : l-1]
		}

		if data.TenantId == "" || data.CfName == "" || data.ChangedKey == "" || data.Operation == "" || data.ValidCacheSize == 0 {
			log.WithFields(fields).Error("unable to load changed data")
			continue
		}
		// Skip entry if it was originated by the same server
		if data.ServerOriginId != "" && data.ServerOriginId == common.ServerOriginId() {
			log.WithFields(fields).Debugf("sync cache, skipping %v for table '%v': %v", data.Operation, data.CfName, data.ChangedKey)
			continue
		}

		log.WithFields(fields).Debugf("sync cache, processing %v for table '%v': %v", data.Operation, data.CfName, data.ChangedKey)

		tableInfo, err := GetTableInfo(data.CfName)
		if err != nil {
			log.WithFields(fields).Errorf("failed to apply changed data %v: %v", data.ChangedKey, err)
			return err
		}

		cache, err := cm.getCache(data.TenantId, data.CfName)
		if err != nil {
			log.WithFields(fields).Errorf("failed to apply changed data %v: %v", data.ChangedKey, err)
			return err
		}

		switch data.Operation {
		case CREATE_OPERATION, UPDATE_OPERATION: // fetch modified entry to cache
			cache.Refresh(data.ChangedKey)
			val, err := cache.Get(data.ChangedKey)
			if err != nil || val == nil {
				log.WithFields(fields).Errorf("failed to apply changed data %v: %v", data.ChangedKey, err)
				return err
			}
		case DELETE_OPERATION: // evict entry from cache
			cache.Invalidate(data.ChangedKey)
		case TRUNCATE_OPERATION:
			cache.InvalidateAll()
		}

		changedSet := changedTables[data.TenantId]
		if changedSet == nil {
			changedSet = util.Set{}
			changedTables[data.TenantId] = changedSet
		}
		changedSet.Add(tableInfo.TableName)

		cacheSize := cache.Size()
		if cacheSize < int(data.ValidCacheSize) {
			log.WithFields(fields).Warnf("cache size difference, got %v instead of %v, scheduling full refresh for %v", cacheSize, data.ValidCacheSize, tableInfo.TableName)

			refreshSet := tablesToRefresh[data.TenantId]
			if refreshSet == nil {
				refreshSet = util.Set{}
				tablesToRefresh[data.TenantId] = refreshSet
			}
			refreshSet.Add(tableInfo.TableName)
		}
	}

	// Invalidate application cache entries for tables that were changed
	for tenantId, tables := range changedTables {
		for _, tableName := range tables.ToSlice() {
			cm.ApplicationCacheDeleteAll(tenantId, tableName)
		}
	}

	return nil
}

// Write changed data to ChangeEvents table asynchronously
func (cm CacheManager) WriteCacheLog(tenantId string, tableName string, changedKey string, operation OperationType) {
	if cache, err := cm.getCache(tenantId, tableName); err == nil {
		cm.writeCacheLog(tenantId, tableName, changedKey, operation, int32(cache.Size()))
	} else {
		log.WithFields(log.Fields{"tenantId": tenantId}).Errorf("failed to write cache changed log: %v", err)
	}
}

// Write changed data to ChangeEvents table async
func (cm CacheManager) writeCacheLog(tenantId string, tableName string, changedKey string, operation OperationType, cacheSize int32) {
	go func() {
		currentTS := util.GetTimestamp(time.Now().UTC())
		key := currentTS - (currentTS % int64(cm.settings.changedKeysTimeWindowSize))

		changedData := ChangedData{
			ColumnName:     gocql.TimeUUID(),
			CfName:         tableName,
			ChangedKey:     changedKey,
			Operation:      operation,
			ValidCacheSize: cacheSize,
			UserName:       "XConf",
			ServerOriginId: common.ServerOriginId(),
			TenantId:       tenantId,
		}

		jsonData, err := json.Marshal(changedData)
		if err == nil {
			log.WithFields(log.Fields{"tenantId": tenantId}).Debugf("write cache changed log for table %s: %v %v %v", tableName, operation, key, changedData.ValidCacheSize)
			// TODO: ensure SetOne support empty tenantId
			err = GetListingDao().SetOne("", TABLE_CHANGE_EVENTS, key, changedData.ColumnName, jsonData)
		}

		if err != nil {
			log.WithFields(log.Fields{"tenantId": tenantId}).Errorf("failed to write cache changed log for table %s key %s: %v", tableName, changedKey, err)
		}

		// Send changed event to a registered observer, if one exists
		if val := cm.cacheChangeNotifier.Load(); val != nil {
			if notifier, ok := val.(CacheChangeNotifier); ok {
				notifier.Notify(tenantId, tableName, changedKey, operation)
			}
		}
	}()
}

// Generates a load function for the cache
func generateLoadFunction(tenantId string, tableName string) func(k cache.Key) (cache.Value, error) {
	loadFn := func(k cache.Key) (cache.Value, error) {
		tableInfo, err := GetTableInfo(tableName)
		if err != nil {
			return nil, err
		}

		// Use the appropriate DAO based on compression policy
		if tableInfo.IsCompressAndSplit() {
			ret, err := GetCompressingDataDao().GetOne(tenantId, tableName, k.(string))
			return ret, err
		} else if tableInfo.IsCompressOnly() {
			twoKeys, err := NewTwoKeysFromString(k.(string))
			if err != nil {
				return nil, err
			}
			return GetListingDao().GetOne(tenantId, tableName, twoKeys.Key, twoKeys.Key2)
		} else {
			return GetSimpleDao().GetOne(tenantId, tableName, k.(string))
		}
	}

	return loadFn
}

// Valid value for timeUnit: NANOSECONDS, MICROSECONDS, MILLISECONDS, SECONDS, MINUTES, HOURS, DAYS
func getDuration(duration int64, timeUnit string) time.Duration {
	var value time.Duration

	switch strings.ToUpper(timeUnit) {
	case "NANOSECONDS":
		value = time.Duration(duration) * time.Nanosecond
	case "MICROSECONDS":
		value = time.Duration(duration) * time.Microsecond
	case "MILLISECONDS":
		value = time.Duration(duration) * time.Millisecond
	case "SECONDS":
		value = time.Duration(duration) * time.Second
	case "MINUTES":
		value = time.Duration(duration) * time.Minute
	case "HOURS":
		value = time.Duration(duration) * time.Hour
	case "DAYS":
		value = time.Duration(duration*24) * time.Hour
	default:
		err := fmt.Errorf("invalid value for param timeUnit: '%v'", timeUnit)
		panic(err)
	}

	return value
}

func buildLogForRanges(ranges map[int64]*RangeInfo) string {
	var buf strings.Builder

	for rowKey, columnRange := range ranges {
		if buf.Len() == 0 {
			fmt.Fprintf(&buf, "Row Key: %d; ", rowKey)
		} else {
			fmt.Fprintf(&buf, " @ Row Key: %d; ", rowKey)
		}

		if columnRange == nil {
			buf.WriteString("Start Column Name: nil")
		} else {
			fmt.Fprintf(&buf, "Start Column Name: %v", columnRange.StartValue)
		}
	}

	return buf.String()
}

// Refresh cache for these tables
func refreshTables(tenantId string, tableNames []string) {
	for _, tableName := range tableNames {
		err := GetCachedSimpleDao().RefreshAll(tenantId, tableName)
		if err != nil {
			log.WithFields(log.Fields{"tenantId": tenantId}).Errorf("failed to refresh cache for %s: %v", tableName, err)
		}
	}
}

// ApplicationCacheGet value for the specified table and key
func (cm CacheManager) ApplicationCacheGet(tenantId string, tableName string, key string) interface{} {
	if !cm.applicationCacheEnabled {
		return nil
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Debugf("get from ApplicationCache for table %s key %s", tableName, key)

	cache := cm.applicationCaches[tenantId]
	if cache == nil {
		log.WithFields(log.Fields{"tenantId": tenantId}).Error("application cache not found")
		return nil
	}

	appKey := getApplicationCacheKey(tableName, key)
	value, _ := cache.GetIfPresent(appKey)
	return value
}

// Set value for the specified table and key
func (cm CacheManager) ApplicationCacheSet(tenantId string, tableName string, key string, value interface{}) {
	if !cm.applicationCacheEnabled {
		return
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Debugf("set ApplicationCache for table %s key %s", tableName, key)

	cache := cm.applicationCaches[tenantId]
	if cache == nil {
		log.WithFields(log.Fields{"tenantId": tenantId}).Error("application cache not found")
		return
	}

	appKey := getApplicationCacheKey(tableName, key)
	cache.Put(appKey, value)
}

// Delete an entry from the application cache
func (cm CacheManager) ApplicationCacheDelete(tenantId string, tableName string, key string) {
	if !cm.applicationCacheEnabled {
		return
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Debugf("delete ApplicationCache for table %s key %s", tableName, key)

	cache := cm.applicationCaches[tenantId]
	if cache == nil {
		log.WithFields(log.Fields{"tenantId": tenantId}).Error("application cache not found")
		return
	}

	appKey := getApplicationCacheKey(tableName, key)
	cache.Invalidate(appKey)
}

// Delete all entries for the given table
func (cm CacheManager) ApplicationCacheDeleteAll(tenantId string, tableName string) {
	if !cm.applicationCacheEnabled {
		return
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Debugf("delete all ApplicationCache entries for table %s", tableName)

	cache := cm.applicationCaches[tenantId]
	if cache == nil {
		log.WithFields(log.Fields{"tenantId": tenantId}).Error("application cache not found")
		return
	}

	keys := cache.GetAllKeys()
	for _, key := range keys {
		if strings.HasPrefix(key.(string), tableName) {
			cache.Invalidate(key)
		}
	}
}

// Delete all entries
func (cm CacheManager) ApplicationCacheInvalidateAll(tenantId string) {
	if !cm.applicationCacheEnabled {
		return
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Debug("invalidate all ApplicationCache")

	cache := cm.applicationCaches[tenantId]
	if cache == nil {
		log.WithFields(log.Fields{"tenantId": tenantId}).Error("application cache not found")
		return
	}

	cache.InvalidateAll()
}

func getApplicationCacheKey(tableName string, name string) string {
	return tableName + CACHE_KEY_DELIMITER + name
}

func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "404")
}

func (cm CacheManager) GetGroupServiceFeatureTags(tenantId string, cacheKey string) map[string]string {
	featureCache := cm.groupServiceFeatureCaches[tenantId]
	if featureCache == nil {
		log.WithFields(log.Fields{"cacheKey": cacheKey, "tenantId": tenantId}).Error("group service feature cache not found")
		return map[string]string{}
	}

	featureTags, err := featureCache.(cache.LoadingCache).Get(cacheKey)
	if err != nil {
		if isNotFoundError(err) {
			log.WithFields(log.Fields{"cacheKey": cacheKey, "tenantId": tenantId}).Infof("Cache miss for feature tags")
			return nil
		}
		log.WithFields(log.Fields{"cacheKey": cacheKey, "tenantId": tenantId}).Errorf("Error retrieving cache: %v", err)
		return nil
	}

	if featureTags == nil {
		log.WithFields(log.Fields{"cacheKey": cacheKey, "tenantId": tenantId}).Info("No feature tags found")
		return nil
	}
	tags, ok := featureTags.(map[string]string)
	if !ok {
		log.WithFields(log.Fields{"key": cacheKey, "tenantId": tenantId, "expected": "map[string]string", "actual": fmt.Sprintf("%T", featureTags)}).Error("Unexpected type")
		return nil
	}
	return tags
}

func (cm CacheManager) SetGroupServiceFeatureTags(tenantId string, cacheKey string, tags map[string]string) error {
	cache := cm.groupServiceFeatureCaches[tenantId]
	if cache == nil {
		err := errors.New("group service feature cache not found")
		log.WithFields(log.Fields{"cacheKey": cacheKey, "tenantId": tenantId}).Error(err)
		return err
	}

	cache.Put(cacheKey, tags)
	return nil
}

func (cm CacheManager) DeleteGroupServiceFeatureTags(tenantId string, cacheKey string) error {
	cache := cm.groupServiceFeatureCaches[tenantId]
	if cache == nil {
		err := errors.New("group service feature cache not found")
		log.WithFields(log.Fields{"cacheKey": cacheKey, "tenantId": tenantId}).Error(err)
		return err
	}

	cache.Invalidate(cacheKey)
	return nil
}

func SetGrpCacheLoadFunc(f cache.LoaderFunc) {
	grpCacheLoadFunc = f
}

func GetGrpCacheLoadFunc() cache.LoaderFunc {
	return grpCacheLoadFunc
}
