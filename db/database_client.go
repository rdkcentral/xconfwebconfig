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
	Key2 any
}

func NewTwoKeys(key string, key2 any) *TwoKeys {
	return &TwoKeys{Key: key, Key2: key2}
}

func NewTwoKeysFromString(tk string) (*TwoKeys, error) {
	parts := strings.Split(tk, TwowKeysDelimiter)
	if len(parts) == 2 {
		return NewTwoKeys(parts[0], parts[1]), nil
	}

	return nil, fmt.Errorf("Invalid TwoKeys: value=%v", tk)
}

func (tk *TwoKeys) String() string {
	return GetTwoKeysAsString(tk.Key, tk.Key2)
}

// GetTwoKeysAsString returns a string representation of two keys, e.g. "key1::key2"
func GetTwoKeysAsString(key string, key2 any) string {
	return key + TwowKeysDelimiter + fmt.Sprint(key2)
}

// RangeInfo Xconf key2 filtering
type RangeInfo struct {
	StartValue any
	EndValue   any
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
	QueryXconfDataRows(query string, queryParams ...string) ([]map[string]any, error)
	ModifyXconfData(query string, queryParameters ...string) error

	// Batch operations
	NewBatch(batchType int) BatchOperation
	ExecuteBatch(batch BatchOperation) error

	SetXconfData(tenantId string, tableName string, key string, value []byte, ttl int) error
	GetXconfData(tenantId string, tableName string, key string) ([]byte, error)
	GetAllXconfDataByKeys(tenantId string, tableName string, keys []string) [][]byte
	GetAllXconfKeys(tenantId string, tableName string) []string
	GetAllXconfDataAsList(tenantId string, tableName string, maxResults int) [][]byte
	GetAllXconfDataAsMap(tenantId string, tableName string, maxResults int) map[string][]byte
	DeleteXconfData(tenantId string, tableName string, key string) error
	DeleteAllXconfData(tenantId string, tableName string) error

	// Xconf TwoKeys
	GetAllXconfData(tenantId string, tableName string, key string) [][]byte
	GetAllXconfDataTwoKeysRange(tenantId string, tableName string, key any, key2FieldName string, rangeInfo *RangeInfo) [][]byte
	GetAllXconfDataTwoKeysAsMap(tenantId string, tableName string, key string, key2FieldName string, key2List []any) map[any][]byte
	SetXconfDataTwoKeys(tenantId string, tableName string, key any, key2FieldName string, key2 any, value []byte, ttl int) error
	GetXconfDataTwoKeys(tenantId string, tableName string, key string, key2FieldName string, key2 any) ([]byte, error)
	DeleteXconfDataTwoKeys(tenantId string, tableName string, key string, key2FieldName string, key2 any) error
	GetAllXconfTwoKeys(tenantId string, tableName string, key2FieldName string) []TwoKeys
	GetAllXconfKey2s(tenantId string, tableName string, key string, key2FieldName string) []any

	// Xconf compressed data
	SetXconfCompressedData(tenantId string, tableName string, key string, values [][]byte, ttl int) error
	GetXconfCompressedData(tenantId string, tableName string, key string) ([]byte, error)
	GetAllXconfCompressedDataAsMap(tenantId string, tableName string) map[string][]byte

	// Pod table lookup estbMac from pod serialNum
	GetEcmMacFromPodTable(string) (string, error)

	// not found
	IsDbNotFound(error) bool

	// Penetration Metrics
	GetPenetrationMetrics(macAddress string) (map[string]any, error)
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
	AcquireLock(tenantId string, lockName string, lockedBy string, ttlSeconds int) error
	ReleaseLock(tenantId string, lockName string, lockedBy string) error
	GetLockInfo(tenantId string, lockName string) (map[string]any, error)
}

// BatchOperation interface for database batch operations
type BatchOperation interface {
	Query(stmt string, args ...any)
	Size() int
}

// Batch types constants
const (
	LoggedBatch = iota
	UnloggedBatch
	CounterBatch
)
