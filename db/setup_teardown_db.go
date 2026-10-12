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

// Functions here are used to setup() and teardown() tables for unit test

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	testSchemaMaxRetries = 3
)

func (c *CassandraClient) SetUp() error {
	if !c.testOnly {
		err := errors.New("DB Setup() can only be invoked from unit test")
		fmt.Println(err.Error())
		return err
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// Ensure all expected test tables exist.
	for _, t := range AllTables {
		var stmt string
		switch t {
		case TABLE_TENANTS:
			stmt = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."%s" (id text, name text, updated timestamp, PRIMARY KEY (id))`, c.Keyspace, t)
		case TABLE_CHANGE_EVENTS:
			stmt = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."%s" (key bigint, key2 timeuuid, value blob, PRIMARY KEY (key, key2))`, c.Keyspace, t)
		case TABLE_LOGS:
			stmt = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."%s" (key text, column1 text, value blob, PRIMARY KEY (key, column1))`, c.Keyspace, t)
		case TABLE_GENERIC_NS_LIST, TABLE_CONFIG_CHANGE_LOGS:
			stmt = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."%s" (tenant_id text, shard_id int, key text, key2 text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key, key2))`, c.Keyspace, t)
		case TABLE_LOCKS:
			stmt = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."%s" (tenant_id text, shard_id int, name text, locked_by text, locked_at timestamp, expires_at timestamp, PRIMARY KEY ((tenant_id, shard_id), name))`, c.Keyspace, t)
		case PenetrationDataTable, PenetrationMetricsTable:
			stmt = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."%s" (tenant_id text, estb_mac text, ecm_mac text, serial_number text, partner text, model text, fw_filename text, fw_version text, fw_reported_version text, fw_additional_version_info text, fw_applied_rule text, rfc_applied_rules text, rfc_features text, rfc_ts timestamp, fw_ts timestamp, time_zone text, rfc_account_hash text, rfc_account_id text, rfc_account_mgmt text, titan_account_id text, rfc_partner text, titan_partner text, rfc_model text, rfc_fw_reported_version text, rfc_env text, rfc_application_type text, rfc_experience text, rfc_time_zone text, precook_rfc_rules text, rfc_configsethash text, precook_configsethash text, precook_rfc_features text, rfc_post_proc text, rfc_query_params text, rfc_tags text, rfc_estb_ip text, client_cert_expiry text, recovery_cert_expiry text, PRIMARY KEY (estb_mac))`, c.Keyspace, t)
		default:
			stmt = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."%s" (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key))`, c.Keyspace, t)
		}
		if err := c.execSchemaStmtWithRetry(stmt); err != nil {
			fmt.Printf("error at stmt=%v: %v\n", stmt, err)
			return err
		}
	}

	// Best-effort clean slate. Truncate errors are logged inside truncateTestTables.
	_ = c.truncateTestTables()

	return nil
}

func (c *CassandraClient) TearDown() error {
	if !c.testOnly {
		err := errors.New("DB TearDown() can only be invoked from unit test")
		fmt.Println(err.Error())
		return err
	}

	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	// NOTE: TRUNCATE cannot be used in a batch
	if err := c.truncateTestTables(); err != nil {
		return err
	}

	return nil
}

func (c *CassandraClient) truncateTestTables() error {
	for _, tableName := range AllTables {
		stmt := fmt.Sprintf(`TRUNCATE "%s"`, tableName)
		if err := c.execSchemaStmtWithRetry(stmt); err != nil {
			// Best-effort truncate for tests: transient replica issues should not fail test setup.
			fmt.Printf("warn at stmt=%v: %v\n", stmt, err)
		}
	}

	return nil
}

func isRetryableSchemaError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "truncate failed on replica") ||
		strings.Contains(errMsg, "no response received from cassandra within timeout period") ||
		strings.Contains(errMsg, "operation timed out") ||
		strings.Contains(errMsg, "connection closed") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "unavailable")
}

func (c *CassandraClient) execSchemaStmtWithRetry(stmt string, values ...any) error {
	var err error
	for attempt := 1; attempt <= testSchemaMaxRetries; attempt++ {
		err = c.Query(stmt, values...).Exec()
		if err == nil {
			return nil
		}
		if !isRetryableSchemaError(err) || attempt == testSchemaMaxRetries {
			return err
		}
		time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
	}

	return err
}
