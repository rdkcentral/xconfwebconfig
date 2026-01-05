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
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	"gotest.tools/assert"
)

func TestTimeFilter_Creation(t *testing.T) {
	filter := &TimeFilter{
		Id:    "filter-1",
		Name:  "Test Time Filter",
		Start: "08:00",
		End:   "18:00",
	}
	
	assert.Equal(t, "filter-1", filter.Id)
	assert.Equal(t, "Test Time Filter", filter.Name)
	assert.Equal(t, "08:00", filter.Start)
	assert.Equal(t, "18:00", filter.End)
	assert.Assert(t, !filter.NeverBlockRebootDecoupled)
	assert.Assert(t, !filter.NeverBlockHttpDownload)
	assert.Assert(t, !filter.LocalTime)
}

func TestTimeFilter_WithIpWhiteList(t *testing.T) {
	ipGroup := &shared.IpAddressGroup{
		Id:   "ip-group-1",
		Name: "Whitelist Group",
	}
	
	filter := &TimeFilter{
		Id:          "filter-2",
		Name:        "Filter with IP Whitelist",
		IpWhiteList: ipGroup,
	}
	
	assert.Assert(t, filter.IpWhiteList != nil)
	assert.Equal(t, "ip-group-1", filter.IpWhiteList.Id)
	assert.Equal(t, "Whitelist Group", filter.IpWhiteList.Name)
}

func TestTimeFilter_WithEnvModelRuleBean(t *testing.T) {
	envModel := EnvModelRuleBean{
		Id:            "rule-1",
		Name:          "Env Model Rule",
		EnvironmentId: "PROD",
		ModelId:       "MODEL-X1",
	}
	
	filter := &TimeFilter{
		Id:               "filter-3",
		Name:             "Filter with Env Model",
		EnvModelRuleBean: envModel,
	}
	
	assert.Equal(t, "rule-1", filter.EnvModelRuleBean.Id)
	assert.Equal(t, "Env Model Rule", filter.EnvModelRuleBean.Name)
	assert.Equal(t, "PROD", filter.EnvModelRuleBean.EnvironmentId)
	assert.Equal(t, "MODEL-X1", filter.EnvModelRuleBean.ModelId)
}

func TestTimeFilter_WithFirmwareConfig(t *testing.T) {
	firmwareConfig := &FirmwareConfig{
		ID:          "config-1",
		Description: "Test Config",
	}
	
	envModel := EnvModelRuleBean{
		FirmwareConfig: firmwareConfig,
	}
	
	filter := &TimeFilter{
		Id:               "filter-4",
		Name:             "Filter with Firmware Config",
		EnvModelRuleBean: envModel,
	}
	
	assert.Assert(t, filter.EnvModelRuleBean.FirmwareConfig != nil)
	assert.Equal(t, "config-1", filter.EnvModelRuleBean.FirmwareConfig.ID)
	assert.Equal(t, "Test Config", filter.EnvModelRuleBean.FirmwareConfig.Description)
}

func TestTimeFilter_BlockingFlags(t *testing.T) {
	filter := &TimeFilter{
		Id:                        "filter-5",
		Name:                      "Filter with Blocking Flags",
		NeverBlockRebootDecoupled: true,
		NeverBlockHttpDownload:    true,
	}
	
	assert.Assert(t, filter.NeverBlockRebootDecoupled)
	assert.Assert(t, filter.NeverBlockHttpDownload)
}

func TestTimeFilter_LocalTime(t *testing.T) {
	filter := &TimeFilter{
		Id:        "filter-6",
		Name:      "Filter with Local Time",
		Start:     "09:00",
		End:       "17:00",
		LocalTime: true,
	}
	
	assert.Assert(t, filter.LocalTime)
	assert.Equal(t, "09:00", filter.Start)
	assert.Equal(t, "17:00", filter.End)
}

func TestTimeFilter_CompleteFilter(t *testing.T) {
	ipGroup := &shared.IpAddressGroup{Id: "ip-1", Name: "IP Group"}
	firmwareConfig := &FirmwareConfig{ID: "fw-1", Description: "Firmware"}
	
	filter := &TimeFilter{
		Id:          "complete-filter",
		Name:        "Complete Time Filter",
		IpWhiteList: ipGroup,
		EnvModelRuleBean: EnvModelRuleBean{
			Id:             "env-model-1",
			Name:           "Env Model",
			EnvironmentId:  "QA",
			ModelId:        "MODEL-Y",
			FirmwareConfig: firmwareConfig,
		},
		NeverBlockRebootDecoupled: true,
		NeverBlockHttpDownload:    false,
		Start:                     "06:00",
		End:                       "22:00",
		LocalTime:                 true,
	}
	
	// Verify all fields
	assert.Equal(t, "complete-filter", filter.Id)
	assert.Equal(t, "Complete Time Filter", filter.Name)
	assert.Assert(t, filter.IpWhiteList != nil)
	assert.Equal(t, "ip-1", filter.IpWhiteList.Id)
	assert.Equal(t, "env-model-1", filter.EnvModelRuleBean.Id)
	assert.Equal(t, "QA", filter.EnvModelRuleBean.EnvironmentId)
	assert.Equal(t, "MODEL-Y", filter.EnvModelRuleBean.ModelId)
	assert.Assert(t, filter.EnvModelRuleBean.FirmwareConfig != nil)
	assert.Assert(t, filter.NeverBlockRebootDecoupled)
	assert.Assert(t, !filter.NeverBlockHttpDownload)
	assert.Equal(t, "06:00", filter.Start)
	assert.Equal(t, "22:00", filter.End)
	assert.Assert(t, filter.LocalTime)
}

func TestParseStringTime_Standard(t *testing.T) {
	result := parseStringTime("'08:30:00'")
	assert.Equal(t, "08:30", result)
	
	result2 := parseStringTime("'23:45:59'")
	assert.Equal(t, "23:45", result2)
	
	result3 := parseStringTime("'00:00:00'")
	assert.Equal(t, "00:00", result3)
}

func TestParseStringTime_WithoutQuotes(t *testing.T) {
	result := parseStringTime("12:30:00")
	assert.Equal(t, "12:30", result)
	
	result2 := parseStringTime("06:15:45")
	assert.Equal(t, "06:15", result2)
}

func TestParseStringTime_EdgeCases(t *testing.T) {
	// Single quotes on one side
	result := parseStringTime("'14:20:30")
	assert.Equal(t, "14:20", result)
	
	result2 := parseStringTime("16:40:50'")
	assert.Equal(t, "16:40", result2)
}

func TestTrimSingleQuote_WithQuotes(t *testing.T) {
	result := trimSingleQuote("'PROD'")
	assert.Equal(t, "PROD", result)
	
	result2 := trimSingleQuote("'MODEL-X1'")
	assert.Equal(t, "MODEL-X1", result2)
}

func TestTrimSingleQuote_WithoutQuotes(t *testing.T) {
	result := trimSingleQuote("DEV")
	assert.Equal(t, "DEV", result)
	
	result2 := trimSingleQuote("MODEL-Y2")
	assert.Equal(t, "MODEL-Y2", result2)
}

func TestTrimSingleQuote_MultipleQuotes(t *testing.T) {
	result := trimSingleQuote("'te'st'")
	assert.Equal(t, "test", result)
	
	result2 := trimSingleQuote("'a'b'c'")
	assert.Equal(t, "abc", result2)
}

func TestTrimSingleQuote_EmptyString(t *testing.T) {
	result := trimSingleQuote("")
	assert.Equal(t, "", result)
	
	result2 := trimSingleQuote("''")
	assert.Equal(t, "", result2)
}

func TestEnvModelRuleBean_EmptyCreation(t *testing.T) {
	bean := EnvModelRuleBean{}
	
	assert.Equal(t, "", bean.Id)
	assert.Equal(t, "", bean.Name)
	assert.Equal(t, "", bean.EnvironmentId)
	assert.Equal(t, "", bean.ModelId)
	assert.Assert(t, bean.FirmwareConfig == nil)
}

func TestEnvModelRuleBean_WithAllFields(t *testing.T) {
	config := &FirmwareConfig{
		ID:          "fw-config-1",
		Description: "Firmware Configuration",
	}
	
	bean := EnvModelRuleBean{
		Id:             "bean-1",
		Name:           "Bean Name",
		EnvironmentId:  "STAGE",
		ModelId:        "MODEL-Z",
		FirmwareConfig: config,
	}
	
	assert.Equal(t, "bean-1", bean.Id)
	assert.Equal(t, "Bean Name", bean.Name)
	assert.Equal(t, "STAGE", bean.EnvironmentId)
	assert.Equal(t, "MODEL-Z", bean.ModelId)
	assert.Assert(t, bean.FirmwareConfig != nil)
	assert.Equal(t, "fw-config-1", bean.FirmwareConfig.ID)
}
