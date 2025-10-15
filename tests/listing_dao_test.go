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
	"fmt"
	"testing"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"gotest.tools/assert"
)

var (
	configChangeLogJsonTemplate1 = `{
		"id": "A4:F3:E8:79:C8:60",
		"updated": 1,
		"input": {
			"estbMac": "A4:F3:E8:79:C8:60",
			"ecmMac": "A4:F3:E8:79:C8:60",
			"env": "TEST",
			"model": "AABBXD7",
			"firmwareVersion": "PY200AN_5.4p36s6_VBN_DYBsd",
			"receiverId": "receiverId",
			"controllerId": 1,
			"channelMapId": 1,
			"vodId": 1,
			"accountId": "accountId",
			"bypassFilters": ["bypassFilters"],
			"forceFilters": ["forceFilters"],
			"capabilities": ["RCDL","supportsFullHttpUrl","rebootDecoupled"],
			"timeZone": "UTC",
            "time": "04/08/2021 18:45:00",
			"ipAddress": "48.26.240.152",
			"rcdl": true,
			"supportsFullHttpUrl": true,
			"rebootDecoupled": true
		},
		"rule": {
			"id": "595cd34d-f572-4f86-b5e2-3ded98113874",
			"type": "MAC_RULE",
			"name": "XconfTest",
			"noop": true,
			"blocking": true
		},
		"filters": [
			{
				"noop": false,
				"blocking": false
			},
			{
				"id": "99c5aa54-95c5-423e-bd7e-e91046e89354",
				"type": "RI_3",
				"name": "XCONFRI_3",
				"noop": true,
				"blocking": false
			}
		],
		"explanation": "Request: firmwareVersion=abc\ncapabilities=RCDL\nenv=TEST\nmodel=AABBXD7\nipAddress=48.26.240.152\neStbMac=A4:F3:E8:79:C8:60\napplicationType=stb\nHA-Haproxy-xconf-http=\ntime=6/4/2021 15:25\n\\n matched MAC_RULE 595cd34d-f572-4f86-b5e2-3ded98113874: XconfTest\n received config: &{Properties:map[description:PY200AN_5.4p36s6_VBN_DYBsd Signed firmwareDownloadProtocol:http firmwareFilename:PY200AN_5.4p36s6_VBN_DYBsd-signed.bin firmwareLocation:dac15cdlserver.ae.ccp.xcal.tv firmwareVersion:PY200AN_5.4p36s6_VBN_DYBsd id:38db58a7-94d6-43e6-90a1-91b2b511e5c2 rebootImmediately:true supportedModelIds:[PX001ANC PX001ANM] updated:1492179526599 upgradeDelay:0]}\n was blocked/modified by filter RI_3[ FirmwareRule{id=99c5aa54-95c5-423e-bd7e-e91046e89354, name=XCONFRI_3, type=RI_3} ]",
		"config": {
			"firmwareDownloadProtocol": "http",
			"firmwareFilename": "PY200AN_5.4p36s6_VBN_DYBsd-signed.bin",
			"firmwareLocation": "dac15cdlserver.ae.ccp.xcal.tv",
			"firmwareVersion": "PY200AN_5.4p36s6_VBN_DYBsd",
			"rebootImmediately": true
		},
		"hasMinimumFirmware": true
	}`
)

func TestListingCRUD(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	truncateTable(ds.TABLE_LOGS)

	rowKey := "A4:F3:E8:79:C8:60"

	// test create
	err := ds.GetListingDao().SetOne(ds.TABLE_LOGS, rowKey, coreef.LAST_CONFIG_LOG_ID, []byte(configChangeLogJsonTemplate1))
	assert.NilError(t, err)
	err = ds.GetListingDao().SetOne(ds.TABLE_LOGS, rowKey, "tvxads-de-k8-xconfds-0153e903b521e2a9e_1", []byte(configChangeLogJsonTemplate1))
	assert.NilError(t, err)
	err = ds.GetListingDao().SetOne(ds.TABLE_LOGS, rowKey, "tvxads-de-k8-xconfds-0153e903b521e2a9e_2", []byte(configChangeLogJsonTemplate1))
	assert.NilError(t, err)

	// test retrieve
	obj, err := ds.GetListingDao().GetOne(ds.TABLE_LOGS, rowKey, coreef.LAST_CONFIG_LOG_ID)
	assert.NilError(t, err)
	assert.Assert(t, obj != nil)
	changeLog := obj.(*coreef.ConfigChangeLog)

	assert.Assert(t, changeLog.Input != nil)
	assert.Equal(t, changeLog.Input.EstbMac, rowKey)
	assert.Equal(t, changeLog.Input.EcmMac, rowKey)
	assert.Equal(t, changeLog.Input.Env, "TEST")
	assert.Equal(t, changeLog.Input.Model, "AABBXD7")
	assert.Equal(t, changeLog.Input.FirmwareVersion, "PY200AN_5.4p36s6_VBN_DYBsd")
	assert.Equal(t, changeLog.Input.ReceiverId, "receiverId")
	assert.Equal(t, changeLog.Input.AccountId, "accountId")
	assert.Equal(t, changeLog.Input.IpAddress, "48.26.240.152")
	assert.Equal(t, changeLog.Input.ControllerId, int64(1))
	assert.Equal(t, changeLog.Input.ChannelMapId, int64(1))
	assert.Equal(t, changeLog.Input.VodId, int64(1))
	assert.Equal(t, changeLog.Input.Rcdl, true)
	assert.Equal(t, changeLog.Input.SupportsFullHttpUrl, true)
	assert.Equal(t, changeLog.Input.RebootDecoupled, true)
	assert.Assert(t, changeLog.Input.TimeZone != nil)
	assert.Assert(t, changeLog.Input.Time != nil)
	assert.Assert(t, changeLog.Input.BypassFilters["bypassFilters"] == struct{}{})
	assert.Assert(t, changeLog.Input.ForceFilters["forceFilters"] == struct{}{})
	assert.Assert(t, changeLog.Input.Capabilities[0] == "RCDL")

	assert.Assert(t, changeLog.Rule != nil)
	assert.Assert(t, changeLog.Rule.NoOp)
	assert.Assert(t, changeLog.Rule.Blocking)
	assert.Equal(t, changeLog.Rule.ID, "595cd34d-f572-4f86-b5e2-3ded98113874")
	assert.Equal(t, changeLog.Rule.Type, "MAC_RULE")
	assert.Equal(t, changeLog.Rule.Name, "XconfTest")

	assert.Assert(t, changeLog.Filters != nil)
	assert.Assert(t, len(changeLog.Filters) == 2)

	assert.Assert(t, changeLog.FirmwareConfig != nil)
	assert.Assert(t, changeLog.FirmwareConfig.Properties["rebootImmediately"])
	assert.Equal(t, changeLog.FirmwareConfig.Properties["firmwareDownloadProtocol"], "http")
	assert.Equal(t, changeLog.FirmwareConfig.Properties["firmwareFilename"], "PY200AN_5.4p36s6_VBN_DYBsd-signed.bin")
	assert.Equal(t, changeLog.FirmwareConfig.Properties["firmwareLocation"], "dac15cdlserver.ae.ccp.xcal.tv")
	assert.Equal(t, changeLog.FirmwareConfig.Properties["firmwareVersion"], "PY200AN_5.4p36s6_VBN_DYBsd")

	assert.Assert(t, len(changeLog.Explanation) > 100)
	assert.Assert(t, changeLog.HasMinimumFirmware)

	list, err := ds.GetListingDao().GetAll(ds.TABLE_LOGS, rowKey)
	assert.NilError(t, err)
	assert.Assert(t, list != nil)
	assert.Assert(t, len(list) == 3)

	// test delete
	err = ds.GetListingDao().DeleteOne(ds.TABLE_LOGS, rowKey, coreef.LAST_CONFIG_LOG_ID)
	assert.NilError(t, err)

	list, err = ds.GetListingDao().GetKey2AsList(ds.TABLE_LOGS, rowKey)
	assert.NilError(t, err)
	assert.Assert(t, list != nil)
	assert.Assert(t, len(list) == 2)

	keys := []string{"tvxads-de-k8-xconfds-0153e903b521e2a9e_1", "tvxads-de-k8-xconfds-0153e903b521e2a9e_2"}
	assert.Assert(t, util.Contains(keys, list[0]))
	assert.Assert(t, util.Contains(keys, list[1]))

	err = ds.GetListingDao().DeleteAll(ds.TABLE_LOGS, rowKey)
	assert.NilError(t, err)
}

func TestListingGetRange(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	tableInfo, err := ds.GetTableInfo(ds.TABLE_MODEL)
	assert.NilError(t, err)

	daoId := tableInfo.DaoId
	assert.Assert(t, daoId != 0)

	changedData := ds.ChangedData{
		ColumnName:     gocql.TimeUUID(),
		CfName:         ds.TABLE_MODEL,
		ChangedKey:     fmt.Sprintf("Model-%s", uuid.New().String()),
		Operation:      ds.CREATE_OPERATION,
		DaoId:          daoId,
		ValidCacheSize: 1000,
		UserName:       "DataService",
		ServerOriginId: common.ServerOriginId(),
	}

	// test create
	jsonData, err := json.Marshal(changedData)
	assert.NilError(t, err)

	currentTS := util.GetTimestamp(time.Now().UTC())
	rowKey := currentTS - (currentTS % int64(10000)) // 10 secs window

	err = ds.GetListingDao().SetOne(ds.TABLE_XCONF_CHANGED_KEYS, rowKey, changedData.ColumnName, jsonData)
	assert.NilError(t, err)

	// test retrieve
	startUuid, err := util.UUIDFromTime(currentTS, 0, 0)
	assert.NilError(t, err)

	rangeInfo := &ds.RangeInfo{StartValue: startUuid}
	list, err := ds.GetListingDao().GetRange(ds.TABLE_XCONF_CHANGED_KEYS, rowKey, rangeInfo)
	assert.NilError(t, err)
	assert.Assert(t, len(list) == 1)

	inst := *list[0].(*ds.ChangedData)
	assert.Equal(t, inst.ColumnName, changedData.ColumnName)
	assert.Equal(t, inst.CfName, changedData.CfName)
	assert.Equal(t, inst.ChangedKey, changedData.ChangedKey)
	assert.Equal(t, inst.Operation, changedData.Operation)
	assert.Equal(t, inst.DaoId, changedData.DaoId)
	assert.Equal(t, inst.ValidCacheSize, changedData.ValidCacheSize)
	assert.Equal(t, inst.UserName, changedData.UserName)
	assert.Equal(t, inst.ServerOriginId, changedData.ServerOriginId)
}
