/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
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
package tests

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/gocql/gocql"
	"github.com/google/uuid"

	"gotest.tools/assert"
)

func TestCRUD(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	model := shared.NewModel(fmt.Sprintf("Model-%s", uuid.New().String()), "TestCacheCRUD")

	// Verify record does not exist
	_, err := ds.GetSimpleDao().GetOne(ds.TABLE_MODEL, model.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))

	// test create
	jsonData, err := json.Marshal(model)
	assert.NilError(t, err)

	err = ds.GetSimpleDao().SetOne(ds.TABLE_MODEL, model.ID, jsonData)
	assert.NilError(t, err)

	// test retrieve
	obj, err := ds.GetSimpleDao().GetOne(ds.TABLE_MODEL, model.ID)
	assert.NilError(t, err)
	assert.Assert(t, obj != nil)

	m := *obj.(*shared.Model)
	assert.Equal(t, m.ID, model.ID)
	assert.Equal(t, m.Description, model.Description)

	// test update
	model.Description = "obsolete model"
	jsonData, err = json.Marshal(model)
	assert.NilError(t, err)

	err = ds.GetSimpleDao().SetOne(ds.TABLE_MODEL, model.ID, jsonData)
	assert.NilError(t, err)

	obj, err = ds.GetSimpleDao().GetOne(ds.TABLE_MODEL, model.ID)
	assert.NilError(t, err)
	assert.Assert(t, obj != nil)

	m = *obj.(*shared.Model)
	assert.Equal(t, m.ID, model.ID)
	assert.Equal(t, m.Description, model.Description)

	// test delete
	err = ds.GetSimpleDao().DeleteOne(ds.TABLE_MODEL, model.ID)
	assert.NilError(t, err)

	_, err = ds.GetSimpleDao().GetOne(ds.TABLE_MODEL, model.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))
}

func TestGetAllByKeys(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	// generate some data
	keys, err := generateTestModels(5)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 5)

	rowKeys := keys[0:3]
	models, err := ds.GetSimpleDao().GetAllByKeys(ds.TABLE_MODEL, rowKeys)
	assert.NilError(t, err)
	assert.Equal(t, len(models), len(rowKeys))

	for _, obj := range models {
		m := *obj.(*shared.Model)
		assert.Assert(t, util.Contains(rowKeys, m.ID))
	}
}

func TestGetAllAsList(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	truncateTable(ds.TABLE_MODEL)

	// generate some data
	keys, err := generateTestModels(5)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 5)

	models, err := ds.GetSimpleDao().GetAllAsList(ds.TABLE_MODEL, 0)
	assert.NilError(t, err)
	assert.Equal(t, len(models), len(keys))

	for _, obj := range models {
		m := *obj.(*shared.Model)
		assert.Assert(t, util.Contains(keys, m.ID))
	}

	models, err = ds.GetSimpleDao().GetAllAsList(ds.TABLE_MODEL, 3)
	assert.NilError(t, err)
	assert.Equal(t, len(models), 3)
}

func TestGetAllAsMap(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	truncateTable(ds.TABLE_MODEL)

	// generate some data
	keys, err := generateTestModels(5)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 5)

	modelMap, err := ds.GetSimpleDao().GetAllAsMap(ds.TABLE_MODEL, 0)
	assert.NilError(t, err)
	assert.Equal(t, len(modelMap), len(keys))

	for _, key := range keys {
		assert.Assert(t, modelMap[key] != nil)
	}

	modelMap, err = ds.GetSimpleDao().GetAllAsMap(ds.TABLE_MODEL, 3)
	assert.NilError(t, err)
	assert.Equal(t, len(modelMap), 3)
}

func TestGetKeys(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	truncateTable(ds.TABLE_MODEL)

	// generate some data
	keys, err := generateTestModels(5)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 5)

	rowKeys := ds.GetSimpleDao().GetKeys(ds.TABLE_MODEL)
	assert.Equal(t, len(rowKeys), len(keys))

	assert.Assert(t, util.StringElementsMatch(keys, rowKeys), fmt.Sprintf("%v : %v", keys, rowKeys))
}

func generateTestModels(num int) ([]string, error) {
	var keys []string
	for i := 0; i < num; i++ {
		id := uuid.New().String()
		model := shared.NewModel(id, "a test model")
		jsonData, err := json.Marshal(model)
		if err != nil {
			return nil, err
		}

		err = ds.GetSimpleDao().SetOne(ds.TABLE_MODEL, model.ID, jsonData)
		if err != nil {
			return nil, err
		}

		keys = append(keys, model.ID)
	}
	return keys, nil
}

func generateTestEnvironments(num int) ([]string, error) {
	var keys []string
	for i := 0; i < num; i++ {
		id := uuid.New().String()
		env := shared.NewEnvironment(id, "a test env")
		jsonData, err := json.Marshal(env)
		if err != nil {
			return nil, err
		}

		err = ds.GetSimpleDao().SetOne(ds.TABLE_ENVIRONMENT, env.ID, jsonData)
		if err != nil {
			return nil, err
		}

		keys = append(keys, env.ID)
	}
	return keys, nil
}

func truncateTable(tableName string) error {
	dbClient := ds.GetDatabaseClient()
	cassandraClient, ok := dbClient.(*ds.CassandraClient)
	if ok {
		return cassandraClient.DeleteAllXconfData(tableName)
	}
	return nil
}
