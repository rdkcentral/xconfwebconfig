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
	"strings"

	copy "github.com/mitchellh/copystructure"
)

/*
Retrieving and processing Xconf data from cache:

The following code illustrates how to retrieve a specific Model from the Model table:

	inmport (
		"github.com/rdkcentral/xconfwebconfig/shared"
		coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
		"github.com/rdkcentral/xconfwebconfig/db"
	)

	obj, err := db.GetListingDao().GetOne(db.TABLE_LOGS, "85:B5:61:4A:26:D5", "tvxads-do-77-xconfdataservice-0.do.xcal.tv_5")
	if err != nil {
		// Handle error!
	}

	var ccl coreef.ConfigChangeLog
	ccl := *obj.(*coreef.ConfigChangeLog)
*/

// CachedListingDao interface for ListingDao
type CachedListingDao interface {
	GetOne(tableName string, rowKey string, key2 interface{}) (interface{}, error)
	SetOne(tableName string, rowKey string, key2 interface{}, entity interface{}) error
	DeleteOne(tableName string, rowKey string, key2 interface{}) error
	DeleteAll(tableName string, rowKey string) error
	GetAll(tableName string, rowKey string) ([]interface{}, error)
	GetAllAsMap(tableName string, rowKey string, key2List []interface{}) (map[interface{}]interface{}, error)
	GetKeys(tableName string) ([]TwoKeys, error)
	GetKey2AsList(tableName string, rowKey string) ([]interface{}, error)
}

type cachedListingDaoImpl struct{}

var cachedListingDao = cachedListingDaoImpl{}

// GetCachedListingDao returns CachedListingDao
func GetCachedListingDao() CachedListingDao {
	return cachedListingDao
}

// GetOne get one Xconf record from cache
func (cld cachedListingDaoImpl) GetOne(tableName string, rowKey string, key2 interface{}) (interface{}, error) {
	cache, err := GetCacheManager().getCache(tableName)
	if err != nil {
		return nil, err
	}

	tkStr := NewTwoKeys(rowKey, key2).String()
	obj, err := cache.Get(tkStr)
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

// SetOne set Xconf record in DB and cache where entity param is the model/struct
func (cld cachedListingDaoImpl) SetOne(tableName string, rowKey string, key2 interface{}, entity interface{}) error {
	cache, err := GetCacheManager().getCache(tableName)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	// 1st update the DB
	err = GetListingDao().SetOne(tableName, rowKey, key2, jsonData)
	if err == nil {
		// Next update the cache
		tkStr := NewTwoKeys(rowKey, key2).String()
		cache.Put(tkStr, entity)
		// Write cache changed log
		GetCacheManager().writeCacheLog(tableName, tkStr, CREATE_OPERATION, int32(cache.Size()))
	}

	return err
}

// DeleteOne delete Xconf record from DB and cache
func (cld cachedListingDaoImpl) DeleteOne(tableName string, rowKey string, key2 interface{}) error {
	cache, err := GetCacheManager().getCache(tableName)
	if err != nil {
		return err
	}

	// 1st delete from DB
	err = GetListingDao().DeleteOne(tableName, rowKey, key2)
	if err == nil {
		// Calculate cache size since removal doesn't take place immediately
		cacheSize := int32(cache.Size() - 1)
		// Next invalidate entry from the cache
		tkStr := NewTwoKeys(rowKey, key2).String()
		cache.Invalidate(tkStr)
		// Write cache changed log
		GetCacheManager().writeCacheLog(tableName, tkStr, DELETE_OPERATION, cacheSize)
	}

	return err
}

// DeleteAll delete Xconf record from DB and cache
func (cld cachedListingDaoImpl) DeleteAll(tableName string, rowKey string) error {
	cache, err := GetCacheManager().getCache(tableName)
	if err != nil {
		return err
	}

	key2List, err := GetListingDao().GetKey2AsList(tableName, rowKey)
	if err != nil {
		return err
	}

	// 1st delete from DB
	err = GetListingDao().DeleteAll(tableName, rowKey)
	if err == nil {
		// Next invalidate entry from the cache
		for _, key2 := range key2List {
			tkStr := NewTwoKeys(rowKey, key2).String()
			cache.Invalidate(tkStr)
		}
	}

	return err
}

// GetAll get all Xconf records from fache for the specified rowKey
func (cld cachedListingDaoImpl) GetAll(tableName string, rowKey string) ([]interface{}, error) {
	cache, err := GetCacheManager().getCache(tableName)
	if err != nil {
		return nil, err
	}

	cloneData := GetCacheManager().settings.cloneDataEnabled
	result := make([]interface{}, len(rowKey))

	// find all records in cache with key that has rowKey + delimiter as the prefix
	keyPrefix := fmt.Sprintf("%s%s", rowKey, TwowKeysDelimiter)
	objMap := cache.GetAll()
	for key, obj := range objMap {
		if strings.HasPrefix(key.(string), keyPrefix) {
			if cloneData {
				// Create a copy to prevent modification of cached object
				copyObj, err := copy.Copy(obj)
				if err != nil {
					return nil, err
				}
				result = append(result, copyObj)
			} else {
				result = append(result, obj)
			}
		}
	}

	return result, nil
}

// GetAllAsMap get a map of all Xconf records for the specified key2 list
func (cld cachedListingDaoImpl) GetAllAsMap(tableName string, rowKey string, key2List []interface{}) (map[interface{}]interface{}, error) {
	cache, err := GetCacheManager().getCache(tableName)
	if err != nil {
		return nil, err
	}

	cloneData := GetCacheManager().settings.cloneDataEnabled
	var result = make(map[interface{}]interface{})

	for _, key2 := range key2List {
		tkStr := NewTwoKeys(rowKey, key2).String()
		obj, err := cache.Get(tkStr)
		if err != nil {
			return nil, err
		}

		if cloneData {
			// Create a copy to prevent modification of cached object
			cloneObj, err := copy.Copy(obj)
			if err != nil {
				return nil, err
			}
			result[key2] = cloneObj
		} else {
			result[key2] = obj
		}
	}

	return result, nil
}

// GetKeys get all Xconf two keys from cache
func (cld cachedListingDaoImpl) GetKeys(tableName string) ([]TwoKeys, error) {
	cache, err := GetCacheManager().getCache(tableName)
	if err != nil {
		return nil, err
	}

	var result []TwoKeys

	keys := cache.GetAllKeys()
	for _, key := range keys {
		twoKeys, err := NewTwoKeysFromString(key.(string))
		if err != nil {
			return nil, err
		}
		result = append(result, *twoKeys)
	}

	return result, nil
}

// GetKey2AsList get a list of Xconf key2 for the specified rowKey
func (cld cachedListingDaoImpl) GetKey2AsList(tableName string, rowKey string) ([]interface{}, error) {
	cache, err := GetCacheManager().getCache(tableName)
	if err != nil {
		return nil, err
	}

	var result []interface{}

	// find all records in cache with key that has rowKey + delimiter as the prefix
	keyPrefix := fmt.Sprintf("%s%s", rowKey, TwowKeysDelimiter)
	keys := cache.GetAllKeys()
	for _, key := range keys {
		if strings.HasPrefix(key.(string), keyPrefix) {
			twoKeys, err := NewTwoKeysFromString(key.(string))
			if err != nil {
				return nil, err
			}
			// TODO handle usecase when key2 not a string, i.e. cast to appropriate type?
			result = append(result, twoKeys.Key2)
		}
	}

	return result, nil
}
