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

	"xconfwebconfig/shared/firmware"
	"xconfwebconfig/util"

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
