/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
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

package telemetry

import (
	"testing"

	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestNewTelemetryProfileService tests the constructor
func TestNewTelemetryProfileService(t *testing.T) {
	service := NewTelemetryProfileService()
	assert.NotNil(t, service)
}

// TestConvertToDescriptor tests the ConvertToDescriptor method
func TestConvertToDescriptor(t *testing.T) {
	service := NewTelemetryProfileService()

	t.Run("ConvertValidTelemetryRule", func(t *testing.T) {
		rule := logupload.TelemetryRule{
			ID:              "test-rule-id",
			Name:            "Test Rule",
			ApplicationType: "stb",
		}

		descriptor := service.ConvertToDescriptor(rule)

		assert.NotNil(t, descriptor)
		assert.Equal(t, "test-rule-id", descriptor.RuleId)
		assert.Equal(t, "Test Rule", descriptor.RuleName)
	})

	t.Run("ConvertEmptyTelemetryRule", func(t *testing.T) {
		rule := logupload.TelemetryRule{}

		descriptor := service.ConvertToDescriptor(rule)

		assert.NotNil(t, descriptor)
		assert.Equal(t, "", descriptor.RuleId)
		assert.Equal(t, "", descriptor.RuleName)
	})

	t.Run("ConvertTelemetryRuleWithXHomeAppType", func(t *testing.T) {
		rule := logupload.TelemetryRule{
			ID:              "xhome-rule",
			Name:            "XHome Rule",
			ApplicationType: "xhome",
		}

		descriptor := service.ConvertToDescriptor(rule)

		assert.NotNil(t, descriptor)
		assert.Equal(t, "xhome-rule", descriptor.RuleId)
		assert.Equal(t, "XHome Rule", descriptor.RuleName)
	})
}

// TestConvertToProfileDescriptor tests the ConvertToProfileDescriptor method
func TestConvertToProfileDescriptor(t *testing.T) {
	service := NewTelemetryProfileService()

	t.Run("ConvertValidTelemetryProfile", func(t *testing.T) {
		profile := logupload.TelemetryProfile{
			ID:   "profile-123",
			Name: "Test Profile",
		}

		descriptor := service.ConvertToProfileDescriptor(profile)

		assert.NotNil(t, descriptor)
		assert.Equal(t, "profile-123", descriptor.ID)
		assert.Equal(t, "Test Profile", descriptor.Name)
	})

	t.Run("ConvertEmptyTelemetryProfile", func(t *testing.T) {
		profile := logupload.TelemetryProfile{}

		descriptor := service.ConvertToProfileDescriptor(profile)

		assert.NotNil(t, descriptor)
		assert.Equal(t, "", descriptor.ID)
		assert.Equal(t, "", descriptor.Name)
	})

	t.Run("ConvertTelemetryProfileWithNameAndAppType", func(t *testing.T) {
		profile := logupload.TelemetryProfile{
			ID:              "profile-multi",
			Name:            "Multi Profile",
			ApplicationType: "xhome",
		}

		descriptor := service.ConvertToProfileDescriptor(profile)

		assert.NotNil(t, descriptor)
		assert.Equal(t, "profile-multi", descriptor.ID)
		assert.Equal(t, "Multi Profile", descriptor.Name)
	})
}

// TestCreateRuleForAttribute tests the CreateRuleForAttribute method
func TestCreateRuleForAttribute(t *testing.T) {
	service := NewTelemetryProfileService()

	t.Run("CreateRuleWithValidAttributes", func(t *testing.T) {
		rule := service.CreateRuleForAttribute("model", "XG1v4")

		assert.NotNil(t, rule)
		assert.NotNil(t, rule.Rule)
		assert.NotNil(t, rule.Rule.Condition)
		// Timestamp field exists but may be 0 if not explicitly set
	})

	t.Run("CreateRuleWithEmptyAttribute", func(t *testing.T) {
		rule := service.CreateRuleForAttribute("", "value")

		assert.NotNil(t, rule)
		assert.NotNil(t, rule.Rule)
	})

	t.Run("CreateRuleWithEmptyValue", func(t *testing.T) {
		rule := service.CreateRuleForAttribute("attribute", "")

		assert.NotNil(t, rule)
		assert.NotNil(t, rule.Rule)
	})

	t.Run("CreateRuleWithBothEmpty", func(t *testing.T) {
		rule := service.CreateRuleForAttribute("", "")

		assert.NotNil(t, rule)
		assert.NotNil(t, rule.Rule)
	})

	t.Run("CreateRuleWithSpecialCharacters", func(t *testing.T) {
		rule := service.CreateRuleForAttribute("device.type", "special@#$%")

		assert.NotNil(t, rule)
		assert.NotNil(t, rule.Rule)
	})
}

// TestGetMaxRule tests the GetMaxRule method
func TestGetMaxRule(t *testing.T) {
	service := NewTelemetryProfileService()

	t.Run("GetMaxFromMultipleRulesWithValidConditions", func(t *testing.T) {
		// Create rules with proper conditions
		rule1 := &logupload.TelemetryRule{ID: "rule1", Name: "First Rule"}
		rule1.Rule.Condition = re.NewCondition(
			re.NewFreeArg("STRING", "model"),
			"IS",
			re.NewFixedArg("XG1v3"),
		)

		rule2 := &logupload.TelemetryRule{ID: "rule2", Name: "Second Rule"}
		rule2.Rule.Condition = re.NewCondition(
			re.NewFreeArg("STRING", "model"),
			"IS",
			re.NewFixedArg("XG1v4"),
		)

		rules := []*logupload.TelemetryRule{rule1, rule2}

		maxRule := service.GetMaxRule(rules)

		assert.NotNil(t, maxRule)
		// Just verify we get a rule back - the comparison logic is in rulesengine
		assert.Contains(t, []string{"rule1", "rule2"}, maxRule.ID)
	})

	t.Run("GetMaxFromSingleRule", func(t *testing.T) {
		rule := &logupload.TelemetryRule{ID: "only-rule", Name: "Only Rule"}
		rule.Rule.Condition = re.NewCondition(
			re.NewFreeArg("STRING", "model"),
			"IS",
			re.NewFixedArg("XG1v4"),
		)

		rules := []*logupload.TelemetryRule{rule}

		maxRule := service.GetMaxRule(rules)

		assert.NotNil(t, maxRule)
		assert.Equal(t, "only-rule", maxRule.ID)
	})

	t.Run("GetMaxFromEmptyList", func(t *testing.T) {
		rules := []*logupload.TelemetryRule{}

		maxRule := service.GetMaxRule(rules)

		assert.Nil(t, maxRule)
	})

	t.Run("GetMaxFromNilList", func(t *testing.T) {
		var rules []*logupload.TelemetryRule = nil

		maxRule := service.GetMaxRule(rules)

		assert.Nil(t, maxRule)
	})
}

// TestGetTelemetryTwoProfileByTelemetryRules tests the deduplication logic
func TestGetTelemetryTwoProfileByTelemetryRules(t *testing.T) {
	service := NewTelemetryProfileService()

	// Save original function
	originalGetOneTelemetryTwoProfileFunc := GetOneTelemetryTwoProfileFunc
	defer func() { GetOneTelemetryTwoProfileFunc = originalGetOneTelemetryTwoProfileFunc }()

	t.Run("GetProfilesWithDeduplication", func(t *testing.T) {
		rules := []*logupload.TelemetryTwoRule{
			{ID: "rule1", Name: "Rule 1", BoundTelemetryIDs: []string{"profile1", "profile2"}},
			{ID: "rule2", Name: "Rule 2", BoundTelemetryIDs: []string{"profile2", "profile3"}},
		}

		// Mock the database call to return profiles
		GetOneTelemetryTwoProfileFunc = func(id string) *logupload.TelemetryTwoProfile {
			profiles := map[string]*logupload.TelemetryTwoProfile{
				"profile1": {ID: "profile1", Name: "Profile 1"},
				"profile2": {ID: "profile2", Name: "Profile 2"},
				"profile3": {ID: "profile3", Name: "Profile 3"},
			}
			return profiles[id]
		}

		fields := log.Fields{}
		profiles := service.GetTelemetryTwoProfileByTelemetryRules(rules, fields)

		assert.NotNil(t, profiles)
		assert.Len(t, profiles, 3) // Should have 3 unique profiles (deduplication)

		// Verify all profiles are present
		profileIDs := make(map[string]bool)
		for _, p := range profiles {
			profileIDs[p.ID] = true
		}
		assert.True(t, profileIDs["profile1"])
		assert.True(t, profileIDs["profile2"])
		assert.True(t, profileIDs["profile3"])
	})

	t.Run("GetProfilesWithEmptyRules", func(t *testing.T) {
		rules := []*logupload.TelemetryTwoRule{}

		GetOneTelemetryTwoProfileFunc = func(id string) *logupload.TelemetryTwoProfile {
			return &logupload.TelemetryTwoProfile{ID: id}
		}

		fields := log.Fields{}
		profiles := service.GetTelemetryTwoProfileByTelemetryRules(rules, fields)

		assert.NotNil(t, profiles)
		assert.Len(t, profiles, 0)
	})

	t.Run("GetProfilesWithNilProfile", func(t *testing.T) {
		rules := []*logupload.TelemetryTwoRule{
			{ID: "rule1", Name: "Rule 1", BoundTelemetryIDs: []string{"profile1", "missing-profile"}},
		}

		// Mock to return nil for missing profile
		GetOneTelemetryTwoProfileFunc = func(id string) *logupload.TelemetryTwoProfile {
			if id == "profile1" {
				return &logupload.TelemetryTwoProfile{ID: "profile1", Name: "Profile 1"}
			}
			return nil
		}

		fields := log.Fields{}
		profiles := service.GetTelemetryTwoProfileByTelemetryRules(rules, fields)

		assert.NotNil(t, profiles)
		assert.Len(t, profiles, 1) // Only profile1 should be included
		assert.Equal(t, "profile1", profiles[0].ID)
	})

	t.Run("GetProfilesWithDuplicateIDs", func(t *testing.T) {
		rules := []*logupload.TelemetryTwoRule{
			{ID: "rule1", Name: "Rule 1", BoundTelemetryIDs: []string{"profile1", "profile1", "profile1"}},
		}

		GetOneTelemetryTwoProfileFunc = func(id string) *logupload.TelemetryTwoProfile {
			return &logupload.TelemetryTwoProfile{ID: id, Name: "Test Profile"}
		}

		fields := log.Fields{}
		profiles := service.GetTelemetryTwoProfileByTelemetryRules(rules, fields)

		assert.NotNil(t, profiles)
		assert.Len(t, profiles, 1) // Should deduplicate to 1 profile
		assert.Equal(t, "profile1", profiles[0].ID)
	})
}
