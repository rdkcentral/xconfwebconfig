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
)

func (c *CassandraClient) SetUp() error {
	if !c.testOnly {
		err := errors.New("DB Setup() can only be invoked from unit test")
		fmt.Println(err.Error())
		return err
	}

	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	// NOTE: CREATE cannot be used in batch
	for _, t := range AllTables {
		var stmt string
		if t == "XconfChangedKeys4" {
			stmt = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."%s" (key bigint, columnName timeuuid, value blob, PRIMARY KEY (key, columnName))`, c.Keyspace, t)
		} else {
			stmt = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."%s" (key text, column1 text, value blob, PRIMARY KEY (key, column1))`, c.Keyspace, t)
		}
		if err := c.Query(stmt).Exec(); err != nil {
			fmt.Printf("error at stmt=%v: %v\n", stmt, err)
			return err
		}
	}

	return nil
}

func (c *CassandraClient) TearDown() error {
	if !c.testOnly {
		err := errors.New("DB TearDown() can only be invoked from unit test")
		fmt.Println(err.Error())
		return err
	}

	c.concurrentQueries <- true
	defer func() { <-c.concurrentQueries }()

	// NOTE: TRUNCATE cannot be used in a batch
	for _, tableName := range AllTables {
		stmt := fmt.Sprintf(`TRUNCATE "%s"`, tableName)
		if err := c.Query(stmt).Exec(); err != nil {
			fmt.Printf("error at stmt=%v: %v\n", stmt, err)
			return err
		}
	}

	return nil
}
