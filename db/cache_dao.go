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

    import 	"xconfwebconfig/db"

	obj, err := db.GetCachedSimpleDao().GetOne("Model", "PX013ANM")
	if err != nil {
		// Handle error!
	}

	var m db.Model
	m = *obj.(*db.Model)
*/

// CachedSimpleDao interface for SimpleDao and CompressingDataDao
type CachedSimpleDao interface {
	GetOne(tableName string, rowKey string) (interface{}, error)
	GetOneFromCacheOnly(tableName string, rowKey string) (interface{}, error)
	SetOne(tableName string, rowKey string, entity interface{}) error
	DeleteOne(tableName string, rowKey string) error
	GetAllByKeys(tableName string, rowKeys []string) ([]interface{}, error)
	GetAllAsList(tableName string, maxResults int) ([]interface{}, error)
	GetAllAsMap(tableName string) (map[interface{}]interface{}, error)
	GetKeys(tableName string) ([]interface{}, error)
	RefreshAll(tableName string) error
}

type cachedSimpleDaoImpl struct{}

var cachedSimpleDao = cachedSimpleDaoImpl{}

// GetCachedSimpleDao returns CachedSimpleDao
func GetCachedSimpleDao() CachedSimpleDao {
	return cachedSimpleDao
}

// GetOne get one Xconf record from cache
func (csd cachedSimpleDaoImpl) GetOne(tableName string, rowKey string) (interface{}, error) {
	cache, err := GetCacheManager().getCache(tableName)
	if err != nil {
		return nil, err
	}

	obj, err := cache.Get(rowKey)
	if err != nil {
		return nil, err
	}

	if !GetCacheManager().Settings.cloneDataEnabled {
		return obj, nil
	}

	// Create a copy to prevent modification of cached object
	cloneObj, err := copy.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj, nil
}

// GetOne get one Xconf record from cache if exists and skips loading from DB.
// Returns gocql.ErrNotFound if there is no cache value for the rowKey.
func (csd cachedSimpleDaoImpl) GetOneFromCacheOnly(tableName string, rowKey string) (interface{}, error) {
	cache, err := GetCacheManager().getCache(tableName)
	if err != nil {
		return nil, err
	}

	obj, exists := cache.GetIfPresent(rowKey)
	if !exists {
		return nil, gocql.ErrNotFound
	}

	if !GetCacheManager().Settings.cloneDataEnabled {
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
func (csd cachedSimpleDaoImpl) SetOne(tableName string, rowKey string, entity interface{}) error {
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

	cache, err := cm.getCache(tableName)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	// 1st update data the DB as Json Data
	if tableInfo.IsCompressAndSplit() {
		err = GetCompressingDataDao().SetOne(tableName, rowKey, jsonData)
	} else {
		err = GetSimpleDao().SetOne(tableName, rowKey, jsonData)
	}

	// Next update the cache with the actual model/struct
	if err == nil {
		cache.Put(rowKey, entity)
		// Invalidate application cache entries for this table
		cm.ApplicationCacheDeleteAll(tableName)
		// Write cache changed log
		cm.writeCacheLog(tableName, rowKey, CREATE_OPERATION, int32(cache.Size()))
	}

	return err
}

// DeleteOne delete Xconf record from DB and cache
func (csd cachedSimpleDaoImpl) DeleteOne(tableName string, rowKey string) error {
	cm := GetCacheManager()

	cache, err := cm.getCache(tableName)
	if err != nil {
		return err
	}

	// 1st delete from DB
	err = GetDatabaseClient().DeleteXconfData(tableName, rowKey)
	if err == nil {
		// Calculate cache size since removal doesn't take place immediately
		cacheSize := int32(cache.Size() - 1)
		// Next invalidate entry from the cache
		cache.Invalidate(rowKey)
		// Invalidate application cache entries for this table
		cm.ApplicationCacheDeleteAll(tableName)
		// Write cache changed log
		cm.writeCacheLog(tableName, rowKey, DELETE_OPERATION, cacheSize)
	}

	return err
}

// GetAllByKeys get all Xconf keys from cache
func (csd cachedSimpleDaoImpl) GetAllByKeys(tableName string, rowKeys []string) ([]interface{}, error) {
	result := make([]interface{}, len(rowKeys))

	for i, rowKey := range rowKeys {
		obj, err := csd.GetOne(tableName, rowKey)
		if err != nil {
			return nil, err
		}

		result[i] = obj
	}

	return result, nil
}

// GetAllAsList get multiple Xconf records from cache where a maxResults value of 0 indicates no limit
func (csd cachedSimpleDaoImpl) GetAllAsList(tableName string, maxResults int) ([]interface{}, error) {
	cache, err := GetCacheManager().getCache(tableName)
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

	if !GetCacheManager().Settings.cloneDataEnabled {
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
func (csd cachedSimpleDaoImpl) GetAllAsMap(tableName string) (map[interface{}]interface{}, error) {
	cache, err := GetCacheManager().getCache(tableName)
	if err != nil {
		return nil, err
	}

	objMap := cache.GetAll()

	if !GetCacheManager().Settings.cloneDataEnabled {
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
func (csd cachedSimpleDaoImpl) GetKeys(tableName string) ([]interface{}, error) {
	cache, err := GetCacheManager().getCache(tableName)
	if err != nil {
		return nil, err
	}

	keys := cache.GetAllKeys()

	return keys, nil
}

func (csd cachedSimpleDaoImpl) RefreshAll(tableName string) error {
	log.Debug(fmt.Sprintf("Refresh cache for '%v'...", tableName))

	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return err
	}

	cm := GetCacheManager()

	cacheInfo, err := cm.getCacheInfo(tableName)
	if err != nil {
		return err
	}
	cache := cacheInfo.cache

	// First, invalidate all entries in the cache then reload from DB
	cache.InvalidateAll()

	var entries map[string]interface{}
	if tableInfo.IsCompressAndSplit() {
		entries, err = GetCompressingDataDao().GetAllAsMap(tableName)
	} else {
		entries, err = GetSimpleDao().GetAllAsMap(tableName, 0)
	}

	if err != nil {
		return err
	}

	for k, v := range entries {
		cache.Put(k, v)
	}

	cacheInfo.Stats.DaoRefreshTime = time.Now().UTC()

	// Invalidate application cache entries for this table
	cm.ApplicationCacheDeleteAll(tableName)

	return nil
}
