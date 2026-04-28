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
	DefaultXconfRecookingStatusTableName = "RecookingStatus"
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
	if len(localDc) > 0 {
		cluster.PoolConfig.HostSelectionPolicy = gocql.DCAwareRoundRobinPolicy(localDc)
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

func (c *CassandraClient) GetPenetrationMetrics(estbMac string) (map[string]any, error) {
	dict := util.Dict{}
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()
	stmt := fmt.Sprintf("SELECT * FROM \"%s\" WHERE %s=?", PenetrationMetricsTable, EstbMacColumnValue)
	qry := c.Query(stmt, estbMac)
	err := qry.MapScan(dict)

	if err != nil {
		return dict, err
	}

	return dict, nil
}

func (c *CassandraClient) SetPenetrationMetrics(pMetrics *PenetrationMetrics) error {
	values := []any{pMetrics.EstbMac, pMetrics.Partner, pMetrics.Model, pMetrics.FwVersion, pMetrics.FwReportedVersion, pMetrics.FwAdditionalVersionInfo, pMetrics.FwAppliedRule, pMetrics.FwTs, pMetrics.RfcAppliedRules, pMetrics.RfcFeatures, pMetrics.RfcTs}
	stmt := fmt.Sprintf(`INSERT INTO "%s" (estb_mac,partner,model,fw_version,fw_reported_version,fw_additional_version_info,fw_applied_rule,fw_ts,rfc_features,rfc_applied_rules,rfc_ts) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, PenetrationMetricsTable)
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()
	qry := c.Query(stmt, values...)
	err := qry.Exec()

	if err != nil {
		return err
	}
	return nil
}

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
func (c *CassandraClient) SetXconfData(tenantId string, tableName string, key string, value []byte, ttl int) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var stmt string
	if ttl > 0 {
		stmt = fmt.Sprintf(`INSERT INTO "%s"(tenant_id, shard_id, key, value) VALUES(?,?,?,?) USING TTL %d`, tableName, ttl)
	} else {
		stmt = fmt.Sprintf(`INSERT INTO "%s"(tenant_id, shard_id, key, value) VALUES(?,?,?,?)`, tableName)
	}

	if err := c.Query(stmt, tenantId, GetShardId(key), key, value).Exec(); err != nil {
		return err
	}

	return nil
}

// GetXconfData Get one row where return value is JSON data
func (c *CassandraClient) GetXconfData(tenantId string, tableName string, key string) ([]byte, error) {
	start := time.Now()

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var value []byte

	stmt := fmt.Sprintf(`SELECT value FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND key = ? LIMIT 1`, tableName)
	err := c.Query(stmt, tenantId, GetShardId(key), key).Scan(&value)
	if err != nil {
		return value, err
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Debug(fmt.Sprintf("CassandraClient.GetXconfData: table %s key %s in %v", tableName, key, time.Since(start)))

	return value, nil
}

// GetAllXconfDataByKeys Get all rows as a list of values for the specified keys, where value is JSON data
func (c *CassandraClient) GetAllXconfDataByKeys(tenantId string, tableName string, keys []string) [][]byte {
	start := time.Now()
	var resultData [][]byte

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

	log.WithFields(log.Fields{"tenantId": tenantId}).Debug(fmt.Sprintf("CassandraClient.GetAllXconfDataByKeys: table %s keys %v in %v", tableName, keys, time.Since(start)))

	return resultData
}

// GetAllXconfKeys Get all keys
func (c *CassandraClient) GetAllXconfKeys(tenantId string, tableName string) []string {
	start := time.Now()

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	resultData := util.Set{}
	stmt := fmt.Sprintf(`SELECT key FROM "%s" WHERE tenant_id = ? AND shard_id IN ?`, tableName)
	iter := c.Query(stmt, tenantId, shardIds).Iter()
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		resultData.Add(row["key"].(string))
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Debug(fmt.Sprintf("CassandraClient.GetAllXconfKeys: table %s in %v", tableName, time.Since(start)))

	return resultData.ToSlice()
}

// GetAllXconfDataAsList Get all rows as a list of values, where value is JSON data
func (c *CassandraClient) GetAllXconfDataAsList(tenantId string, tableName string, maxResults int) [][]byte {
	start := time.Now()

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var resultData [][]byte
	var stmt string
	var iter *gocql.Iter

	// If tenantId is empty, it means the table is not sharded and does not have tenant_id and shard_id columns
	if tenantId == "" {
		if maxResults > 0 {
			stmt = fmt.Sprintf(`SELECT value FROM "%s" LIMIT %v`, tableName, maxResults)
		} else {
			stmt = fmt.Sprintf(`SELECT value FROM "%s"`, tableName)
		}
		iter = c.Query(stmt).Iter()
	} else {
		if maxResults > 0 {
			stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE tenant_id = ? AND shard_id IN ? LIMIT %v`, tableName, maxResults)
		} else {
			stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE tenant_id = ? AND shard_id IN ?`, tableName)
		}
		iter = c.Query(stmt, tenantId, shardIds).Iter()
	}

	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row["value"].([]byte))
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Debug(fmt.Sprintf("CassandraClient.GetAllXconfDataAsList: table %s in %v", tableName, time.Since(start)))

	return resultData
}

// GetAllXconfDataAsMap Get all rows as a map of key to value, where value is JSON data
func (c *CassandraClient) GetAllXconfDataAsMap(tenantId string, tableName string, maxResults int) map[string][]byte {
	start := time.Now()

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var resultData = make(map[string][]byte)
	var stmt string
	var iter *gocql.Iter

	// If tenantId is empty, it means the table is not sharded and does not have tenant_id and shard_id columns
	if tenantId == "" {
		if maxResults > 0 {
			stmt = fmt.Sprintf(`SELECT key, value FROM "%s" LIMIT %v`, tableName, maxResults)
		} else {
			stmt = fmt.Sprintf(`SELECT key, value FROM "%s"`, tableName)
		}
		iter = c.Query(stmt).Iter()
	} else {
		if maxResults > 0 {
			stmt = fmt.Sprintf(`SELECT key, value FROM "%s" WHERE tenant_id = ? AND shard_id IN ? LIMIT %v`, tableName, maxResults)
		} else {
			stmt = fmt.Sprintf(`SELECT key, value FROM "%s" WHERE tenant_id = ? AND shard_id IN ?`, tableName)
		}
		iter = c.Query(stmt, tenantId, shardIds).Iter()
	}

	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		resultData[row["key"].(string)] = row["value"].([]byte)
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Debug(fmt.Sprintf("CassandraClient.GetAllXconfDataAsMap: table %s in %v", tableName, time.Since(start)))

	return resultData
}

// DeleteXconfData Delete XconfData for the specified tenant, table, and key
func (c *CassandraClient) DeleteXconfData(tenantId string, tableName string, key string) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// If tenantId is empty, it means the table is not sharded and does not have tenant_id and shard_id columns
	if tenantId == "" {
		stmt := fmt.Sprintf(`DELETE FROM "%s" WHERE key = ?`, tableName)
		return c.Query(stmt, key).Exec()
	} else {
		stmt := fmt.Sprintf(`DELETE FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND key = ?`, tableName)
		return c.Query(stmt, tenantId, GetShardId(key), key).Exec()
	}
}

// DeleteAllXconfData Delete all XconfData for the specified tenant and table
func (c *CassandraClient) DeleteAllXconfData(tenantId string, tableName string) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// If tenantId is empty, it means the table is not sharded and does not have tenant_id and shard_id columns
	if tenantId == "" {
		stmt := fmt.Sprintf(`TRUNCATE "%s"`, tableName)
		return c.Query(stmt).Exec()
	} else {
		stmt := fmt.Sprintf(`DELETE FROM "%s" WHERE tenant_id = ? AND shard_id IN ?`, tableName)
		return c.Query(stmt, tenantId, shardIds).Exec()
	}
}

// Two keys support

// GetAllXconfData Get multiple rows as a list of values, where value is JSON data
func (c *CassandraClient) GetAllXconfData(tenantId string, tableName string, key string) [][]byte {
	start := time.Now()

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var resultData [][]byte
	var iter *gocql.Iter

	// If tenantId is empty, it means the table is not sharded and does not have tenant_id and shard_id columns
	if tenantId == "" {
		stmt := fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ?`, tableName)
		iter = c.Query(stmt, key).Iter()
	} else {
		stmt := fmt.Sprintf(`SELECT value FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND key = ?`, tableName)
		iter = c.Query(stmt, tenantId, GetShardId(key), key).Iter()
	}

	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row["value"].([]byte))
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Debug(fmt.Sprintf("CassandraClient.GetAllXconfData: table %s key %s in %v", tableName, key, time.Since(start)))

	return resultData
}

// GetAllXconfDataTwoKeysRange Get multiple rows for the specified key and key2 range as list of values, where value is JSON data
func (c *CassandraClient) GetAllXconfDataTwoKeysRange(tenantId string, tableName string, key any, key2FieldName string, rangeInfo *RangeInfo) [][]byte {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var resultData [][]byte
	var stmt string
	var iter *gocql.Iter

	nilStartValue := true
	nilEndValue := true
	if rangeInfo != nil {
		nilStartValue = rangeInfo.IsNilStartValue()
		nilEndValue = rangeInfo.IsNilEndValue()
	}

	// If tenantId is empty, it means the table is not sharded and does not have tenant_id and shard_id columns
	if tenantId == "" {
		if nilStartValue && nilEndValue {
			stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ? ALLOW FILTERING`, tableName)
			iter = c.Query(stmt, key).Iter()
		} else {
			if nilStartValue {
				if !nilEndValue {
					stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ? and %s < ? ALLOW FILTERING`, tableName, key2FieldName)
					iter = c.Query(stmt, key, rangeInfo.EndValue).Iter()
				}
			} else {
				if nilEndValue {
					stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ? and %s > ? ALLOW FILTERING`, tableName, key2FieldName)
					iter = c.Query(stmt, key, rangeInfo.StartValue).Iter()
				} else {
					stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ? and %s > ? and %s < ? ALLOW FILTERING`, tableName, key2FieldName, key2FieldName)
					iter = c.Query(stmt, key, rangeInfo.StartValue, rangeInfo.EndValue).Iter()
				}
			}
		}
	} else {
		if nilStartValue && nilEndValue {
			stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND key = ? ALLOW FILTERING`, tableName)
			iter = c.Query(stmt, tenantId, GetShardId(key), key).Iter()
		} else {
			if nilStartValue {
				if !nilEndValue {
					stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND key = ? and %s < ? ALLOW FILTERING`, tableName, key2FieldName)
					iter = c.Query(stmt, tenantId, GetShardId(key), key, rangeInfo.EndValue).Iter()
				}
			} else {
				if nilEndValue {
					stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND key = ? and %s > ? ALLOW FILTERING`, tableName, key2FieldName)
					iter = c.Query(stmt, tenantId, GetShardId(key), key, rangeInfo.StartValue).Iter()
				} else {
					stmt = fmt.Sprintf(`SELECT value FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND key = ? and %s > ? and %s < ? ALLOW FILTERING`, tableName, key2FieldName, key2FieldName)
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
func (c *CassandraClient) GetAllXconfDataTwoKeysAsMap(tenantId string, tableName string, key string, key2FieldName string, key2List []any) map[any][]byte {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var resultData = make(map[any][]byte)
	var iter *gocql.Iter

	// If tenantId is empty, it means the table is not sharded and does not have tenant_id and shard_id columns
	if tenantId == "" {
		stmt := fmt.Sprintf(`SELECT %s, value FROM "%s" WHERE key = ? and %s IN ?`, key2FieldName, tableName, key2FieldName)
		iter = c.Query(stmt, key, key2List).Iter()
	} else {
		stmt := fmt.Sprintf(`SELECT %s, value FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND key = ? and %s IN ?`, key2FieldName, tableName, key2FieldName)
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
func (c *CassandraClient) SetXconfDataTwoKeys(tenantId string, tableName string, key any, key2FieldName string, key2 any, value []byte, ttl int) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var stmt string

	// If tenantId is empty, it means the table is not sharded and does not have tenant_id and shard_id columns
	if tenantId == "" {
		if ttl > 0 {
			stmt = fmt.Sprintf(`INSERT INTO "%s"(key, %s, value) VALUES(?,?,?) USING TTL %d`, tableName, key2FieldName, ttl)
		} else {
			stmt = fmt.Sprintf(`INSERT INTO "%s"(key, %s, value) VALUES(?,?,?)`, tableName, key2FieldName)
		}

		return c.Query(stmt, key, key2, value).Exec()
	} else {
		if ttl > 0 {
			stmt = fmt.Sprintf(`INSERT INTO "%s"(tenant_id, shard_id, key, %s, value) VALUES(?,?,?,?,?) USING TTL %d`, tableName, key2FieldName, ttl)
		} else {
			stmt = fmt.Sprintf(`INSERT INTO "%s"(tenant_id, shard_id, key, %s, value) VALUES(?,?,?,?,?)`, tableName, key2FieldName)
		}

		return c.Query(stmt, tenantId, GetShardId(key), key, key2, value).Exec()
	}
}

// GetXconfDataTwoKeys Get one row where return value is JSON data
func (c *CassandraClient) GetXconfDataTwoKeys(tenantId string, tableName string, key string, key2FieldName string, key2 any) ([]byte, error) {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var value []byte
	var err error

	// If tenantId is empty, it means the table is not sharded and does not have tenant_id and shard_id columns
	if tenantId == "" {
		stmt := fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ? AND %s = ? LIMIT 1`, tableName, key2FieldName)
		err = c.Query(stmt, key, key2).Scan(&value)
	} else {
		stmt := fmt.Sprintf(`SELECT value FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND key = ? AND %s = ? LIMIT 1`, tableName, key2FieldName)
		err = c.Query(stmt, tenantId, GetShardId(key), key, key2).Scan(&value)
	}

	return value, err
}

// DeleteXconfDataTwoKeys Delete XconfData for the specified two keys
func (c *CassandraClient) DeleteXconfDataTwoKeys(tenantId string, tableName string, key string, key2FieldName string, key2 any) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// If tenantId is empty, it means the table is not sharded and does not have tenant_id and shard_id columns
	if tenantId == "" {
		stmt := fmt.Sprintf(`DELETE FROM "%s" WHERE key = ? AND %s = ?`, tableName, key2FieldName)
		return c.Query(stmt, key, key2).Exec()
	} else {
		stmt := fmt.Sprintf(`DELETE FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND key = ? AND %s = ?`, tableName, key2FieldName)
		return c.Query(stmt, tenantId, GetShardId(key), key, key2).Exec()
	}
}

// GetAllXconfTwoKeys Get all TwoKeys
func (c *CassandraClient) GetAllXconfTwoKeys(tenantId string, tableName string, key2FieldName string) []TwoKeys {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var resultData []TwoKeys
	var iter *gocql.Iter

	// If tenantId is empty, it means the table is not sharded and does not have tenant_id and shard_id columns
	if tenantId == "" {
		stmt := fmt.Sprintf(`SELECT key, "%s" FROM "%s"`, key2FieldName, tableName)
		iter = c.Query(stmt).Iter()
	} else {
		stmt := fmt.Sprintf(`SELECT key, "%s" FROM "%s" WHERE tenant_id = ? AND shard_id IN ?`, key2FieldName, tableName)
		iter = c.Query(stmt, tenantId, shardIds).Iter()
	}

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
func (c *CassandraClient) GetAllXconfKey2s(tenantId string, tableName string, key string, key2FieldName string) []any {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var resultData []any
	var iter *gocql.Iter

	// If tenantId is empty, it means the table is not sharded and does not have tenant_id and shard_id columns
	if tenantId == "" {
		stmt := fmt.Sprintf(`SELECT %s FROM "%s" WHERE key = ?`, key2FieldName, tableName)
		iter = c.Query(stmt, key).Iter()
	} else {
		stmt := fmt.Sprintf(`SELECT %s FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND key = ?`, key2FieldName, tableName)
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
func (c *CassandraClient) SetXconfCompressedData(tenantId string, tableName string, key string, values [][]byte, ttl int) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	shardId := GetShardId(key)
	batch := c.NewBatch(LoggedBatch)

	// Add a record that specifies the number of compressed data chunks
	var stmt string
	if ttl > 0 {
		stmt = fmt.Sprintf(`INSERT INTO "%s"(tenant_id, shard_id, key, %s, value) VALUES(?,?,?,?,intAsBlob(?)) USING TTL %d`, tableName, Key2FieldNameForList, ttl)
	} else {
		stmt = fmt.Sprintf(`INSERT INTO "%s"(tenant_id, shard_id, key, %s, value) VALUES(?,?,?,?,intAsBlob(?))`, tableName, Key2FieldNameForList)
	}
	batch.Query(stmt, tenantId, shardId, key, NamedListCountColumnValue, len(values))

	for i, value := range values {
		// Add a record for each compressed data chunk where key has the format: NamedListData_part_0, ...
		partColumnValue := NamedListPartColumnValue + strconv.Itoa(i)
		if ttl > 0 {
			stmt = fmt.Sprintf(`INSERT INTO "%s"(tenant_id, shard_id, key, %s, value) VALUES(?,?,?,?,?) USING TTL %d`, tableName, Key2FieldNameForList, ttl)
		} else {
			stmt = fmt.Sprintf(`INSERT INTO "%s"(tenant_id, shard_id, key, %s, value) VALUES(?,?,?,?,?)`, tableName, Key2FieldNameForList)
		}
		batch.Query(stmt, tenantId, shardId, key, partColumnValue, value)
	}

	if err := c.ExecuteBatch(batch); err != nil {
		return err
	}

	return nil
}

// GetXconfCompressedData Get one row where return value is compressed JSON data
func (c *CassandraClient) GetXconfCompressedData(tenantId string, tableName string, key string) ([]byte, error) {
	start := time.Now()

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// Get the number of compressed data chunks
	var partsCount int
	shardId := GetShardId(key)
	stmt := fmt.Sprintf(`SELECT blobAsInt(value) FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND key = ? AND %s = ? LIMIT 1`, tableName, Key2FieldNameForList)
	err := c.Query(stmt, tenantId, shardId, key, NamedListCountColumnValue).Scan(&partsCount)
	if err != nil {
		return nil, err
	}

	// Get all the compressed data chunks
	var partsMap = make(map[string][]byte)
	stmt = fmt.Sprintf(`SELECT key, %s, value FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND key = ?`, Key2FieldNameForList, tableName)
	iter := c.Query(stmt, tenantId, shardId, key).Iter()
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}

		partName := row[Key2FieldNameForList].(string)
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

	log.WithFields(log.Fields{"tenantId": tenantId}).Debug(fmt.Sprintf("CassandraClient.GetXconfCompressedData: table %s key %s in %v", tableName, key, time.Since(start)))

	return resultData, nil
}

// GetAllXconfCompressedDataAsMap Get all rows as a map of key to value, where value is compressed JSON data
func (c *CassandraClient) GetAllXconfCompressedDataAsMap(tenantId string, tableName string) map[string][]byte {
	start := time.Now()

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

	log.WithFields(log.Fields{"tenantId": tenantId}).Debug(fmt.Sprintf("CassandraClient.GetAllXconfCompressedDataAsMap: table %s in %v", tableName, time.Since(start)))

	return resultData
}

// GetXconfCompressedDataRaw Get all rows as a map of key to another map,
// where key specifies part number and value is compressed JSON data chunk.
//
// Sample data for one record in GenericXconfNamedList table:
//
// tenant_id | shard_id | key               | column1                   | value
// ----------+----------+-------------------+---------------------------+-----------------------------
// COMCAST   | 0        | Test_Mac_List     |      NamedListData_part_0 | 0x7df05a7b226964223a2241...
// COMCAST   | 0        | Test_Mac_List     |      NamedListData_part_1 | 0x60f05f7b226964223a2231...
// COMCAST   | 0        | Test_Mac_List     | NamedListData_parts_count |                  0x00000002
func (c *CassandraClient) GetXconfCompressedDataRaw(tenantId string, tableName string) map[string]map[string][]byte {
	start := time.Now()

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var resultData = make(map[string]map[string][]byte)
	var countMap = make(map[string]int)

	// Get all the count records
	stmt := fmt.Sprintf(`SELECT key, blobAsInt(value) as count FROM "%s" where tenant_id = ? AND shard_id IN ? AND %s = ? ALLOW FILTERING`, tableName, Key2FieldNameForList)

	iter := c.Query(stmt, tenantId, shardIds, NamedListCountColumnValue).Iter()
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		countMap[row["key"].(string)] = row["count"].(int)
	}

	// Get all the compressed data chunks
	stmt = fmt.Sprintf(`SELECT key, %s, value FROM "%s" WHERE tenant_id = ? AND shard_id IN ?`, Key2FieldNameForList, tableName)
	iter = c.Query(stmt, tenantId, shardIds).Iter()
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}

		column1 := row[Key2FieldNameForList].(string)
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
			log.WithFields(log.Fields{"tenantId": tenantId}).Warn(fmt.Sprintf("Inconsistent compressed data for table '%s' key '%s': expected %v record(s) got %v",
				tableName, key, partsCount, len(partsMap)))

			// Deleting the wrong data! Need to delete partsmap[key][extra_NamedList_data_part_1,2,3..]
			// delete(partsMap, key) // Ignored invalid record
		}
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Debug(fmt.Sprintf("CassandraClient.GetXconfCompressedDataRaw: table %s in %v", tableName, time.Since(start)))

	return resultData
}

func (c *CassandraClient) QueryXconfDataRows(query string, queryParameters ...string) ([]map[string]any, error) {
	start := time.Now()

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// Convert string slice to interface slice
	params := make([]any, len(queryParameters))
	for i, v := range queryParameters {
		params[i] = v
	}

	var resultData []map[string]any
	iter := c.Query(query, params...).Iter()
	for {
		row := make(map[string]any)
		if !iter.MapScan(row) {
			break
		}
		resultData = append(resultData, row)
	}
	log.Debug(fmt.Sprintf("CassandraClient.QueryXconfDataRows executed query=%q parameters=%v duration=%s", query, queryParameters, time.Since(start)))
	return resultData, nil
}

func (c *CassandraClient) ModifyXconfData(query string, queryParameters ...string) error {
	start := time.Now()

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// Convert string slice to interface slice
	params := make([]any, len(queryParameters))
	for i, v := range queryParameters {
		params[i] = v
	}

	if err := c.Query(query, params...).Exec(); err != nil {
		return err
	}
	log.Debug(fmt.Sprintf("CassandraClient.ModifyXconfData executed query=%q parameters=%v duration=%s", query, queryParameters, time.Since(start)))
	return nil
}

// NewBatch creates a new batch operation
func (c *CassandraClient) NewBatch(batchType int) BatchOperation {
	start := time.Now()

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

	batch := &BatchWrapper{c.Session.NewBatch(gocqlBatchType)}
	log.Debug(fmt.Sprintf("CassandraClient.NewBatch created batch_type=%d duration=%s",
		batchType, time.Since(start)))

	return batch
}

// ExecuteBatch executes a batch operation
func (c *CassandraClient) ExecuteBatch(batch BatchOperation) error {
	start := time.Now()

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	batchWrapper := batch.(*BatchWrapper)
	err := c.Session.ExecuteBatch(batchWrapper.Batch)

	log.Debug(fmt.Sprintf("CassandraClient.ExecuteBatch executed batch_size=%d duration=%s error=%v",
		batch.Size(), time.Since(start), err))

	return err
}

func (c *CassandraClient) GetAllTenants() []*Tenant {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var tenants []*Tenant
	stmt := fmt.Sprintf(`SELECT id, name, updated FROM %s`, TABLE_TENANTS)
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

	stmt := fmt.Sprintf(`INSERT INTO "%s"(id, name, updated) VALUES(?,?,?)`, TABLE_TENANTS)
	if err := c.Query(stmt, tenant.ID, tenant.Name, tenant.Updated).Exec(); err != nil {
		return err
	}
	return nil
}

func (c *CassandraClient) DeleteTenant(tenantId string) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	stmt := fmt.Sprintf(`DELETE FROM "%s" WHERE id = ?`, TABLE_TENANTS)
	if err := c.Query(stmt, tenantId).Exec(); err != nil {
		return err
	}
	return nil
}

func (c *CassandraClient) AcquireLock(tenantId string, lockName string, lockedBy string, ttl int) error {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	lockedAt := time.Now()
	expiresAt := lockedAt.Add(time.Duration(ttl) * time.Second)

	// First, try to insert a new lock (if no lock exists)
	existingLock := make(map[string]any)
	stmt := fmt.Sprintf(`INSERT INTO "%s"(tenant_id, shard_id, name, locked_by, locked_at, expires_at) VALUES(?,?,?,?,?,?) IF NOT EXISTS`, TABLE_LOCKS)
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

	stmt = fmt.Sprintf(`UPDATE "%s" SET locked_by = ?, locked_at = ?, expires_at = ? WHERE tenant_id = ? AND shard_id = ? AND name = ? IF expires_at < ?`, TABLE_LOCKS)
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
	stmt := fmt.Sprintf(`DELETE FROM "%s" WHERE tenant_id = ? AND shard_id = ? AND name = ? IF locked_by = ?`, TABLE_LOCKS)
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
	stmt := fmt.Sprintf(`SELECT * FROM "%s" WHERE tenant_id = ? and shard_id = ? AND name=?`, TABLE_LOCKS)
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
