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
	"testing"

	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"

	"gotest.tools/assert"
)

func TestConvertFirmwareRuleToIpRuleBean(t *testing.T) {
	// create rule like the following
	ruleStr := []byte(`{
		"id": "2123455666",
		"name": "000ipPerformanceTestRule",
		"rule": {
		  "negated": false,
		  "compoundParts": [
			{
			  "negated": false,
			  "condition": {
				"freeArg": {
				  "type": "STRING",
				  "name": "ipAddress"
				},
				"operation": "IN_LIST",
				"fixedArg": {
				  "bean": {
					"value": {
					  "java.lang.String": "%v"
					}
				  }
				}
			  }
			}
		  ],
		  "condition": {
				"freeArg": {
				  "type": "STRING",
				  "name": "ipAddress"
				},
				"operation": "IN_LIST",
				"fixedArg": {
				  "bean": {
					"value": {
					  "java.lang.String": "127.0.0.1"
					}
				  }
				}
			}
		},
		"applicableAction": {
		  "type": ".RuleAction",
		  "actionType": "RULE",
		  "configId": "234567",
		  "configEntries": [],
		  "active": true,
		  "useAccountPercentage": false,
		  "firmwareCheckRequired": false,
		  "rebootImmediately": false,
		  "properties": {
			  "firmwareLocation": "http://127.0.1.1/app/download",
			  "ipv6FirmwareLocation": "http://127.0.1.1/app/downloadv6",
			  "irmwareDownloadProtocol": "https"
		  }
		},
		"type": "IP_RULE",
		"active": true,
		"applicationType": "stb"
	  }`)
	var rule firmware.FirmwareRule
	err := json.Unmarshal(ruleStr, &rule)
	assert.NilError(t, err)
	bean := ConvertFirmwareRuleToIpRuleBean(&rule)
	assert.Assert(t, bean != nil)
	assert.Assert(t, bean.IpAddressGroup != nil)
}

func TestIsLegacyIpFreeArg(t *testing.T) {
	// Test with legacy IP free arg
	freeArg := &re.FreeArg{}
	freeArg.SetType(re.AuxFreeArgTypeIpAddress)
	freeArg.SetName("ipAddress")
	
	result := IsLegacyIpFreeArg(freeArg)
	assert.Assert(t, result)
}

func TestIsLegacyIpFreeArg_NotLegacy(t *testing.T) {
	// Test with non-legacy free arg (different type)
	freeArg := &re.FreeArg{}
	freeArg.SetType(re.AuxFreeArgTypeMacAddress)
	freeArg.SetName("ipAddress")
	
	result := IsLegacyIpFreeArg(freeArg)
	assert.Assert(t, !result)
	
	// Test with wrong name
	freeArg2 := &re.FreeArg{}
	freeArg2.SetType(re.AuxFreeArgTypeIpAddress)
	freeArg2.SetName("model")
	
	result2 := IsLegacyIpFreeArg(freeArg2)
	assert.Assert(t, !result2)
}

func TestIsLegacyMacFreeArg(t *testing.T) {
	// Test with legacy MAC free arg
	freeArg := re.FreeArg{}
	freeArg.SetType(re.AuxFreeArgTypeMacAddress)
	freeArg.SetName("eStbMac")
	
	result := IsLegacyMacFreeArg(freeArg)
	assert.Assert(t, result)
}

func TestIsLegacyMacFreeArg_NotLegacy(t *testing.T) {
	// Test with non-legacy free arg (different type)
	freeArg := re.FreeArg{}
	freeArg.SetType(re.AuxFreeArgTypeIpAddress)
	freeArg.SetName("eStbMac")
	
	result := IsLegacyMacFreeArg(freeArg)
	assert.Assert(t, !result)
	
	// Test with wrong name
	freeArg2 := re.FreeArg{}
	freeArg2.SetType(re.AuxFreeArgTypeMacAddress)
	freeArg2.SetName("model")
	
	result2 := IsLegacyMacFreeArg(freeArg2)
	assert.Assert(t, !result2)
}

func TestIsLegacyLocalTimeFreeArg(t *testing.T) {
	// Test with legacy local time free arg
	freeArg := re.FreeArg{}
	freeArg.SetType(re.AuxFreeArgTypeTime)
	freeArg.SetName("time")
	
	result := IsLegacyLocalTimeFreeArg(freeArg)
	assert.Assert(t, result)
}

func TestIsLegacyLocalTimeFreeArg_NotLegacy(t *testing.T) {
	// Test with non-legacy free arg (different type)
	freeArg := re.FreeArg{}
	freeArg.SetType(re.AuxFreeArgTypeIpAddress)
	freeArg.SetName("time")
	
	result := IsLegacyLocalTimeFreeArg(freeArg)
	assert.Assert(t, !result)
	
	// Test with wrong name
	freeArg2 := re.FreeArg{}
	freeArg2.SetType(re.AuxFreeArgTypeTime)
	freeArg2.SetName("model")
	
	result2 := IsLegacyLocalTimeFreeArg(freeArg2)
	assert.Assert(t, !result2)
}

func TestConvertIpRuleBeanToFirmwareRule(t *testing.T) {
	// Test basic conversion
	bean := &IpRuleBean{
		Id:   "test-id-123",
		Name: "test-rule",
	}
	bean.EnvironmentId = "prod"
	bean.ModelId = "MODEL-X1"
	bean.IpAddressGroup = shared.NewEmptyIpAddressGroup()
	bean.IpAddressGroup.Name = "test-ip-group"
	
	config := &FirmwareConfig{}
	config.ID = "config-123"
	bean.FirmwareConfig = config
	
	rule := ConvertIpRuleBeanToFirmwareRule(bean)
	
	assert.Assert(t, rule != nil)
	assert.Equal(t, "test-id-123", rule.ID)
	assert.Equal(t, "test-rule", rule.Name)
	assert.Equal(t, IP_RULE, rule.Type)
	assert.Assert(t, rule.ApplicableAction != nil)
	assert.Equal(t, firmware.RULE, rule.ApplicableAction.ActionType)
	assert.Equal(t, "config-123", rule.ApplicableAction.ConfigId)
}

func TestConvertIpRuleBeanToFirmwareRule_NoConfig(t *testing.T) {
	// Test conversion without firmware config
	bean := &IpRuleBean{
		Id:   "test-id-456",
		Name: "test-rule-no-config",
	}
	bean.EnvironmentId = "dev"
	bean.ModelId = "MODEL-Y2"
	bean.IpAddressGroup = shared.NewEmptyIpAddressGroup()
	bean.IpAddressGroup.Name = "test-ip-group-2"
	
	rule := ConvertIpRuleBeanToFirmwareRule(bean)
	
	assert.Assert(t, rule != nil)
	assert.Equal(t, "test-id-456", rule.ID)
	assert.Equal(t, IP_RULE, rule.Type)
	assert.Assert(t, rule.ApplicableAction != nil)
	assert.Equal(t, "", rule.ApplicableAction.ConfigId)
}

func TestConvertPercentageBeanToFirmwareRule(t *testing.T) {
	// Test basic conversion with environment and model
	bean := PercentageBean{
		Environment: "prod",
		Model:       "MODEL-A1",
	}
	
	distribution1 := &firmware.ConfigEntry{
		ConfigId:   "config-1",
		Percentage: 50.0,
	}
	distribution2 := &firmware.ConfigEntry{
		ConfigId:   "config-2",
		Percentage: 50.0,
	}
	bean.Distributions = []*firmware.ConfigEntry{distribution1, distribution2}
	
	rule := ConvertPercentageBeanToFirmwareRule(bean)
	
	assert.Assert(t, rule != nil)
	assert.Equal(t, ENV_MODEL_RULE, rule.Type)
	assert.Assert(t, rule.ApplicableAction != nil)
}

func TestConvertPercentageBeanToFirmwareRule_ModelOnly(t *testing.T) {
	// Test conversion with only model (no environment)
	bean := PercentageBean{
		Model: "MODEL-B2",
	}
	
	distribution := &firmware.ConfigEntry{
		ConfigId:   "config-xyz",
		Percentage: 100.0,
	}
	bean.Distributions = []*firmware.ConfigEntry{distribution}
	
	rule := ConvertPercentageBeanToFirmwareRule(bean)
	
	assert.Assert(t, rule != nil)
	assert.Equal(t, ENV_MODEL_RULE, rule.Type)
	assert.Assert(t, rule.ApplicableAction != nil)
}

func TestConvertPercentageBeanToFirmwareRule_WithOptionalConditions(t *testing.T) {
	// Test conversion with optional conditions
	bean := PercentageBean{
		Environment: "staging",
		Model:       "MODEL-C3",
	}
	
	// Create optional conditions
	condition := &re.Condition{}
	freeArg := re.NewFreeArg(re.StandardFreeArgTypeString, "firmwareVersion")
	fixedArg := re.NewFixedArg("2.0")
	condition.SetFreeArg(freeArg)
	condition.SetOperation(re.StandardOperationIs)
	condition.SetFixedArg(fixedArg)
	
	optionalConditions := &re.Rule{}
	optionalConditions.SetCondition(condition)
	bean.OptionalConditions = optionalConditions
	
	distribution := &firmware.ConfigEntry{
		ConfigId:   "config-abc",
		Percentage: 100.0,
	}
	bean.Distributions = []*firmware.ConfigEntry{distribution}
	
	rule := ConvertPercentageBeanToFirmwareRule(bean)
	
	assert.Assert(t, rule != nil)
	assert.Equal(t, ENV_MODEL_RULE, rule.Type)
}

func TestConvertIntoPercentRange(t *testing.T) {
	// Test with basic config entries
	entries := []firmware.ConfigEntry{
		{
			ConfigId:   "config-1",
			Percentage: 30.0,
		},
		{
			ConfigId:   "config-2",
			Percentage: 70.0,
		},
	}
	
	result := ConvertIntoPercentRange(entries)
	
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "config-1", result[0].ConfigId)
	assert.Equal(t, 30.0, result[0].Percentage)
	assert.Equal(t, 0.0, result[0].StartPercentRange)
	assert.Equal(t, 30.0, result[0].EndPercentRange)
	
	assert.Equal(t, "config-2", result[1].ConfigId)
	assert.Equal(t, 70.0, result[1].Percentage)
	assert.Equal(t, 30.0, result[1].StartPercentRange)
	assert.Equal(t, 100.0, result[1].EndPercentRange)
}

func TestConvertIntoPercentRange_WithExistingRanges(t *testing.T) {
	// Test with pre-defined ranges
	entries := []firmware.ConfigEntry{
		{
			ConfigId:          "config-1",
			Percentage:        50.0,
			StartPercentRange: 0.0,
			EndPercentRange:   50.0,
		},
		{
			ConfigId:          "config-2",
			Percentage:        50.0,
			StartPercentRange: 50.0,
			EndPercentRange:   100.0,
		},
	}
	
	result := ConvertIntoPercentRange(entries)
	
	assert.Equal(t, 2, len(result))
	// Should preserve existing ranges
	assert.Equal(t, 0.0, result[0].StartPercentRange)
	assert.Equal(t, 50.0, result[0].EndPercentRange)
	assert.Equal(t, 50.0, result[1].StartPercentRange)
	assert.Equal(t, 100.0, result[1].EndPercentRange)
}

func TestConvertIntoPercentRange_WithFlags(t *testing.T) {
	// Test with canary and pause flags
	entries := []firmware.ConfigEntry{
		{
			ConfigId:         "config-1",
			Percentage:       100.0,
			IsCanaryDisabled: true,
			IsPaused:         true,
		},
	}
	
	result := ConvertIntoPercentRange(entries)
	
	assert.Equal(t, 1, len(result))
	assert.Assert(t, result[0].IsCanaryDisabled)
	assert.Assert(t, result[0].IsPaused)
}

func TestConvertIntoPercentRange_Empty(t *testing.T) {
	// Test with empty slice
	entries := []firmware.ConfigEntry{}
	
	result := ConvertIntoPercentRange(entries)
	
	assert.Equal(t, 0, len(result))
}

func TestGetWhitelistName(t *testing.T) {
	// Test with valid IP address group
	ipGroup := &shared.IpAddressGroup{
		Id:   "whitelist-123",
		Name: "Production Whitelist",
	}
	
	name := GetWhitelistName(ipGroup)
	
	assert.Equal(t, "Production Whitelist", name)
}

func TestGetWhitelistName_Nil(t *testing.T) {
	// Test with nil
	name := GetWhitelistName(nil)
	
	assert.Equal(t, "", name)
}

func TestGetWhitelistName_EmptyName(t *testing.T) {
	// Test with empty name
	ipGroup := &shared.IpAddressGroup{
		Id: "whitelist-456",
	}
	
	name := GetWhitelistName(ipGroup)
	
	assert.Equal(t, "", name)
}

