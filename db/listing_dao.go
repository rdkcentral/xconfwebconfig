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
)

/*
Retrieving and processing Xconf data:

Xconf getter functions require a constructor to instantiate a model in order
for the method to unmarshal the raw JSON data (i.e. []byte) to the proper struct.
Therefore, a table name and a corresponding constructor function need to be
configured for the TableConfig variable.

The returned value is an empty interface{} so the caller needs to cast the value
to the target data type.
*/

// ListingDao interface
type ListingDao interface {
	GetOne(tableName string, rowKey string, key2 interface{}) (interface{}, error)
	SetOne(tableName string, rowKey interface{}, key2 interface{}, value []byte) error
	DeleteOne(tableName string, rowKey string, key2 interface{}) error
	DeleteAll(tableName string, rowKey string) error
	GetAll(tableName string, rowKey string) ([]interface{}, error)
	GetAllAsList(tableName string) ([]interface{}, error)
	GetAllAsMap(tableName string, rowKey string, key2List []interface{}) (map[interface{}]interface{}, error)
	GetRange(tableName string, rowKey interface{}, rangeInfo *RangeInfo) ([]interface{}, error)
	GetKeys(tableName string) ([]TwoKeys, error)
	GetKey2AsList(tableName string, rowKey string) ([]interface{}, error)
}

type listingDaoImpl struct{}

var listingDao = listingDaoImpl{}

// GetListingDao return an implementation of ListingDao
func GetListingDao() ListingDao {
	return listingDao
}

// GetOne get one Xconf record for two keys
func (ld listingDaoImpl) GetOne(tableName string, rowKey string, key2 interface{}) (interface{}, error) {
	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	// Get data from DB as raw JSON []byte
	data, err := GetDatabaseClient().GetXconfDataTwoKeys(tableName, rowKey, tableInfo.Key2FieldName, key2)
	if err != nil {
		return nil, err
	}

	var jsonData []byte
	if tableInfo.Compress {
		jsonData, err = decompress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress rowKey '%s': %w", rowKey, err)
		}
	} else {
		jsonData = data
	}

	obj := tableInfo.ConstructorFunc()  // Instantiate a new model/struct
	err = json.Unmarshal(jsonData, obj) // Deserialize the raw JSON []byte to a struct
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// SetOne set Xconf record for two keys
func (ld listingDaoImpl) SetOne(tableName string, rowKey interface{}, key2 interface{}, value []byte) error {
	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return err
	}

	// Compress the JSON data if required
	var data []byte
	if tableInfo.Compress {
		data = compress(value)
	} else {
		data = value
	}

	err = GetDatabaseClient().SetXconfDataTwoKeys(tableName, rowKey, tableInfo.Key2FieldName, key2, data, tableInfo.TTL)
	return err
}

// DeleteOne delete Xconf record for two keys
func (ld listingDaoImpl) DeleteOne(tableName string, rowKey string, key2 interface{}) error {
	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return err
	}

	return GetDatabaseClient().DeleteXconfDataTwoKeys(tableName, rowKey, tableInfo.Key2FieldName, key2)
}

func (ld listingDaoImpl) DeleteAll(tableName string, rowKey string) error {
	err := GetDatabaseClient().DeleteXconfData(tableName, rowKey)
	return err
}

// GetAll get multiple Xconf records for the specified rowKey
func (ld listingDaoImpl) GetAll(tableName string, rowKey string) ([]interface{}, error) {
	var result []interface{}

	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	// Get data from DB as a list of raw JSON []byte
	rows := GetDatabaseClient().GetAllXconfData(tableName, rowKey)
	for _, data := range rows {
		var jsonData []byte
		if tableInfo.Compress {
			jsonData, err = decompress(data)
			if err != nil {
				return nil, fmt.Errorf("failed to decompress rowKey '%s': %w", rowKey, err)
			}
		} else {
			jsonData = data
		}

		obj := tableInfo.ConstructorFunc()  // Instantiate a new model/struct
		err = json.Unmarshal(jsonData, obj) // Deserialize the raw JSON []byte to a struct
		if err != nil {
			return nil, err
		}
		result = append(result, obj)
	}

	return result, err
}

// GetAllAsList get a list of all Xconf records
func (ld listingDaoImpl) GetAllAsList(tableName string) ([]interface{}, error) {
	var result []interface{}

	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	rows := GetDatabaseClient().GetAllXconfDataAsList(tableName, 0)
	for _, data := range rows {
		var jsonData []byte
		if tableInfo.Compress {
			jsonData, err = decompress(data)
			if err != nil {
				return nil, err
			}
		} else {
			jsonData = data
		}

		obj := tableInfo.ConstructorFunc()  // Instantiate a new model/struct
		err = json.Unmarshal(jsonData, obj) // Deserialize the raw JSON []byte to a struct
		if err != nil {
			return nil, err
		}
		result = append(result, obj)
	}

	return result, err
}

// GetAllAsMap get a map of all Xconf records for the specified key2 list
func (ld listingDaoImpl) GetAllAsMap(tableName string, rowKey string, key2List []interface{}) (map[interface{}]interface{}, error) {
	var result = make(map[interface{}]interface{})

	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	// Get data from DB as a map of key2 and raw JSON []byte
	dataMap := GetDatabaseClient().GetAllXconfDataTwoKeysAsMap(tableName, rowKey, tableInfo.Key2FieldName, key2List)
	for key2, data := range dataMap {
		var jsonData []byte
		if tableInfo.Compress {
			jsonData, err = decompress(data)
			if err != nil {
				return nil, fmt.Errorf("failed to decompress rowKey '%s': %w", rowKey, err)
			}
		} else {
			jsonData = data
		}

		obj := tableInfo.ConstructorFunc()  // Instantiate a new model/struct
		err = json.Unmarshal(jsonData, obj) // Deserialize the raw JSON []byte to a struct
		if err != nil {
			return nil, err
		}
		result[key2] = obj
	}

	return result, err
}

func (ld listingDaoImpl) GetRange(tableName string, rowKey interface{}, rangeInfo *RangeInfo) ([]interface{}, error) {
	var result []interface{}

	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	// Get data from DB as a list of raw JSON []byte
	rows := GetDatabaseClient().GetAllXconfDataTwoKeysRange(tableName, rowKey, tableInfo.Key2FieldName, rangeInfo)
	for _, data := range rows {
		var jsonData []byte
		if tableInfo.Compress {
			jsonData, err = decompress(data)
			if err != nil {
				return nil, fmt.Errorf("failed to decompress rowKey '%s': %w", rowKey, err)
			}
		} else {
			jsonData = data
		}

		obj := tableInfo.ConstructorFunc()  // Instantiate a new model/struct
		err = json.Unmarshal(jsonData, obj) // Deserialize the raw JSON []byte to a struct
		if err != nil {
			return nil, err
		}
		result = append(result, obj)
	}

	return result, err
}

// GetKeys get all Xconf two keys
func (ld listingDaoImpl) GetKeys(tableName string) ([]TwoKeys, error) {
	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	return GetDatabaseClient().GetAllXconfTwoKeys(tableName, tableInfo.Key2FieldName), nil
}

// GetKeys get a list of Xconf key2 for the specified rowKey
func (ld listingDaoImpl) GetKey2AsList(tableName string, rowKey string) ([]interface{}, error) {
	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	return GetDatabaseClient().GetAllXconfKey2s(tableName, rowKey, tableInfo.Key2FieldName), nil
}
