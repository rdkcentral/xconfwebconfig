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

	"xconfwebconfig/db"
	"xconfwebconfig/shared/firmware"
	"xconfwebconfig/util"
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

// ConfigChangeLog Logs2 table
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

func GetLastConfigLog(mac string) *ConfigChangeLog {
	var lastConfigLog *ConfigChangeLog
	data, err := db.GetListingDao().GetOne(db.TABLE_LOGS, mac, LAST_CONFIG_LOG_ID)
	if err == nil {
		config, ok := data.(*ConfigChangeLog)
		if ok {
			lastConfigLog = config
		}
	}
	return lastConfigLog
}

func GetConfigChangeLogsOnly(mac string) []*ConfigChangeLog {
	configChangeLogs := make([]*ConfigChangeLog, 0)
	data, err := db.GetListingDao().GetAll(db.TABLE_LOGS, mac)
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

func SetLastConfigLog(mac string, configChangeLog *ConfigChangeLog) error {
	jsonData, err := json.Marshal(configChangeLog)
	if err != nil {
		return err
	}
	return db.GetListingDao().SetOne(db.TABLE_LOGS, mac, LAST_CONFIG_LOG_ID, []byte(jsonData))
}

func SetConfigChangeLog(mac string, configChangeLog *ConfigChangeLog) error {
	id, err := GetCurrentId(mac)
	if err == nil {
		configChangeLog.ID = id
		jsonData, err := json.Marshal(configChangeLog)
		if err == nil {
			return db.GetListingDao().SetOne(db.TABLE_LOGS, mac, id, []byte(jsonData))
		}
	}
	return err
}

func GetCurrentId(mac string) (string, error) {
	// Get count from DB
	rangeInfo := &db.RangeInfo{
		StartValue: numberToColumnName(0),
		EndValue:   numberToColumnName(BOUNDS + 1),
	}
	data, err := db.GetListingDao().GetRange(db.TABLE_LOGS, mac, rangeInfo)
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

	return numberToColumnName(count), nil
}

func numberToColumnName(number int) string {
	return fmt.Sprintf("%s_%d", prefix, number)
}
