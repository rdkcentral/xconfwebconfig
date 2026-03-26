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
	// Common
	TABLE_APPLICATION_TYPES  = "application_types"
	TABLE_ENVIRONMENTS       = "environments"
	TABLE_MODELS             = "models"
	TABLE_GENERIC_NS_LIST    = "generic_named_lists"
	TABLE_APP_SETTINGS       = "app_settings"
	TABLE_LOCKS              = "locks"
	TABLE_CONFIG_CHANGE_LOGS = "config_change_logs"
	TABLE_CHANGE_EVENTS      = "change_events"

	// Firmware
	TABLE_FIRMWARE_RULES          = "firmware_rules"
	TABLE_FIRMWARE_CONFIGS        = "firmware_configs"
	TABLE_FIRMWARE_RULE_TEMPLATES = "firmware_rule_templates"
	TABLE_SINGLETON_FILTER_VALUES = "singleton_filter_values"

	// RFC
	TABLE_FEATURE_CONTROL_RULES = "feature_control_rules"
	TABLE_FEATURES              = "features"

	// Setting
	TABLE_SETTING_PROFILES = "setting_profiles"
	TABLE_SETTING_RULES    = "setting_rules"

	// DCM - LogUpload
	TABLE_DCM_RULES           = "dcm_rules"
	TABLE_UPLOAD_REPOSITORIES = "upload_repositories"
	TABLE_LOG_UPLOAD_SETTINGS = "log_upload_settings"
	TABLE_LOG_FILES           = "log_files"
	TABLE_LOG_FILE_LISTS      = "log_file_lists"
	TABLE_DEVICE_SETTINGS     = "device_settings"
	TABLE_VOD_SETTINGS        = "vod_settings"

	// Telemetry
	TABLE_TELEMETRY_PROFILES           = "telemetry_profiles"
	TABLE_TELEMETRY_RULES              = "telemetry_rules"
	TABLE_TELEMETRY_TWO_PROFILES       = "telemetry_two_profiles"
	TABLE_TELEMETRY_TWO_RULES          = "telemetry_two_rules"
	TABLE_PERMANENT_TELEMETRY_PROFILES = "permanent_telemetry_profiles"

	// Telemetry - Changes
	TABLE_TELEMETRY_CHANGES              = "telemetry_changes"
	TABLE_TELEMETRY_APPROVED_CHANGES     = "telemetry_approved_changes"
	TABLE_TELEMETRY_TWO_CHANGES          = "telemetry_two_changes"
	TABLE_TELEMETRY_APPROVED_TWO_CHANGES = "telemetry_approved_two_changes"

	// Old tables for backwards compatibility
	TABLE_LOGS = "Logs2"
)

var AllTables = []string{
	TABLE_DCM_RULES,
	TABLE_UPLOAD_REPOSITORIES,
	TABLE_LOG_UPLOAD_SETTINGS,
	TABLE_LOG_FILES,
	TABLE_LOG_FILE_LISTS,
	TABLE_DEVICE_SETTINGS,
	TABLE_VOD_SETTINGS,
	TABLE_SETTING_PROFILES,
	TABLE_SETTING_RULES,
	TABLE_TELEMETRY_PROFILES,
	TABLE_TELEMETRY_RULES,
	TABLE_TELEMETRY_TWO_RULES,
	TABLE_TELEMETRY_TWO_PROFILES,
	TABLE_PERMANENT_TELEMETRY_PROFILES,
	TABLE_FIRMWARE_RULES,
	TABLE_FIRMWARE_RULE_TEMPLATES,
	TABLE_FIRMWARE_CONFIGS,
	TABLE_SINGLETON_FILTER_VALUES,
	TABLE_FEATURE_CONTROL_RULES,
	TABLE_FEATURES,
	TABLE_TELEMETRY_CHANGES,
	TABLE_TELEMETRY_APPROVED_CHANGES,
	TABLE_TELEMETRY_TWO_CHANGES,
	TABLE_TELEMETRY_APPROVED_TWO_CHANGES,
	TABLE_ENVIRONMENTS,
	TABLE_MODELS,
	TABLE_GENERIC_NS_LIST,
	TABLE_CONFIG_CHANGE_LOGS,
	TABLE_CHANGE_EVENTS,
	TABLE_APPLICATION_TYPES,
}

// Two possible values for Key2FieldName that is used for list types of tables
// (e.g. Logs2, XconfChangedKeys4) where we need to specify the column name for the second key
const (
	Key2FieldNameForList        = "column1"
	Key2FieldNameForChangedKeys = "columnName"
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
	Key2FieldName   string             // column name for list types of tables, e.g. Logs2, XconfChangedKeys4
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
	}
}

/*
 * RegisterTableConfig registers information related to the table name,
 * i.e. constructor function, compression policy, and TTL for data.
 */
func RegisterTableConfig(tableInfo *TableInfo) {
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
