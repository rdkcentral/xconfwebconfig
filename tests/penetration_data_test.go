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
	"testing"
	"time"

	"github.com/rdkcentral/xconfwebconfig/db"

	"gotest.tools/assert"
)

func TestFwPenetrationDataCRUD(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	client := db.GetDatabaseClient()
	tenantId := db.GetDefaultTenantId()
	estbMac := "A4:F3:E8:79:C8:60"

	pData := &db.FwPenetrationData{
		TenantId:                tenantId,
		EstbMac:                 estbMac,
		Partner:                 "comcast",
		Model:                   "Modelxyz",
		FwFilename:              "TEST_1.1p1s1_VBN_sstb.bin",
		FwVersion:               "TEST_1.1p1s1_VBN",
		FwReportedVersion:       "TEST_1.1p1s1_VBN",
		FwAdditionalVersionInfo: "extra-info",
		FwAppliedRule:           "MAC_RULE:test-rule-id",
		FwTs:                    time.Now().Unix(),
		ClientCertExpiry:        "2026-12-31",
		RecoveryCertExpiry:      "2026-12-31",
	}

	// test create
	err := client.SetFwPenetrationData(pData)
	assert.NilError(t, err)

	// test retrieve
	result, err := client.GetFwPenetrationData(estbMac)
	assert.NilError(t, err)
	assert.Assert(t, result != nil)

	assert.Equal(t, result.EstbMac, estbMac)
	assert.Equal(t, result.Partner, pData.Partner)
	assert.Equal(t, result.Model, pData.Model)
	assert.Equal(t, result.FwFilename, pData.FwFilename)
	assert.Equal(t, result.FwVersion, pData.FwVersion)
	assert.Equal(t, result.FwReportedVersion, pData.FwReportedVersion)
	assert.Equal(t, result.FwAdditionalVersionInfo, pData.FwAdditionalVersionInfo)
	assert.Equal(t, result.FwAppliedRule, pData.FwAppliedRule)

	// test update — overwrite with new firmware info
	pData.FwVersion = "TEST_1.1p1s1_VBN"
	pData.FwFilename = "TEST_1.1p1s1_VBN_sstb.bin"
	err = client.SetFwPenetrationData(pData)
	assert.NilError(t, err)

	updated, err := client.GetFwPenetrationData(estbMac)
	assert.NilError(t, err)
	assert.Equal(t, updated.FwVersion, "TEST_1.1p1s1_VBN")
	assert.Equal(t, updated.FwFilename, "TEST_1.1p1s1_VBN_sstb.bin")
}

func TestFwPenetrationDataEmptyFields(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	client := db.GetDatabaseClient()
	tenantId := db.GetDefaultTenantId()
	estbMac := "B5:C6:D7:E8:F9:01"

	// empty/unknown partner and model should not be written
	pData := &db.FwPenetrationData{
		TenantId:  tenantId,
		EstbMac:   estbMac,
		Partner:   "unknown",
		Model:     "",
		FwVersion: "TEST_VERSION",
		FwTs:      time.Now().Unix(),
	}

	err := client.SetFwPenetrationData(pData)
	assert.NilError(t, err)

	result, err := client.GetFwPenetrationData(estbMac)
	assert.NilError(t, err)
	assert.Equal(t, result.EstbMac, estbMac)
	assert.Equal(t, result.FwVersion, pData.FwVersion)
	// partner "unknown" and empty model are treated as empty and not written
	assert.Equal(t, result.Partner, "")
	assert.Equal(t, result.Model, "")
}

func TestRfcPenetrationDataCRUD(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	client := db.GetDatabaseClient()
	tenantId := db.GetDefaultTenantId()
	estbMac := "C6:D7:E8:F9:01:A2"

	pData := &db.RfcPenetrationData{
		TenantId:             tenantId,
		EstbMac:              estbMac,
		EcmMac:               "C6:D7:E8:F9:01:A3",
		SerialNum:            "SN123456789",
		Partner:              "comcast",
		Model:                "Modelxyz",
		RfcPartner:           "comcast",
		TitanPartner:         "titan-partner",
		RfcModel:             "Modelxyz",
		RfcFwReportedVersion: "TEST_1.1p1s1_VBN",
		RfcAppliedRules:      "rule1,rule2",
		RfcAccountMgmt:       "acct-mgmt",
		RfcFeatures:          "feature1,feature2",
		RfcTs:                time.Now().Unix(),
		RfcAccountHash:       "hash123",
		RfcAccountId:         "acct-123",
		TitanAccountId:       "titan-acct-456",
		RfcEnv:               "PROD",
		RfcApplicationType:   "stb",
		RfcExperience:        "X1",
		RfcTimeZone:          "US/Eastern",
		RfcConfigsetHash:     "configset-hash-abc",
		RfcQueryParams:       "param1=val1&param2=val2",
		RfcTags:              "tag1,tag2",
		RfcEstbIp:            "192.168.1.100",
		RfcPostProc:          "post-proc-val",
		ClientCertExpiry:     "2026-12-31",
		RecoveryCertExpiry:   "2026-12-31",
	}

	// test create (is304FromPrecook=false: features and applied rules ARE written)
	err := client.SetRfcPenetrationData(pData, false)
	assert.NilError(t, err)

	// test retrieve
	result, err := client.GetRfcPenetrationData(estbMac)
	assert.NilError(t, err)
	assert.Assert(t, result != nil)
	assert.Equal(t, result.EstbMac, estbMac)
	assert.Equal(t, result.Partner, pData.Partner)
	assert.Equal(t, result.Model, pData.Model)
	assert.Equal(t, result.RfcAppliedRules, pData.RfcAppliedRules)
	assert.Equal(t, result.RfcFeatures, pData.RfcFeatures)

}

func TestRfcPenetrationData304SkipsFeaturesAndRules(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	client := db.GetDatabaseClient()
	tenantId := db.GetDefaultTenantId()
	estbMac := "D7:E8:F9:01:A2:B3"

	// first write with real features/rules
	initial := &db.RfcPenetrationData{
		TenantId:        tenantId,
		EstbMac:         estbMac,
		RfcAppliedRules: "initial-rule",
		RfcFeatures:     "initial-feature",
		RfcTs:           time.Now().Unix(),
	}
	err := client.SetRfcPenetrationData(initial, false)
	assert.NilError(t, err)

	// now write with is304FromPrecook=true — features/rules must not be overwritten
	updated := &db.RfcPenetrationData{
		TenantId:        tenantId,
		EstbMac:         estbMac,
		RfcAppliedRules: "new-rule",
		RfcFeatures:     "new-feature",
		RfcTs:           time.Now().Unix(),
	}
	err = client.SetRfcPenetrationData(updated, true)
	assert.NilError(t, err)

	result, err := client.GetRfcPenetrationData(estbMac)
	assert.NilError(t, err)
	// features and applied rules should still be the initial values
	assert.Equal(t, result.RfcAppliedRules, "initial-rule")
	assert.Equal(t, result.RfcFeatures, "initial-feature")
}

func TestSetPenetrationDataWithMap(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	client := db.GetDatabaseClient()
	estbMac := "E8:F9:01:A2:B3:C4"

	kvmap := map[string]string{
		db.TenantIdColumnName:  db.GetDefaultTenantId(),
		db.EstbMacColumnName:   estbMac,
		db.FwVersionColumnName: "KVMap_FW_1.0",
		db.PartnerColumnName:   "comcast",
	}

	err := client.SetPenetrationData(kvmap)
	assert.NilError(t, err)

	result, err := client.GetFwPenetrationData(estbMac)
	assert.NilError(t, err)
	assert.Equal(t, result.EstbMac, estbMac)
	assert.Equal(t, result.FwVersion, "KVMap_FW_1.0")
	assert.Equal(t, result.Partner, "comcast")
}
