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
	"time"

	"github.com/gocql/gocql"
	copy "github.com/mitchellh/copystructure"
	log "github.com/sirupsen/logrus"
)

/*
Retrieving and processing Xconf data:

The following code illustrates how to retrieve a specific Model from the Model table:

    import 	"github.com/rdkcentral/xconfwebconfig/db"

	obj, err := db.GetCachedSimpleDao().GetOne("COMCAST", "Model", "PX013ANM")
	if err != nil {
		// Handle error!
	}

	var m db.Model
	m = *obj.(*db.Model)
*/

// CachedSimpleDao interface for SimpleDao and CompressingDataDao
type CachedSimpleDao interface {
	GetOne(tenantId string, tableName string, key string) (interface{}, error)
	GetOneFromCacheOnly(tenantId string, tableName string, key string) (interface{}, error)
	SetOne(tenantId string, tableName string, key string, entity interface{}) error
	DeleteOne(tenantId string, tableName string, key string) error
	GetAllByKeys(tenantId string, tableName string, keys []string) ([]interface{}, error)
	GetAllAsList(tenantId string, tableName string, maxResults int) ([]interface{}, error)
	GetAllAsMap(tenantId string, tableName string) (map[interface{}]interface{}, error)
	GetKeys(tenantId string, tableName string) ([]interface{}, error)
	RefreshAll(tenantId string, tableName string) error
	RefreshOne(tenantId string, tableName string, key string) error
}

type cachedSimpleDaoImpl struct{}

var cachedSimpleDao = cachedSimpleDaoImpl{}

// GetCachedSimpleDao returns CachedSimpleDao
func GetCachedSimpleDao() CachedSimpleDao {
	return cachedSimpleDao
}

// GetOne get one Xconf record from cache
func (csd cachedSimpleDaoImpl) GetOne(tenantId string, tableName string, key string) (interface{}, error) {
	cache, err := GetCacheManager().getCache(tenantId, tableName)
	if err != nil {
		return nil, err
	}

	obj, err := cache.Get(key)
	if err != nil {
		return nil, err
	}

	if !GetCacheManager().settings.cloneDataEnabled {
		return obj, nil
	}

	// Create a copy to prevent modification of cached object
	cloneObj, err := copy.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj, nil
}

func (csd cachedSimpleDaoImpl) RefreshOne(tenantId string, tableName string, key string) error {
	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return err
	}

	cache, err := GetCacheManager().getCache(tenantId, tableName)
	if err != nil {
		return err
	}

	// First, invalidate the entry in the cache then reload from DB
	cache.Invalidate(key)

	var entry interface{}
	if tableInfo.IsCompressAndSplit() {
		entry, err = GetCompressingDataDao().GetOne(tenantId, tableName, key)
	} else {
		entry, err = GetSimpleDao().GetOne(tenantId, tableName, key)
	}

	if err != nil {
		return err
	}

	cache.Put(key, entry)

	return nil
}

// GetOne get one Xconf record from cache if exists and skips loading from DB.
// Returns gocql.ErrNotFound if there is no cache value for the rowKey.
func (csd cachedSimpleDaoImpl) GetOneFromCacheOnly(tenantId string, tableName string, key string) (interface{}, error) {
	cache, err := GetCacheManager().getCache(tenantId, tableName)
	if err != nil {
		return nil, err
	}

	obj, exists := cache.GetIfPresent(key)
	if !exists {
		return nil, gocql.ErrNotFound
	}

	if !GetCacheManager().settings.cloneDataEnabled {
		return obj, nil
	}

	// Create a copy to prevent modification of cached object
	cloneObj, err := copy.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj, nil
}

// SetOne set Xconf record in DB and cache where entity param is the *struct
func (csd cachedSimpleDaoImpl) SetOne(tenantId string, tableName string, key string, entity interface{}) error {
	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return err
	}

	// Ensure entity param is ptr to a struct
	// valid := false
	// kind := reflect.TypeOf(entity).Kind()
	// if kind == reflect.Ptr {
	// 	kind = reflect.ValueOf(entity).Elem().Kind()
	// 	valid = (kind == reflect.Struct)
	// }
	// if !valid {
	// 	return fmt.Errorf("Invalid object type: %v. Entity param must be a *struct", kind)
	// }

	cm := GetCacheManager()

	cache, err := cm.getCache(tenantId, tableName)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	// 1st update data the DB as Json Data
	if tableInfo.IsCompressAndSplit() {
		err = GetCompressingDataDao().SetOne(tenantId, tableName, key, jsonData)
	} else {
		err = GetSimpleDao().SetOne(tenantId, tableName, key, jsonData)
	}

	// Next update the cache with the actual model/struct
	if err == nil {
		cache.Put(key, entity)
		// Invalidate application cache entries for this table
		cm.ApplicationCacheDeleteAll(tenantId, tableName)
		// Write cache changed log
		cm.writeCacheLog(tenantId, tableName, key, CREATE_OPERATION, int32(cache.Size()))
	}

	return err
}

// DeleteOne delete Xconf record from DB and cache
func (csd cachedSimpleDaoImpl) DeleteOne(tenantId string, tableName string, key string) error {
	cm := GetCacheManager()

	cache, err := cm.getCache(tenantId, tableName)
	if err != nil {
		return err
	}

	// 1st delete from DB
	err = GetDatabaseClient().DeleteXconfData(tenantId, tableName, key)
	if err == nil {
		// Calculate cache size since removal doesn't take place immediately
		cacheSize := int32(cache.Size() - 1)
		// Next invalidate entry from the cache
		cache.Invalidate(key)
		// Invalidate application cache entries for this table
		cm.ApplicationCacheDeleteAll(tenantId, tableName)
		// Write cache changed log
		cm.writeCacheLog(tenantId, tableName, key, DELETE_OPERATION, cacheSize)
	}

	return err
}

// GetAllByKeys get all Xconf keys from cache
func (csd cachedSimpleDaoImpl) GetAllByKeys(tenantId string, tableName string, keys []string) ([]interface{}, error) {
	result := make([]interface{}, len(keys))

	for i, key := range keys {
		obj, err := csd.GetOne(tenantId, tableName, key)
		if err != nil {
			return nil, err
		}

		result[i] = obj
	}

	return result, nil
}

// GetAllAsList get multiple Xconf records from cache where a maxResults value of 0 indicates no limit
func (csd cachedSimpleDaoImpl) GetAllAsList(tenantId string, tableName string, maxResults int) ([]interface{}, error) {
	cache, err := GetCacheManager().getCache(tenantId, tableName)
	if err != nil {
		return nil, err
	}

	var objs []interface{}
	values := cache.GetAllValues()
	if maxResults > 0 && maxResults <= len(values) {
		objs = values[0:maxResults]
	} else {
		objs = values
	}

	if !GetCacheManager().settings.cloneDataEnabled {
		return objs, nil
	}

	result := make([]interface{}, len(objs))

	for i, obj := range objs {
		// Create a copy to prevent modification of cached object
		copyObj, err := copy.Copy(obj)
		if err != nil {
			return nil, err
		}
		result[i] = copyObj
	}

	return result, nil
}

// GetAllAsMap get all Xconf records as a map from cache
func (csd cachedSimpleDaoImpl) GetAllAsMap(tenantId string, tableName string) (map[interface{}]interface{}, error) {
	cache, err := GetCacheManager().getCache(tenantId, tableName)
	if err != nil {
		return nil, err
	}

	objMap := cache.GetAll()

	if !GetCacheManager().settings.cloneDataEnabled {
		return objMap, nil
	}

	result := make(map[interface{}]interface{})

	for key, obj := range objMap {
		// Create a copy to prevent modification of cached object
		copyObj, err := copy.Copy(obj)
		if err != nil {
			return nil, err
		}
		result[key] = copyObj
	}

	return result, nil
}

// GetKeys get all Xconf keys from cache
func (csd cachedSimpleDaoImpl) GetKeys(tenantId string, tableName string) ([]interface{}, error) {
	cache, err := GetCacheManager().getCache(tenantId, tableName)
	if err != nil {
		return nil, err
	}

	keys := cache.GetAllKeys()

	return keys, nil
}

func (csd cachedSimpleDaoImpl) RefreshAll(tenantId string, tableName string) error {
	log.Debug(fmt.Sprintf("Refresh cache for tenantId %s table %s...", tenantId, tableName))

	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return err
	}

	cm := GetCacheManager()
	cacheInfo, err := cm.getCacheInfo(tenantId, tableName)
	if err != nil {
		return err
	}
	cache := cacheInfo.cache

	// First, invalidate all entries in the cache then reload from DB
	cache.InvalidateAll()

	var entries map[string]interface{}
	if tableInfo.IsCompressAndSplit() {
		entries, err = GetCompressingDataDao().GetAllAsMap(tenantId, tableName, true)
	} else {
		entries, err = GetSimpleDao().GetAllAsMap(tenantId, tableName, 0)
	}

	if err != nil {
		return err
	}

	for k, v := range entries {
		cache.Put(k, v)
	}

	cacheInfo.DaoRefreshTime = time.Now().UTC()

	// Invalidate application cache entries for this table
	cm.ApplicationCacheDeleteAll(tenantId, tableName)

	return nil
}
