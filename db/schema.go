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

import (
	"fmt"
)

const (
	// LogUpload
	TABLE_DCM_RULE            = "DcmRule"
	TABLE_UPLOAD_REPOSITORY   = "UploadRepository"
	TABLE_LOG_UPLOAD_SETTINGS = "LogUploadSettings2"
	TABLE_LOG_FILE            = "LogFile"
	TABLE_LOG_FILE_LIST       = "LogFileList"
	TABLE_INDEXED_LOG_FILES   = "IndexedLogFiles"
	TABLE_LOG_FILES_GROUPS    = "LogFilesGroups"
	TABLE_DEVICE_SETTINGS     = "DeviceSettings2"
	TABLE_VOD_SETTINGS        = "VodSettings2"

	// Setting
	TABLE_SETTING_PROFILES = "SettingProfiles"
	TABLE_SETTING_RULES    = "SettingRules"

	// Telemetry
	TABLE_TELEMETRY              = "Telemetry"
	TABLE_TELEMETRY_RULES        = "TelemetryRules"
	TABLE_TELEMETRY_TWO_RULES    = "TelemetryTwoRules"
	TABLE_TELEMETRY_TWO_PROFILES = "TelemetryTwoProfiles"
	TABLE_PERMANENT_TELEMETRY    = "PermanentTelemetry"

	// Firmware
	TABLE_FIRMWARE_RULE          = "FirmwareRule4"
	TABLE_FIRMWARE_RULE_TEMPLATE = "FirmwareRuleTemplate"
	TABLE_FIRMWARE_CONFIG        = "FirmwareConfig"
	TABLE_SINGLETON_FILTER_VALUE = "SingletonFilterValue"

	// RFC
	TABLE_FEATURE_CONTROL_RULE = "FeatureControlRule2"
	TABLE_XCONF_FEATURE        = "XconfFeature"

	// Change
	TABLE_XCONF_CHANGE                        = "XconfChange"
	TABLE_XCONF_APPROVED_CHANGE               = "XconfApprovedChange"
	TABLE_XCONF_TELEMETRY_TWO_CHANGE          = "XconfTelemetryTwoChange"
	TABLE_XCONF_APPROVED_TELEMETRY_TWO_CHANGE = "XconfApprovedTelemetryTwoChange"

	// Common
	TABLE_ENVIRONMENT        = "Environment"
	TABLE_MODEL              = "Model"
	TABLE_IP_ADDRESS_GROUP   = "IpAddressGroupExtended"
	TABLE_NS_LIST            = "XconfNamedList"
	TABLE_GENERIC_NS_LIST    = "GenericXconfNamedList"
	TABLE_LOGS               = "Logs2"
	TABLE_XCONF_CHANGED_KEYS = "XconfChangedKeys4"
	TABLE_APP_SETTINGS       = "AppSettings"
	TABLE_TAG                = "Tag"
)

var AllTables = []string{
	TABLE_DCM_RULE,
	TABLE_UPLOAD_REPOSITORY,
	TABLE_LOG_UPLOAD_SETTINGS,
	TABLE_LOG_FILE,
	TABLE_LOG_FILE_LIST,
	TABLE_INDEXED_LOG_FILES,
	TABLE_LOG_FILES_GROUPS,
	TABLE_DEVICE_SETTINGS,
	TABLE_VOD_SETTINGS,
	TABLE_SETTING_PROFILES,
	TABLE_SETTING_RULES,
	TABLE_TELEMETRY,
	TABLE_TELEMETRY_RULES,
	TABLE_TELEMETRY_TWO_RULES,
	TABLE_TELEMETRY_TWO_PROFILES,
	TABLE_PERMANENT_TELEMETRY,
	TABLE_FIRMWARE_RULE,
	TABLE_FIRMWARE_RULE_TEMPLATE,
	TABLE_FIRMWARE_CONFIG,
	TABLE_SINGLETON_FILTER_VALUE,
	TABLE_FEATURE_CONTROL_RULE,
	TABLE_XCONF_FEATURE,
	TABLE_XCONF_CHANGE,
	TABLE_XCONF_APPROVED_CHANGE,
	TABLE_XCONF_TELEMETRY_TWO_CHANGE,
	TABLE_XCONF_APPROVED_TELEMETRY_TWO_CHANGE,
	TABLE_ENVIRONMENT,
	TABLE_MODEL,
	TABLE_IP_ADDRESS_GROUP,
	TABLE_NS_LIST,
	TABLE_GENERIC_NS_LIST,
	TABLE_LOGS,
	TABLE_XCONF_CHANGED_KEYS,
	TABLE_TAG,
}

// Two possible values for Key2FieldName
const (
	DefaultKey2FieldName     = "column1"
	ChangedKeysKey2FieldName = "columnName"
)

/*
	Table configuration
*/

type TableInfo struct {
	TableName       string
	ConstructorFunc func() interface{} // model/struct constructor function
	Compress        bool               // data is compressed
	Split           bool               // data is split into multiple chunks
	CacheData       bool               // specifies whether to cache the data
	TTL             int                // TTL for the data
	Key2FieldName   string             // column name for listing table, e.g. Logs2, XconfChangedKeys4
	DaoId           int32              // Xconf DAO ID
}

// tableConfig is a map of table name to TableInfo
var tableConfig = make(map[string]TableInfo)

/*
 * RegisterTableConfigSimple registers constructor function for a table.
 */
func RegisterTableConfigSimple(tableName string, fn func() interface{}) {
	tableConfig[tableName] = TableInfo{
		TableName:       tableName,
		ConstructorFunc: fn,
		CacheData:       true,
		Key2FieldName:   DefaultKey2FieldName,
		DaoId:           xconfDaoIdMap[tableName],
	}
}

/*
 * RegisterTableConfig registers information related to the table name,
 * i.e. constructor function, compression policy, and TTL for data.
 */
func RegisterTableConfig(tableInfo *TableInfo) {
	if tableInfo.Key2FieldName == "" {
		tableInfo.Key2FieldName = DefaultKey2FieldName
	}
	tableInfo.DaoId = xconfDaoIdMap[tableInfo.TableName]
	tableConfig[tableInfo.TableName] = *tableInfo
}

func GetTableInfo(tableName string) (*TableInfo, error) {
	tableInfo := tableConfig[tableName]
	if tableInfo.TableName == "" || tableInfo.ConstructorFunc == nil {
		err := fmt.Errorf("Table configuration not found for table '%v'", tableName)
		return nil, err
	}
	return &tableInfo, nil
}

func GetAllTableInfo() []TableInfo {
	var result []TableInfo
	for _, tableInfo := range tableConfig {
		result = append(result, tableInfo)
	}
	return result
}

// IsCompressOnly checks if data is compressed, i.e. ListingDao
func (ti *TableInfo) IsCompressOnly() bool {
	return ti.Compress && !ti.Split
}

// IsCompressAndSplit checks if data is compressed and split, i.e. CompressingDataDao
func (ti *TableInfo) IsCompressAndSplit() bool {
	return ti.Compress && ti.Split
}

var xconfDaoIdMap = map[string]int32{
	TABLE_FIRMWARE_CONFIG:        1586263717,
	TABLE_DCM_RULE:               97791402,
	TABLE_GENERIC_NS_LIST:        834213306,
	TABLE_TELEMETRY_RULES:        1229706247,
	TABLE_FIRMWARE_RULE:          720917275,
	TABLE_FIRMWARE_RULE_TEMPLATE: 1895299265,
	TABLE_SETTING_RULES:          -1826597405,
	TABLE_XCONF_FEATURE:          1413526283,
	TABLE_FEATURE_CONTROL_RULE:   -638116398,
	TABLE_IP_ADDRESS_GROUP:       505086088,
	TABLE_LOG_FILE_LIST:          -293241179,
	TABLE_DEVICE_SETTINGS:        803752845,
	TABLE_LOG_UPLOAD_SETTINGS:    1872416451,
	TABLE_LOG_FILE:               -1330571547,
	TABLE_UPLOAD_REPOSITORY:      -1493573423,
	TABLE_VOD_SETTINGS:           1003179471,
	TABLE_TELEMETRY:              -270293656,
	TABLE_PERMANENT_TELEMETRY:    1528055203,
	TABLE_TELEMETRY_TWO_PROFILES: -865570487,
	TABLE_TELEMETRY_TWO_RULES:    1765731811,
	TABLE_SETTING_PROFILES:       621503215,
	TABLE_ENVIRONMENT:            1586640622,
	TABLE_SINGLETON_FILTER_VALUE: 235925253,
	TABLE_MODEL:                  -795610003,
	TABLE_LOG_FILES_GROUPS:       -867320790,
	TABLE_NS_LIST:                1409490260,
	TABLE_APP_SETTINGS:           1,
	TABLE_TAG:                    1698455800,
}
