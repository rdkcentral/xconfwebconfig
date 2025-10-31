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
package firmware

import (
	"testing"

	"gotest.tools/assert"
)

// Test IsValidApplicableActionType
func TestIsValidApplicableActionType(t *testing.T) {
	// Valid types
	assert.Assert(t, IsValidApplicableActionType(RULE))
	assert.Assert(t, IsValidApplicableActionType(DEFINE_PROPERTIES))
	assert.Assert(t, IsValidApplicableActionType(BLOCKING_FILTER))
	assert.Assert(t, IsValidApplicableActionType(RULE_TEMPLATE))
	assert.Assert(t, IsValidApplicableActionType(DEFINE_PROPERTIES_TEMPLATE))
	assert.Assert(t, IsValidApplicableActionType(BLOCKING_FILTER_TEMPLATE))

	// Invalid type
	assert.Assert(t, !IsValidApplicableActionType("INVALID"))
	assert.Assert(t, !IsValidApplicableActionType(""))
}

// Test ApplicableActionTypeToString
func TestApplicableActionTypeToString(t *testing.T) {
	assert.Equal(t, "RULE", ApplicableActionTypeToString(RULE))
	assert.Equal(t, "DEFINE_PROPERTIES", ApplicableActionTypeToString(DEFINE_PROPERTIES))
	assert.Equal(t, "BLOCKING_FILTER", ApplicableActionTypeToString(BLOCKING_FILTER))
	assert.Equal(t, "RULE_TEMPLATE", ApplicableActionTypeToString(RULE_TEMPLATE))
	assert.Equal(t, "DEFINE_PROPERTIES_TEMPLATE", ApplicableActionTypeToString(DEFINE_PROPERTIES_TEMPLATE))
	assert.Equal(t, "BLOCKING_FILTER_TEMPLATE", ApplicableActionTypeToString(BLOCKING_FILTER_TEMPLATE))

	// Invalid type returns empty string
	assert.Equal(t, "", ApplicableActionTypeToString("INVALID"))
}

// Test CaseIgnoreEquals
func TestApplicableActionType_CaseIgnoreEquals(t *testing.T) {
	rule := RULE

	assert.Assert(t, rule.CaseIgnoreEquals(RULE))
	assert.Assert(t, rule.CaseIgnoreEquals("rule"))
	assert.Assert(t, rule.CaseIgnoreEquals("RULE"))
	assert.Assert(t, rule.CaseIgnoreEquals("RuLe"))
	assert.Assert(t, !rule.CaseIgnoreEquals(DEFINE_PROPERTIES))
	assert.Assert(t, !rule.CaseIgnoreEquals("different"))
}

// Test IsSuperSetOf
func TestApplicableActionType_IsSuperSetOf(t *testing.T) {
	ruleTemplate := RULE_TEMPLATE
	rule := RULE
	defPropsTemplate := DEFINE_PROPERTIES_TEMPLATE
	defProps := DEFINE_PROPERTIES

	// RULE_TEMPLATE is superset of RULE
	assert.Assert(t, ruleTemplate.IsSuperSetOf(&rule))

	// DEFINE_PROPERTIES_TEMPLATE is superset of DEFINE_PROPERTIES
	assert.Assert(t, defPropsTemplate.IsSuperSetOf(&defProps))

	// RULE is not superset of RULE_TEMPLATE
	assert.Assert(t, !rule.IsSuperSetOf(&ruleTemplate))

	// RULE is superset of itself
	rule2 := RULE
	assert.Assert(t, rule.IsSuperSetOf(&rule2))
}

// Test isValidApplicableClass
func TestIsValidApplicableClass(t *testing.T) {
	assert.Assert(t, isValidApplicableClass("com.comcast.xconf.ApplicableAction"))
	assert.Assert(t, isValidApplicableClass("com.comcast.xconf.RuleAction"))
	assert.Assert(t, isValidApplicableClass("com.comcast.xconf.DefinePropertiesAction"))
	assert.Assert(t, isValidApplicableClass("com.comcast.xconf.DefinePropertiesTemplateAction"))
	assert.Assert(t, isValidApplicableClass("com.comcast.xconf.BlockingFilterAction"))

	assert.Assert(t, !isValidApplicableClass(""))
	assert.Assert(t, !isValidApplicableClass("InvalidClass"))
	assert.Assert(t, !isValidApplicableClass("com.comcast.xconf.OtherAction"))
}

// Test ApplicableActionType constants
func TestApplicableActionType_Constants(t *testing.T) {
	assert.Equal(t, "RULE", string(RULE))
	assert.Equal(t, "DEFINE_PROPERTIES", string(DEFINE_PROPERTIES))
	assert.Equal(t, "BLOCKING_FILTER", string(BLOCKING_FILTER))
	assert.Equal(t, "RULE_TEMPLATE", string(RULE_TEMPLATE))
	assert.Equal(t, "DEFINE_PROPERTIES_TEMPLATE", string(DEFINE_PROPERTIES_TEMPLATE))
	assert.Equal(t, "BLOCKING_FILTER_TEMPLATE", string(BLOCKING_FILTER_TEMPLATE))
}

// Test class name constants
func TestApplicableClass_Constants(t *testing.T) {
	assert.Equal(t, ".ApplicableAction", ApplicableActionClass)
	assert.Equal(t, ".RuleAction", RuleActionClass)
	assert.Equal(t, ".DefinePropertiesAction", DefinePropertiesActionClass)
	assert.Equal(t, ".DefinePropertiesTemplateAction", DefinePropertiesTemplateActionClass)
	assert.Equal(t, ".BlockingFilterAction", BlockingFilterActionClass)
}

// Test RuleAction struct
func TestRuleAction_Fields(t *testing.T) {
	configEntry := &ConfigEntry{
		ConfigId:          "config-1",
		Percentage:        50.5,
		StartPercentRange: 0.0,
		EndPercentRange:   50.5,
		IsPaused:          false,
		IsCanaryDisabled:  true,
	}

	ruleAction := &RuleAction{
		ConfigId:              "config-123",
		ConfigEntries:         []ConfigEntry{*configEntry},
		Active:                true,
		UseAccountPercentage:  true,
		FirmwareCheckRequired: false,
		RebootImmediately:     true,
	}

	assert.Equal(t, "config-123", ruleAction.ConfigId)
	assert.Equal(t, 1, len(ruleAction.ConfigEntries))
	assert.Assert(t, ruleAction.Active)
	assert.Assert(t, ruleAction.UseAccountPercentage)
	assert.Assert(t, !ruleAction.FirmwareCheckRequired)
	assert.Assert(t, ruleAction.RebootImmediately)
}

// Test NewRuleAction constructor
func TestNewRuleAction(t *testing.T) {
	ruleAction := NewRuleAction()

	assert.Assert(t, ruleAction != nil)
	assert.Assert(t, ruleAction.Active)
	assert.Assert(t, !ruleAction.UseAccountPercentage)
	assert.Assert(t, !ruleAction.FirmwareCheckRequired)
	assert.Assert(t, !ruleAction.RebootImmediately)
}

// Test NewConfigEntry
func TestNewConfigEntry(t *testing.T) {
	entry := NewConfigEntry("config-1", 10.5, 60.8)

	assert.Assert(t, entry != nil)
	assert.Equal(t, "config-1", entry.ConfigId)
	assert.Equal(t, 10.5, entry.StartPercentRange)
	assert.Equal(t, 60.8, entry.EndPercentRange)
	// Percentage should be calculated as (60.8 - 10.5) = 50.3
	assert.Equal(t, 50.3, entry.Percentage)
}

// Test ConfigEntry Equals
func TestConfigEntry_Equals_Same(t *testing.T) {
	entry1 := &ConfigEntry{
		ConfigId:          "config-1",
		Percentage:        50.0,
		StartPercentRange: 0.0,
		EndPercentRange:   50.0,
		IsPaused:          false,
	}

	entry2 := &ConfigEntry{
		ConfigId:          "config-1",
		Percentage:        50.0,
		StartPercentRange: 0.0,
		EndPercentRange:   50.0,
		IsPaused:          false,
	}

	assert.Assert(t, entry1.Equals(entry2))
}

func TestConfigEntry_Equals_Different(t *testing.T) {
	entry1 := &ConfigEntry{
		ConfigId:          "config-1",
		Percentage:        50.0,
		StartPercentRange: 0.0,
		EndPercentRange:   50.0,
	}

	entry2 := &ConfigEntry{
		ConfigId:          "config-2",
		Percentage:        50.0,
		StartPercentRange: 0.0,
		EndPercentRange:   50.0,
	}

	assert.Assert(t, !entry1.Equals(entry2))
}

func TestConfigEntry_Equals_Nil(t *testing.T) {
	entry := &ConfigEntry{ConfigId: "config-1"}

	assert.Assert(t, !entry.Equals(nil))
}

// Test ConfigEntry CompareTo
func TestConfigEntry_CompareTo_Greater(t *testing.T) {
	entry1 := &ConfigEntry{StartPercentRange: 50.0}
	entry2 := &ConfigEntry{StartPercentRange: 25.0}

	assert.Equal(t, 1, entry1.CompareTo(entry2))
}

func TestConfigEntry_CompareTo_Less(t *testing.T) {
	entry1 := &ConfigEntry{StartPercentRange: 25.0}
	entry2 := &ConfigEntry{StartPercentRange: 50.0}

	assert.Equal(t, -1, entry1.CompareTo(entry2))
}

func TestConfigEntry_CompareTo_Equal(t *testing.T) {
	entry1 := &ConfigEntry{StartPercentRange: 50.0}
	entry2 := &ConfigEntry{StartPercentRange: 50.0}

	assert.Equal(t, 0, entry1.CompareTo(entry2))
}

func TestConfigEntry_CompareTo_Nil(t *testing.T) {
	entry := &ConfigEntry{StartPercentRange: 50.0}

	assert.Equal(t, 0, entry.CompareTo(nil))
}

func TestConfigEntry_CompareTo_ZeroRange(t *testing.T) {
	entry1 := &ConfigEntry{StartPercentRange: 0.0}
	entry2 := &ConfigEntry{StartPercentRange: 50.0}

	assert.Equal(t, 0, entry1.CompareTo(entry2))
}
