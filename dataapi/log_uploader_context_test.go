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
package dataapi

import (
	"net/http/httptest"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/common"
	loguploader "github.com/rdkcentral/xconfwebconfig/dataapi/dcm/logupload"
	xhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeLogUploaderContext(t *testing.T) {
	ws := &xhttp.XconfServer{}
	fields := log.Fields{}

	t.Run("NormalizeWithBasicContext", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"

		contextMap := map[string]string{
			common.ESTB_MAC_ADDRESS: "AA:BB:CC:DD:EE:FF",
			common.ECM_MAC_ADDRESS:  "11:22:33:44:55:66",
			common.ESTB_IP:          "10.0.0.1",
			common.APPLICATION_TYPE: shared.STB,
		}

		NormalizeLogUploaderContext(ws, req, contextMap, false, fields)

		// Verify MAC addresses are normalized
		assert.NotEmpty(t, contextMap[common.ESTB_MAC_ADDRESS])
		assert.NotEmpty(t, contextMap[common.ESTB_IP])
		assert.Equal(t, shared.STB, contextMap[common.APPLICATION_TYPE])
	})

	t.Run("NormalizeWithPartnerAppType", func(t *testing.T) {
		// Set up Xc for partner-based conversion
		originalXc := Xc
		defer func() { Xc = originalXc }()

		Xc = &XconfConfigs{
			DeriveAppTypeFromPartnerId: true,
			PartnerApplicationTypes:    []string{"cox", "shaw"},
		}

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"

		contextMap := map[string]string{
			common.ESTB_MAC_ADDRESS: "AA:BB:CC:DD:EE:FF",
			common.ECM_MAC_ADDRESS:  "11:22:33:44:55:66",
			common.ESTB_IP:          "10.0.0.1",
			common.APPLICATION_TYPE: shared.STB,
			common.PARTNER_ID:       "cox-partner",
		}

		NormalizeLogUploaderContext(ws, req, contextMap, true, fields)

		assert.NotEmpty(t, contextMap[common.ESTB_IP])
		// Partner ID "cox-partner" should convert STB to "cox"
		assert.Equal(t, "cox", contextMap[common.APPLICATION_TYPE])
	})

	t.Run("NormalizeWithoutPartnerAppType", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.5:8080"

		contextMap := map[string]string{
			common.ESTB_MAC_ADDRESS: "AA:BB:CC:DD:EE:FF",
			common.APPLICATION_TYPE: shared.STB,
		}

		NormalizeLogUploaderContext(ws, req, contextMap, false, fields)

		// Should not modify application type when usePartnerAppType is false
		assert.Equal(t, shared.STB, contextMap[common.APPLICATION_TYPE])
	})

	t.Run("NormalizeWithEmptyIP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.50:9090"

		contextMap := map[string]string{
			common.ESTB_MAC_ADDRESS: "AA:BB:CC:DD:EE:FF",
			common.ESTB_IP:          "",
			common.APPLICATION_TYPE: shared.STB,
		}

		NormalizeLogUploaderContext(ws, req, contextMap, false, fields)

		// Should extract IP from request
		assert.NotEmpty(t, contextMap[common.ESTB_IP])
	})
}

func TestToTelemetry2Profile(t *testing.T) {
	t.Run("ConvertWithComponent", func(t *testing.T) {
		telemetryProfile := []logupload.TelemetryElement{
			{
				ID:        "test-id-1",
				Header:    "test-header",
				Content:   "test-content",
				Type:      "test-type",
				Component: "test-component",
			},
		}

		result := ToTelemetry2Profile(telemetryProfile)

		assert.Len(t, result, 1)
		assert.Equal(t, "test-component", result[0].Content)
		assert.Equal(t, "<event>", result[0].Type)
		assert.Equal(t, "", result[0].Component)
		assert.Equal(t, "test-id-1", result[0].ID)
		assert.Equal(t, "test-header", result[0].Header)
	})

	t.Run("ConvertWithoutComponent", func(t *testing.T) {
		telemetryProfile := []logupload.TelemetryElement{
			{
				ID:        "test-id-2",
				Header:    "test-header-2",
				Content:   "original-content",
				Type:      "original-type",
				Component: "",
			},
		}

		result := ToTelemetry2Profile(telemetryProfile)

		assert.Len(t, result, 1)
		// Should not modify when component is empty
		assert.Equal(t, "original-content", result[0].Content)
		assert.Equal(t, "original-type", result[0].Type)
		assert.Equal(t, "", result[0].Component)
	})

	t.Run("ConvertMultipleElements", func(t *testing.T) {
		telemetryProfile := []logupload.TelemetryElement{
			{
				ID:        "id-1",
				Component: "component-1",
				Content:   "content-1",
				Type:      "type-1",
			},
			{
				ID:        "id-2",
				Component: "component-2",
				Content:   "content-2",
				Type:      "type-2",
			},
			{
				ID:        "id-3",
				Component: "", // No component
				Content:   "content-3",
				Type:      "type-3",
			},
		}

		result := ToTelemetry2Profile(telemetryProfile)

		assert.Len(t, result, 3)
		// First element should be converted
		assert.Equal(t, "component-1", result[0].Content)
		assert.Equal(t, "<event>", result[0].Type)
		assert.Equal(t, "", result[0].Component)
		// Second element should be converted
		assert.Equal(t, "component-2", result[1].Content)
		assert.Equal(t, "<event>", result[1].Type)
		assert.Equal(t, "", result[1].Component)
		// Third element should remain unchanged
		assert.Equal(t, "content-3", result[2].Content)
		assert.Equal(t, "type-3", result[2].Type)
	})

	t.Run("ConvertEmptySlice", func(t *testing.T) {
		telemetryProfile := []logupload.TelemetryElement{}

		result := ToTelemetry2Profile(telemetryProfile)

		assert.Len(t, result, 0)
	})
}

func TestNullifyUnwantedFields(t *testing.T) {
	t.Run("NullifyWithValidProfile", func(t *testing.T) {
		profile := &logupload.PermanentTelemetryProfile{
			ID:   "test-profile-id",
			Name: "test-profile",
			TelemetryProfile: []logupload.TelemetryElement{
				{
					ID:        "element-id-1",
					Header:    "header-1",
					Content:   "content-1",
					Type:      "type-1",
					Component: "component-1",
				},
				{
					ID:        "element-id-2",
					Header:    "header-2",
					Content:   "content-2",
					Type:      "type-2",
					Component: "component-2",
				},
			},
		}

		NullifyUnwantedFields(profile)

		// Profile-level fields should be preserved
		assert.Equal(t, "test-profile-id", profile.ID)
		assert.Equal(t, "test-profile", profile.Name)

		// Element ID and Component should be nullified
		assert.Equal(t, "", profile.TelemetryProfile[0].ID)
		assert.Equal(t, "", profile.TelemetryProfile[0].Component)
		assert.Equal(t, "", profile.TelemetryProfile[1].ID)
		assert.Equal(t, "", profile.TelemetryProfile[1].Component)

		// Other element fields should be preserved
		assert.Equal(t, "header-1", profile.TelemetryProfile[0].Header)
		assert.Equal(t, "content-1", profile.TelemetryProfile[0].Content)
		assert.Equal(t, "type-1", profile.TelemetryProfile[0].Type)
	})

	t.Run("NullifyWithNilProfile", func(t *testing.T) {
		// Should not panic with nil profile
		NullifyUnwantedFields(nil)
	})

	t.Run("NullifyWithEmptyTelemetryProfile", func(t *testing.T) {
		profile := &logupload.PermanentTelemetryProfile{
			ID:               "test-id",
			Name:             "test-name",
			TelemetryProfile: []logupload.TelemetryElement{},
		}

		NullifyUnwantedFields(profile)

		assert.Equal(t, "test-id", profile.ID)
		assert.Equal(t, "test-name", profile.Name)
		assert.Len(t, profile.TelemetryProfile, 0)
	})
}

func TestCleanupLusUploadRepository(t *testing.T) {
	t.Run("CleanupForVersion2OrHigher", func(t *testing.T) {
		settings := &logupload.Settings{
			LusUploadRepositoryURL:            "http://old-url.com",
			LusUploadRepositoryUploadProtocol: "http",
			LusUploadRepositoryURLNew:         "http://new-url.com",
		}

		CleanupLusUploadRepository(settings, "2.0")

		// For version >= 2.0, LusUploadRepositoryURL should be cleared
		assert.Equal(t, "", settings.LusUploadRepositoryURL)
		// Other fields should remain
		assert.Equal(t, "http", settings.LusUploadRepositoryUploadProtocol)
		assert.Equal(t, "http://new-url.com", settings.LusUploadRepositoryURLNew)
	})

	t.Run("CleanupForVersion3", func(t *testing.T) {
		settings := &logupload.Settings{
			LusUploadRepositoryURL:            "http://old-url.com",
			LusUploadRepositoryUploadProtocol: "https",
			LusUploadRepositoryURLNew:         "http://new-url-v3.com",
		}

		CleanupLusUploadRepository(settings, "3.0")

		// For version >= 2.0, LusUploadRepositoryURL should be cleared
		assert.Equal(t, "", settings.LusUploadRepositoryURL)
		assert.Equal(t, "https", settings.LusUploadRepositoryUploadProtocol)
		assert.Equal(t, "http://new-url-v3.com", settings.LusUploadRepositoryURLNew)
	})

	t.Run("CleanupForVersion1", func(t *testing.T) {
		settings := &logupload.Settings{
			LusUploadRepositoryURL:            "http://old-url.com",
			LusUploadRepositoryUploadProtocol: "http",
			LusUploadRepositoryURLNew:         "http://new-url.com",
		}

		CleanupLusUploadRepository(settings, "1.9")

		// For version < 2.0, protocol and new URL should be cleared
		assert.Equal(t, "http://old-url.com", settings.LusUploadRepositoryURL)
		assert.Equal(t, "", settings.LusUploadRepositoryUploadProtocol)
		assert.Equal(t, "", settings.LusUploadRepositoryURLNew)
	})

	t.Run("CleanupForVersion1_5", func(t *testing.T) {
		settings := &logupload.Settings{
			LusUploadRepositoryURL:            "http://legacy-url.com",
			LusUploadRepositoryUploadProtocol: "ftp",
			LusUploadRepositoryURLNew:         "http://new-url.com",
		}

		CleanupLusUploadRepository(settings, "1.5")

		// For version < 2.0, protocol and new URL should be cleared
		assert.Equal(t, "http://legacy-url.com", settings.LusUploadRepositoryURL)
		assert.Equal(t, "", settings.LusUploadRepositoryUploadProtocol)
		assert.Equal(t, "", settings.LusUploadRepositoryURLNew)
	})

	t.Run("CleanupWithNilSettings", func(t *testing.T) {
		// Should not panic with nil settings
		CleanupLusUploadRepository(nil, "2.0")
	})

	t.Run("CleanupWithEmptyVersion", func(t *testing.T) {
		settings := &logupload.Settings{
			LusUploadRepositoryURL:            "http://url.com",
			LusUploadRepositoryUploadProtocol: "http",
			LusUploadRepositoryURLNew:         "http://new-url.com",
		}

		CleanupLusUploadRepository(settings, "")

		// Empty version should be treated as < 2.0
		assert.Equal(t, "http://url.com", settings.LusUploadRepositoryURL)
		assert.Equal(t, "", settings.LusUploadRepositoryUploadProtocol)
		assert.Equal(t, "", settings.LusUploadRepositoryURLNew)
	})
}

func TestLogResultSettings(t *testing.T) {
	// Save original function and restore after test
	originalGetOneDcmRuleFunc := loguploader.GetOneDcmRuleFunc
	defer func() {
		loguploader.GetOneDcmRuleFunc = originalGetOneDcmRuleFunc
	}()

	t.Run("LogWithValidSettings", func(t *testing.T) {
		// Mock GetOneDcmRuleFunc
		loguploader.GetOneDcmRuleFunc = func(ruleId string) *logupload.DCMGenericRule {
			if ruleId == "rule-1" {
				return &logupload.DCMGenericRule{
					ID:   "rule-1",
					Name: "Test Rule 1",
				}
			}
			if ruleId == "rule-2" {
				return &logupload.DCMGenericRule{
					ID:   "rule-2",
					Name: "Test Rule 2",
				}
			}
			return nil
		}

		settings := &logupload.Settings{
			RuleIDs: map[string]string{
				"rule-1": "",
				"rule-2": "",
			},
		}

		telemetryRule := &logupload.TelemetryRule{
			ID:   "telemetry-1",
			Name: "Test Telemetry Rule",
		}

		settingRules := []*logupload.SettingRule{
			{
				ID:   "setting-1",
				Name: "Setting Rule 1",
			},
			{
				ID:   "setting-2",
				Name: "Setting Rule 2",
			},
		}

		fields := log.Fields{}

		LogResultSettings(settings, telemetryRule, settingRules, fields)

		// Verify fields are set
		assert.NotNil(t, fields["formulaNames"])
		assert.Equal(t, "Test Telemetry Rule", fields["telemetryRuleName"])
		assert.NotNil(t, fields["settingRuleNames"])

		settingRuleNames := fields["settingRuleNames"].([]string)
		assert.Len(t, settingRuleNames, 2)
		assert.Contains(t, settingRuleNames, "Setting Rule 1")
		assert.Contains(t, settingRuleNames, "Setting Rule 2")
	})

	t.Run("LogWithNilTelemetryRule", func(t *testing.T) {
		loguploader.GetOneDcmRuleFunc = func(ruleId string) *logupload.DCMGenericRule {
			return &logupload.DCMGenericRule{
				ID:   ruleId,
				Name: "Rule " + ruleId,
			}
		}

		settings := &logupload.Settings{
			RuleIDs: map[string]string{
				"rule-1": "",
			},
		}

		fields := log.Fields{}

		LogResultSettings(settings, nil, []*logupload.SettingRule{}, fields)

		// When telemetry rule is nil, should log "NoMatch"
		assert.Equal(t, "NoMatch", fields["telemetryRuleName"])
	})

	t.Run("LogWithEmptySettingRules", func(t *testing.T) {
		loguploader.GetOneDcmRuleFunc = func(ruleId string) *logupload.DCMGenericRule {
			return nil
		}

		settings := &logupload.Settings{
			RuleIDs: map[string]string{},
		}

		telemetryRule := &logupload.TelemetryRule{
			Name: "Test Rule",
		}

		fields := log.Fields{}

		LogResultSettings(settings, telemetryRule, []*logupload.SettingRule{}, fields)

		assert.Equal(t, "Test Rule", fields["telemetryRuleName"])
		settingRuleNames := fields["settingRuleNames"].([]string)
		assert.Len(t, settingRuleNames, 0)
	})

	t.Run("LogWithNilDcmRule", func(t *testing.T) {
		// Mock returns nil
		loguploader.GetOneDcmRuleFunc = func(ruleId string) *logupload.DCMGenericRule {
			return nil
		}

		settings := &logupload.Settings{
			RuleIDs: map[string]string{
				"nonexistent-rule": "",
			},
		}

		telemetryRule := &logupload.TelemetryRule{
			Name: "Test Rule",
		}

		fields := log.Fields{}

		LogResultSettings(settings, telemetryRule, []*logupload.SettingRule{}, fields)

		// Should handle nil dcmRule gracefully
		formulaNames := fields["formulaNames"].([]string)
		assert.Len(t, formulaNames, 0) // Nil rules are skipped
	})
}

func TestGetTelemetryTwoProfileResponeDicts(t *testing.T) {
	t.Run("GetProfilesWithInvalidJSON", func(t *testing.T) {
		contextMap := map[string]string{
			common.ESTB_MAC_ADDRESS: "AA:BB:CC:DD:EE:FF",
			common.MODEL:            "TEST_MODEL",
		}

		fields := log.Fields{}

		// This will return empty results since no rules/profiles are configured
		result, err := GetTelemetryTwoProfileResponeDicts(contextMap, fields)

		// Should handle gracefully even with no configured rules
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.RulesMatched)
		assert.Len(t, result.ProfilesData, 0)
	})

	t.Run("GetProfilesWithEmptyContext", func(t *testing.T) {
		contextMap := map[string]string{}
		fields := log.Fields{}

		result, err := GetTelemetryTwoProfileResponeDicts(contextMap, fields)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.RulesMatched)
		assert.Len(t, result.ProfilesData, 0)
	})

	t.Run("TestTelemetryEvaluationResultStructure", func(t *testing.T) {
		// Test the structure directly
		result := &TelemetryEvaluationResult{
			RulesMatched: true,
			ProfilesData: []util.Dict{
				{
					"name":        "test-profile",
					"versionHash": uint32(12345),
					"value": util.Dict{
						"key": "value",
					},
				},
			},
		}

		assert.True(t, result.RulesMatched)
		assert.Len(t, result.ProfilesData, 1)
		assert.Equal(t, "test-profile", result.ProfilesData[0]["name"])
	})
}

func TestAddLogUploaderContext(t *testing.T) {
	t.Run("AddLogUploaderContextSkipped", func(t *testing.T) {
		// This test is skipped because AddLogUploaderContext requires complex mocking:
		// 1. SAT token retrieval via xhttp.GetLocalSatToken
		// 2. Partner service call via GetPartnerFromAccountServiceByHostMac
		// 3. Tagging service call via AddContextFromTaggingService
		// These require HTTP server mocks and are better suited for integration tests.
		t.Skip("Skipping AddLogUploaderContext - requires complex HTTP service mocking for SAT token, account service, and tagging service")
	})
}
