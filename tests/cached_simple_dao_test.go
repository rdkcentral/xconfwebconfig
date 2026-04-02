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

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/util"

	"gotest.tools/assert"
)

// cacheNotifierImpl provides a sample implementation for cache change notifications
type cacheNotifierImpl struct {
	ch chan string
}

// Notify logs the cache change event to a channel
func (n *cacheNotifierImpl) Notify(tenantId string, tableName string, changedKey string, operation db.OperationType) {
	msg := fmt.Sprintf("%s: tenantId=%s, tableName=%s, changedKey=%s, operation=%s\n", common.ServerOriginId(), tenantId, tableName, changedKey, operation)
	n.ch <- msg
}

func TestCacheCRUD(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	model := shared.NewModel(fmt.Sprintf("Model-%s", uuid.New().String()), "TestCacheCRUD")

	// verify record not in cache
	obj, err := db.GetCachedSimpleDao().GetOne(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, model.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))
	assert.Assert(t, obj == nil)

	// create record in DB
	jsonData, err := json.Marshal(model)
	assert.NilError(t, err)

	err = db.GetSimpleDao().SetOne(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, model.ID, jsonData)
	assert.NilError(t, err)

	// test retrieve from cache
	obj, err = db.GetCachedSimpleDao().GetOne(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, model.ID)
	assert.NilError(t, err)

	m := *obj.(*shared.Model)
	assert.Equal(t, m.ID, model.ID)
	assert.Equal(t, m.Description, model.Description)

	// test update
	model.Description = "obsolete model"
	err = db.GetCachedSimpleDao().SetOne(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, model.ID, model)
	assert.NilError(t, err)

	// verify against cache
	obj, err = db.GetCachedSimpleDao().GetOne(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, model.ID)
	assert.NilError(t, err)

	m = *obj.(*shared.Model)
	assert.Equal(t, m.ID, model.ID)
	assert.Equal(t, m.Description, model.Description)

	// verify against db
	obj, err = db.GetSimpleDao().GetOne(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, model.ID)
	assert.NilError(t, err)

	m = *obj.(*shared.Model)
	assert.Equal(t, m.ID, model.ID)
	assert.Equal(t, m.Description, model.Description)

	// test delete
	err = db.GetCachedSimpleDao().DeleteOne(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, model.ID)
	assert.NilError(t, err)

	// entry is not immediatly removed from cache so we check db first
	_, err = db.GetSimpleDao().GetOne(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, model.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))

	obj, err = db.GetCachedSimpleDao().GetOne(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, model.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))
	assert.Assert(t, obj == nil)
}

func TestCacheCompressingDataCRUD(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	nl := shared.NewGenericNamespacedList(fmt.Sprintf("NL-%s", uuid.New().String()), "STRING", humptyStrList)

	// verify record not in cache
	obj, err := db.GetCachedSimpleDao().GetOne(db.DEFAULT_TENANT_ID, db.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))
	assert.Assert(t, obj == nil)

	// create record in DB
	jsonData, err := json.Marshal(nl)
	assert.NilError(t, err)

	err = db.GetCompressingDataDao().SetOne(db.DEFAULT_TENANT_ID, db.TABLE_GENERIC_NS_LIST, nl.ID, jsonData)
	assert.NilError(t, err)

	// test retreive from cache only
	obj, err = db.GetCachedSimpleDao().GetOneFromCacheOnly(db.DEFAULT_TENANT_ID, db.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))
	assert.Assert(t, obj == nil)

	// test retrieve from cache
	obj, err = db.GetCachedSimpleDao().GetOne(db.DEFAULT_TENANT_ID, db.TABLE_GENERIC_NS_LIST, nl.ID)
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

	err = db.GetCachedSimpleDao().SetOne(db.DEFAULT_TENANT_ID, db.TABLE_GENERIC_NS_LIST, nl.ID, nl)
	assert.NilError(t, err)

	// verify against cache
	obj, err = db.GetCachedSimpleDao().GetOne(db.DEFAULT_TENANT_ID, db.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.NilError(t, err)

	resNL = *obj.(*shared.GenericNamespacedList)
	assert.Equal(t, resNL.ID, nl.ID)
	assert.Equal(t, len(resNL.Data), len(macList))
	assert.Assert(t, util.StringElementsMatch(resNL.Data, macList))

	// verify against db
	obj, err = db.GetCompressingDataDao().GetOne(db.DEFAULT_TENANT_ID, db.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.NilError(t, err)
	assert.Assert(t, obj != nil)

	resNL = *obj.(*shared.GenericNamespacedList)
	assert.Equal(t, resNL.ID, nl.ID)
	assert.Equal(t, len(resNL.Data), len(macList))
	assert.Assert(t, util.StringElementsMatch(resNL.Data, macList))

	// test delete
	err = db.GetCachedSimpleDao().DeleteOne(db.DEFAULT_TENANT_ID, db.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.NilError(t, err)

	// entry is not immediatly removed from cache so we check db first
	_, err = db.GetCompressingDataDao().GetOne(db.DEFAULT_TENANT_ID, db.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))

	obj, err = db.GetCachedSimpleDao().GetOne(db.DEFAULT_TENANT_ID, db.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))
	assert.Assert(t, obj == nil)
}

func TestCacheGetAllByKeys(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	// generate some data
	keys, err := generateCacheTestModels(5)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 5)

	rowKeys := keys[0:3]
	models, err := db.GetCachedSimpleDao().GetAllByKeys(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, rowKeys)
	assert.NilError(t, err)
	assert.Equal(t, len(models), len(rowKeys))

	for _, obj := range models {
		m := *obj.(*shared.Model)
		assert.Assert(t, util.Contains(rowKeys, m.ID))
	}
}

func TestCacheGetAll(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	existingKeys, err := db.GetCachedSimpleDao().GetKeys(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)
	assert.NilError(t, err)

	// generate some data
	newKeys, err := generateCacheTestModels(3)
	assert.NilError(t, err)
	assert.Assert(t, len(newKeys) == 3)

	// test GetKeys
	allKeys, err := db.GetCachedSimpleDao().GetKeys(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)
	assert.NilError(t, err)
	assert.Assert(t, (len(existingKeys)+len(newKeys)) == len(allKeys))

	// test GetAllAsList
	modelList, err := db.GetCachedSimpleDao().GetAllAsList(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, 0)
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
	modelMap, err := db.GetCachedSimpleDao().GetAllAsMap(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)
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
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	truncateTable(db.TABLE_MODELS)

	// ensure no data in cache
	db.GetCacheManager().RefreshAll(db.DEFAULT_TENANT_ID)

	keys, err := db.GetCachedSimpleDao().GetKeys(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 0)

	keys, err = db.GetCachedSimpleDao().GetKeys(db.DEFAULT_TENANT_ID, db.TABLE_ENVIRONMENTS)
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
	db.GetCacheManager().Refresh(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)

	cacheModelKeys, err := db.GetCachedSimpleDao().GetKeys(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)
	assert.NilError(t, err)
	assert.Assert(t, len(cacheModelKeys) == 3)
	for _, key := range cacheModelKeys {
		assert.Assert(t, util.Contains(modelKeys, key.(string)))
	}

	cacheEnvKeys, err := db.GetCachedSimpleDao().GetKeys(db.DEFAULT_TENANT_ID, db.TABLE_ENVIRONMENTS)
	assert.NilError(t, err)
	assert.Assert(t, len(cacheEnvKeys) == 0)

	// test refresh all tables
	db.GetCacheManager().RefreshAll(db.DEFAULT_TENANT_ID)

	cacheEnvKeys, err = db.GetCachedSimpleDao().GetKeys(db.DEFAULT_TENANT_ID, db.TABLE_ENVIRONMENTS)
	assert.NilError(t, err)
	assert.Assert(t, len(cacheEnvKeys) == 3)
	for _, key := range cacheEnvKeys {
		assert.Assert(t, util.Contains(envKeys, key.(string)))
	}
}

func TestCacheInvalidate(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	truncateTable(db.TABLE_MODELS)

	// ensure no data in cache
	db.GetCacheManager().Refresh(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)

	keys, err := db.GetCachedSimpleDao().GetKeys(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 0)

	// generate some data
	modelKeys, err := generateTestModels(3)
	assert.NilError(t, err)
	assert.Assert(t, len(modelKeys) == 3)
	db.GetCacheManager().Refresh(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)

	cacheModelKeys, err := db.GetCachedSimpleDao().GetKeys(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)
	assert.NilError(t, err)
	assert.Assert(t, len(cacheModelKeys) == 3)

	// invalidate one key
	keyToInvalidate := cacheModelKeys[0]
	db.GetCacheManager().Invalidate(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, keyToInvalidate.(string))
	time.Sleep(500 * time.Millisecond) // wait for async invalidation to complete

	cacheModelKeys, err = db.GetCachedSimpleDao().GetKeys(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)
	assert.NilError(t, err)
	assert.Assert(t, len(cacheModelKeys) == 2)
	assert.Assert(t, !util.Contains(cacheModelKeys, keyToInvalidate))

	// invalidate all keys
	db.GetCacheManager().InvalidateAll(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)
	time.Sleep(500 * time.Millisecond) // wait for async invalidation to complete

	cacheModelKeys, err = db.GetCachedSimpleDao().GetKeys(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)
	assert.NilError(t, err)
	assert.Assert(t, len(cacheModelKeys) == 0)
}

func TestCacheChangedKeys(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	truncateTable(db.TABLE_MODELS)
	truncateTable(db.TABLE_CHANGE_EVENTS)

	db.GetCacheManager().Refresh(db.DEFAULT_TENANT_ID, db.TABLE_MODELS)

	// create record
	model := shared.NewModel(fmt.Sprintf("Model-%s", uuid.New().String()), "TestCacheChangedKeys")
	err := db.GetCachedSimpleDao().SetOne(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, model.ID, model)
	assert.NilError(t, err)

	// need to wait since changed record is written async
	time.Sleep(1 * time.Second)

	// verify changed key record is created
	changedList, err := db.GetListingDao().GetAllAsList(db.DEFAULT_TENANT_ID, db.TABLE_CHANGE_EVENTS)
	assert.NilError(t, err)
	assert.Assert(t, len(changedList) == 1)

	data := *changedList[0].(*db.ChangedData)
	assert.Equal(t, data.Operation, db.CREATE_OPERATION)
	assert.Equal(t, data.CfName, db.TABLE_MODELS)
	assert.Equal(t, data.ChangedKey, model.ID)
	assert.Equal(t, data.ServerOriginId, common.ServerOriginId())
}

func TestCacheChangeNotifier(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	truncateTable(db.TABLE_MODELS)
	truncateTable(db.TABLE_CHANGE_EVENTS)

	testNotifier := cacheNotifierImpl{
		ch: make(chan string, 1),
	}
	db.GetCacheManager().SetCacheChangeNotifier(&testNotifier)

	// create record to trigger notification
	model := shared.NewModel(fmt.Sprintf("Model-%s", uuid.New().String()), "TestCacheChangedKeys")
	err := db.GetCachedSimpleDao().SetOne(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, model.ID, model)
	assert.NilError(t, err, "SetOne should not fail")

	// Wait for the notification on the channel, with a timeout
	select {
	case msg := <-testNotifier.ch:
		assert.Assert(t, strings.HasPrefix(msg, fmt.Sprintf("%s: tableName=Model,", common.ServerOriginId())))
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for cache notification")
	}
}

func generateCacheTestModels(num int) ([]string, error) {
	var keys []string
	for i := 0; i < num; i++ {
		id := strings.ToUpper(uuid.New().String())
		model := shared.NewModel(id, "a test model")
		err := db.GetCachedSimpleDao().SetOne(db.DEFAULT_TENANT_ID, db.TABLE_MODELS, model.ID, model)
		if err != nil {
			return nil, err
		}

		keys = append(keys, id)
	}
	return keys, nil
}
