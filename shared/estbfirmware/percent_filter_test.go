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

func TestNewEnvModelPercentage(t *testing.T) {
	emp := NewEnvModelPercentage()
	
	assert.Assert(t, emp != nil)
	assert.Assert(t, !emp.RebootImmediately)
	assert.Assert(t, !emp.Active)
	assert.Assert(t, !emp.FirmwareCheckRequired)
	assert.Equal(t, float32(0), emp.Percentage)
}

func TestNewEmptyPercentFilterValue(t *testing.T) {
	pfv := NewEmptyPercentFilterValue()
	
	assert.Assert(t, pfv != nil)
	assert.Equal(t, PERCENT_FILTER_SINGLETON_ID, pfv.ID)
	assert.Equal(t, PercentFilterClass, pfv.Type)
	assert.Equal(t, float32(100.0), pfv.Percentage)
	assert.Equal(t, 100, pfv.Percent)
	assert.Assert(t, pfv.EnvModelPercentages != nil)
	assert.Equal(t, 0, len(pfv.EnvModelPercentages))
}

func TestNewPercentFilterValue(t *testing.T) {
	whitelist := &shared.IpAddressGroup{Id: "wl-1", Name: "Whitelist"}
	envModelPercentages := map[string]EnvModelPercentage{
		"key1": {Percentage: 50.0, Active: true},
	}
	
	pfv := NewPercentFilterValue(whitelist, 75.5, envModelPercentages)
	
	assert.Assert(t, pfv != nil)
	assert.Equal(t, PERCENT_FILTER_SINGLETON_ID, pfv.ID)
	assert.Equal(t, PercentFilterClass, pfv.Type)
	assert.Equal(t, float32(75.5), pfv.Percentage)
	assert.Assert(t, pfv.Whitelist != nil)
	assert.Equal(t, "wl-1", pfv.Whitelist.Id)
	assert.Equal(t, 1, len(pfv.EnvModelPercentages))
}

func TestPercentFilterValue_SetId_Valid(t *testing.T) {
	pfv := NewEmptyPercentFilterValue()
	
	err := pfv.SetId(PERCENT_FILTER_SINGLETON_ID)
	assert.NilError(t, err)
	assert.Equal(t, PERCENT_FILTER_SINGLETON_ID, pfv.ID)
}

func TestPercentFilterValue_SetId_Invalid(t *testing.T) {
	pfv := NewEmptyPercentFilterValue()
	
	err := pfv.SetId("INVALID_ID")
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "PercentFilterValue id is PERCENT_FILTER_VALUE")
}

func TestPercentFilterValue_GetId(t *testing.T) {
	pfv := NewEmptyPercentFilterValue()
	
	id := pfv.GetId()
	assert.Equal(t, PERCENT_FILTER_SINGLETON_ID, id)
}

func TestNewGlobalPercentage(t *testing.T) {
	gp := NewGlobalPercentage()
	
	assert.Assert(t, gp != nil)
	assert.Equal(t, float32(100.0), gp.Percentage)
	assert.Equal(t, shared.STB, gp.ApplicationType)
	assert.Equal(t, "", gp.Whitelist)
}

func TestGlobalPercentage_WithFields(t *testing.T) {
	gp := &GlobalPercentage{
		Whitelist:       "whitelist-1",
		Percentage:      50.0,
		ApplicationType: "xhome",
	}
	
	assert.Equal(t, "whitelist-1", gp.Whitelist)
	assert.Equal(t, float32(50.0), gp.Percentage)
	assert.Equal(t, "xhome", gp.ApplicationType)
}

func TestNewDefaultPercentFilterVo(t *testing.T) {
	pfvo := NewDefaultPercentFilterVo()
	
	assert.Assert(t, pfvo != nil)
	assert.Assert(t, pfvo.PercentageBeans != nil)
	assert.Equal(t, 0, len(pfvo.PercentageBeans))
	assert.Assert(t, pfvo.GlobalPercentage == nil)
}

func TestNewPercentFilterVo(t *testing.T) {
	gp := &GlobalPercentage{Percentage: 80.0}
	beans := []PercentageBean{
		{ID: "bean1", Active: true},
		{ID: "bean2", Active: false},
	}
	
	pfvo := NewPercentFilterVo(gp, beans)
	
	assert.Assert(t, pfvo != nil)
	assert.Assert(t, pfvo.GlobalPercentage != nil)
	assert.Equal(t, float32(80.0), pfvo.GlobalPercentage.Percentage)
	assert.Equal(t, 2, len(pfvo.PercentageBeans))
	assert.Equal(t, "bean1", pfvo.PercentageBeans[0].ID)
	assert.Equal(t, "bean2", pfvo.PercentageBeans[1].ID)
}

func TestPercentageBean_Creation(t *testing.T) {
	bean := PercentageBean{
		ID:                    "bean-1",
		Name:                  "Test Bean",
		Active:                true,
		FirmwareCheckRequired: false,
		RebootImmediately:     true,
	}
	
	assert.Equal(t, "bean-1", bean.ID)
	assert.Equal(t, "Test Bean", bean.Name)
	assert.Assert(t, bean.Active)
	assert.Assert(t, !bean.FirmwareCheckRequired)
	assert.Assert(t, bean.RebootImmediately)
}

func TestPercentageBean_WithVersions(t *testing.T) {
	bean := PercentageBean{
		ID:                  "bean-2",
		LastKnownGood:       "1.0.0",
		IntermediateVersion: "1.5.0",
		FirmwareVersions:    []string{"2.0.0", "2.1.0"},
	}
	
	assert.Equal(t, "1.0.0", bean.LastKnownGood)
	assert.Equal(t, "1.5.0", bean.IntermediateVersion)
	assert.Equal(t, 2, len(bean.FirmwareVersions))
	assert.Equal(t, "2.0.0", bean.FirmwareVersions[0])
	assert.Equal(t, "2.1.0", bean.FirmwareVersions[1])
}

func TestPercentageBean_WithEnvironmentModel(t *testing.T) {
	bean := PercentageBean{
		ID:              "bean-3",
		Environment:     "PROD",
		Model:           "MODEL-X",
		ApplicationType: "stb",
	}
	
	assert.Equal(t, "PROD", bean.Environment)
	assert.Equal(t, "MODEL-X", bean.Model)
	assert.Equal(t, "stb", bean.ApplicationType)
}

func TestPercentageBean_WithWhitelist(t *testing.T) {
	bean := PercentageBean{
		ID:        "bean-4",
		Whitelist: "whitelist-group-1",
	}
	
	assert.Equal(t, "whitelist-group-1", bean.Whitelist)
}

func TestPercentageBean_UseAccountIdPercentage(t *testing.T) {
	bean := PercentageBean{
		ID:                     "bean-5",
		UseAccountIdPercentage: true,
	}
	
	assert.Assert(t, bean.UseAccountIdPercentage)
}

func TestEnvModelPercentage_Creation(t *testing.T) {
	emp := EnvModelPercentage{
		Percentage:            25.5,
		Active:                true,
		FirmwareCheckRequired: true,
		RebootImmediately:     false,
		Name:                  "Test Env Model",
	}
	
	assert.Equal(t, float32(25.5), emp.Percentage)
	assert.Assert(t, emp.Active)
	assert.Assert(t, emp.FirmwareCheckRequired)
	assert.Assert(t, !emp.RebootImmediately)
	assert.Equal(t, "Test Env Model", emp.Name)
}

func TestEnvModelPercentage_WithVersions(t *testing.T) {
	emp := EnvModelPercentage{
		LastKnownGood:       "3.0.0",
		IntermediateVersion: "3.5.0",
		FirmwareVersions:    []string{"4.0.0", "4.1.0", "4.2.0"},
	}
	
	assert.Equal(t, "3.0.0", emp.LastKnownGood)
	assert.Equal(t, "3.5.0", emp.IntermediateVersion)
	assert.Equal(t, 3, len(emp.FirmwareVersions))
	assert.Equal(t, "4.0.0", emp.FirmwareVersions[0])
	assert.Equal(t, "4.2.0", emp.FirmwareVersions[2])
}

func TestEnvModelPercentage_WithWhitelist(t *testing.T) {
	whitelist := &shared.IpAddressGroup{Id: "wl-group", Name: "Whitelist Group"}
	emp := EnvModelPercentage{
		Whitelist: whitelist,
		Active:    true,
	}
	
	assert.Assert(t, emp.Whitelist != nil)
	assert.Equal(t, "wl-group", emp.Whitelist.Id)
	assert.Equal(t, "Whitelist Group", emp.Whitelist.Name)
}

func TestPercentFilterValue_WithEnvModelPercentages(t *testing.T) {
	envModelPercentages := map[string]EnvModelPercentage{
		"prod-model1": {
			Percentage:            30.0,
			Active:                true,
			FirmwareCheckRequired: true,
			Name:                  "Production Model 1",
		},
		"qa-model2": {
			Percentage:        70.0,
			Active:            false,
			RebootImmediately: true,
			Name:              "QA Model 2",
		},
	}
	
	pfv := &PercentFilterValue{
		ID:                  PERCENT_FILTER_SINGLETON_ID,
		EnvModelPercentages: envModelPercentages,
	}
	
	assert.Equal(t, 2, len(pfv.EnvModelPercentages))
	assert.Equal(t, float32(30.0), pfv.EnvModelPercentages["prod-model1"].Percentage)
	assert.Assert(t, pfv.EnvModelPercentages["prod-model1"].Active)
	assert.Equal(t, float32(70.0), pfv.EnvModelPercentages["qa-model2"].Percentage)
	assert.Assert(t, !pfv.EnvModelPercentages["qa-model2"].Active)
}

func TestNewPercentageBean(t *testing.T) {
	bean := NewPercentageBean()
	
	assert.Assert(t, bean != nil)
	assert.Equal(t, shared.STB, bean.ApplicationType)
	assert.Assert(t, bean.FirmwareVersions != nil)
	assert.Equal(t, 0, len(bean.FirmwareVersions))
	assert.Assert(t, bean.Distributions != nil)
	assert.Equal(t, 0, len(bean.Distributions))
}

func TestPercentageBean_GetTemplateId(t *testing.T) {
	bean := &PercentageBean{}
	
	templateId := bean.GetTemplateId()
	assert.Equal(t, "ENV_MODEL_RULE", templateId)
}

func TestPercentageBean_GetRuleType(t *testing.T) {
	bean := &PercentageBean{}
	
	ruleType := bean.GetRuleType()
	assert.Equal(t, "PercentFilter", ruleType)
}

func TestPercentageBean_ValidateAll_NoDuplicates(t *testing.T) {
	bean1 := &PercentageBean{
		ID:          "bean1",
		Name:        "Bean 1",
		Environment: "PROD",
		Model:       "MODEL-A",
	}
	bean2 := &PercentageBean{
		ID:          "bean2",
		Name:        "Bean 2",
		Environment: "QA",
		Model:       "MODEL-B",
	}
	beans := []*PercentageBean{bean1, bean2}
	
	err := bean1.ValidateAll(beans)
	assert.NilError(t, err)
}

func TestPercentageBean_ValidateAll_DuplicateName(t *testing.T) {
	bean1 := &PercentageBean{
		ID:   "bean1",
		Name: "Duplicate Name",
	}
	bean2 := &PercentageBean{
		ID:   "bean2",
		Name: "duplicate name", // Same name, different case
	}
	beans := []*PercentageBean{bean1, bean2}
	
	err := bean1.ValidateAll(beans)
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "This name")
	assert.ErrorContains(t, err, "is already used")
}

func TestPercentageBean_ValidateAll_DuplicateEnvModel(t *testing.T) {
	bean1 := &PercentageBean{
		ID:          "bean1",
		Name:        "Bean 1",
		Environment: "PROD",
		Model:       "MODEL-A",
	}
	bean2 := &PercentageBean{
		ID:          "bean2",
		Name:        "Bean 2",
		Environment: "PROD", // Same env/model
		Model:       "MODEL-A",
	}
	beans := []*PercentageBean{bean1, bean2}
	
	err := bean1.ValidateAll(beans)
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "PercentageBean already exists")
}

func TestPercentageBean_Validate_BlankName(t *testing.T) {
	bean := &PercentageBean{
		Name:            "",
		Model:           "MODEL-A",
		ApplicationType: "stb",
	}
	
	err := bean.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Name could not be blank")
}

func TestPercentageBean_Validate_BlankModel(t *testing.T) {
	bean := &PercentageBean{
		Name:            "Test Bean",
		Model:           "",
		ApplicationType: "stb",
	}
	
	err := bean.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Model could not be blank")
}

func TestPercentFilterValue_GetEnvModelPercentage_Exists(t *testing.T) {
	emp := EnvModelPercentage{
		Percentage: 50.0,
		Active:     true,
	}
	
	pfv := &PercentFilterValue{
		EnvModelPercentages: map[string]EnvModelPercentage{
			"test-key": emp,
		},
	}
	
	result := pfv.GetEnvModelPercentage("test-key")
	assert.Assert(t, result != nil)
	assert.Equal(t, float32(50.0), result.Percentage)
	assert.Assert(t, result.Active)
}

func TestPercentFilterValue_GetEnvModelPercentage_NotExists(t *testing.T) {
	pfv := &PercentFilterValue{
		EnvModelPercentages: map[string]EnvModelPercentage{},
	}
	
	result := pfv.GetEnvModelPercentage("non-existent")
	assert.Assert(t, result == nil)
}

func TestEnvModelPercentage_AllFields(t *testing.T) {
	whitelist := &shared.IpAddressGroup{Id: "wl-1"}
	emp := EnvModelPercentage{
		Percentage:            75.5,
		Active:                true,
		FirmwareCheckRequired: true,
		RebootImmediately:     false,
		LastKnownGood:         "lkg-version",
		IntermediateVersion:   "intermediate-version",
		Whitelist:             whitelist,
		FirmwareVersions:      []string{"v1", "v2", "v3"},
		Name:                  "Env Model Name",
	}
	
	assert.Equal(t, float32(75.5), emp.Percentage)
	assert.Assert(t, emp.Active)
	assert.Assert(t, emp.FirmwareCheckRequired)
	assert.Assert(t, !emp.RebootImmediately)
	assert.Equal(t, "lkg-version", emp.LastKnownGood)
	assert.Equal(t, "intermediate-version", emp.IntermediateVersion)
	assert.Assert(t, emp.Whitelist != nil)
	assert.Equal(t, 3, len(emp.FirmwareVersions))
	assert.Equal(t, "Env Model Name", emp.Name)
}

func TestPercentFilterVo_EmptyPercentageBeans(t *testing.T) {
	gp := &GlobalPercentage{
		Percentage:      100.0,
		ApplicationType: "stb",
	}
	
	pfvo := &PercentFilterVo{
		GlobalPercentage: gp,
		PercentageBeans:  []PercentageBean{},
	}
	
	assert.Assert(t, pfvo.GlobalPercentage != nil)
	assert.Equal(t, 0, len(pfvo.PercentageBeans))
}

func TestGlobalPercentage_AllFields(t *testing.T) {
	gp := &GlobalPercentage{
		Whitelist:       "whitelist-id",
		Percentage:      85.5,
		ApplicationType: "xhome",
	}
	
	assert.Equal(t, "whitelist-id", gp.Whitelist)
	assert.Equal(t, float32(85.5), gp.Percentage)
	assert.Equal(t, "xhome", gp.ApplicationType)
}
