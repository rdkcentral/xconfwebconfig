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
package logupload

import (
	"encoding/json"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/db"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"

	"gotest.tools/assert"
)

type cachedSimpleDaoMock struct{}

func (dao cachedSimpleDaoMock) GetOne(tableName string, rowKey string) (interface{}, error) {
	return getOneMock(tableName, rowKey)
}

func (dao cachedSimpleDaoMock) GetOneFromCacheOnly(tableName string, rowKey string) (interface{}, error) {
	return getOneMock(tableName, rowKey)
}

func (dao cachedSimpleDaoMock) SetOne(tableName string, rowKey string, entity interface{}) error {
	return setOneMock(tableName, rowKey, entity)
}

func (dao cachedSimpleDaoMock) DeleteOne(tableName string, rowKey string) error {
	return deleteOneMock(tableName, rowKey)
}

func (dao cachedSimpleDaoMock) GetAllByKeys(tableName string, rowKeys []string) ([]interface{}, error) {
	return getAllByKeysMock(tableName, rowKeys)
}

func (dao cachedSimpleDaoMock) GetAllAsList(tableName string, maxResults int) ([]interface{}, error) {
	return getAllAsListMock(tableName, maxResults)
}

func (dao cachedSimpleDaoMock) GetAllAsMap(tableName string) (map[interface{}]interface{}, error) {
	return getAllAsMapMock(tableName)
}

func (dao cachedSimpleDaoMock) GetKeys(tableName string) ([]interface{}, error) {
	return getKeysMock(tableName)
}

func (dao cachedSimpleDaoMock) RefreshAll(tableName string) error {
	return refreshAllMock(tableName)
}

func (dao cachedSimpleDaoMock) RefreshOne(tableName string, rowKey string) error {
	return refreshOne(tableName, rowKey)
}

var getOneMock func(tableName string, rowKey string) (interface{}, error)
var setOneMock func(tableName string, rowKey string, entity interface{}) error
var deleteOneMock func(tableName string, rowKey string) error
var getAllByKeysMock func(tableName string, rowKeys []string) ([]interface{}, error)
var getAllAsListMock func(tableName string, maxResults int) ([]interface{}, error)
var getAllAsMapMock func(tableName string) (map[interface{}]interface{}, error)
var getKeysMock func(tableName string) ([]interface{}, error)
var refreshAllMock func(tableName string) error
var refreshOne func(tableName string, rowKey string) error

func TestGetOne(t *testing.T) {
	GetCachedSimpleDaoFunc = func() db.CachedSimpleDao {
		return cachedSimpleDaoMock{}
	}
	getOneMock = func(tableName string, rowKey string) (interface{}, error) {
		if tableName == db.TABLE_TELEMETRY {
			telemetryProfile := TelemetryProfile{
				ID:               "id",
				Name:             "name",
				TelemetryProfile: nil,
				Schedule:         "Schedule",
				Expires:          123456,
				UploadRepository: "uploadRepository:URL",
				ApplicationType:  "ApplicationType",
			}
			return telemetryProfile, nil
		}
		return nil, nil
	}
	telemetryProfile := GetOneTelemetryProfile("rowKey")
	assert.Equal(t, telemetryProfile.ID, "id")
	assert.Equal(t, telemetryProfile.Schedule, "Schedule")
	var a int64 = 123456
	assert.Equal(t, telemetryProfile.Expires, a)
	assert.Equal(t, telemetryProfile.UploadRepository, "uploadRepository:URL")
}

func TestGetTelemetryProfileList(t *testing.T) {
	GetCachedSimpleDaoFunc = func() db.CachedSimpleDao {
		return cachedSimpleDaoMock{}
	}
	getAllAsListMock = func(tableName string, maxResults int) ([]interface{}, error) {
		if tableName == db.TABLE_TELEMETRY {
			telemetryProfile1 := TelemetryProfile{
				ID:               "id1",
				Name:             "name1",
				TelemetryProfile: nil,
				Schedule:         "Schedule1",
				Expires:          123451,
				UploadRepository: "uploadRepository:URL1",
				ApplicationType:  "ApplicationType1",
			}
			telemetryProfile2 := TelemetryProfile{
				ID:               "id2",
				Name:             "name2",
				TelemetryProfile: nil,
				Schedule:         "Schedule2",
				Expires:          123452,
				UploadRepository: "uploadRepository:URL2",
				ApplicationType:  "ApplicationType2",
			}
			tpList := make([]interface{}, 0)
			tpList = append(tpList, telemetryProfile1)
			tpList = append(tpList, telemetryProfile2)
			return tpList, nil
		}
		return nil, nil
	}
	//[]*TelemetryProfile
	telemetryProfileList := GetTelemetryProfileList()
	assert.Equal(t, len(telemetryProfileList), 2)
	assert.Equal(t, telemetryProfileList[0].ApplicationType, "ApplicationType1")
	assert.Equal(t, telemetryProfileList[1].UploadRepository, "uploadRepository:URL2")
}

func TestGetTelemetryProfileMap(t *testing.T) {

	GetCachedSimpleDaoFunc = func() db.CachedSimpleDao {
		return cachedSimpleDaoMock{}
	}
	getAllAsMapMock = func(tableName string) (map[interface{}]interface{}, error) {
		if tableName == db.TABLE_TELEMETRY {
			telemetryProfile1 := TelemetryProfile{
				ID:               "id1",
				Name:             "name1",
				TelemetryProfile: nil,
				Schedule:         "Schedule1",
				Expires:          123451,
				UploadRepository: "uploadRepository:URL1",
				ApplicationType:  "ApplicationType1",
			}
			telemetryProfile2 := TelemetryProfile{
				ID:               "id2",
				Name:             "name2",
				TelemetryProfile: nil,
				Schedule:         "Schedule2",
				Expires:          123452,
				UploadRepository: "uploadRepository:URL2",
				ApplicationType:  "ApplicationType2",
			}
			rule := re.Rule{
				Negated:  true,
				Relation: "Relation",
			}
			timestampedRule1 := TimestampedRule{
				Rule:      rule,
				Timestamp: 1234561,
			}
			timestampedRule2 := TimestampedRule{
				Rule:      rule,
				Timestamp: 1234562,
			}
			timestampedRuleBytes1, _ := json.Marshal(timestampedRule1)
			timestampedRuleBytes2, _ := json.Marshal(timestampedRule2)

			map1 := make(map[interface{}]interface{})
			map1[string(timestampedRuleBytes1)] = telemetryProfile1
			map1[string(timestampedRuleBytes2)] = telemetryProfile2
			return map1, nil
		}
		return nil, nil
	}
	finalMap := GetTelemetryProfileMap()
	assert.Equal(t, len(*finalMap), 2)
	var a1 int64 = 1234561
	for k, v := range *finalMap {
		bytes := []byte(k)
		var timestampedRule TimestampedRule
		json.Unmarshal(bytes, &timestampedRule)
		if timestampedRule.Timestamp == a1 {
			assert.Equal(t, v.ApplicationType, "ApplicationType1")
		}
	}
}
