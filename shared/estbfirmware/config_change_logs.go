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
package estbfirmware

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rdkcentral/xconfwebconfig/util"
)

const (
	DEFAULT_PREFIX     string = "XCONF"
	BOUNDS             int    = 5
	LAST_CONFIG_LOG_ID        = "0"
)

var prefix string

func init() {
	name, err := os.Hostname()
	if err == nil {
		prefix = name
	} else {
		prefix = DEFAULT_PREFIX
	}
}

// RuleInfo ...
type RuleInfo struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Name     string `json:"name,omitempty"`
	NoOp     bool   `json:"noop"`
	Blocking bool   `json:"blocking,omitempty"`
}

// ConfigChangeLog config_change_logs table (old name: Logs2)
type ConfigChangeLog struct {
	ID                 string                `json:"id,omitempty"`
	Updated            int64                 `json:"updated,omitempty"`
	Input              *ConvertedContext     `json:"input"`
	Rule               *RuleInfo             `json:"rule"`
	Filters            []*RuleInfo           `json:"filters"`
	Explanation        string                `json:"explanation,omitempty"`
	FirmwareConfig     *FirmwareConfigFacade `json:"config"`
	HasMinimumFirmware bool                  `json:"hasMinimumFirmware"`
}

func (obj *ConfigChangeLog) GetUpdated() int64 {
	return obj.Updated
}

func (obj *ConfigChangeLog) SetUpdated(ts int64) {
	obj.Updated = ts
}

func NewRuleInfo(filterOrRule interface{}) *RuleInfo {
	switch t := filterOrRule.(type) {
	case *firmware.FirmwareRule:
		isBlocking := false
		if t.ApplicableAction != nil && t.ApplicableAction.ActionType == firmware.BLOCKING_FILTER {
			isBlocking = true
		}
		return &RuleInfo{
			ID:       t.ID,
			Type:     t.Type,
			Name:     t.Name,
			NoOp:     t.IsNoop(),
			Blocking: isBlocking,
		}
	case *SingletonFilterValue:
		i := strings.LastIndex(t.ID, "_VALUE")
		var id string
		if i == -1 {
			id = fmt.Sprintf("SINGLETON_%s", t.ID)
		} else {
			id = fmt.Sprintf("SINGLETON_%s", t.ID[:i])
		}
		return &RuleInfo{
			ID:       id,
			Type:     "SingletonFilter",
			Name:     t.ID,
			NoOp:     true,
			Blocking: false,
		}
	case *firmware.RuleAction:
		return &RuleInfo{
			ID:       "DistributionPercentInRuleAction",
			Type:     "DistributionPercentInRuleAction",
			Name:     "DistributionPercentInRuleAction",
			NoOp:     false,
			Blocking: false,
		}
	case *PercentageBean:
		return &RuleInfo{
			ID:       "",
			Type:     "PercentageBean",
			Name:     t.Name,
			NoOp:     false,
			Blocking: false,
		}
	default:
		return &RuleInfo{}
	}
}

// NewConfigChangeLogInf constructor
func NewConfigChangeLogInf() interface{} {
	return &ConfigChangeLog{}
}

func NewConfigChangeLog(convertedContext *ConvertedContext, explanation string, firmwareConfig *FirmwareConfigFacade, appliedFilters []interface{}, evaluatedRule *firmware.FirmwareRule, isLastLog bool) *ConfigChangeLog {
	var rule *RuleInfo
	if evaluatedRule != nil {
		rule = NewRuleInfo(evaluatedRule)
	}
	filters := []*RuleInfo{}
	for i := range appliedFilters {
		filters = append(filters, NewRuleInfo(appliedFilters[i]))
	}
	var updated int64
	if !isLastLog {
		updated = util.GetTimestamp(time.Now())
	}
	return &ConfigChangeLog{
		ID:             LAST_CONFIG_LOG_ID,
		Updated:        updated,
		Input:          convertedContext,
		Rule:           rule,
		Filters:        filters,
		Explanation:    explanation,
		FirmwareConfig: firmwareConfig,
	}
}

func GetLastConfigLog(tenantId string, mac string) *ConfigChangeLog {
	var lastConfigLog *ConfigChangeLog
	tableName := db.TABLE_CONFIG_CHANGE_LOGS
	if db.IsDualWriteEnabled() {
		// When dual write is enabled, read from old Logs2 table for backward compatibility,
		// until Logs2 table is fully migrated
		tableName = db.TABLE_LOGS
	}
	data, err := db.GetListingDao().GetOne(tenantId, tableName, mac, LAST_CONFIG_LOG_ID)
	if err == nil {
		config, ok := data.(*ConfigChangeLog)
		if ok {
			lastConfigLog = config
		}
	}
	return lastConfigLog
}

func GetConfigChangeLogsOnly(tenantId string, mac string) []*ConfigChangeLog {
	configChangeLogs := make([]*ConfigChangeLog, 0)
	tableName := db.TABLE_CONFIG_CHANGE_LOGS
	if db.IsDualWriteEnabled() {
		// When dual write is enabled, read from old Logs2 table for backward compatibility,
		// until Logs2 table is fully migrated
		tableName = db.TABLE_LOGS
	}
	data, err := db.GetListingDao().GetAll(tenantId, tableName, mac)
	if err == nil {
		configLogs := []*ConfigChangeLog{}
		for _, log := range data {
			configLog, ok := log.(*ConfigChangeLog)
			if ok && configLog.ID != LAST_CONFIG_LOG_ID {
				configLogs = append(configLogs, configLog)
			}
		}
		// sort by descending updated time
		sort.Slice(configLogs, func(i, j int) bool {
			return configLogs[i].Updated > configLogs[j].Updated
		})
		return configLogs
	}
	return configChangeLogs
}

func SetLastConfigLog(tenantId string, mac string, configChangeLog *ConfigChangeLog) error {
	jsonData, err := json.Marshal(configChangeLog)
	if err != nil {
		return err
	}
	if db.IsDualWriteEnabled() {
		// Write to Logs2 table for backward compatibility, but Logs2 will be eventually removed
		err = db.GetListingDao().SetOne(tenantId, db.TABLE_LOGS, mac, LAST_CONFIG_LOG_ID, []byte(jsonData), configChangeLog.Updated)
		if err != nil {
			return err
		}
	}
	return db.GetListingDao().SetOne(tenantId, db.TABLE_CONFIG_CHANGE_LOGS, mac, LAST_CONFIG_LOG_ID, []byte(jsonData), configChangeLog.Updated)
}

func SetConfigChangeLog(tenantId string, mac string, configChangeLog *ConfigChangeLog) error {
	logTables := []string{db.TABLE_CONFIG_CHANGE_LOGS}
	if db.IsDualWriteEnabled() {
		// Write to Logs2 table for backward compatibility, but Logs2 will be eventually removed
		logTables = append(logTables, db.TABLE_LOGS)
	}
	for _, tableName := range logTables {
		id, err := GetCurrentChangeLogId(tenantId, tableName, mac)
		if err == nil {
			configChangeLog.ID = id
			jsonData, err := json.Marshal(configChangeLog)
			if err == nil {
				err = db.GetListingDao().SetOne(tenantId, tableName, mac, id, []byte(jsonData), configChangeLog.Updated)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func GetCurrentChangeLogId(tenantId string, tableName string, mac string) (string, error) {
	// Get count from DB
	rangeInfo := &db.RangeInfo{
		StartValue: GetChangeLogColumnName(0),
		EndValue:   GetChangeLogColumnName(BOUNDS + 1),
	}
	data, err := db.GetListingDao().GetRange(tenantId, tableName, mac, rangeInfo)
	if err != nil {
		return "", err
	}
	var configLogs []*ConfigChangeLog
	for _, log := range data {
		configLog, ok := log.(*ConfigChangeLog)
		if ok {
			configLogs = append(configLogs, configLog)
		}
	}

	var count int = 1
	if len(configLogs) > 0 {
		// sort by descending updated time
		sort.Slice(configLogs, func(i, j int) bool {
			return configLogs[i].Updated > configLogs[j].Updated
		})
		prefixLength := len(fmt.Sprintf("%s_", prefix))
		suffix := configLogs[0].ID[prefixLength:len(configLogs[0].ID)]
		count, err = strconv.Atoi(suffix)
		if err != nil {
			return "", err
		}
	}
	if count == 1 {
		count = BOUNDS
	} else {
		count--
	}

	return GetChangeLogColumnName(count), nil
}

func GetChangeLogColumnName(number int) string {
	return fmt.Sprintf("%s_%d", prefix, number)
}
