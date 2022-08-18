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
)

/*
Retrieving and processing Xconf data:

Xconf getter functions require a constructor to instantiate a model in order
for the method to unmarshal the raw JSON data (i.e. []byte) to the proper struct.
Therefore, a table name and a corresponding constructor function need to be
configured for the TableConfig variable.

The returned value is an empty interface{} so the caller needs to cast the value
to the target data type.

The following code illustrates how to retrieve a specific Model from the Model table:

    import "xconfwebconfig/db"

	obj, err := db.GetSimpleDao().GetOne("Model", "PX013ANM")
	if err != nil {
		// Handle error!
	}
	var m db.Model
	m = *obj.(*db.Model)
*/

// SimpleDao interface
type SimpleDao interface {
	GetOne(tableName string, rowKey string) (interface{}, error)
	SetOne(tableName string, rowKey string, value []byte) error
	DeleteOne(tableName string, rowKey string) error
	GetAllByKeys(tableName string, rowKeys []string) ([]interface{}, error)
	GetAllAsList(tableName string, maxResults int) ([]interface{}, error)
	GetAllAsMap(tableName string, maxResults int) (map[string]interface{}, error)
	GetKeys(tableName string) []string
}

type simpleDaoImpl struct{}

var simpleDao = simpleDaoImpl{}

// GetSimpleDao return an implementation of SimpleDao
func GetSimpleDao() SimpleDao {
	return simpleDao
}

// GetOne get one Xconf record
func (sd simpleDaoImpl) GetOne(tableName string, rowKey string) (interface{}, error) {
	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	// Get data from DB as raw JSON []byte
	jsonData, err := GetDatabaseClient().GetXconfData(tableName, rowKey)
	if err != nil {
		return nil, err
	}

	obj := tableInfo.ConstructorFunc()  // Instantiate a new model/struct
	err = json.Unmarshal(jsonData, obj) // Deserialize the raw JSON []byte to a struct
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// SetOne set Xconf record
func (sd simpleDaoImpl) SetOne(tableName string, rowKey string, value []byte) error {
	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return err
	}

	err = GetDatabaseClient().SetXconfData(tableName, rowKey, value, tableInfo.TTL)
	return err
}

// DeleteOne delete Xconf record
func (sd simpleDaoImpl) DeleteOne(tableName string, rowKey string) error {
	err := GetDatabaseClient().DeleteXconfData(tableName, rowKey)
	return err
}

// GetAllByKeys get Xconf records for the specified list of rowKeys
func (sd simpleDaoImpl) GetAllByKeys(tableName string, rowKeys []string) ([]interface{}, error) {
	var result []interface{}

	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	// Get data from DB as a list of raw JSON []byte
	rows := GetDatabaseClient().GetAllXconfDataByKeys(tableName, rowKeys)
	for _, jsonData := range rows {
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
func (sd simpleDaoImpl) GetAllAsList(tableName string, maxResults int) ([]interface{}, error) {
	var result []interface{}

	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	// Get data from DB as a list of raw JSON []byte
	rows := GetDatabaseClient().GetAllXconfDataAsList(tableName, maxResults)
	for _, jsonData := range rows {
		obj := tableInfo.ConstructorFunc()  // Instantiate a new model/struct
		err = json.Unmarshal(jsonData, obj) // Deserialize the raw JSON []byte to a struct
		if err != nil {
			return nil, err
		}
		result = append(result, obj)
	}

	return result, err
}

// GetAllAsMap get a map of all Xconf records
func (sd simpleDaoImpl) GetAllAsMap(tableName string, maxResults int) (map[string]interface{}, error) {
	var result = make(map[string]interface{})

	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	// Get data from DB as a map of key and raw JSON []byte
	dataMap := GetDatabaseClient().GetAllXconfDataAsMap(tableName, maxResults)
	for key, jsonData := range dataMap {
		obj := tableInfo.ConstructorFunc()  // Instantiate a new model/struct
		err = json.Unmarshal(jsonData, obj) // Deserialize the raw JSON []byte to a struct
		if err != nil {
			return nil, err
		}
		result[key] = obj
	}

	return result, err
}

// GetKeys get all Xconf keys
func (sd simpleDaoImpl) GetKeys(tableName string) []string {
	return GetDatabaseClient().GetAllXconfKeys(tableName)
}
