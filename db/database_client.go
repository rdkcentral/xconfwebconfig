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
	"fmt"
	"strings"
	"time"
)

var dbClient DatabaseClient

func SetDatabaseClient(c DatabaseClient) {
	dbClient = c
}

func GetDatabaseClient() DatabaseClient {
	return dbClient
}

func IsCassandraClient() bool {
	_, ok := dbClient.(*CassandraClient)
	return ok
}

// TwoKeys Xconf key values
const TwowKeysDelimiter = "::"

type TwoKeys struct {
	Key  string
	Key2 interface{}
}

func NewTwoKeys(Key string, Key2 interface{}) *TwoKeys {
	return &TwoKeys{Key: Key, Key2: Key2}
}

func NewTwoKeysFromString(tk string) (*TwoKeys, error) {
	parts := strings.Split(tk, TwowKeysDelimiter)
	if len(parts) == 2 {
		return NewTwoKeys(parts[0], parts[1]), nil
	}

	return nil, fmt.Errorf("Invalid TwoKeys: value=%v", tk)
}

func (tk *TwoKeys) String() string {
	return fmt.Sprintf("%s%s%v", tk.Key, TwowKeysDelimiter, tk.Key2)
}

// RangeInfo Xconf key2 filtering
type RangeInfo struct {
	StartValue interface{}
	EndValue   interface{}
}

func (ri *RangeInfo) IsNilStartValue() bool {
	return ri.StartValue == nil || ri.StartValue == "" || ri.StartValue == 0
}

func (ri *RangeInfo) IsNilEndValue() bool {
	return ri.EndValue == nil || ri.EndValue == "" || ri.EndValue == 0
}

type DatabaseClient interface {
	SetUp() error
	TearDown() error
	Close() error
	Sleep()

	// Xconf
	QueryXconfDataRows(query string, queryParams ...string) ([]map[string]interface{}, error)
	ModifyXconfData(query string, queryParameters ...string) error

	// Batch operations
	NewBatch(batchType int) BatchOperation
	ExecuteBatch(batch BatchOperation) error

	SetXconfData(tableName string, rowKey string, value []byte, ttl int) error
	GetXconfData(tableName string, rowKey string) ([]byte, error)
	GetAllXconfDataByKeys(tableName string, rowKeys []string) [][]byte
	GetAllXconfKeys(tableName string) []string
	GetAllXconfDataAsList(tableName string, maxResults int) [][]byte
	GetAllXconfDataAsMap(tableName string, maxResults int) map[string][]byte
	DeleteXconfData(tableName string, rowKey string) error
	DeleteAllXconfData(tableName string) error

	// Xconf TwoKeys
	GetAllXconfData(tableName string, rowKey string) [][]byte
	GetAllXconfDataTwoKeysRange(tableName string, rowKey interface{}, key2FieldName string, rangeInfo *RangeInfo) [][]byte
	GetAllXconfDataTwoKeysAsMap(tableName string, rowKey string, key2FieldName string, key2List []interface{}) map[interface{}][]byte
	SetXconfDataTwoKeys(tableName string, rowKey interface{}, key2FieldName string, key2 interface{}, value []byte, ttl int) error
	GetXconfDataTwoKeys(tableName string, rowKey string, key2FieldName string, key2 interface{}) ([]byte, error)
	DeleteXconfDataTwoKeys(tableName string, rowKey string, key2FieldName string, key2 interface{}) error
	GetAllXconfTwoKeys(tableName string, key2FieldName string) []TwoKeys
	GetAllXconfKey2s(tableName string, rowKey string, key2FieldName string) []interface{}
	// Xconf compressed data
	SetXconfCompressedData(tableName string, rowKey string, values [][]byte, ttl int) error
	GetXconfCompressedData(tableName string, rowKey string) ([]byte, error)
	GetAllXconfCompressedDataAsMap(tableName string) map[string][]byte

	// Pod table lookup estbMac from pod serialNum
	GetEcmMacFromPodTable(string) (string, error)

	// not found
	IsDbNotFound(error) bool

	// Penetration Metrics
	GetPenetrationMetrics(macAddress string) (map[string]interface{}, error)
	SetPenetrationMetrics(penetrationmetrics *PenetrationMetrics) error
	SetFwPenetrationMetrics(*FwPenetrationMetrics) error
	GetFwPenetrationMetrics(string) (*FwPenetrationMetrics, error)
	SetRfcPenetrationMetrics(pMetrics *RfcPenetrationMetrics, is304FromPrecook bool) error
	GetRfcPenetrationMetrics(string) (*RfcPenetrationMetrics, error)
	UpdateFwPenetrationMetrics(map[string]string) error
	GetEstbIp(string) (string, error)

	SetRecookingStatus(module string, partitionId string, state int) error
	GetRecookingStatus(module string, partitionId string) (int, time.Time, error)
	CheckFinalRecookingStatus(module string) (bool, time.Time, error)

	// XPC precook reference data
	SetPrecookDataInXPC(RfcPrecookHash string, RfcPrecookPayload []byte) error
	GetPrecookDataFromXPC(RfcPrecookHash string) ([]byte, string, error)

	// Locks
	AcquireLock(lockName string, lockedBy string, ttlSeconds int) error
	ReleaseLock(lockName string, lockedBy string) error
	GetLockInfo(lockName string) (map[string]interface{}, error)
}

// BatchOperation interface for database batch operations
type BatchOperation interface {
	Query(stmt string, args ...interface{})
	Size() int
}

// Batch types constants
const (
	LoggedBatch = iota
	UnloggedBatch
	CounterBatch
)
