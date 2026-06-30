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

	"github.com/golang/snappy"
	log "github.com/sirupsen/logrus"
)

const (
	ColumnValuePrefix     = "NamedListData"
	PartsCountColumnValue = ColumnValuePrefix + "_parts_count"
	CompressionChunkSize  = (64 * 1024) // In bytes
)

// CompressingDataDao interface
type CompressingDataDao interface {
	GetOne(tenantId string, tableName string, key string) (any, error)
	SetOne(tenantId string, tableName string, key string, value []byte) error
	DeleteOne(tenantId string, tableName string, key string) error
	GetAllByKeys(tenantId string, tableName string, keys []string) ([]any, error)
	GetAllAsList(tenantId string, tableName string, continueOnError bool) ([]any, error)
	GetAllAsMap(tenantId string, tableName string, continueOnError bool) (map[string]any, error)
	GetKeys(tenantId string, tableName string) []string
}

type compressingDataDaoImpl struct{}

var compressingDataDao = compressingDataDaoImpl{}

// GetCompressingDataDao return an implementation of CompressingDataDao
func GetCompressingDataDao() CompressingDataDao {
	return compressingDataDao
}

// GetOne get one compressed Xconf record
func (cd compressingDataDaoImpl) GetOne(tenantId string, tableName string, key string) (any, error) {
	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	// Get data from DB as compressed JSON []byte
	compressedData, err := GetDatabaseClient().GetXconfCompressedData(tenantId, tableName, key)
	if err != nil {
		return nil, err
	}

	jsonData, err := decompress(compressedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress tenantId %s table %s key %s: %w", tenantId, tableName, key, err)
	}

	obj := tableInfo.ConstructorFunc()  // Instantiate a new model/struct
	err = json.Unmarshal(jsonData, obj) // Deserialize the raw JSON []byte to a struct
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// SetOne set compressed Xconf record
func (cd compressingDataDaoImpl) SetOne(tenantId string, tableName string, key string, value []byte) error {
	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return err
	}

	// First compress the JSON data then split it into multiple chunks
	compressedData := compress(value)
	values := split(compressedData, CompressionChunkSize)

	err = GetDatabaseClient().SetXconfCompressedData(tenantId, tableName, key, values, tableInfo.TTL)
	return err
}

// DeleteOne delete Xconf record
func (cd compressingDataDaoImpl) DeleteOne(tenantId string, tableName string, key string) error {
	err := GetDatabaseClient().DeleteXconfData(tenantId, tableName, key)
	return err
}

// GetAllByKeys get Xconf records for the specified list of keys
func (cd compressingDataDaoImpl) GetAllByKeys(tenantId string, tableName string, keys []string) ([]any, error) {
	var result []any

	// Process one compressed record at a time
	for _, key := range keys {
		obj, err := cd.GetOne(tenantId, tableName, key)
		if err != nil {
			return nil, err
		}
		result = append(result, obj)
	}

	return result, nil
}

// GetAllAsList get a list of all Xconf records
func (cd compressingDataDaoImpl) GetAllAsList(tenantId string, tableName string, continueOnError bool) ([]any, error) {
	resultMap, err := cd.GetAllAsMap(tenantId, tableName, continueOnError)
	if err != nil {
		return nil, err
	}

	result := make([]any, 0, len(resultMap))
	for _, value := range resultMap {
		result = append(result, value)
	}

	return result, nil
}

// GetAllAsMap get a map of all Xconf records
func (cd compressingDataDaoImpl) GetAllAsMap(tenantId string, tableName string, continueOnError bool) (map[string]any, error) {
	var result = make(map[string]any)

	tableInfo, err := GetTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	// Get data from DB as a map of key and compressed JSON []byte
	compressedDataMap := GetDatabaseClient().GetAllXconfCompressedDataAsMap(tenantId, tableName)
	for key, compressedData := range compressedDataMap {
		jsonData, err := decompress(compressedData)
		if err != nil {
			err = fmt.Errorf("failed to decompress table %s key %s: %w", tableName, key, err)
			if continueOnError {
				log.WithFields(log.Fields{"tenantId": tenantId}).Error(err)
				continue
			}
			return nil, err
		}

		obj := tableInfo.ConstructorFunc()  // Instantiate a new model/struct
		err = json.Unmarshal(jsonData, obj) // Deserialize the raw JSON []byte to a struct
		if err != nil {
			return nil, err
		}
		result[key] = obj
	}

	return result, nil
}

// GetKeys get all Xconf keys
func (cd compressingDataDaoImpl) GetKeys(tenantId string, tableName string) []string {
	return GetDatabaseClient().GetAllXconfKeys(tenantId, tableName)
}

func compress(data []byte) []byte {
	compressedData := snappy.Encode(nil, data)
	return compressedData
}

func decompress(data []byte) ([]byte, error) {
	decompressedData, err := snappy.Decode(nil, data)
	return decompressedData, err
}

func split(data []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunks = append(chunks, data[i:end])
	}

	return chunks
}
