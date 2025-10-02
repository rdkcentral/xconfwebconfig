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
	"strings"
	"testing"
	"time"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/gocql/gocql"
	"github.com/google/uuid"

	"gotest.tools/assert"
)

func TestCacheCRUD(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	model := shared.NewModel(fmt.Sprintf("Model-%s", uuid.New().String()), "TestCacheCRUD")

	// verify record not in cache
	obj, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_MODEL, model.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))
	assert.Assert(t, obj == nil)

	// create record in DB
	jsonData, err := json.Marshal(model)
	assert.NilError(t, err)

	err = ds.GetSimpleDao().SetOne(ds.TABLE_MODEL, model.ID, jsonData)
	assert.NilError(t, err)

	// test retrieve from cache
	obj, err = ds.GetCachedSimpleDao().GetOne(ds.TABLE_MODEL, model.ID)
	assert.NilError(t, err)

	m := *obj.(*shared.Model)
	assert.Equal(t, m.ID, model.ID)
	assert.Equal(t, m.Description, model.Description)

	// test update
	model.Description = "obsolete model"
	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_MODEL, model.ID, model)
	assert.NilError(t, err)

	// verify against cache
	obj, err = ds.GetCachedSimpleDao().GetOne(ds.TABLE_MODEL, model.ID)
	assert.NilError(t, err)

	m = *obj.(*shared.Model)
	assert.Equal(t, m.ID, model.ID)
	assert.Equal(t, m.Description, model.Description)

	// verify against db
	obj, err = ds.GetSimpleDao().GetOne(ds.TABLE_MODEL, model.ID)
	assert.NilError(t, err)

	m = *obj.(*shared.Model)
	assert.Equal(t, m.ID, model.ID)
	assert.Equal(t, m.Description, model.Description)

	// test delete
	err = ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_MODEL, model.ID)
	assert.NilError(t, err)

	// entry is not immediatly removed from cache so we check db first
	_, err = ds.GetSimpleDao().GetOne(ds.TABLE_MODEL, model.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))

	obj, err = ds.GetCachedSimpleDao().GetOne(ds.TABLE_MODEL, model.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))
	assert.Assert(t, obj == nil)
}

func TestCacheCompressingDataCRUD(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	nl := shared.NewGenericNamespacedList(fmt.Sprintf("NL-%s", uuid.New().String()), "STRING", humptyStrList)

	// verify record not in cache
	obj, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))
	assert.Assert(t, obj == nil)

	// create record in DB
	jsonData, err := json.Marshal(nl)
	assert.NilError(t, err)

	err = ds.GetCompressingDataDao().SetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID, jsonData)
	assert.NilError(t, err)

	// test retreive from cache only
	obj, err = ds.GetCachedSimpleDao().GetOneFromCacheOnly(ds.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))
	assert.Assert(t, obj == nil)

	// test retrieve from cache
	obj, err = ds.GetCachedSimpleDao().GetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.NilError(t, err)

	resNL := *obj.(*shared.GenericNamespacedList)
	assert.Equal(t, resNL.ID, nl.ID)
	assert.Equal(t, len(resNL.Data), len(humptyStrList))
	assert.Assert(t, util.StringElementsMatch(resNL.Data, humptyStrList))

	// test update
	macList := []string{
		util.GenerateRandomCpeMac(),
		util.GenerateRandomCpeMac(),
		util.GenerateRandomCpeMac(),
	}

	nl.TypeName = "MAC_LIST"
	nl.Data = macList

	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID, nl)
	assert.NilError(t, err)

	// verify against cache
	obj, err = ds.GetCachedSimpleDao().GetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.NilError(t, err)

	resNL = *obj.(*shared.GenericNamespacedList)
	assert.Equal(t, resNL.ID, nl.ID)
	assert.Equal(t, len(resNL.Data), len(macList))
	assert.Assert(t, util.StringElementsMatch(resNL.Data, macList))

	// verify against db
	obj, err = ds.GetCompressingDataDao().GetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.NilError(t, err)
	assert.Assert(t, obj != nil)

	resNL = *obj.(*shared.GenericNamespacedList)
	assert.Equal(t, resNL.ID, nl.ID)
	assert.Equal(t, len(resNL.Data), len(macList))
	assert.Assert(t, util.StringElementsMatch(resNL.Data, macList))

	// test delete
	err = ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.NilError(t, err)

	// entry is not immediatly removed from cache so we check db first
	_, err = ds.GetCompressingDataDao().GetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))

	obj, err = ds.GetCachedSimpleDao().GetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))
	assert.Assert(t, obj == nil)
}

func TestCacheGetAllByKeys(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	// generate some data
	keys, err := generateCacheTestModels(5)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 5)

	rowKeys := keys[0:3]
	models, err := ds.GetCachedSimpleDao().GetAllByKeys(ds.TABLE_MODEL, rowKeys)
	assert.NilError(t, err)
	assert.Equal(t, len(models), len(rowKeys))

	for _, obj := range models {
		m := *obj.(*shared.Model)
		assert.Assert(t, util.Contains(rowKeys, m.ID))
	}
}

func TestCacheGetAll(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	existingKeys, err := ds.GetCachedSimpleDao().GetKeys(ds.TABLE_MODEL)
	assert.NilError(t, err)

	// generate some data
	newKeys, err := generateCacheTestModels(3)
	assert.NilError(t, err)
	assert.Assert(t, len(newKeys) == 3)

	// test GetKeys
	allKeys, err := ds.GetCachedSimpleDao().GetKeys(ds.TABLE_MODEL)
	assert.NilError(t, err)
	assert.Assert(t, (len(existingKeys)+len(newKeys)) == len(allKeys))

	// test GetAllAsList
	modelList, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_MODEL, 0)
	assert.NilError(t, err)
	assert.Assert(t, (len(existingKeys)+len(newKeys)) == len(modelList))

	for _, key := range newKeys {
		found := false
		for _, model := range modelList {
			m := *model.(*shared.Model)
			if m.ID == key {
				found = true
			}
		}
		assert.Assert(t, found)
	}

	// test GetAllAsMap
	modelMap, err := ds.GetCachedSimpleDao().GetAllAsMap(ds.TABLE_MODEL)
	assert.NilError(t, err)
	assert.Assert(t, (len(existingKeys)+len(newKeys)) == len(modelMap))
	for _, key := range newKeys {
		found := false
		for k, v := range modelMap {
			m := *v.(*shared.Model)
			assert.Assert(t, k == m.ID)
			if m.ID == key {
				found = true
			}
		}
		assert.Assert(t, found)
	}
}

func TestCacheRefresh(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	truncateTable(ds.TABLE_MODEL)

	// ensure no data in cache
	ds.GetCacheManager().RefreshAll()

	keys, err := ds.GetCachedSimpleDao().GetKeys(ds.TABLE_MODEL)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 0)

	keys, err = ds.GetCachedSimpleDao().GetKeys(ds.TABLE_ENVIRONMENT)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 0)

	// generate some data
	modelKeys, err := generateTestModels(3)
	assert.NilError(t, err)
	assert.Assert(t, len(modelKeys) == 3)

	envKeys, err := generateTestEnvironments(3)
	assert.NilError(t, err)
	assert.Assert(t, len(envKeys) == 3)

	// test refresh cache for a single table
	ds.GetCacheManager().Refresh(ds.TABLE_MODEL)

	cacheModelKeys, err := ds.GetCachedSimpleDao().GetKeys(ds.TABLE_MODEL)
	assert.NilError(t, err)
	assert.Assert(t, len(cacheModelKeys) == 3)
	for _, key := range cacheModelKeys {
		assert.Assert(t, util.Contains(modelKeys, key.(string)))
	}

	cacheEnvKeys, err := ds.GetCachedSimpleDao().GetKeys(ds.TABLE_ENVIRONMENT)
	assert.NilError(t, err)
	assert.Assert(t, len(cacheEnvKeys) == 0)

	// test refresh all tables
	ds.GetCacheManager().RefreshAll()

	cacheEnvKeys, err = ds.GetCachedSimpleDao().GetKeys(ds.TABLE_ENVIRONMENT)
	assert.NilError(t, err)
	assert.Assert(t, len(cacheEnvKeys) == 3)
	for _, key := range cacheEnvKeys {
		assert.Assert(t, util.Contains(envKeys, key.(string)))
	}
}

func TestCacheChangedKeys(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	truncateTable(ds.TABLE_MODEL)
	truncateTable(ds.TABLE_XCONF_CHANGED_KEYS)

	ds.GetCacheManager().Refresh(ds.TABLE_MODEL)

	// create record
	model := shared.NewModel(fmt.Sprintf("Model-%s", uuid.New().String()), "TestCacheChangedKeys")
	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_MODEL, model.ID, model)
	assert.NilError(t, err)

	// need to wait since changed record is written async
	time.Sleep(500 * time.Millisecond)

	// verify changed key record is created
	changedList, err := ds.GetListingDao().GetAllAsList(ds.TABLE_XCONF_CHANGED_KEYS)
	assert.NilError(t, err)
	assert.Assert(t, len(changedList) == 1)

	data := *changedList[0].(*ds.ChangedData)
	assert.Equal(t, data.Operation, ds.CREATE_OPERATION)
	assert.Equal(t, data.CfName, ds.TABLE_MODEL)
	assert.Equal(t, data.ChangedKey, model.ID)

	tableInfo, err := ds.GetTableInfo(ds.TABLE_MODEL)
	assert.NilError(t, err)
	assert.Equal(t, data.DaoId, tableInfo.DaoId)
}

func generateCacheTestModels(num int) ([]string, error) {
	var keys []string
	for i := 0; i < num; i++ {
		id := strings.ToUpper(uuid.New().String())
		model := shared.NewModel(id, "a test model")
		err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_MODEL, model.ID, model)
		if err != nil {
			return nil, err
		}

		keys = append(keys, id)
	}
	return keys, nil
}
