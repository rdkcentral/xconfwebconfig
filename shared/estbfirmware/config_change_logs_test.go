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
	"fmt"
	"testing"
	"time"

	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rdkcentral/xconfwebconfig/util"

	"gotest.tools/assert"
)

// TODO create unit tests where data is found in db
// TODO create unit tests for setting data?

func TestNewRuleInfo(t *testing.T) {
	// FirmwareRule
	applicableAction := &firmware.ApplicableAction{
		ActionType: "BLOCKING_FILTER",
	}
	firmwareRule := &firmware.FirmwareRule{
		ID:               "id",
		Type:             "type",
		Name:             "name",
		ApplicableAction: applicableAction,
	}
	ruleInfo := NewRuleInfo(firmwareRule)
	assert.Equal(t, ruleInfo.ID, "id")
	assert.Equal(t, ruleInfo.Type, "type")
	assert.Equal(t, ruleInfo.Name, "name")
	assert.Equal(t, ruleInfo.NoOp, true)
	assert.Equal(t, ruleInfo.Blocking, true)

	// SingletonFilterValue
	singletonFilterValue := &SingletonFilterValue{
		ID: "id_VALUE",
	}
	ruleInfo = NewRuleInfo(singletonFilterValue)
	assert.Equal(t, ruleInfo.ID, "SINGLETON_id")
	assert.Equal(t, ruleInfo.Type, "SingletonFilter")
	assert.Equal(t, ruleInfo.Name, "id_VALUE")
	assert.Equal(t, ruleInfo.NoOp, true)
	assert.Equal(t, ruleInfo.Blocking, false)

	// RuleAction
	ruleAction := &firmware.RuleAction{}
	ruleInfo = NewRuleInfo(ruleAction)
	assert.Equal(t, ruleInfo.ID, "DistributionPercentInRuleAction")
	assert.Equal(t, ruleInfo.Type, "DistributionPercentInRuleAction")
	assert.Equal(t, ruleInfo.Name, "DistributionPercentInRuleAction")
	assert.Equal(t, ruleInfo.NoOp, false)
	assert.Equal(t, ruleInfo.Blocking, false)

	// PercentageBean
	percentageBean := &PercentageBean{
		Name: "name",
	}
	ruleInfo = NewRuleInfo(percentageBean)
	assert.Equal(t, ruleInfo.ID, "")
	assert.Equal(t, ruleInfo.Type, "PercentageBean")
	assert.Equal(t, ruleInfo.Name, "name")
	assert.Equal(t, ruleInfo.NoOp, false)
	assert.Equal(t, ruleInfo.Blocking, false)

	// other interface
	ruleInfo = NewRuleInfo("")
	assert.Equal(t, ruleInfo.ID, "")
	assert.Equal(t, ruleInfo.Type, "")
	assert.Equal(t, ruleInfo.Name, "")
	assert.Equal(t, ruleInfo.NoOp, false)
	assert.Equal(t, ruleInfo.Blocking, false)
}

func TestNewConfigChangeLog(t *testing.T) {
	// isLastLog = true
	contextMap := map[string]string{
		"eStbMac": "F4:4B:2A:14:17:D9",
	}
	convertedContext := GetContextConverted(contextMap)
	explanation := "explanation"
	firmwareConfig := &FirmwareConfigFacade{}
	appliedFilters := []interface{}{"appliedFilters"}
	evaluatedRule := &firmware.FirmwareRule{
		ID:   "id",
		Type: "type",
		Name: "name",
	}
	configChangeLog := NewConfigChangeLog(convertedContext, explanation, firmwareConfig, appliedFilters, evaluatedRule, true)
	assert.Equal(t, configChangeLog.ID, "0")
	assert.Equal(t, configChangeLog.Updated, int64(0))
	assert.Equal(t, configChangeLog.Input, convertedContext)
	assert.Equal(t, configChangeLog.Rule.ID, "id")
	assert.Equal(t, configChangeLog.Rule.Type, "type")
	assert.Equal(t, configChangeLog.Rule.Name, "name")
	assert.Equal(t, len(configChangeLog.Filters), 1)
	assert.Equal(t, configChangeLog.Explanation, explanation)
	assert.Equal(t, configChangeLog.FirmwareConfig, firmwareConfig)

	// isLastLog = false
	past := util.GetTimestamp(time.Now())
	configChangeLog = NewConfigChangeLog(convertedContext, explanation, firmwareConfig, appliedFilters, evaluatedRule, false)
	future := util.GetTimestamp(time.Now())
	fmt.Printf("Past: %d, Present: %d, Future: %d\n", past, configChangeLog.Updated, future)
	assert.Equal(t, configChangeLog.ID, "0")
	assert.Equal(t, configChangeLog.Updated >= past, true)
	assert.Equal(t, configChangeLog.Updated <= future, true)
	assert.Equal(t, configChangeLog.Input, convertedContext)
	assert.Equal(t, configChangeLog.Rule.ID, "id")
	assert.Equal(t, configChangeLog.Rule.Type, "type")
	assert.Equal(t, configChangeLog.Rule.Name, "name")
	assert.Equal(t, len(configChangeLog.Filters), 1)
	assert.Equal(t, configChangeLog.Explanation, explanation)
	assert.Equal(t, configChangeLog.FirmwareConfig, firmwareConfig)

}

func TestGetLastConfigLog(t *testing.T) {
	lastConfigLog := GetLastConfigLog("testMac")
	assert.Equal(t, lastConfigLog == nil, true)
}

func TestGetChangeLogsOnly(t *testing.T) {
	lastConfigLogs := GetConfigChangeLogsOnly("testMac")
	assert.Equal(t, len(lastConfigLogs), 0)
}

func TestGetCurrentId(t *testing.T) {
	id, err := GetCurrentId("testMac")
	assert.Equal(t, id, "")
	assert.Equal(t, err != nil, true)
}

func TestRuleInfo_Creation(t *testing.T) {
	ruleInfo := &RuleInfo{
		ID:       "rule-123",
		Type:     "IP_RULE",
		Name:     "Test Rule",
		NoOp:     false,
		Blocking: true,
	}
	
	assert.Equal(t, "rule-123", ruleInfo.ID)
	assert.Equal(t, "IP_RULE", ruleInfo.Type)
	assert.Equal(t, "Test Rule", ruleInfo.Name)
	assert.Assert(t, !ruleInfo.NoOp)
	assert.Assert(t, ruleInfo.Blocking)
}

func TestRuleInfo_Empty(t *testing.T) {
	ruleInfo := &RuleInfo{}
	
	assert.Equal(t, "", ruleInfo.ID)
	assert.Equal(t, "", ruleInfo.Type)
	assert.Equal(t, "", ruleInfo.Name)
	assert.Assert(t, !ruleInfo.NoOp)
	assert.Assert(t, !ruleInfo.Blocking)
}

func TestConfigChangeLog_Creation(t *testing.T) {
	contextMap := map[string]string{
		"eStbMac": "AA:BB:CC:DD:EE:FF",
		"model":   "MODEL-X",
	}
	convertedContext := GetContextConverted(contextMap)
	
	ruleInfo := &RuleInfo{
		ID:   "rule-1",
		Name: "Test Rule",
		Type: "MAC_RULE",
	}
	
	filters := []*RuleInfo{
		{ID: "filter-1", Name: "Filter 1"},
		{ID: "filter-2", Name: "Filter 2"},
	}
	
	log := &ConfigChangeLog{
		ID:           "log-123",
		Updated:      123456789,
		Input:        convertedContext,
		Rule:         ruleInfo,
		Filters:      filters,
		Explanation:  "Test explanation",
		HasMinimumFirmware: true,
	}
	
	assert.Equal(t, "log-123", log.ID)
	assert.Equal(t, int64(123456789), log.Updated)
	assert.Assert(t, log.Input != nil)
	assert.Assert(t, log.Rule != nil)
	assert.Equal(t, 2, len(log.Filters))
	assert.Equal(t, "Test explanation", log.Explanation)
	assert.Assert(t, log.HasMinimumFirmware)
}

func TestConfigChangeLog_WithFirmwareConfig(t *testing.T) {
	facade := &FirmwareConfigFacade{
		Properties: map[string]interface{}{
			"firmwareVersion": "1.0.0",
		},
	}
	
	log := &ConfigChangeLog{
		ID:             "log-456",
		FirmwareConfig: facade,
	}
	
	assert.Assert(t, log.FirmwareConfig != nil)
	assert.Equal(t, 1, len(log.FirmwareConfig.Properties))
}

func TestNewRuleInfo_WithBlockingFilter(t *testing.T) {
	applicableAction := &firmware.ApplicableAction{
		ActionType: firmware.BLOCKING_FILTER,
	}
	firmwareRule := &firmware.FirmwareRule{
		ID:               "blocking-rule",
		Type:             "IP_RULE",
		Name:             "Blocking Rule",
		ApplicableAction: applicableAction,
	}
	
	ruleInfo := NewRuleInfo(firmwareRule)
	
	assert.Equal(t, "blocking-rule", ruleInfo.ID)
	assert.Equal(t, "IP_RULE", ruleInfo.Type)
	assert.Equal(t, "Blocking Rule", ruleInfo.Name)
	assert.Assert(t, ruleInfo.Blocking)
}

func TestNewRuleInfo_WithoutBlockingFilter(t *testing.T) {
	applicableAction := &firmware.ApplicableAction{
		ActionType: "RULE",
	}
	firmwareRule := &firmware.FirmwareRule{
		ID:               "non-blocking-rule",
		Type:             "MAC_RULE",
		Name:             "Non-Blocking Rule",
		ApplicableAction: applicableAction,
	}
	
	ruleInfo := NewRuleInfo(firmwareRule)
	
	assert.Equal(t, "non-blocking-rule", ruleInfo.ID)
	assert.Equal(t, "MAC_RULE", ruleInfo.Type)
	assert.Equal(t, "Non-Blocking Rule", ruleInfo.Name)
	assert.Assert(t, !ruleInfo.Blocking)
}

func TestNewRuleInfo_PercentageBean(t *testing.T) {
	bean := &PercentageBean{
		ID:   "bean-123",
		Name: "Test Percentage Bean",
	}
	
	ruleInfo := NewRuleInfo(bean)
	
	assert.Equal(t, "", ruleInfo.ID)
	assert.Equal(t, "PercentageBean", ruleInfo.Type)
	assert.Equal(t, "Test Percentage Bean", ruleInfo.Name)
	assert.Assert(t, !ruleInfo.NoOp)
	assert.Assert(t, !ruleInfo.Blocking)
}

func TestNewRuleInfo_SingletonFilterValue(t *testing.T) {
	sfv := &SingletonFilterValue{
		ID: "PERCENT_FILTER_VALUE",
	}
	
	ruleInfo := NewRuleInfo(sfv)
	
	assert.Equal(t, "SINGLETON_PERCENT_FILTER", ruleInfo.ID)
	assert.Equal(t, "SingletonFilter", ruleInfo.Type)
	assert.Equal(t, "PERCENT_FILTER_VALUE", ruleInfo.Name)
	assert.Assert(t, ruleInfo.NoOp)
	assert.Assert(t, !ruleInfo.Blocking)
}

func TestNewRuleInfo_RuleAction(t *testing.T) {
	ruleAction := &firmware.RuleAction{
		ConfigId: "config-123",
	}
	
	ruleInfo := NewRuleInfo(ruleAction)
	
	assert.Equal(t, "DistributionPercentInRuleAction", ruleInfo.ID)
	assert.Equal(t, "DistributionPercentInRuleAction", ruleInfo.Type)
	assert.Equal(t, "DistributionPercentInRuleAction", ruleInfo.Name)
	assert.Assert(t, !ruleInfo.NoOp)
	assert.Assert(t, !ruleInfo.Blocking)
}

