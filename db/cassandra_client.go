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
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-akka/configuration"
	"github.com/gocql/gocql"

	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

const (
	ProtocolVersion               = 4
	DefaultKeyspace               = "ApplicationsDiscoveryDataService"
	DefaultTestKeyspace           = "ApplicationsDiscoveryDataServiceTest"
	DefaultDeviceKeyspace         = "odp"
	DefaultDeviceTestKeyspace     = "odp_test_keyspace"
	DefaultDevicePodTableName     = "pod_cpe_account"
	PenetrationMetricsTable       = "PenetrationMetrics"
	EstbMacColumnValue            = "estb_mac"
	DisableInitialHostLookup      = false
	DefaultSleepTimeInMillisecond = 10
	DefaultConnections            = 2
	DefaultColumnValue            = "data"
	NamedListPartColumnValue      = "NamedListData_part_"
	NamedListCountColumnValue     = "NamedListData_parts_count"
	DefaultPort                   = 9042
)

type CassandraClient struct {
	*gocql.Session
	*gocql.ClusterConfig
	sleepTime          int32
	concurrentQueries  chan bool
	localDc            string
	deviceKeyspace     string
	devicePodTableName string
	testOnly           bool
}

type PenetrationMetrics struct {
	EstbMac                 string
	Partner                 string
	Model                   string
	FwVersion               string
	FwReportedVersion       string
	FwAdditionalVersionInfo string
	FwAppliedRule           string
	FwTs                    time.Time
	RfcAppliedRules         string
	RfcFeatures             string
	RfcTs                   time.Time
}

func NewCassandraClient(conf *configuration.Config, testOnly bool) (*CassandraClient, error) {
	isIamBasedAccessEnabled := conf.GetBoolean("xconfwebconfig.database.iam_based_access_enabled")
	if isIamBasedAccessEnabled {
		return awsKeyspaceClient(conf, testOnly)
	} else {
		return cassandraClient(conf, testOnly)
	}
}

// Cassandra Impl of DatabaseClient
func (c *CassandraClient) GetPenetrationMetrics(estbMac string) (map[string]interface{}, error) {
	dict := util.Dict{}
	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()
	stmt := fmt.Sprintf("SELECT * FROM \"%s\" WHERE %s=?", PenetrationMetricsTable, EstbMacColumnValue)
	qry := c.Query(stmt, estbMac)
	err := qry.MapScan(dict)

	if err != nil {
		return dict, err
	}

	return dict, nil
}

func (c *CassandraClient) SetPenetrationMetrics(pMetrics *PenetrationMetrics) error {
	values := []interface{}{pMetrics.EstbMac, pMetrics.Partner, pMetrics.Model, pMetrics.FwVersion, pMetrics.FwReportedVersion, pMetrics.FwAdditionalVersionInfo, pMetrics.FwAppliedRule, pMetrics.FwTs, pMetrics.RfcAppliedRules, pMetrics.RfcFeatures, pMetrics.RfcTs}
	stmt := fmt.Sprintf(`INSERT INTO "%s" (estb_mac,partner,model,fw_version,fw_reported_version,fw_additional_version_info,fw_applied_rule,fw_ts,rfc_features,rfc_applied_rules,rfc_ts) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, PenetrationMetricsTable)
	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()
	qry := c.Query(stmt, values...)
	err := qry.Exec()

	if err != nil {
		return err
	}
	return nil
}

func (c *CassandraClient) Sleep() {
	time.Sleep(time.Duration(c.sleepTime) * time.Millisecond)
}

func (c *CassandraClient) LocalDc() string {
	return c.localDc
}

func (c *CassandraClient) Close() error {
	c.Session.Close()
	return nil
}

func (c *CassandraClient) IsDbNotFound(err error) bool {
	return errors.Is(err, gocql.ErrNotFound)
}

func (c *CassandraClient) IsTestOnly() bool {
	return c.testOnly
}

func (c *CassandraClient) DeviceKeyspace() string {
	return c.deviceKeyspace
}

func (c *CassandraClient) DevicePodTableName() string {
	return c.devicePodTableName
}

// SetXconfData Create XconfData for the specified key and value, where value is JSON data
func (c *CassandraClient) SetXconfData(tableName string, rowKey string, value []byte, ttl int) error {
	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var stmt string
	if ttl > 0 {
		stmt = fmt.Sprintf(`INSERT INTO "%s"(key, column1, value) VALUES(?,?,?) USING TTL %d`, tableName, ttl)
	} else {
		stmt = fmt.Sprintf(`INSERT INTO "%s"(key, column1, value) VALUES(?,?,?)`, tableName)
	}

	if err := c.Query(stmt, rowKey, DefaultColumnValue, value).Exec(); err != nil {
		return err
	}

	return nil
}

// GetXconfData Get one row where return value is JSON data
func (c *CassandraClient) GetXconfData(tableName string, rowKey string) ([]byte, error) {
	start := time.Now()

	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var value []byte

	stmt := fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ? AND column1 = ? LIMIT 1`, tableName)
	err := c.Query(stmt, rowKey, DefaultColumnValue).Scan(&value)
	if err != nil {
		return value, err
	}

	log.Debug(fmt.Sprintf("CassandraClient.GetXconfData: table %v rowKey %v in %v", tableName, rowKey, time.Since(start)))

	return value, nil
}

// GetAllXconfDataByKeys Get all rows as a list of values for the specified keys, where value is JSON data
func (c *CassandraClient) GetAllXconfDataByKeys(tableName string, rowKeys []string) [][]byte {
	start := time.Now()

	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var resultData [][]byte

	stmt := fmt.Sprintf(`SELECT key, value FROM "%s" WHERE KEY IN ?`, tableName)
	iter := c.Query(stmt, rowKeys).Iter()
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row["value"].([]byte))
	}

	log.Debug(fmt.Sprintf("CassandraClient.GetAllXconfDataByKeys: table %v rowKeys %v in %v", tableName, rowKeys, time.Since(start)))

	return resultData
}

// GetAllXconfKeys Get all keys
func (c *CassandraClient) GetAllXconfKeys(tableName string) []string {
	start := time.Now()

	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var resultData []string

	stmt := fmt.Sprintf(`SELECT DISTINCT key FROM "%s"`, tableName)
	iter := c.Query(stmt).Iter()
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row["key"].(string))
	}

	log.Debug(fmt.Sprintf("CassandraClient.GetAllXconfKeys: table %v in %v", tableName, time.Since(start)))

	return resultData
}

// GetAllXconfDataAsList Get all rows as a list of values, where value is JSON data
func (c *CassandraClient) GetAllXconfDataAsList(tableName string, maxResults int) [][]byte {
	start := time.Now()

	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var resultData [][]byte
	var stmt string
	if maxResults > 0 {
		stmt = fmt.Sprintf(`SELECT value FROM "%s" LIMIT %v`, tableName, maxResults)
	} else {
		stmt = fmt.Sprintf(`SELECT value FROM "%s"`, tableName)
	}

	iter := c.Query(stmt).Iter()
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row["value"].([]byte))
	}

	log.Debug(fmt.Sprintf("CassandraClient.GetAllXconfDataAsList: table %v in %v", tableName, time.Since(start)))

	return resultData
}

// GetAllXconfDataAsMap Get all rows as a map of key to value, where value is JSON data
func (c *CassandraClient) GetAllXconfDataAsMap(tableName string, maxResults int) map[string][]byte {
	start := time.Now()

	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var resultData = make(map[string][]byte)
	var stmt string
	if maxResults > 0 {
		stmt = fmt.Sprintf(`SELECT key, value FROM "%s" LIMIT %v`, tableName, maxResults)
	} else {
		stmt = fmt.Sprintf(`SELECT key, value FROM "%s"`, tableName)
	}

	iter := c.Query(stmt).Iter()
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		resultData[row["key"].(string)] = row["value"].([]byte)
	}

	log.Debug(fmt.Sprintf("CassandraClient.GetAllXconfDataAsMap: table %v in %v", tableName, time.Since(start)))

	return resultData
}

// DeleteXconfData Delete XconfData for the specified key
func (c *CassandraClient) DeleteXconfData(tableName string, rowKey string) error {
	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	stmt := fmt.Sprintf(`DELETE FROM "%s" WHERE key = ?`, tableName)
	if err := c.Query(stmt, rowKey).Exec(); err != nil {
		return err
	}

	return nil
}

// DeleteAllXconfData Delete all XconfData
func (c *CassandraClient) DeleteAllXconfData(tableName string) error {
	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	stmt := fmt.Sprintf(`TRUNCATE "%s"`, tableName)
	if err := c.Query(stmt).Exec(); err != nil {
		return err
	}

	return nil
}

// Two keys support

// GetAllXconfData Get multiple rows as a list of values, where value is JSON data
func (c *CassandraClient) GetAllXconfData(tableName string, rowKey string) [][]byte {
	start := time.Now()

	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var resultData [][]byte

	stmt := fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ?`, tableName)
	iter := c.Query(stmt, rowKey).Iter()
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row["value"].([]byte))
	}

	log.Debug(fmt.Sprintf("CassandraClient.GetAllXconfData: table %v rowKey %v in %v", tableName, rowKey, time.Since(start)))

	return resultData
}

// GetAllXconfDataTwoKeysRange Get multiple rows for the specified key and key2 range as list of values, where value is JSON data
func (c *CassandraClient) GetAllXconfDataTwoKeysRange(tableName string, rowKey interface{}, key2FieldName string, rangeInfo *RangeInfo) [][]byte {
	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var resultData [][]byte
	var stmt string
	var iter *gocql.Iter

	nilStartValue := true
	nilEndValue := true
	if rangeInfo != nil {
		nilStartValue = rangeInfo.IsNilStartValue()
		nilEndValue = rangeInfo.IsNilEndValue()
	}

	if nilStartValue && nilEndValue {
		stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ? ALLOW FILTERING`, tableName)
		iter = c.Query(stmt, rowKey).Iter()
	} else {
		if nilStartValue {
			if !nilEndValue {
				stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ? and %s < ? ALLOW FILTERING`, tableName, key2FieldName)
				iter = c.Query(stmt, rowKey, rangeInfo.EndValue).Iter()
			}
		} else {
			if nilEndValue {
				stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ? and %s > ? ALLOW FILTERING`, tableName, key2FieldName)
				iter = c.Query(stmt, rowKey, rangeInfo.StartValue).Iter()
			} else {
				stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ? and %s > ? and %s < ? ALLOW FILTERING`, tableName, key2FieldName, key2FieldName)
				iter = c.Query(stmt, rowKey, rangeInfo.StartValue, rangeInfo.EndValue).Iter()
			}
		}
	}

	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row["value"].([]byte))
	}

	return resultData
}

// GetAllXconfDataTwoKeysAsMap Get multiple rows for the specified key and key2 list as map of values, where value is JSON data
func (c *CassandraClient) GetAllXconfDataTwoKeysAsMap(tableName string, rowKey string, key2FieldName string, key2List []interface{}) map[interface{}][]byte {
	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var resultData = make(map[interface{}][]byte)

	stmt := fmt.Sprintf(`SELECT %s, value FROM "%s" WHERE key = ? and %s IN ?`, key2FieldName, tableName, key2FieldName)
	iter := c.Query(stmt, rowKey, key2List).Iter()
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		resultData[row[key2FieldName].(interface{})] = row["value"].([]byte)
	}

	return resultData
}

// SetXconfDataTwoKeys Create XconfData for the specified two keys and value, where value is JSON data
func (c *CassandraClient) SetXconfDataTwoKeys(tableName string, rowKey interface{}, key2FieldName string, key2 interface{}, value []byte, ttl int) error {
	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var stmt string
	if ttl > 0 {
		stmt = fmt.Sprintf(`INSERT INTO "%s"(key, %s, value) VALUES(?,?,?) USING TTL %d`, tableName, key2FieldName, ttl)
	} else {
		stmt = fmt.Sprintf(`INSERT INTO "%s"(key, %s, value) VALUES(?,?,?)`, tableName, key2FieldName)
	}

	if err := c.Query(stmt, rowKey, key2, value).Exec(); err != nil {
		return err
	}

	return nil
}

// GetXconfDataTwoKeys Get one row where return value is JSON data
func (c *CassandraClient) GetXconfDataTwoKeys(tableName string, rowKey string, key2FieldName string, key2 interface{}) ([]byte, error) {
	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var value []byte

	stmt := fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ? AND %s = ? LIMIT 1`, tableName, key2FieldName)
	err := c.Query(stmt, rowKey, key2).Scan(&value)
	if err != nil {
		return value, err
	}

	return value, nil
}

// DeleteXconfDataTwoKeys Delete XconfData for the specified two keys
func (c *CassandraClient) DeleteXconfDataTwoKeys(tableName string, rowKey string, key2FieldName string, key2 interface{}) error {
	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	stmt := fmt.Sprintf(`DELETE FROM "%s" WHERE key = ? AND %s = ?`, tableName, key2FieldName)
	if err := c.Query(stmt, rowKey, key2).Exec(); err != nil {
		return err
	}

	return nil
}

// GetAllXconfTwoKeys Get all TwoKeys
func (c *CassandraClient) GetAllXconfTwoKeys(tableName string, key2FieldName string) []TwoKeys {
	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var resultData []TwoKeys

	stmt := fmt.Sprintf(`SELECT key, "%s" FROM "%s"`, key2FieldName, tableName)
	iter := c.Query(stmt).Iter()
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}

		twoKeys := TwoKeys{
			Key:  row["key"].(string),
			Key2: row[key2FieldName].(interface{}),
		}
		resultData = append(resultData, twoKeys)
	}

	return resultData
}

// GetAllXconfKey2s Get a list of Xconf key2 for the specified rowKey
func (c *CassandraClient) GetAllXconfKey2s(tableName string, rowKey string, key2FieldName string) []interface{} {
	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var resultData []interface{}

	stmt := fmt.Sprintf(`SELECT %s FROM "%s" WHERE key = ?`, key2FieldName, tableName)
	iter := c.Query(stmt, rowKey).Iter()
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row[key2FieldName].(interface{}))
	}

	return resultData
}

// SetXconfCompressedData Create XconfData for the specified key and values, where values is compressed JSON data
func (c *CassandraClient) SetXconfCompressedData(tableName string, rowKey string, values [][]byte, ttl int) error {
	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	batch := c.NewBatch(gocql.LoggedBatch)

	// Add a record that specifies the number of compressed data chunks
	var stmt string
	if ttl > 0 {
		stmt = fmt.Sprintf(`INSERT INTO "%s"(key, column1, value) VALUES(?,?,intAsBlob(?)) USING TTL %d`, tableName, ttl)
	} else {
		stmt = fmt.Sprintf(`INSERT INTO "%s"(key, column1, value) VALUES(?,?,intAsBlob(?))`, tableName)
	}
	batch.Query(stmt, rowKey, NamedListCountColumnValue, len(values))

	for i, value := range values {
		// Add a record for each compressed data chunk where key has the format: NamedListData_part_0, ...
		partColumnValue := NamedListPartColumnValue + strconv.Itoa(i)
		if ttl > 0 {
			stmt = fmt.Sprintf(`INSERT INTO "%s"(key, column1, value) VALUES(?,?,?) USING TTL %d`, tableName, ttl)
		} else {
			stmt = fmt.Sprintf(`INSERT INTO "%s"(key, column1, value) VALUES(?,?,?)`, tableName)
		}
		batch.Query(stmt, rowKey, partColumnValue, value)
	}

	if err := c.ExecuteBatch(batch); err != nil {
		return err
	}

	return nil
}

// GetXconfCompressedData Get one row where return value is compressed JSON data
func (c *CassandraClient) GetXconfCompressedData(tableName string, rowKey string) ([]byte, error) {
	start := time.Now()

	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	// Get the number of compressed data chunks
	var partsCount int
	stmt := fmt.Sprintf(`SELECT blobAsInt(value) FROM "%s" WHERE key = ? AND column1 = ? LIMIT 1`, tableName)
	err := c.Query(stmt, rowKey, NamedListCountColumnValue).Scan(&partsCount)
	if err != nil {
		return nil, err
	}

	// Get all the compressed data chunks
	var partsMap = make(map[string][]byte)
	stmt = fmt.Sprintf(`SELECT key, column1, value FROM "%s" WHERE key = ?`, tableName)
	iter := c.Query(stmt, rowKey).Iter()
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}

		partName := row["column1"].(string)
		if partName != NamedListCountColumnValue {
			partsMap[partName] = row["value"].([]byte)
		}
	}

	// Ensure all the parts are loaded
	if partsCount != len(partsMap) {
		err := fmt.Errorf("Inconsistent compressed data for key '%v' from '%v': expected %v record(s) got %v",
			rowKey, tableName, partsCount, len(partsMap))
		return nil, err
	}

	// Combine all the compressed data chunks into one
	var chunks [][]byte
	for i := 0; i < partsCount; i++ {
		key := NamedListPartColumnValue + strconv.Itoa(i)
		chunk := partsMap[key]
		chunks = append(chunks, chunk)
	}

	resultData := bytes.Join(chunks, []byte(""))

	log.Debug(fmt.Sprintf("CassandraClient.GetXconfCompressedData: table %v rowKey %v in %v", tableName, rowKey, time.Since(start)))

	return resultData, nil
}

// GetAllXconfCompressedDataAsMap Get all rows as a map of key to value, where value is compressed JSON data
func (c *CassandraClient) GetAllXconfCompressedDataAsMap(tableName string) map[string][]byte {
	start := time.Now()

	var resultData = make(map[string][]byte)

	rawData := c.GetXconfCompressedDataRaw(tableName)
	for key, partsMap := range rawData {
		// Combine all the compressed data chunks into one
		partsCount := len(partsMap)
		var chunks [][]byte
		for i := 0; i < partsCount; i++ {
			partKey := NamedListPartColumnValue + strconv.Itoa(i)
			chunk := partsMap[partKey]
			chunks = append(chunks, chunk)
		}
		data := bytes.Join(chunks, []byte(""))
		resultData[key] = data
	}

	log.Debug(fmt.Sprintf("CassandraClient.GetAllXconfCompressedDataAsMap: table %v in %v", tableName, time.Since(start)))

	return resultData
}

// GetXconfCompressedDataRaw Get all rows as a map of key to another map,
// where key specifies part number and value is compressed JSON data chunk.
//
// Sample data for one record in GenericXconfNamedList table:
//
// key               | column1                   | value
// -------------------+---------------------------+-----------------------------
// Test_Mac_List     |      NamedListData_part_0 | 0x7df05a7b226964223a2241...
// Test_Mac_List     |      NamedListData_part_1 | 0x60f05f7b226964223a2231...
// Test_Mac_List     | NamedListData_parts_count |                  0x00000002
func (c *CassandraClient) GetXconfCompressedDataRaw(tableName string) map[string]map[string][]byte {
	start := time.Now()

	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	var resultData = make(map[string]map[string][]byte)
	var countMap = make(map[string]int)

	// Get all the count records
	stmt := fmt.Sprintf(`SELECT key, blobAsInt(value) as count FROM "%s" where column1 = ? ALLOW FILTERING`, tableName)
	iter := c.Query(stmt, NamedListCountColumnValue).Iter()
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		countMap[row["key"].(string)] = row["count"].(int)
	}

	// Get all the compressed data chunks
	stmt = fmt.Sprintf(`SELECT key, column1, value FROM "%s"`, tableName)
	iter = c.Query(stmt).Iter()
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}

		column1 := row["column1"].(string)
		if column1 == NamedListCountColumnValue {
			continue // Ignored count record which has already been processed
		} else {
			key := row["key"].(string)
			partsMap := resultData[key]
			if partsMap == nil {
				partsMap = make(map[string][]byte)
				resultData[key] = partsMap
			}
			count := countMap[key]
			if len(partsMap) >= count {
				continue // skip extra data
			}
			partsMap[column1] = row["value"].([]byte)
		}
	}

	// Ensure all the parts are loaded
	for key, partsMap := range resultData {
		partsCount := countMap[key]
		if partsCount != len(partsMap) {
			log.Warn(fmt.Sprintf("Inconsistent compressed data for key '%v' from '%v': expected %v record(s) got %v",
				key, tableName, partsCount, len(partsMap)))

			// Deleting the wrong data! Need to delete partsmap[key][extra_NamedList_data_part_1,2,3..]
			// delete(partsMap, key) // Ignored invalid record
		}
	}

	log.Debug(fmt.Sprintf("CassandraClient.GetXconfCompressedDataRaw: table %v in %v", tableName, time.Since(start)))

	return resultData
}
