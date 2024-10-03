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
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-akka/configuration"
	"github.com/gocql/gocql"
	log "github.com/sirupsen/logrus"

	"xconfwebconfig/util"

	cache "github.com/Comcast/goburrow-cache"
)

var Conf *configuration.Config

// ConfigInjection - dependency injection
func ConfigInjection(conf *configuration.Config) {
	Conf = conf
}

// CacheRefreshTask background task to refresh cache
type CacheRefreshTask struct {
	lastRefreshedTimestamp time.Time
	refreshAttemptsLeft    int32
	stopped                chan bool
	ticker                 *time.Ticker
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
					failedTables := GetCacheManager().RefreshAll()
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
						t.refreshAttemptsLeft = cacheManager.Settings.retryCountUntilFullRefresh
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

// Statistics cache statistics
type Statistics struct {
	CacheMap map[string]CacheStats `json:"Statistics"`
}

type CacheInfo struct {
	cache cache.LoadingCache
	Stats *CacheStats
}

// CacheManager a cache manager
type CacheManager struct {
	Settings                CacheSettings
	cacheMap                map[string]CacheInfo
	refreshCacheTask        CacheRefreshTask
	applicationCache        cache.Cache
	applicationCacheEnabled bool
}

// CacheManager a cache manager
var cacheManager CacheManager
var initOnce sync.Once
var refreshCacheMutex sync.Mutex

// GetCacheManager Initializes a CacheManager
func GetCacheManager() CacheManager {
	initOnce.Do(func() {
		cacheManager = CacheManager{}
		cacheManager.cacheMap = make(map[string]CacheInfo)

		if Conf == nil {
			// Handle ServerConfig not initialized yet
			cacheManager.Settings.tickDuration = 60000
			cacheManager.Settings.retryCountUntilFullRefresh = 10
			cacheManager.Settings.changedKeysTimeWindowSize = 900000
			cacheManager.Settings.reloadCacheEntries = false
			cacheManager.Settings.reloadCacheEntriesTimeout = 1
			cacheManager.Settings.reloadCacheEntriesTimeUnit = "DAYS"
			cacheManager.Settings.numberOfEntriesToProcessSequentially = 10000
			cacheManager.Settings.keysetChunkSizeForMassCacheLoad = 500
			cacheManager.Settings.cloneDataEnabled = false
			cacheManager.Settings.applicationCacheEnabled = false
		} else {
			cacheManager.Settings.tickDuration = Conf.GetInt32("xconfwebconfig.xconf.cache_tickDuration", 60000)
			cacheManager.Settings.retryCountUntilFullRefresh = Conf.GetInt32("xconfwebconfig.xconf.cache_retryCountUntilFullRefresh", 10)
			cacheManager.Settings.changedKeysTimeWindowSize = Conf.GetInt32("xconfwebconfig.xconf.cache_changedKeysTimeWindowSize", 900000)
			cacheManager.Settings.reloadCacheEntries = Conf.GetBoolean("xconfwebconfig.xconf.cache_reloadCacheEntries", false)
			cacheManager.Settings.reloadCacheEntriesTimeout = Conf.GetInt64("xconfwebconfig.xconf.cache_reloadCacheEntriesTimeout", 1)
			cacheManager.Settings.reloadCacheEntriesTimeUnit = Conf.GetString("xconfwebconfig.xconf.cache_reloadCacheEntriesTimeUnit", "DAYS")
			cacheManager.Settings.numberOfEntriesToProcessSequentially = Conf.GetInt32("xconfwebconfig.xconf.cache_numberOfEntriesToProcessSequentially", 10000)
			cacheManager.Settings.keysetChunkSizeForMassCacheLoad = Conf.GetInt32("xconfwebconfig.xconf.cache_keysetChunkSizeForMassCacheLoad", 500)
			cacheManager.Settings.cloneDataEnabled = Conf.GetBoolean("xconfwebconfig.xconf.cache_clone_data_enabled", false)
			cacheManager.Settings.applicationCacheEnabled = Conf.GetBoolean("xconfwebconfig.xconf.application_cache_enabled", false)
		}

		// Initialize the cache
		for tableName, tableInfo := range tableConfig {
			if !tableInfo.CacheData {
				continue // No caching for this table
			}

			// Generate a load function for the table
			loadFn := generateLoadFunction(tableName)

			// Create a LoadingCache for each table
			var loadingCache cache.LoadingCache
			if cacheManager.Settings.reloadCacheEntries {
				freshAfterWrite := getDuration(
					cacheManager.Settings.reloadCacheEntriesTimeout, cacheManager.Settings.reloadCacheEntriesTimeUnit)
				loadingCache = cache.NewLoadingCache(loadFn,
					cache.WithMaximumSize(0),                     // Unlimited number of entries in the cache.
					cache.WithRefreshAfterWrite(freshAfterWrite), // Expire entries after specified duration since last created.
				)
			} else {
				loadingCache = cache.NewLoadingCache(loadFn,
					cache.WithMaximumSize(0), // Unlimited number of entries in the cache.
				)
			}
			cacheManager.cacheMap[tableName] = CacheInfo{cache: loadingCache, Stats: &CacheStats{}}
		}

		cc, ok := GetDatabaseClient().(*CassandraClient)
		if ok && !cc.IsTestOnly() {
			// Initiate precaching
			cacheManager.initiatePrecaching()

			// Start background task to refresh cache
			cacheManager.refreshCacheTask = CacheRefreshTask{
				lastRefreshedTimestamp: time.Now().UTC(),
				refreshAttemptsLeft:    cacheManager.Settings.retryCountUntilFullRefresh,
				stopped:                make(chan bool),
				ticker:                 time.NewTicker(time.Duration(cacheManager.Settings.tickDuration) * time.Millisecond),
			}
			cacheManager.refreshCacheTask.Run()
		}

		// Initialize application cache
		cacheManager.applicationCache = cache.New(cache.WithMaximumSize(0))
		cacheManager.applicationCacheEnabled = cacheManager.Settings.applicationCacheEnabled
	})

	return cacheManager
}

func (cm CacheManager) GetCacheStats(tableName string) (*CacheStats, error) {
	if err := cacheManager.updateCacheStats(tableName); err != nil {
		return nil, err
	}
	return cm.cacheMap[tableName].Stats, nil
}

func (cm CacheManager) updateCacheStats(tableName string) error {
	cacheInfo := cm.cacheMap[tableName]
	if cacheInfo.cache == nil {
		return fmt.Errorf("cache not found or configured for table '%v'", tableName)
	}

	// Get current cache stats
	stats := cache.Stats{}
	cacheInfo.cache.Stats(&stats)

	cacheInfo.Stats.CacheSize = cacheInfo.cache.Size()
	cacheInfo.Stats.RequestCount = stats.RequestCount()
	cacheInfo.Stats.EvictionCount = stats.EvictionCount
	cacheInfo.Stats.HitRate = stats.HitRate()
	cacheInfo.Stats.MissRate = stats.MissRate()
	cacheInfo.Stats.TotalLoadTime = stats.TotalLoadTime

	nonAbsentCount := 0
	values := cacheInfo.cache.GetAllValues()
	for _, v := range values {
		if v != nil {
			nonAbsentCount++
		}
	}
	cacheInfo.Stats.NonAbsentCount = nonAbsentCount

	return nil
}

// GetStatistics returns cache statistics
func (cm CacheManager) GetStatistics() *Statistics {
	statistics := Statistics{CacheMap: make(map[string]CacheStats)}

	for tableName, cacheInfo := range cm.cacheMap {
		// Get current cache stats
		stats := cache.Stats{}
		cacheInfo.cache.Stats(&stats)

		cacheInfo.Stats.CacheSize = cacheInfo.cache.Size()
		cacheInfo.Stats.RequestCount = stats.RequestCount()
		cacheInfo.Stats.EvictionCount = stats.EvictionCount
		cacheInfo.Stats.HitRate = stats.HitRate()
		cacheInfo.Stats.MissRate = stats.MissRate()
		cacheInfo.Stats.TotalLoadTime = stats.TotalLoadTime

		statistics.CacheMap[tableName] = *cacheInfo.Stats
	}

	return &statistics
}

// getCacheInfo returns CacheInfo for the specified table
func (cm CacheManager) getCacheInfo(tableName string) (*CacheInfo, error) {
	cacheInfo := cm.cacheMap[tableName]
	if cacheInfo.cache == nil {
		err := fmt.Errorf("cache not found or configured for table '%v'", tableName)
		return nil, err
	}
	return &cacheInfo, nil
}

// getCache returns a LoadingCache for the specified table
func (cm CacheManager) getCache(tableName string) (cache.LoadingCache, error) {
	cacheInfo, err := cm.getCacheInfo(tableName)
	if err != nil {
		return nil, err
	}
	return cacheInfo.cache, nil
}

// Preloading caches
func (cm CacheManager) initiatePrecaching() {
	log.Debug("initializing cache...")

	// Just a debugging convenience, load the tables in the same order every time
	tableList := make([]string, 0, len(tableConfig))
	for tableName := range tableConfig {
		tableList = append(tableList, tableName)
	}
	sort.Strings(tableList)

	var wg sync.WaitGroup

	precachingStart := time.Now()

	for _, tableName := range tableList {
		wg.Add(1)

		go func(tableName string) {
			defer wg.Done()

			tableInfo := tableConfig[tableName]
			if tableInfo.CacheData {
				log.Debugf("precaching for '%v'...", tableName)
			} else {
				log.Debugf("skipped precaching for '%v'", tableName)
				return
			}

			start := time.Now()
			var entries map[string]interface{}
			var err error
			if tableInfo.IsCompressAndSplit() {
				entries, err = GetCompressingDataDao().GetAllAsMap(tableName)
			} else {
				entries, err = GetSimpleDao().GetAllAsMap(tableName, 0)
			}

			if err != nil {
				log.Fatalf("failed to preload cache for table '%v': %v", tableName, err)
				panic(err)
			}

			cache, err := cm.getCache(tableName)
			if err != nil {
				log.Fatalf("failed to preload cache for table '%v': %v", tableName, err)
				panic(err)
			}

			for k, v := range entries {
				cache.Put(k, v)
			}

			log.Debugf("'%v' precached %v entries in %v", tableName, len(entries), time.Since(start))
		}(tableName)
	}

	wg.Wait()

	log.WithFields(log.Fields{"duration": time.Since(precachingStart).String()}).Info("precache duration")
}

func (cm CacheManager) GetChangedKeysTimeWindowSize() int32 {
	return cm.Settings.changedKeysTimeWindowSize
}

// RefreshAll Refresh all caches and return list of table names which were not refreshed
func (cm CacheManager) RefreshAll() []string {
	refreshCacheMutex.Lock()
	defer refreshCacheMutex.Unlock()

	log.Debug("starting cache refresh...")

	var failedToRefreshTables []string
	for _, tableInfo := range tableConfig {
		var err error
		var start = time.Now()
		if tableInfo.CacheData {
			err = GetCachedSimpleDao().RefreshAll(tableInfo.TableName)
		} else {
			continue // Skip since data is not cache
		}

		duration := time.Since(start)

		if err == nil {
			cache, _ := cm.getCache(tableInfo.TableName)
			log.Debugf("cache refreshed: '%v' precached %v entries in %v", tableInfo.TableName, cache.Size(), duration)
		} else {
			log.Errorf("failed to refresh cache for table '%v': %v", tableInfo.TableName, err)
			failedToRefreshTables = append(failedToRefreshTables, tableInfo.TableName)
		}
	}

	return failedToRefreshTables
}

// Refresh Refresh cache for the specified table
func (cm CacheManager) Refresh(tableName string) error {
	refreshCacheMutex.Lock()
	defer refreshCacheMutex.Unlock()

	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		log.Errorf("failed to refresh cache for table '%v': %v", tableName, err)
		return err
	}

	if tableInfo.CacheData {
		err = GetCachedSimpleDao().RefreshAll(tableInfo.TableName)
		if err != nil {
			log.Errorf("failed to refresh cache for table '%v': %v", tableInfo.TableName, err)
			return err
		}
	} else {
		log.Debugf("unable to refresh cache for table '%v', data is not cached", tableInfo.TableName)
	}

	return nil
}

// SyncChanges Updates changes for given time-window defined by
// start (lower bound, inclusive) and end (upper bound, exclusive)
func (cm CacheManager) SyncChanges(startTime time.Time, endTime time.Time, apply bool) (changedList []interface{}, err error) {
	startTS := util.GetTimestamp(startTime)
	currentRowKey := startTS - (startTS % int64(cm.Settings.changedKeysTimeWindowSize))

	endTS := util.GetTimestamp(endTime)
	endRowKey := endTS - (endTS % int64(cm.Settings.changedKeysTimeWindowSize))

	startUuid, err := util.UUIDFromTime(startTS, 0, 0)
	if err != nil {
		return nil, err
	}

	ranges := make(map[int64]*RangeInfo)
	ranges[currentRowKey] = &RangeInfo{StartValue: startUuid}
	currentRowKey += int64(cm.Settings.changedKeysTimeWindowSize)
	for currentRowKey <= endRowKey {
		ranges[currentRowKey] = nil
		currentRowKey += int64(cm.Settings.changedKeysTimeWindowSize)
	}

	// Load all changes from DB
	log.Debugf("sync cache, getting changed keys [%v - %v]: %v", startTS, endTS, buildLogForRanges(ranges))
	changedList, err = cm.loadChanges(ranges)
	if err != nil {
		return nil, err
	}

	// Apply changes to cache
	err = cm.applyChanges(changedList)
	if apply {
		log.Debugf("sync cache, getting changed keys [%v - %v]: %v", startTS, endTS, buildLogForRanges(ranges))
		err = cm.applyChanges(changedList)
	}

	return changedList, err
}

func (cm CacheManager) loadChanges(ranges map[int64]*RangeInfo) ([]interface{}, error) {
	var result []interface{}

	for rowKey, rangeInfo := range ranges {
		list, err := GetListingDao().GetRange(TABLE_XCONF_CHANGED_KEYS, rowKey, rangeInfo)
		if err != nil {
			return nil, err
		}
		result = append(result, list...)
	}

	return result, nil
}

// Apply changed data to keep cache in sync
func (cm CacheManager) applyChanges(changedDataList []interface{}) error {
	changedTables := util.Set{}   // Keep track of tables which were changed
	tablesToRefresh := util.Set{} // Keep track of tables which need to be refreshed

	// Ensure whatever's already in the list will get refresh
	defer func() {
		refreshTables(tablesToRefresh.ToSlice())
	}()

	for _, obj := range changedDataList {
		data := *obj.(*ChangedData)

		// Fixed issue w/ Xconf AS wrote changedKey field with double quotation marks ("")
		l := len(data.ChangedKey)
		if l > 0 && data.ChangedKey[0] == '"' && data.ChangedKey[l-1] == '"' {
			data.ChangedKey = data.ChangedKey[1 : l-1]
		}

		if data.CfName == "" || data.ChangedKey == "" || data.Operation == "" || data.DaoId == 0 || data.ValidCacheSize == 0 {
			log.Error("unable to load changed data")
			continue
		}

		log.Debugf("sync cache, processing %v for table '%v': %v", data.Operation, data.CfName, data.ChangedKey)

		tableInfo, err := GetTableInfo(data.CfName)
		if err != nil {
			log.Errorf("failed to apply changed data %v: %v", data.ChangedKey, err)
			return err
		}

		cache, err := cm.getCache(tableInfo.TableName)
		if err != nil {
			log.Errorf("failed to apply changed data %v: %v", data.ChangedKey, err)
			return err
		}

		if tableInfo.DaoId != data.DaoId {
			// Log a warning message since DaoId mapping table need to be updated
			log.Warnf("DaoId (%v) for table %v does not match the expected value (%v)",
				tableInfo.DaoId, tableInfo.TableName, data.DaoId)
		}

		switch data.Operation {
		case CREATE_OPERATION, UPDATE_OPERATION: // fetch modified entry to cache
			cache.Refresh(data.ChangedKey)
			val, err := cache.Get(data.ChangedKey)
			if err != nil || val == nil {
				log.Errorf("failed to apply changed data %v: %v", data.ChangedKey, err)
				return err
			}
		case DELETE_OPERATION: // evict entry from cache
			cache.Invalidate(data.ChangedKey)
		case TRUNCATE_OPERATION:
			cache.InvalidateAll()
			return nil
		}

		changedTables.Add(tableInfo.TableName)

		cacheSize := cache.Size()
		if cacheSize < int(data.ValidCacheSize) {
			log.Warnf("cache size difference, got %v instead of %v, scheduling full refresh for %v",
				cacheSize, data.ValidCacheSize, tableInfo.TableName)
			tablesToRefresh.Add(tableInfo.TableName)
		}
	}

	// Invalidate application cache entries for tables that were changed
	for _, tableName := range changedTables.ToSlice() {
		cm.ApplicationCacheDeleteAll(tableName)
	}

	return nil
}

// Write changed data to XconfChangedKeys4 table asynchronously
func (cm CacheManager) WriteCacheLog(tableName string, changedKey string, operation OperationType) {
	if cache, err := GetCacheManager().getCache(tableName); err == nil {
		cm.writeCacheLog(tableName, changedKey, operation, int32(cache.Size()))
	} else {
		log.Errorf("failed to write cache changed log: %v", err)
	}
}

// Write changed data to XconfChangedKeys4 table async
func (cm CacheManager) writeCacheLog(tableName string, changedKey string, operation OperationType, cacheSize int32) {
	go func() {
		currentTS := util.GetTimestamp(time.Now().UTC())
		rowKey := currentTS - (currentTS % int64(cm.Settings.changedKeysTimeWindowSize))

		tableInfo, err := GetTableInfo(tableName)
		if err != nil {
			log.Errorf("failed to write cache changed log: %v", err)
		}

		daoId := tableInfo.DaoId
		if daoId == 0 {
			log.Errorf("failed to write cache changed log: DAOid not configured for table '%v'", tableName)
		}

		changedData := ChangedData{
			ColumnName:     gocql.TimeUUID(),
			CfName:         tableName,
			ChangedKey:     changedKey,
			Operation:      operation,
			DaoId:          daoId,
			ValidCacheSize: cacheSize,
			UserName:       "DataService",
		}

		jsonData, err := json.Marshal(changedData)
		if err == nil {
			log.Debugf("write cache changed log for table '%v' (%v): %v %v %v", tableName, changedData.DaoId, operation, rowKey, changedData.ValidCacheSize)
			err = GetListingDao().SetOne(TABLE_XCONF_CHANGED_KEYS, rowKey, changedData.ColumnName, jsonData)
		}

		if err != nil {
			log.Errorf("failed to write cache changed log: %v", err)
		}
	}()
}

// Generates a load function for the cache
func generateLoadFunction(tableName string) func(k cache.Key) (cache.Value, error) {
	loadFn := func(k cache.Key) (cache.Value, error) {
		name := tableName
		tableInfo, err := GetTableInfo(name)
		if err != nil {
			return nil, err
		}

		// Use the appropriate DAO based on compression policy
		if tableInfo.IsCompressAndSplit() {
			// return GetCompressingDataDao().GetOne(name, k.(string))
			ret, err := GetCompressingDataDao().GetOne(name, k.(string))
			return ret, err
		} else if tableInfo.IsCompressOnly() {
			twoKeys, err := NewTwoKeysFromString(k.(string))
			if err != nil {
				return nil, err
			}
			return GetListingDao().GetOne(name, twoKeys.Key, twoKeys.Key2)
		} else {
			return GetSimpleDao().GetOne(name, k.(string))
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
func refreshTables(tableNames []string) {
	for _, tableName := range tableNames {
		err := GetCachedSimpleDao().RefreshAll(tableName)
		if err != nil {
			log.Errorf("failed to refresh cache for table '%v': %v", tableName, err)
		}
	}
}

// ApplicationCacheGet value for the specified table and key
func (cm CacheManager) ApplicationCacheGet(tableName string, key string) interface{} {
	if !cm.applicationCacheEnabled {
		return nil
	}

	log.Debugf("get from ApplicationCache for table '%v' key '%v'", tableName, key)

	appKey := getApplicationCacheKey(tableName, key)
	value, _ := cm.applicationCache.GetIfPresent(appKey)
	return value
}

// Set value for the specified table and key
func (cm CacheManager) ApplicationCacheSet(tableName string, key string, value interface{}) {
	if !cm.applicationCacheEnabled {
		return
	}

	log.Debugf("set ApplicationCache for table '%v' key '%v'", tableName, key)

	appKey := getApplicationCacheKey(tableName, key)
	cm.applicationCache.Put(appKey, value)
}

// Delete an entry from the application cache
func (cm CacheManager) ApplicationCacheDelete(tableName string, key string) {
	if !cm.applicationCacheEnabled {
		return
	}

	log.Debugf("delete ApplicationCache for table '%v' key '%v'", tableName, key)

	appKey := getApplicationCacheKey(tableName, key)
	cm.applicationCache.Invalidate(appKey)
}

// Delete all entries for the given table
func (cm CacheManager) ApplicationCacheDeleteAll(tableName string) {
	if !cm.applicationCacheEnabled {
		return
	}

	log.Debugf("delete all ApplicationCache for table '%v'", tableName)

	keys := cm.applicationCache.GetAllKeys()
	for _, key := range keys {
		if strings.HasPrefix(key.(string), tableName) {
			cm.applicationCache.Invalidate(key)
		}
	}
}

// Delete all entries
func (cm CacheManager) ApplicationCacheInvalidateAll() {
	if !cm.applicationCacheEnabled {
		return
	}

	cm.applicationCache.InvalidateAll()
}

func getApplicationCacheKey(tableName string, name string) string {
	return fmt.Sprintf("%v::%v", tableName, name)
}
