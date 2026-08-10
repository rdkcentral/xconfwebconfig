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
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-akka/configuration"
	"github.com/gocql/gocql"

	"github.com/rdkcentral/xconfwebconfig/security"
	"github.com/rdkcentral/xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

const (
	ProtocolVersion                      = 4
	DefaultKeyspace                      = "xconf"
	DefaultTestKeyspace                  = "xconf_test"
	DefaultLogKeyspace                   = "ApplicationsDiscoveryDataService"
	DefaultLogTestKeyspace               = "ApplicationsDiscoveryDataServiceTest"
	DefaultDeviceKeyspace                = "odp"
	DefaultDeviceTestKeyspace            = "odp_test_keyspace"
	DefaultDevicePodTableName            = "pod_cpe_account"
	DisableInitialHostLookup             = false
	DefaultSleepTimeInMillisecond        = 10
	DefaultConnections                   = 2
	NamedListPartColumnValue             = "NamedListData_part_"
	NamedListCountColumnValue            = "NamedListData_parts_count"
	DefaultXpcKeyspace                   = "xpc"
	DefaultXpcTestKeyspace               = "xpc_test_keyspace"
	DefaultXpcPrecookTableName           = "reference_document"
	DefaultXconfRecookingStatusTableName = "recooking_status"
	LockNameDelimiter                    = "|"

	// DO NOT CHANGE UNLESS YOU KNOW WHAT YOU ARE DOING
	ScalingFactor = 8 // number of shards (nodes) to distribute data across
)

var shardIds = GetShardIds() // parameter value for IN clause to query across all shards

// Interface used for connecting to Cassandra in a cloud environment
type CassandraConnector interface {
	NewCassandraClient(conf *configuration.Config, testOnly bool) (*CassandraClient, error)
}

// example Default connector
type DefaultCassandraConnection struct {
	Connection_type string
}

type CassandraClient struct {
	*gocql.Session
	*gocql.ClusterConfig
	SleepTime                     int32
	ConcurrentQueries             chan bool
	LocalDc                       string
	Connection_type               string
	testOnly                      bool
	addsKeyspace                  string
	deviceKeyspace                string
	devicePodTableName            string
	xpcKeyspace                   string
	xpcPrecookTableName           string
	xconfRecookingStatusTableName string
}

type DistributedLockSettings struct {
	retries      int
	retryInMsecs int
}

var distributedLockSettings = DistributedLockSettings{}

// BatchWrapper wraps gocql.Batch to implement BatchOperation interface
type BatchWrapper struct {
	*gocql.Batch
}

func (bw *BatchWrapper) Query(stmt string, args ...any) {
	bw.Batch.Query(stmt, args...)
}

func (bw *BatchWrapper) Size() int {
	return bw.Batch.Size()
}

func (ca *DefaultCassandraConnection) NewCassandraClient(conf *configuration.Config, testOnly bool) (*CassandraClient, error) {
	distributedLockSettings.retries = int(conf.GetInt32("xconfwebconfig.xconf.distributed_lock_retries", 0))
	distributedLockSettings.retryInMsecs = int(conf.GetInt32("xconfwebconfig.xconf.distributed_lock_retry_in_msecs", 200))

	// init
	log.Debug("Connecting to Cassandra with DefaultCassandraConnection")
	hosts := conf.GetStringList("xconfwebconfig.database.hosts")
	cluster := gocql.NewCluster(hosts...)

	cluster.Consistency = gocql.LocalQuorum
	cluster.ProtoVersion = int(conf.GetInt32("xconfwebconfig.database.protocolversion", ProtocolVersion))
	cluster.DisableInitialHostLookup = DisableInitialHostLookup
	cluster.Timeout = time.Duration(conf.GetInt32("xconfwebconfig.database.timeout_in_sec", 1)) * time.Second
	cluster.ConnectTimeout = time.Duration(conf.GetInt32("xconfwebconfig.database.connect_timeout_in_sec", 1)) * time.Second
	cluster.NumConns = int(conf.GetInt32("xconfwebconfig.database.connections", DefaultConnections))

	cluster.RetryPolicy = &gocql.DowngradingConsistencyRetryPolicy{
		[]gocql.Consistency{
			gocql.LocalQuorum,
			gocql.LocalOne,
			gocql.One,
		},
	}

	localDc := conf.GetString("xconfwebconfig.database.local_dc")
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
	if len(localDc) > 0 {
		cluster.HostFilter = gocql.DataCentreHostFilter(localDc)
	}

	isSslEnabled := conf.GetBoolean("xconfwebconfig.database.is_ssl_enabled")

	// credentials from environment takes precedence over config file
	user := os.Getenv("DATABASE_USER")
	if util.IsBlank(user) {
		user = conf.GetString("xconfwebconfig.database.user")
		if util.IsBlank(user) {
			return nil, errors.New("no env DATABASE_USER")
		}
	}

	var password string
	var err error

	encryptedPassword := os.Getenv("DATABASE_ENCRYPTED_PASSWORD")
	if util.IsBlank(encryptedPassword) {
		encryptedPassword = conf.GetString("xconfwebconfig.database.encrypted_password")
	}
	if util.IsBlank(encryptedPassword) {
		password = os.Getenv("DATABASE_PASSWORD")
		if util.IsBlank(password) {
			password = conf.GetString("xconfwebconfig.database.password")
			if util.IsBlank(password) {
				return nil, errors.New("no env DATABASE_PASSWORD or DATABASE_ENCRYPTED_PASSWORD")
			}
		}
	} else {
		xpckeyB64 := ""

		envs := os.Environ()
		for _, line := range envs {
			if len(line) > 8 {
				prefix := line[:8]
				if prefix == "XPC_KEY=" {
					xpckeyB64 = line[8:]
					break
				}
			}
			// fmt.Println(v)
		}

		if xpckeyB64 == "" {
			panic(fmt.Errorf("missing env XPC_KEY"))
		}

		codec := security.NewAesCodec(xpckeyB64)
		password, err = codec.Decrypt(encryptedPassword)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: user,
		Password: password,
	}

	if isSslEnabled {
		sslOpts := &gocql.SslOptions{
			EnableHostVerification: false,
		}
		sslServerName := conf.GetString("xconfwebconfig.database.ssl_server_name")
		if len(sslServerName) > 0 {
			sslOpts.Config = &tls.Config{
				ServerName:         sslServerName,
				InsecureSkipVerify: true,
			}
			sslOpts.EnableHostVerification = true
		}
		cluster.SslOpts = sslOpts
	}

	// Use the appropriate keyspace
	var addsKeyspace string
	var deviceKeyspace string
	var session *gocql.Session

	// now point to the real keyspace
	if testOnly {
		cluster.Keyspace = conf.GetString("xconfwebconfig.database.test_keyspace", DefaultTestKeyspace)
		deviceKeyspace = conf.GetString("xconfwebconfig.database.device_test_keyspace", DefaultDeviceTestKeyspace)
		addsKeyspace = conf.GetString("xconfwebconfig.database.test_keyspace", DefaultLogKeyspace)
	} else {
		cluster.Keyspace = conf.GetString("xconfwebconfig.database.keyspace", DefaultKeyspace)
		deviceKeyspace = conf.GetString("xconfwebconfig.database.device_keyspace", DefaultDeviceKeyspace)
		addsKeyspace = conf.GetString("xconfwebconfig.database.adds_keyspace", DefaultLogTestKeyspace)
	}
	log.Debug(fmt.Sprintf("Init CassandraClient with keyspace: %v", cluster.Keyspace))

	xpcKeyspace := conf.GetString("xconfwebconfig.database.xpc_keyspace", DefaultXpcKeyspace)
	xpcPrecookTableName := conf.GetString("xconfwebconfig.database.xpc_precook_table_name", DefaultXpcPrecookTableName)
	xconfRecookingStatusTableName := conf.GetString("xconfwebconfig.database.xconf_recooking_status_table_name", DefaultXconfRecookingStatusTableName)

	session, err = cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	devicePodTableName := conf.GetString("xconfwebconfig.database.device_pod_table_name", DefaultDevicePodTableName)

	return &CassandraClient{
		Session:                       session,
		ClusterConfig:                 cluster,
		SleepTime:                     conf.GetInt32("xconfwebconfig.perftest.sleep_in_msecs", DefaultSleepTimeInMillisecond),
		ConcurrentQueries:             make(chan bool, conf.GetInt32("xconfwebconfig.database.concurrent_queries", 500)),
		LocalDc:                       localDc,
		Connection_type:               ca.Connection_type,
		testOnly:                      testOnly,
		addsKeyspace:                  addsKeyspace,
		deviceKeyspace:                deviceKeyspace,
		devicePodTableName:            devicePodTableName,
		xpcKeyspace:                   xpcKeyspace,
		xpcPrecookTableName:           xpcPrecookTableName,
		xconfRecookingStatusTableName: xconfRecookingStatusTableName,
	}, nil
}

func (c *CassandraClient) XpcKeyspace() string {
	return c.xpcKeyspace
}

func (c *CassandraClient) XpcPrecookTableName() string {
	return c.xpcPrecookTableName
}

func (c *CassandraClient) XconfRecookingStatusTableName() string {
	return c.xconfRecookingStatusTableName
}

// Cassandra Impl of DatabaseClient

func (c *CassandraClient) Sleep() {
	time.Sleep(time.Duration(c.SleepTime) * time.Millisecond)
}

func (c *CassandraClient) GetLocalDc() string {
	return c.LocalDc
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

func (c *CassandraClient) GetDeviceKeyspace() string {
	return c.deviceKeyspace
}

func (c *CassandraClient) GetDevicePodTableName() string {
	return c.devicePodTableName
}

func (c *CassandraClient) GetLogKeyspace() string {
	return c.addsKeyspace
}

// SetXconfData Create XconfData for the specified key and value, where value is JSON data
func (c *CassandraClient) SetXconfData(tenantId string, tableName string, key string, value []byte, updatedAt int64, ttl int) error {
	if tenantId == "" {
		return fmt.Errorf("CassandraClient.SetXconfData: tenantId is empty, table %s", tableName)
	}

	if updatedAt == 0 {
		updatedAt = util.GetTimestamp()
	}

	var stmt string
	if ttl > 0 {
		stmt = fmt.Sprintf(`INSERT INTO %s(tenant_id, shard_id, key, value, updated) VALUES(?,?,?,?,?) USING TTL %d`, tableName, ttl)
	} else {
		stmt = fmt.Sprintf(`INSERT INTO %s(tenant_id, shard_id, key, value, updated) VALUES(?,?,?,?,?)`, tableName)
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	return c.Query(stmt, tenantId, GetShardId(key), key, value, updatedAt).Exec()
}

// GetXconfData Get one row where return value is JSON data
func (c *CassandraClient) GetXconfData(tenantId string, tableName string, key string) (value []byte, err error) {
	if tenantId == "" {
		return value, fmt.Errorf("CassandraClient.GetXconfData: tenantId is empty, table %s", tableName)
	}

	stmt := fmt.Sprintf(`SELECT value FROM %s WHERE tenant_id = ? AND shard_id = ? AND key = ? LIMIT 1`, tableName)

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	err = c.Query(stmt, tenantId, GetShardId(key), key).Scan(&value)

	return value, err
}

// GetAllXconfDataByKeys Get all rows as a list of values for the specified keys, where value is JSON data
func (c *CassandraClient) GetAllXconfDataByKeys(tenantId string, tableName string, keys []string) (resultData [][]byte) {
	for _, key := range keys {
		// concurrency will be handled inside GetXconfData method, so no need to add concurrency here
		data, err := c.GetXconfData(tenantId, tableName, key)
		if err != nil {
			if !c.IsDbNotFound(err) {
				log.WithFields(log.Fields{"tenantId": tenantId}).Warnf("CassandraClient.GetAllXconfDataByKeys: failed to get data for table %s, key %s: %v", tableName, key, err)
			}
			continue
		}
		resultData = append(resultData, data)
	}

	return resultData
}

// GetAllXconfKeys Get all keys
func (c *CassandraClient) GetAllXconfKeys(tenantId string, tableName string) []string {
	if tenantId == "" {
		log.Errorf("CassandraClient.GetAllXconfKeys: tenantId is empty, table %s", tableName)
		return []string{}
	}

	resultData := util.Set{}
	stmt := fmt.Sprintf(`SELECT key FROM %s WHERE tenant_id = ? AND shard_id IN ?`, tableName)

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	iter := c.Query(stmt, tenantId, shardIds).Iter()
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		resultData.Add(row["key"].(string))
	}

	return resultData.ToSlice()
}

// GetAllXconfDataAsList Get all rows as a list of values, where value is JSON data
func (c *CassandraClient) GetAllXconfDataAsList(tenantId string, tableName string, maxResults int) (resultData [][]byte) {
	if tenantId == "" {
		log.Errorf("CassandraClient.GetAllXconfDataAsList: tenantId is empty, table %s", tableName)
		return resultData
	}

	var stmt string
	if maxResults > 0 {
		stmt = fmt.Sprintf(`SELECT value FROM %s WHERE tenant_id = ? AND shard_id IN ? LIMIT %v`, tableName, maxResults)
	} else {
		stmt = fmt.Sprintf(`SELECT value FROM %s WHERE tenant_id = ? AND shard_id IN ?`, tableName)
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	iter := c.Query(stmt, tenantId, shardIds).Iter()
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row["value"].([]byte))
	}

	return resultData
}

// GetAllXconfDataAsMap Get all rows as a map of key to value, where value is JSON data
func (c *CassandraClient) GetAllXconfDataAsMap(tenantId string, tableName string, maxResults int) map[string][]byte {
	var resultData = make(map[string][]byte)

	if tenantId == "" {
		log.Errorf("CassandraClient.GetAllXconfDataAsMap: tenantId is empty, table %s", tableName)
		return resultData
	}

	var stmt string
	if maxResults > 0 {
		stmt = fmt.Sprintf(`SELECT key, value FROM %s WHERE tenant_id = ? AND shard_id IN ? LIMIT %v`, tableName, maxResults)
	} else {
		stmt = fmt.Sprintf(`SELECT key, value FROM %s WHERE tenant_id = ? AND shard_id IN ?`, tableName)
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	iter := c.Query(stmt, tenantId, shardIds).Iter()
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		resultData[row["key"].(string)] = row["value"].([]byte)
	}

	return resultData
}

// DeleteXconfData Delete XconfData for the specified tenant, table, and key
func (c *CassandraClient) DeleteXconfData(tenantId string, tableName string, unsharded bool, key string) error {
	if tableName == TABLE_LOGS {
		tableName = c.GetTableNameFromLogKeyspace(tableName)
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// table is not sharded (no shard_id column), so we only need to delete by key
	if unsharded {
		stmt := fmt.Sprintf(`DELETE FROM %s WHERE key = ?`, tableName)
		return c.Query(stmt, key).Exec()
	} else {
		if tenantId == "" {
			return fmt.Errorf("CassandraClient.DeleteXconfData: tenantId is empty, table %s", tableName)
		}
		stmt := fmt.Sprintf(`DELETE FROM %s WHERE tenant_id = ? AND shard_id = ? AND key = ?`, tableName)
		return c.Query(stmt, tenantId, GetShardId(key), key).Exec()
	}
}

// DeleteAllXconfData Delete all XconfData for the specified tenant and table
func (c *CassandraClient) DeleteAllXconfData(tenantId string, tableName string) error {
	if tenantId == "" {
		return fmt.Errorf("CassandraClient.DeleteAllXconfData: tenantId is empty, table %s", tableName)
	}

	if tableName == TABLE_LOGS {
		tableName = c.GetTableNameFromLogKeyspace(tableName)
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	stmt := fmt.Sprintf(`DELETE FROM %s WHERE tenant_id = ? AND shard_id IN ?`, tableName)
	return c.Query(stmt, tenantId, shardIds).Exec()
}

// Two keys support

// GetAllXconfData Get multiple rows as a list of values, where value is JSON data
func (c *CassandraClient) GetAllXconfData(tenantId string, tableName string, unsharded bool, key string) (resultData [][]byte) {
	var iter *gocql.Iter

	if tableName == TABLE_LOGS {
		tableName = c.GetTableNameFromLogKeyspace(tableName)
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// table is not sharded and does not have shard_id column, so we only need to query by key
	if unsharded {
		stmt := fmt.Sprintf(`SELECT value FROM %s WHERE key = ?`, tableName)
		iter = c.Query(stmt, key).Iter()
	} else {
		if tenantId == "" {
			log.Errorf("CassandraClient.GetAllXconfData: tenantId is empty, table %s", tableName)
			return resultData
		}
		stmt := fmt.Sprintf(`SELECT value FROM %s WHERE tenant_id = ? AND shard_id = ? AND key = ?`, tableName)
		iter = c.Query(stmt, tenantId, GetShardId(key), key).Iter()
	}

	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row["value"].([]byte))
	}

	return resultData
}

// GetAllXconfDataTwoKeysRange Get multiple rows for the specified key and key2 range as list of values, where value is JSON data
func (c *CassandraClient) GetAllXconfDataTwoKeysRange(tenantId string, tableName string, unsharded bool, key any, rangeInfo *RangeInfo) [][]byte {
	var resultData [][]byte
	var stmt string
	var iter *gocql.Iter

	nilStartValue := true
	nilEndValue := true
	if rangeInfo != nil {
		nilStartValue = rangeInfo.IsNilStartValue()
		nilEndValue = rangeInfo.IsNilEndValue()
	}

	key2FieldName := DefaultKey2FieldName
	if tableName == TABLE_LOGS {
		tableName = c.GetTableNameFromLogKeyspace(tableName)
		key2FieldName = Key2FieldNameForLogs2
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// table is not sharded and does not have shard_id column, so we only need to query by key
	if unsharded {
		if nilStartValue && nilEndValue {
			stmt = fmt.Sprintf(`SELECT value FROM %s WHERE key = ? ALLOW FILTERING`, tableName)
			iter = c.Query(stmt, key).Iter()
		} else {
			if nilStartValue {
				if !nilEndValue {
					stmt = fmt.Sprintf(`SELECT value FROM %s WHERE key = ? and %s < ? ALLOW FILTERING`, tableName, key2FieldName)
					iter = c.Query(stmt, key, rangeInfo.EndValue).Iter()
				}
			} else {
				if nilEndValue {
					stmt = fmt.Sprintf(`SELECT value FROM %s WHERE key = ? and %s > ? ALLOW FILTERING`, tableName, key2FieldName)
					iter = c.Query(stmt, key, rangeInfo.StartValue).Iter()
				} else {
					stmt = fmt.Sprintf(`SELECT value FROM %s WHERE key = ? and %s > ? and %s < ? ALLOW FILTERING`, tableName, key2FieldName, key2FieldName)
					iter = c.Query(stmt, key, rangeInfo.StartValue, rangeInfo.EndValue).Iter()
				}
			}
		}
	} else {
		if tenantId == "" {
			log.Errorf("CassandraClient.GetAllXconfDataTwoKeysRange: tenantId is empty, table %s", tableName)
			return resultData
		}

		if nilStartValue && nilEndValue {
			stmt = fmt.Sprintf(`SELECT value FROM %s WHERE tenant_id = ? AND shard_id = ? AND key = ? ALLOW FILTERING`, tableName)
			iter = c.Query(stmt, tenantId, GetShardId(key), key).Iter()
		} else {
			if nilStartValue {
				if !nilEndValue {
					stmt = fmt.Sprintf(`SELECT value FROM %s WHERE tenant_id = ? AND shard_id = ? AND key = ? and %s < ? ALLOW FILTERING`, tableName, key2FieldName)
					iter = c.Query(stmt, tenantId, GetShardId(key), key, rangeInfo.EndValue).Iter()
				}
			} else {
				if nilEndValue {
					stmt = fmt.Sprintf(`SELECT value FROM %s WHERE tenant_id = ? AND shard_id = ? AND key = ? and %s > ? ALLOW FILTERING`, tableName, key2FieldName)
					iter = c.Query(stmt, tenantId, GetShardId(key), key, rangeInfo.StartValue).Iter()
				} else {
					stmt = fmt.Sprintf(`SELECT value FROM %s WHERE tenant_id = ? AND shard_id = ? AND key = ? and %s > ? and %s < ? ALLOW FILTERING`, tableName, key2FieldName, key2FieldName)
					iter = c.Query(stmt, tenantId, GetShardId(key), key, rangeInfo.StartValue, rangeInfo.EndValue).Iter()
				}
			}
		}
	}

	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row["value"].([]byte))
	}

	return resultData
}

// GetAllXconfDataTwoKeysAsMap Get multiple rows for the specified key and key2 list as map of values, where value is JSON data
func (c *CassandraClient) GetAllXconfDataTwoKeysAsMap(tenantId string, tableName string, unsharded bool, key string, key2List []any) map[any][]byte {
	var resultData = make(map[any][]byte)
	var iter *gocql.Iter

	key2FieldName := DefaultKey2FieldName
	if tableName == TABLE_LOGS {
		tableName = c.GetTableNameFromLogKeyspace(tableName)
		key2FieldName = Key2FieldNameForLogs2
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// table is not sharded and does not have shard_id column, so we only need to query by key
	if unsharded {
		stmt := fmt.Sprintf(`SELECT %s, value FROM %s WHERE key = ? and %s IN ?`, key2FieldName, tableName, key2FieldName)
		iter = c.Query(stmt, key, key2List).Iter()
	} else {
		if tenantId == "" {
			log.Errorf("CassandraClient.GetAllXconfDataTwoKeysAsMap: tenantId is empty, table %s", tableName)
			return resultData
		}

		stmt := fmt.Sprintf(`SELECT %s, value FROM %s WHERE tenant_id = ? AND shard_id = ? AND key = ? and %s IN ?`, key2FieldName, tableName, key2FieldName)
		iter = c.Query(stmt, tenantId, GetShardId(key), key, key2List).Iter()
	}

	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		resultData[row[key2FieldName]] = row["value"].([]byte)
	}

	return resultData
}

// SetXconfDataTwoKeys Create XconfData for the specified two keys and value, where value is JSON data
func (c *CassandraClient) SetXconfDataTwoKeys(tenantId string, tableName string, unsharded bool, key any, key2 any, value []byte, updatedAt int64, ttl int) error {
	key2FieldName := DefaultKey2FieldName

	if updatedAt == 0 {
		updatedAt = util.GetTimestamp()
	}

	ttlClause := ""
	if ttl > 0 {
		ttlClause = fmt.Sprintf(" USING TTL %d", ttl)
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// table is not sharded and does not have shard_id column, so we only need to insert by key and key2
	if unsharded {
		if tableName == TABLE_LOGS {
			tableName = c.GetTableNameFromLogKeyspace(tableName)
			key2FieldName = Key2FieldNameForLogs2
			stmt := fmt.Sprintf(`INSERT INTO %s(key, %s, value) VALUES(?,?,?)%s`, tableName, key2FieldName, ttlClause)
			return c.Query(stmt, key, key2, value).Exec()
		} else {
			stmt := fmt.Sprintf(`INSERT INTO %s(tenant_id, key, %s, value, updated) VALUES(?,?,?,?,?)%s`, tableName, key2FieldName, ttlClause)
			return c.Query(stmt, tenantId, key, key2, value, updatedAt).Exec()
		}
	} else {
		if tenantId == "" {
			return fmt.Errorf("CassandraClient.SetXconfDataTwoKeys: tenantId is empty, table %s", tableName)
		}
		stmt := fmt.Sprintf(`INSERT INTO %s(tenant_id, shard_id, key, %s, value, updated) VALUES(?,?,?,?,?,?)%s`, tableName, key2FieldName, ttlClause)
		return c.Query(stmt, tenantId, GetShardId(key), key, key2, value, updatedAt).Exec()
	}
}

// GetXconfDataTwoKeys Get one row where return value is JSON data
func (c *CassandraClient) GetXconfDataTwoKeys(tenantId string, tableName string, unsharded bool, key string, key2 any) (value []byte, err error) {
	key2FieldName := DefaultKey2FieldName
	if tableName == TABLE_LOGS {
		tableName = c.GetTableNameFromLogKeyspace(tableName)
		key2FieldName = Key2FieldNameForLogs2
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// table is not sharded and does not have shard_id column, so we only need to delete by key and key2
	if unsharded {
		stmt := fmt.Sprintf(`SELECT value FROM %s WHERE key = ? AND %s = ? LIMIT 1`, tableName, key2FieldName)
		err = c.Query(stmt, key, key2).Scan(&value)
	} else {
		if tenantId == "" {
			return value, fmt.Errorf("CassandraClient.GetXconfDataTwoKeys: tenantId is empty, table %s", tableName)
		}

		stmt := fmt.Sprintf(`SELECT value FROM %s WHERE tenant_id = ? AND shard_id = ? AND key = ? AND %s = ? LIMIT 1`, tableName, key2FieldName)
		err = c.Query(stmt, tenantId, GetShardId(key), key, key2).Scan(&value)
	}

	return value, err
}

// DeleteXconfDataTwoKeys Delete XconfData for the specified two keys
func (c *CassandraClient) DeleteXconfDataTwoKeys(tenantId string, tableName string, unsharded bool, key string, key2 any) error {
	key2FieldName := DefaultKey2FieldName
	if tableName == TABLE_LOGS {
		tableName = c.GetTableNameFromLogKeyspace(tableName)
		key2FieldName = Key2FieldNameForLogs2
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// table is not sharded and does not have shard_id column, so we only need to delete by key and key2
	if unsharded {
		stmt := fmt.Sprintf(`DELETE FROM %s WHERE key = ? AND %s = ?`, tableName, key2FieldName)
		return c.Query(stmt, key, key2).Exec()
	} else {
		if tenantId == "" {
			return fmt.Errorf("CassandraClient.DeleteXconfDataTwoKeys: tenantId is empty, table %s", tableName)
		}

		stmt := fmt.Sprintf(`DELETE FROM %s WHERE tenant_id = ? AND shard_id = ? AND key = ? AND %s = ?`, tableName, key2FieldName)
		return c.Query(stmt, tenantId, GetShardId(key), key, key2).Exec()
	}
}

// GetAllXconfTwoKeys Get all TwoKeys
func (c *CassandraClient) GetAllXconfTwoKeys(tenantId string, tableName string) (resultData []TwoKeys) {
	if tenantId == "" {
		log.Errorf("CassandraClient.GetAllXconfTwoKeys: tenantId is empty, table %s", tableName)
		return resultData
	}

	key2FieldName := DefaultKey2FieldName
	if tableName == TABLE_LOGS {
		tableName = c.GetTableNameFromLogKeyspace(tableName)
		key2FieldName = Key2FieldNameForLogs2
	}

	stmt := fmt.Sprintf(`SELECT key, %s FROM %s WHERE tenant_id = ? AND shard_id IN ?`, key2FieldName, tableName)

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	iter := c.Query(stmt, tenantId, shardIds).Iter()
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}

		twoKeys := TwoKeys{
			Key:  row["key"].(string),
			Key2: row[key2FieldName],
		}
		resultData = append(resultData, twoKeys)
	}

	return resultData
}

// GetAllXconfKey2s Get a list of Xconf key2 for the specified key
func (c *CassandraClient) GetAllXconfKey2s(tenantId string, tableName string, unsharded bool, key string) (resultData []any) {
	var iter *gocql.Iter

	key2FieldName := DefaultKey2FieldName
	if tableName == TABLE_LOGS {
		tableName = c.GetTableNameFromLogKeyspace(tableName)
		key2FieldName = Key2FieldNameForLogs2
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// table is not sharded and does not have shard_id column, so we only need to query by key
	if unsharded {
		stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE key = ?`, key2FieldName, tableName)
		iter = c.Query(stmt, key).Iter()
	} else {
		if tenantId == "" {
			log.Errorf("CassandraClient.GetAllXconfKey2s: tenantId is empty, table %s", tableName)
			return resultData
		}

		stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE tenant_id = ? AND shard_id = ? AND key = ?`, key2FieldName, tableName)
		iter = c.Query(stmt, tenantId, GetShardId(key), key).Iter()
	}

	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row[key2FieldName])
	}

	return resultData
}

// SetXconfCompressedData Create XconfData for the specified key and values, where values is compressed JSON data
func (c *CassandraClient) SetXconfCompressedData(tenantId string, tableName string, key string, values [][]byte, updatedAt int64, ttl int) error {
	if tenantId == "" {
		return fmt.Errorf("CassandraClient.SetXconfCompressedData: tenantId is empty, table %s", tableName)
	}

	key2FieldName := DefaultKey2FieldName
	if tableName == TABLE_LOGS {
		tableName = c.GetTableNameFromLogKeyspace(tableName)
		key2FieldName = Key2FieldNameForLogs2
	}

	shardId := GetShardId(key)
	batch := c.NewBatch(LoggedBatch)
	if updatedAt == 0 {
		updatedAt = util.GetTimestamp()
	}

	// Add a record that specifies the number of compressed data chunks
	var stmt string
	if ttl > 0 {
		stmt = fmt.Sprintf(`INSERT INTO %s(tenant_id, shard_id, key, %s, value, updated) VALUES(?,?,?,?,intAsBlob(?),?) USING TTL %d`, tableName, key2FieldName, ttl)
	} else {
		stmt = fmt.Sprintf(`INSERT INTO %s(tenant_id, shard_id, key, %s, value, updated) VALUES(?,?,?,?,intAsBlob(?),?)`, tableName, key2FieldName)
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	batch.Query(stmt, tenantId, shardId, key, NamedListCountColumnValue, len(values), updatedAt)

	for i, value := range values {
		// Add a record for each compressed data chunk where key has the format: NamedListData_part_0, ...
		partColumnValue := NamedListPartColumnValue + strconv.Itoa(i)
		if ttl > 0 {
			stmt = fmt.Sprintf(`INSERT INTO %s(tenant_id, shard_id, key, %s, value, updated) VALUES(?,?,?,?,?,?) USING TTL %d`, tableName, key2FieldName, ttl)
		} else {
			stmt = fmt.Sprintf(`INSERT INTO %s(tenant_id, shard_id, key, %s, value, updated) VALUES(?,?,?,?,?,?)`, tableName, key2FieldName)
		}
		batch.Query(stmt, tenantId, shardId, key, partColumnValue, value, updatedAt)
	}

	if err := c.ExecuteBatch(batch); err != nil {
		return err
	}

	return nil
}

// GetXconfCompressedData Get one row where return value is compressed JSON data
func (c *CassandraClient) GetXconfCompressedData(tenantId string, tableName string, key string) ([]byte, error) {
	if tenantId == "" {
		return nil, fmt.Errorf("CassandraClient.GetXconfCompressedData: tenantId is empty, table %s", tableName)
	}

	key2FieldName := DefaultKey2FieldName
	if tableName == TABLE_LOGS {
		tableName = c.GetTableNameFromLogKeyspace(tableName)
		key2FieldName = Key2FieldNameForLogs2
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// Get the number of compressed data chunks
	var partsCount int
	shardId := GetShardId(key)
	stmt := fmt.Sprintf(`SELECT blobAsInt(value) FROM %s WHERE tenant_id = ? AND shard_id = ? AND key = ? AND %s = ? LIMIT 1`, tableName, key2FieldName)
	err := c.Query(stmt, tenantId, shardId, key, NamedListCountColumnValue).Scan(&partsCount)
	if err != nil {
		return nil, err
	}

	// Get all the compressed data chunks
	var partsMap = make(map[string][]byte)
	stmt = fmt.Sprintf(`SELECT key, %s, value FROM %s WHERE tenant_id = ? AND shard_id = ? AND key = ?`, key2FieldName, tableName)
	iter := c.Query(stmt, tenantId, shardId, key).Iter()
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}

		partName := row[key2FieldName].(string)
		if partName != NamedListCountColumnValue {
			partsMap[partName] = row["value"].([]byte)
		}
	}

	// Ensure all the parts are loaded
	if partsCount > len(partsMap) {
		err := fmt.Errorf("Inconsistent compressed data for key '%s' from '%s': expected %d record(s) got %d",
			key, tableName, partsCount, len(partsMap))
		log.WithFields(log.Fields{"tenantId": tenantId}).Error(err)
		return nil, err
	}

	// Combine all the compressed data chunks into one
	var chunks [][]byte
	for i := 0; i < partsCount; i++ {
		keyName := NamedListPartColumnValue + strconv.Itoa(i)
		if chunk, exists := partsMap[keyName]; exists {
			chunks = append(chunks, chunk)
		} else {
			err := fmt.Errorf("Inconsistent compressed data for key '%s' from '%s': missing part '%s'",
				key, tableName, keyName)
			log.WithFields(log.Fields{"tenantId": tenantId}).Error(err)
			return nil, err
		}
	}

	resultData := bytes.Join(chunks, []byte(""))

	return resultData, nil
}

// GetAllXconfCompressedDataAsMap Get all rows as a map of key to value, where value is compressed JSON data
func (c *CassandraClient) GetAllXconfCompressedDataAsMap(tenantId string, tableName string) map[string][]byte {
	var resultData = make(map[string][]byte)

	rawData := c.GetXconfCompressedDataRaw(tenantId, tableName)
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

	return resultData
}

// GetXconfCompressedDataRaw Get all rows as a map of key to another map,
// where key specifies part number and value is compressed JSON data chunk.
//
// Sample data for one record in GenericXconfNamedList table:
//
// tenant_id | shard_id | key               | key2                      | value
// ----------+----------+-------------------+---------------------------+-----------------------------
// COMCAST   | 0        | Test_Mac_List     |      NamedListData_part_0 | 0x7df05a7b226964223a2241...
// COMCAST   | 0        | Test_Mac_List     |      NamedListData_part_1 | 0x60f05f7b226964223a2231...
// COMCAST   | 0        | Test_Mac_List     | NamedListData_parts_count |                  0x00000002
func (c *CassandraClient) GetXconfCompressedDataRaw(tenantId string, tableName string) map[string]map[string][]byte {
	var resultData = make(map[string]map[string][]byte)
	var countMap = make(map[string]int)

	if tenantId == "" {
		log.Errorf("CassandraClient.GetXconfCompressedDataRaw: tenantId is empty, table %s", tableName)
		return resultData
	}

	key2FieldName := DefaultKey2FieldName
	if tableName == TABLE_LOGS {
		tableName = c.GetTableNameFromLogKeyspace(tableName)
		key2FieldName = Key2FieldNameForLogs2
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// Get all the count records
	stmt := fmt.Sprintf(`SELECT key, blobAsInt(value) as count FROM %s where tenant_id = ? AND shard_id IN ? AND %s = ? ALLOW FILTERING`, tableName, key2FieldName)

	iter := c.Query(stmt, tenantId, shardIds, NamedListCountColumnValue).Iter()
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		countMap[row["key"].(string)] = row["count"].(int)
	}

	// Get all the compressed data chunks
	stmt = fmt.Sprintf(`SELECT key, %s, value FROM %s WHERE tenant_id = ? AND shard_id IN ?`, key2FieldName, tableName)
	iter = c.Query(stmt, tenantId, shardIds).Iter()
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}

		key2 := row[key2FieldName].(string)
		if key2 == NamedListCountColumnValue {
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
			partsMap[key2] = row["value"].([]byte)
		}
	}

	// Ensure all the parts are loaded
	for key, partsMap := range resultData {
		partsCount := countMap[key]
		if partsCount != len(partsMap) {
			log.WithFields(log.Fields{"tenantId": tenantId}).Warn(fmt.Sprintf("Inconsistent compressed data for table '%s' key '%s': expected %v record(s) got %v",
				tableName, key, partsCount, len(partsMap)))

			// Deleting the wrong data! Need to delete partsmap[key][extra_NamedList_data_part_1,2,3..]
			// delete(partsMap, key) // Ignored invalid record
		}
	}

	return resultData
}

func (c *CassandraClient) QueryXconfDataRows(query string, queryParameters ...string) ([]map[string]any, error) {
	var resultData []map[string]any

	// Convert string slice to interface slice
	params := make([]any, len(queryParameters))
	for i, v := range queryParameters {
		params[i] = v
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	iter := c.Query(query, params...).Iter()
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row)
	}

	return resultData, nil
}

func (c *CassandraClient) ModifyXconfData(query string, queryParameters ...string) error {
	// Convert string slice to interface slice
	params := make([]any, len(queryParameters))
	for i, v := range queryParameters {
		params[i] = v
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	return c.Query(query, params...).Exec()
}

// NewBatch creates a new batch operation
func (c *CassandraClient) NewBatch(batchType int) BatchOperation {
	var gocqlBatchType gocql.BatchType
	switch batchType {
	case LoggedBatch:
		gocqlBatchType = gocql.LoggedBatch
	case UnloggedBatch:
		gocqlBatchType = gocql.UnloggedBatch
	case CounterBatch:
		gocqlBatchType = gocql.CounterBatch
	default:
		gocqlBatchType = gocql.LoggedBatch
	}

	return &BatchWrapper{c.Session.NewBatch(gocqlBatchType)}
}

// ExecuteBatch executes a batch operation
func (c *CassandraClient) ExecuteBatch(batch BatchOperation) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	batchWrapper := batch.(*BatchWrapper)
	return c.Session.ExecuteBatch(batchWrapper.Batch)
}

func (c *CassandraClient) GetTenant(tenantId string) (*Tenant, error) {
	if util.IsBlank(tenantId) {
		return nil, fmt.Errorf("CassandraClient.GetTenant: tenantId is empty")
	}

	tenant := Tenant{}
	stmt := fmt.Sprintf(`SELECT id, name, updated FROM %s WHERE id = ? LIMIT 1`, TABLE_TENANTS)

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	err := c.Query(stmt, tenantId).Scan(&tenant.ID, &tenant.Name, &tenant.Updated)
	if err != nil {
		if c.IsDbNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &tenant, nil
}

func (c *CassandraClient) GetAllTenants() []*Tenant {
	var tenants []*Tenant
	stmt := fmt.Sprintf(`SELECT id, name, updated FROM %s`, TABLE_TENANTS)

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	iter := c.Query(stmt).Iter()
	for {
		var tenant Tenant
		if !iter.Scan(&tenant.ID, &tenant.Name, &tenant.Updated) {
			break
		}
		tenants = append(tenants, &tenant)
	}

	return tenants
}

func (c *CassandraClient) SetTenant(tenant *Tenant) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	stmt := fmt.Sprintf(`INSERT INTO %s(id, name, updated) VALUES(?,?,?)`, TABLE_TENANTS)
	return c.Query(stmt, tenant.ID, tenant.Name, tenant.Updated).Exec()
}

func (c *CassandraClient) DeleteTenant(tenantId string) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	stmt := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TABLE_TENANTS)
	return c.Query(stmt, tenantId).Exec()
}

func (c *CassandraClient) AcquireLock(tenantId string, lockName string, lockedBy string, ttl int) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	lockedAt := time.Now()
	expiresAt := lockedAt.Add(time.Duration(ttl) * time.Second)

	// First, try to insert a new lock (if no lock exists)
	existingLock := make(map[string]any)
	stmt := fmt.Sprintf(`INSERT INTO %s(tenant_id, shard_id, name, locked_by, locked_at, expires_at) VALUES(?,?,?,?,?,?) IF NOT EXISTS`, TABLE_LOCKS)
	applied, err := c.Query(stmt, tenantId, GetShardId(lockName), lockName, lockedBy, lockedAt, expiresAt).MapScanCAS(existingLock)
	if err != nil {
		return fmt.Errorf("failed to acquire lock '%s': %w", lockName, err)
	}
	if applied {
		log.WithFields(log.Fields{"tenantId": tenantId}).Debug(fmt.Sprintf("Lock '%s' acquired by '%s'", lockName, lockedBy))
		return nil
	}

	// Lock exists, check if it's expired and try to update
	if exExpiresAt, ok := existingLock["expires_at"].(time.Time); ok {
		if time.Now().Before(exExpiresAt) {
			return fmt.Errorf("failed to acquire lock '%s' held by '%s' until %s", lockName, existingLock["locked_by"], exExpiresAt)
		}
	}

	stmt = fmt.Sprintf(`UPDATE %s SET locked_by = ?, locked_at = ?, expires_at = ? WHERE tenant_id = ? AND shard_id = ? AND name = ? IF expires_at < ?`, TABLE_LOCKS)
	applied, err = c.Query(stmt, lockedBy, lockedAt, expiresAt, tenantId, GetShardId(lockName), lockName, lockedAt).MapScanCAS(existingLock)
	if err != nil {
		return fmt.Errorf("failed to acquire expired lock '%s': %w", lockName, err)
	}
	if !applied {
		return fmt.Errorf("failed to acquire expired lock '%s' held by '%s'", lockName, existingLock["locked_by"])
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Debug(fmt.Sprintf("Lock '%s' acquired by '%s'", lockName, lockedBy))
	return nil
}

func (c *CassandraClient) ReleaseLock(tenantId string, lockName string, lockedBy string) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// Try to release the lock by deleting the record only if it is held by the specified lockHolder
	existingLock := make(map[string]any)
	stmt := fmt.Sprintf(`DELETE FROM %s WHERE tenant_id = ? AND shard_id = ? AND name = ? IF locked_by = ?`, TABLE_LOCKS)
	applied, err := c.Query(stmt, tenantId, GetShardId(lockName), lockName, lockedBy).MapScanCAS(existingLock)
	if err != nil {
		return fmt.Errorf("failed to release lock '%s': %w", lockName, err)
	}
	if !applied {
		return fmt.Errorf("failed to release lock '%s' held by '%s'", lockName, existingLock["locked_by"])
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Debug(fmt.Sprintf("Lock '%s' released by '%s'", lockName, lockedBy))
	return nil
}

func (c *CassandraClient) GetLockInfo(tenantId string, lockName string) (map[string]any, error) {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	dict := util.Dict{}
	stmt := fmt.Sprintf(`SELECT * FROM %s WHERE tenant_id = ? and shard_id = ? AND name=?`, TABLE_LOCKS)
	qry := c.Query(stmt, tenantId, GetShardId(lockName), lockName)
	err := qry.MapScan(dict)
	if err != nil {
		return dict, fmt.Errorf("failed to retrieve lock '%s': %w", lockName, err)
	}

	return dict, nil
}

type DistributedLock struct {
	DistributedLockSettings
	name string
	ttl  int
}

func NewDistributedLock(name string, ttl int) *DistributedLock {
	if name == "" || ttl <= 0 {
		return nil
	}
	return &DistributedLock{
		DistributedLockSettings: distributedLockSettings,
		name:                    name,
		ttl:                     ttl,
	}
}

func (dl DistributedLock) Name() string {
	return dl.name
}

func (dl DistributedLock) TTL() int {
	return dl.ttl
}

func (dl DistributedLock) Retries() int {
	return dl.retries
}

func (dl DistributedLock) RetryInMsecs() int {
	return dl.retryInMsecs
}

func (dl *DistributedLock) SetTTL(secs int) {
	dl.ttl = secs
}

func (dl *DistributedLock) SetRetries(retries int) {
	dl.retries = retries
}

func (dl *DistributedLock) SetRetryInMsecs(retryInMsecs int) {
	dl.retryInMsecs = retryInMsecs
}

func (dl DistributedLock) Lock(tenantId string, owner string) (e error) {
	if util.IsBlank(tenantId) {
		e = fmt.Errorf("tenantId is required to lock '%s' table", dl.name)
		return
	}
	if util.IsBlank(owner) {
		e = fmt.Errorf("owner is required to lock '%s' table", dl.name)
		return
	}

	retryWaitTime := time.Duration(dl.retryInMsecs) * time.Millisecond

	var err error
	var attempt int // attempt=0 is NOT considered a retry
	for attempt = 0; attempt <= dl.retries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryWaitTime)
		}
		err = GetDatabaseClient().AcquireLock(tenantId, dl.name, owner, dl.ttl)
		if err == nil {
			return
		}
	}

	if dl.retries > 0 {
		e = fmt.Errorf("unable to lock table '%s' after %d attempts: %w", dl.name, attempt+1, err)
	} else {
		e = fmt.Errorf("unable to lock table '%s': %w", dl.name, err)
	}
	log.WithFields(log.Fields{"tenantId": tenantId}).Error(e)

	return
}

func (dl DistributedLock) Unlock(tenantId string, owner string) (e error) {
	if util.IsBlank(tenantId) {
		e = fmt.Errorf("tenantId is required to unlock table '%s'", dl.name)
		return
	}
	if util.IsBlank(owner) {
		e = fmt.Errorf("owner is required to unlock table '%s'", dl.name)
		return
	}

	if err := GetDatabaseClient().ReleaseLock(tenantId, dl.name, owner); err != nil {
		e = fmt.Errorf("unable to unlock table '%s': %w", dl.name, err)
		log.WithFields(log.Fields{"tenantId": tenantId}).Error(e)
	}

	return
}

// LockRow locks a specific row in the table identified by key.
// The lock name is constructed as "<tableName>|<key>".
// This allows for row-level locking within the same table using the existing locking mechanism.
// For a given resource either resource-level or sub-resource-level locks can be used, but not both.
func (dl DistributedLock) LockRow(tenantId string, owner string, key string) (e error) {
	if util.IsBlank(tenantId) {
		e = fmt.Errorf("tenantId is required to lock '%s' table", dl.name)
		return
	}
	if util.IsBlank(owner) {
		e = fmt.Errorf("owner is required to lock '%s' table", dl.name)
		return
	}
	if util.IsBlank(key) {
		e = fmt.Errorf("rowKey is required to lock '%s' table", dl.name)
		return
	}

	lockName := dl.name + LockNameDelimiter + key
	retryWaitTime := time.Duration(dl.retryInMsecs) * time.Millisecond

	var err error
	var attempt int // attempt=0 is NOT considered a retry
	for attempt = 0; attempt <= dl.retries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryWaitTime)
		}
		err = GetDatabaseClient().AcquireLock(tenantId, lockName, owner, dl.ttl)
		if err == nil {
			return
		}
	}

	if dl.retries > 0 {
		e = fmt.Errorf("unable to lock table '%s' row '%s' after %d attempts: %w", dl.name, key, attempt+1, err)
	} else {
		e = fmt.Errorf("unable to lock table '%s' row '%s': %w", dl.name, key, err)
	}
	log.WithFields(log.Fields{"tenantId": tenantId}).Error(e)

	return
}

func (dl DistributedLock) UnlockRow(tenantId string, owner string, key string) (e error) {
	if util.IsBlank(tenantId) {
		e = fmt.Errorf("tenantId is required to unlock table '%s'", dl.name)
		return
	}
	if util.IsBlank(owner) {
		e = fmt.Errorf("owner is required to unlock table '%s'", dl.name)
		return
	}
	if util.IsBlank(key) {
		e = fmt.Errorf("key is required to unlock table '%s'", dl.name)
		return
	}

	lockName := dl.name + LockNameDelimiter + key
	if err := GetDatabaseClient().ReleaseLock(tenantId, lockName, owner); err != nil {
		e = fmt.Errorf("unable to unlock table '%s' row '%s': %w", dl.name, key, err)
		log.WithFields(log.Fields{"tenantId": tenantId}).Error(e)
	}

	return
}

// forEachShard iterates through each shard and executes the provided function
// until all shards have been processed or an error occurs
func forEachShard(fn func(shardId int) error) error {
	for shardId := 0; shardId < ScalingFactor; shardId++ {
		if err := fn(shardId); err != nil {
			return err
		}
	}
	return nil
}

// Get fully-qualified table name from log keyspace. This is necessary because Logs2 table
// is in the old keyspace. Function can be removed when data is migrated to the new xconf keyspace.
func (c *CassandraClient) GetTableNameFromLogKeyspace(tableName string) string {
	return fmt.Sprintf("\"%s\".\"%s\"", c.GetLogKeyspace(), tableName)
}
