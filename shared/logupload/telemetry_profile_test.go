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
	"github.com/rdkcentral/xconfwebconfig/shared"

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

// Test Clone and New constructor functions for better coverage
func TestUploadRepositoryClone(t *testing.T) {
	original := &UploadRepository{
		ID:              "test-id",
		Name:            "test-name",
		Description:     "test description",
		URL:             "http://test.com",
		ApplicationType: "STB",
		Protocol:        "HTTP",
	}

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.Description, cloned.Description)
	assert.Equal(t, original.URL, cloned.URL)
	assert.Equal(t, original.ApplicationType, cloned.ApplicationType)
	assert.Equal(t, original.Protocol, cloned.Protocol)

	// Verify it's a deep copy
	cloned.ID = "modified-id"
	assert.Assert(t, original.ID != cloned.ID)
}

func TestNewUploadRepositoryInf(t *testing.T) {
	result := NewUploadRepositoryInf()
	repo, ok := result.(*UploadRepository)
	assert.Assert(t, ok, "Should return *UploadRepository")
	assert.Equal(t, repo.ApplicationType, shared.STB)
}

func TestLogFileClone(t *testing.T) {
	original := &LogFile{
		ID:             "log-file-id",
		Updated:        1234567890,
		Name:           "test.log",
		DeleteOnUpload: true,
	}

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.DeleteOnUpload, cloned.DeleteOnUpload)

	// Verify it's a deep copy
	cloned.Name = "modified.log"
	assert.Assert(t, original.Name != cloned.Name)
}

func TestNewLogFileInf(t *testing.T) {
	result := NewLogFileInf()
	logFile, ok := result.(*LogFile)
	assert.Assert(t, ok, "Should return *LogFile")
	assert.Equal(t, logFile.ID, "")
	assert.Equal(t, logFile.Updated, int64(0))
}

func TestLogFilesGroupsClone(t *testing.T) {
	original := &LogFilesGroups{
		ID:         "group-id",
		Updated:    1234567890,
		GroupName:  "group-name",
		LogFileIDs: []string{"file1", "file2"},
	}

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, original.GroupName, cloned.GroupName)
	assert.Equal(t, len(original.LogFileIDs), len(cloned.LogFileIDs))
}

func TestNewLogFilesGroupsInf(t *testing.T) {
	result := NewLogFilesGroupsInf()
	group, ok := result.(*LogFilesGroups)
	assert.Assert(t, ok, "Should return *LogFilesGroups")
	assert.Equal(t, group.ID, "")
}

func TestLogFileListClone(t *testing.T) {
	logFile1 := &LogFile{ID: "file1", Name: "test1.log"}
	logFile2 := &LogFile{ID: "file2", Name: "test2.log"}
	original := &LogFileList{
		Updated: 1234567890,
		Data:    []*LogFile{logFile1, logFile2},
	}

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, len(original.Data), len(cloned.Data))
	assert.Equal(t, original.Data[0].ID, cloned.Data[0].ID)
}

func TestNewLogFileListInf(t *testing.T) {
	result := NewLogFileListInf()
	list, ok := result.(*LogFileList)
	assert.Assert(t, ok, "Should return *LogFileList")
	assert.Equal(t, list.Updated, int64(0))
}

// Test DCMGenericRule methods
func TestDCMGenericRule_GetPriority(t *testing.T) {
	rule := &DCMGenericRule{Priority: 5}
	assert.Equal(t, rule.GetPriority(), 5)
}

func TestDCMGenericRule_SetPriority(t *testing.T) {
	rule := &DCMGenericRule{}
	rule.SetPriority(10)
	assert.Equal(t, rule.Priority, 10)
}

func TestDCMGenericRule_GetID(t *testing.T) {
	rule := &DCMGenericRule{ID: "test-rule-id"}
	assert.Equal(t, rule.GetID(), "test-rule-id")
}

func TestDCMGenericRule_Clone(t *testing.T) {
	original := &DCMGenericRule{
		ID:          "rule-id",
		Name:        "test-rule",
		Description: "test description",
		Priority:    5,
		Percentage:  80,
	}

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.Priority, cloned.Priority)
	assert.Equal(t, original.Percentage, cloned.Percentage)

	// Verify it's a deep copy
	cloned.ID = "modified-id"
	assert.Assert(t, original.ID != cloned.ID)
}

func TestNewDCMGenericRuleInf(t *testing.T) {
	result := NewDCMGenericRuleInf()
	rule, ok := result.(*DCMGenericRule)
	assert.Assert(t, ok, "Should return *DCMGenericRule")
	assert.Equal(t, rule.Percentage, 100)
	assert.Equal(t, rule.ApplicationType, shared.STB)
}

// Test Settings functions
func TestIsValidSettingType(t *testing.T) {
	// Test valid setting types
	assert.Assert(t, IsValidSettingType("PARTNER_SETTINGS"))
	assert.Assert(t, IsValidSettingType("EPON"))
	assert.Assert(t, IsValidSettingType("partnersettings"))
	assert.Assert(t, IsValidSettingType("epon"))

	// Test invalid setting types
	assert.Assert(t, !IsValidSettingType("INVALID"))
	assert.Assert(t, !IsValidSettingType(""))
	assert.Assert(t, !IsValidSettingType("random"))
}

func TestSettingTypeEnum(t *testing.T) {
	// Test valid enums
	assert.Equal(t, SettingTypeEnum("epon"), EPON)
	assert.Equal(t, SettingTypeEnum("EPON"), EPON)
	assert.Equal(t, SettingTypeEnum("partner_settings"), PARTNER_SETTINGS)
	assert.Equal(t, SettingTypeEnum("partnersettings"), PARTNER_SETTINGS)
	assert.Equal(t, SettingTypeEnum("PARTNERSETTINGS"), PARTNER_SETTINGS)

	// Test invalid enum
	assert.Equal(t, SettingTypeEnum("invalid"), 0)
	assert.Equal(t, SettingTypeEnum(""), 0)
}

func TestSettingProfiles_Clone(t *testing.T) {
	original := &SettingProfiles{
		ID:               "settings-id",
		Updated:          1234567890,
		SettingProfileID: "profile-id",
		SettingType:      "EPON",
		Properties:       map[string]string{"key1": "value1", "key2": "value2"},
		ApplicationType:  "STB",
	}

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.SettingProfileID, cloned.SettingProfileID)
	assert.Equal(t, original.SettingType, cloned.SettingType)
	assert.Equal(t, len(original.Properties), len(cloned.Properties))
	assert.Equal(t, original.Properties["key1"], cloned.Properties["key1"])
}

func TestNewSettingProfilesInf(t *testing.T) {
	result := NewSettingProfilesInf()
	profiles, ok := result.(*SettingProfiles)
	assert.Assert(t, ok, "Should return *SettingProfiles")
	assert.Equal(t, profiles.ApplicationType, shared.STB)
}

func TestVodSettings_Clone(t *testing.T) {
	original := &VodSettings{
		ID:           "vod-id",
		Updated:      1234567890,
		Name:         "vod-settings",
		LocationsURL: "http://locations.com",
		IPNames:      []string{"ip1", "ip2"},
		IPList:       []string{"192.168.1.1", "192.168.1.2"},
		SrmIPList:    map[string]string{"srm1": "10.0.0.1"},
	}

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, len(original.IPNames), len(cloned.IPNames))
	assert.Equal(t, original.IPNames[0], cloned.IPNames[0])
}

func TestNewVodSettingsInf(t *testing.T) {
	result := NewVodSettingsInf()
	vod, ok := result.(*VodSettings)
	assert.Assert(t, ok, "Should return *VodSettings")
	assert.Equal(t, vod.ApplicationType, shared.STB)
}

func TestDeviceSettings_Clone(t *testing.T) {
	original := &DeviceSettings{
		ID:                      "device-id",
		Updated:                 1234567890,
		Name:                    "device-settings",
		CheckOnReboot:           true,
		SettingsAreActive:       true,
		Schedule:                Schedule{Type: "cron", Expression: "0 0 * * *"},
		ConfigurationServiceURL: ConfigurationServiceURL{URL: "http://config.com"},
	}

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.CheckOnReboot, cloned.CheckOnReboot)
	assert.Equal(t, original.Schedule.Type, cloned.Schedule.Type)
}

func TestNewDeviceSettingsInf(t *testing.T) {
	result := NewDeviceSettingsInf()
	device, ok := result.(*DeviceSettings)
	assert.Assert(t, ok, "Should return *DeviceSettings")
	assert.Equal(t, device.ApplicationType, shared.STB)
}

func TestLogUploadSettings_Clone(t *testing.T) {
	original := &LogUploadSettings{
		ID:                "upload-id",
		Updated:           1234567890,
		Name:              "upload-settings",
		UploadOnReboot:    true,
		NumberOfDays:      7,
		AreSettingsActive: true,
	}

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.UploadOnReboot, cloned.UploadOnReboot)
	assert.Equal(t, original.NumberOfDays, cloned.NumberOfDays)
}

func TestNewLogUploadSettingsInf(t *testing.T) {
	result := NewLogUploadSettingsInf()
	upload, ok := result.(*LogUploadSettings)
	assert.Assert(t, ok, "Should return *LogUploadSettings")
	assert.Equal(t, upload.ApplicationType, shared.STB)
}

// Test TelemetryProfile functions
func TestTelemetryProfile_Clone(t *testing.T) {
	original := &TelemetryProfile{
		ID:               "telemetry-id",
		Name:             "telemetry-profile",
		Schedule:         "0 0 * * *",
		UploadRepository: "http://upload.com",
		ApplicationType:  "STB",
	}

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.Schedule, cloned.Schedule)
	assert.Equal(t, original.UploadRepository, cloned.UploadRepository)
}

func TestNewTelemetryProfileInf(t *testing.T) {
	result := NewTelemetryProfileInf()
	profile, ok := result.(*TelemetryProfile)
	assert.Assert(t, ok, "Should return *TelemetryProfile")
	assert.Equal(t, profile.ApplicationType, shared.STB)
}

func TestNewTelemetryProfileDescriptor(t *testing.T) {
	descriptor := NewTelemetryProfileDescriptor()
	assert.Assert(t, descriptor != nil, "Should return non-nil descriptor")
}

func TestIsValidUploadProtocol(t *testing.T) {
	assert.Assert(t, IsValidUploadProtocol("HTTP"))
	assert.Assert(t, IsValidUploadProtocol("HTTPS"))
	assert.Assert(t, IsValidUploadProtocol("SFTP"))
	assert.Assert(t, IsValidUploadProtocol("SCP"))
	assert.Assert(t, IsValidUploadProtocol("S3"))

	assert.Assert(t, !IsValidUploadProtocol("FTP"))
	assert.Assert(t, !IsValidUploadProtocol("INVALID"))
	assert.Assert(t, !IsValidUploadProtocol(""))
}

func TestIsValidUrl(t *testing.T) {
	assert.Assert(t, IsValidUrl("http://example.com"))
	assert.Assert(t, IsValidUrl("https://example.com"))
	assert.Assert(t, IsValidUrl("sftp://example.com"))

	assert.Assert(t, !IsValidUrl(""))
	assert.Assert(t, !IsValidUrl("invalid-url"))
	assert.Assert(t, !IsValidUrl("ftp://example.com"))
}

func TestNewPermanentTelemetryProfileInf(t *testing.T) {
	result := NewPermanentTelemetryProfileInf()
	profile, ok := result.(*PermanentTelemetryProfile)
	assert.Assert(t, ok, "Should return *PermanentTelemetryProfile")
	assert.Equal(t, profile.ApplicationType, shared.STB)
}

func TestNewTelemetryRuleInf(t *testing.T) {
	result := NewTelemetryRuleInf()
	rule, ok := result.(*TelemetryRule)
	assert.Assert(t, ok, "Should return *TelemetryRule")
	assert.Equal(t, rule.ApplicationType, shared.STB)
}

func TestNewTelemetryTwoRuleInf(t *testing.T) {
	result := NewTelemetryTwoRuleInf()
	rule, ok := result.(*TelemetryTwoRule)
	assert.Assert(t, ok, "Should return *TelemetryTwoRule")
	assert.Equal(t, rule.ApplicationType, shared.STB)
}

func TestNewTelemetryTwoProfileInf(t *testing.T) {
	result := NewTelemetryTwoProfileInf()
	profile, ok := result.(*TelemetryTwoProfile)
	assert.Assert(t, ok, "Should return *TelemetryTwoProfile")
	assert.Equal(t, profile.ApplicationType, shared.STB)
}
