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
	"testing"

	"github.com/rdkcentral/xconfwebconfig/db"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/util"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestChangeLogs(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	tenantId := db.GetDefaultTenantId()
	key := "A4:F3:E8:79:C8:60"
	dualWriteModes := []bool{true, false}
	for _, mode := range dualWriteModes {
		t.Run(fmt.Sprintf("IsDualWriteEnabled=%v", mode), func(t *testing.T) {
			truncateTable(db.TABLE_CONFIG_CHANGE_LOGS)
			truncateTable(db.TABLE_LOGS)

			db.SetDualWriteEnabled(mode)

			// test create config log
			jsonData := []byte(configChangeLogJsonTemplate1)
			obj := coreef.NewConfigChangeLogInf()
			err := json.Unmarshal(jsonData, obj)
			assert.Nil(t, err)

			configChangeLog := obj.(*coreef.ConfigChangeLog)
			configChangeLog.Updated = util.GetTimestamp()
			err = coreef.SetConfigChangeLog(tenantId, key, configChangeLog)
			assert.Nil(t, err)

			configChangeLog.Updated = util.GetTimestamp()
			err = coreef.SetConfigChangeLog(tenantId, key, configChangeLog)
			assert.Nil(t, err)

			// test create confg last log
			configChangeLog.ID = coreef.LAST_CONFIG_LOG_ID
			configChangeLog.Updated = 0 // last log does not set updated time
			err = coreef.SetLastConfigLog(tenantId, key, configChangeLog)
			assert.Nil(t, err)
			list := coreef.GetConfigChangeLogsOnly(tenantId, key)
			assert.NotNil(t, list)
			assert.Equal(t, len(list), 2)

			// test retrieve config log
			retrievedConfigChangeLog := coreef.GetLastConfigLog(tenantId, key)
			assert.NotNil(t, retrievedConfigChangeLog)
			assert.Equal(t, retrievedConfigChangeLog.ID, configChangeLog.ID)
			assert.Equal(t, retrievedConfigChangeLog.Input.EstbMac, configChangeLog.Input.EstbMac)
			assert.Equal(t, retrievedConfigChangeLog.Input.Env, configChangeLog.Input.Env)
			assert.Equal(t, retrievedConfigChangeLog.Input.Model, configChangeLog.Input.Model)

			if db.IsDualWriteEnabled() {
				// when dual write is enabled, ensure we also write in Logs2 table
				data, err := db.GetListingDao().GetAll(tenantId, db.TABLE_LOGS, configChangeLog.Input.EstbMac)
				assert.Nil(t, err)
				assert.Equal(t, len(data), 3) // 2 config logs + 1 last log
			} else {
				// when dual write is disabled, ensure we do not write in Logs2 table
				data, err := db.GetListingDao().GetAll(tenantId, db.TABLE_LOGS, configChangeLog.Input.EstbMac)
				assert.Nil(t, err)
				assert.Equal(t, len(data), 0)
			}
		})
	}
}
