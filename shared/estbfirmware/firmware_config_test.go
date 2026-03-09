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

func TestFirmwareConfig_SetApplicationType(t *testing.T) {
	config := &FirmwareConfig{}
	config.SetApplicationType("stb")
	
	assert.Equal(t, "stb", config.ApplicationType)
}

func TestFirmwareConfig_GetApplicationType(t *testing.T) {
	config := &FirmwareConfig{ApplicationType: "xhome"}
	
	appType := config.GetApplicationType()
	assert.Equal(t, "xhome", appType)
}

func TestFirmwareConfig_Creation(t *testing.T) {
	config := &FirmwareConfig{
		ID:               "config-1",
		Description:      "Test Firmware Config",
		FirmwareFilename: "firmware-v1.0.bin",
		FirmwareVersion:  "1.0.0",
	}
	
	assert.Equal(t, "config-1", config.ID)
	assert.Equal(t, "Test Firmware Config", config.Description)
	assert.Equal(t, "firmware-v1.0.bin", config.FirmwareFilename)
	assert.Equal(t, "1.0.0", config.FirmwareVersion)
}

func TestFirmwareConfig_WithSupportedModelIds(t *testing.T) {
	config := &FirmwareConfig{
		ID:                "config-2",
		Description:       "Config with Models",
		FirmwareFilename:  "firmware.bin",
		FirmwareVersion:   "2.0.0",
		SupportedModelIds: []string{"MODEL-A", "MODEL-B", "MODEL-C"},
	}
	
	assert.Equal(t, 3, len(config.SupportedModelIds))
	assert.Equal(t, "MODEL-A", config.SupportedModelIds[0])
	assert.Equal(t, "MODEL-B", config.SupportedModelIds[1])
	assert.Equal(t, "MODEL-C", config.SupportedModelIds[2])
}

func TestFirmwareConfig_WithLocations(t *testing.T) {
	config := &FirmwareConfig{
		ID:                   "config-3",
		Description:          "Config with Locations",
		FirmwareFilename:     "firmware.bin",
		FirmwareVersion:      "3.0.0",
		FirmwareLocation:     "http://cdn.example.com/firmware",
		Ipv6FirmwareLocation: "http://[2001:db8::1]/firmware",
	}
	
	assert.Equal(t, "http://cdn.example.com/firmware", config.FirmwareLocation)
	assert.Equal(t, "http://[2001:db8::1]/firmware", config.Ipv6FirmwareLocation)
}

func TestFirmwareConfig_WithProtocol(t *testing.T) {
	config := &FirmwareConfig{
		ID:                       "config-4",
		Description:              "Config with Protocol",
		FirmwareFilename:         "firmware.bin",
		FirmwareVersion:          "4.0.0",
		FirmwareDownloadProtocol: "https",
	}
	
	assert.Equal(t, "https", config.FirmwareDownloadProtocol)
}

func TestFirmwareConfig_WithFlags(t *testing.T) {
	config := &FirmwareConfig{
		ID:                "config-5",
		Description:       "Config with Flags",
		FirmwareFilename:  "firmware.bin",
		FirmwareVersion:   "5.0.0",
		RebootImmediately: true,
		MandatoryUpdate:   true,
	}
	
	assert.Assert(t, config.RebootImmediately)
	assert.Assert(t, config.MandatoryUpdate)
}

func TestFirmwareConfig_WithUpgradeDelay(t *testing.T) {
	config := &FirmwareConfig{
		ID:               "config-6",
		Description:      "Config with Delay",
		FirmwareFilename: "firmware.bin",
		FirmwareVersion:  "6.0.0",
		UpgradeDelay:     3600,
	}
	
	assert.Equal(t, int64(3600), config.UpgradeDelay)
}

func TestFirmwareConfig_WithProperties(t *testing.T) {
	properties := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	
	config := &FirmwareConfig{
		ID:               "config-7",
		Description:      "Config with Properties",
		FirmwareFilename: "firmware.bin",
		FirmwareVersion:  "7.0.0",
		Properties:       properties,
	}
	
	assert.Equal(t, 2, len(config.Properties))
	assert.Equal(t, "value1", config.Properties["key1"])
	assert.Equal(t, "value2", config.Properties["key2"])
}

func TestFirmwareConfig_Clone(t *testing.T) {
	original := &FirmwareConfig{
		ID:                "original-id",
		Description:       "Original Config",
		FirmwareFilename:  "firmware.bin",
		FirmwareVersion:   "1.0.0",
		SupportedModelIds: []string{"MODEL-X"},
		ApplicationType:   "stb",
	}
	
	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Description, cloned.Description)
	assert.Equal(t, original.FirmwareFilename, cloned.FirmwareFilename)
	assert.Equal(t, original.FirmwareVersion, cloned.FirmwareVersion)
	
	// Verify it's a deep copy
	assert.Assert(t, original != cloned)
}

func TestFirmwareConfig_Validate_NilConfig(t *testing.T) {
	var config *FirmwareConfig = nil
	
	err := config.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Firmware config is not present")
}

func TestFirmwareConfig_Validate_EmptyDescription(t *testing.T) {
	config := &FirmwareConfig{
		ID:                "config-empty-desc",
		Description:       "",
		FirmwareFilename:  "firmware.bin",
		FirmwareVersion:   "1.0.0",
		SupportedModelIds: []string{"MODEL-A"},
	}
	
	err := config.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Description is empty")
}

func TestFirmwareConfig_Validate_EmptyFilename(t *testing.T) {
	config := &FirmwareConfig{
		ID:                "config-empty-filename",
		Description:       "Valid Description",
		FirmwareFilename:  "",
		FirmwareVersion:   "1.0.0",
		SupportedModelIds: []string{"MODEL-A"},
	}
	
	err := config.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "File name is empty")
}

func TestFirmwareConfig_Validate_EmptyVersion(t *testing.T) {
	config := &FirmwareConfig{
		ID:                "config-empty-version",
		Description:       "Valid Description",
		FirmwareFilename:  "firmware.bin",
		FirmwareVersion:   "",
		SupportedModelIds: []string{"MODEL-A"},
	}
	
	err := config.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Version is empty")
}

func TestFirmwareConfig_Validate_EmptySupportedModels(t *testing.T) {
	config := &FirmwareConfig{
		ID:                "config-empty-models",
		Description:       "Valid Description",
		FirmwareFilename:  "firmware.bin",
		FirmwareVersion:   "1.0.0",
		SupportedModelIds: []string{},
	}
	
	err := config.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Supported model list is empty")
}

func TestFirmwareConfigToFirmwareConfigForMacRuleBeanResponse(t *testing.T) {
	properties := map[string]string{"key": "value"}
	config := &FirmwareConfig{
		ID:                       "fw-1",
		Updated:                  123456789,
		Description:              "Test Config",
		SupportedModelIds:        []string{"MODEL-1"},
		FirmwareFilename:         "firmware.bin",
		FirmwareVersion:          "1.0.0",
		ApplicationType:          "stb",
		FirmwareDownloadProtocol: "https",
		FirmwareLocation:         "http://example.com",
		Ipv6FirmwareLocation:     "http://[::1]",
		UpgradeDelay:             3600,
		RebootImmediately:        true,
		MandatoryUpdate:          false,
		Properties:               properties,
	}
	
	response := FirmwareConfigToFirmwareConfigForMacRuleBeanResponse(config)
	
	assert.Assert(t, response != nil)
	assert.Equal(t, "fw-1", response.ID)
	assert.Equal(t, int64(123456789), response.Updated)
	assert.Equal(t, "Test Config", response.Description)
	assert.Equal(t, 1, len(response.SupportedModelIds))
	assert.Equal(t, "firmware.bin", response.FirmwareFilename)
	assert.Equal(t, "1.0.0", response.FirmwareVersion)
	assert.Equal(t, "stb", response.ApplicationType)
	assert.Equal(t, "https", response.FirmwareDownloadProtocol)
	assert.Equal(t, "http://example.com", response.FirmwareLocation)
	assert.Equal(t, "http://[::1]", response.Ipv6FirmwareLocation)
	assert.Equal(t, int64(3600), response.UpgradeDelay)
	assert.Assert(t, response.RebootImmediately)
	assert.Assert(t, !response.MandatoryUpdate)
	assert.Equal(t, 1, len(response.Properties))
}

func TestMacRuleBeanToMacRuleBeanResponse_WithFirmwareConfig(t *testing.T) {
	firmwareConfig := &FirmwareConfig{
		ID:               "fw-config-1",
		Description:      "Firmware Config",
		FirmwareFilename: "firmware.bin",
		FirmwareVersion:  "2.0.0",
	}
	
	targetedModels := []string{"MODEL-A", "MODEL-B"}
	macList := []string{"AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66"}
	
	macRuleBean := &MacRuleBean{
		Id:               "rule-1",
		Name:             "Test Rule",
		MacAddresses:     "AA:BB:CC:DD:EE:FF",
		MacListRef:       "mac-list-ref",
		FirmwareConfig:   firmwareConfig,
		TargetedModelIds: &targetedModels,
		MacList:          &macList,
	}
	
	response := MacRuleBeanToMacRuleBeanResponse(macRuleBean)
	
	assert.Assert(t, response != nil)
	assert.Equal(t, "rule-1", response.Id)
	assert.Equal(t, "Test Rule", response.Name)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", response.MacAddresses)
	assert.Equal(t, "mac-list-ref", response.MacListRef)
	assert.Assert(t, response.TargetedModelIds != nil)
	assert.Equal(t, 2, len(*response.TargetedModelIds))
	assert.Assert(t, response.MacList != nil)
	assert.Equal(t, 2, len(*response.MacList))
	assert.Assert(t, response.FirmwareConfig != nil)
	assert.Equal(t, "fw-config-1", response.FirmwareConfig.ID)
}

func TestMacRuleBeanToMacRuleBeanResponse_WithoutFirmwareConfig(t *testing.T) {
	macRuleBean := &MacRuleBean{
		Id:             "rule-2",
		Name:           "Test Rule Without Config",
		MacAddresses:   "11:22:33:44:55:66",
		FirmwareConfig: nil,
	}
	
	response := MacRuleBeanToMacRuleBeanResponse(macRuleBean)
	
	assert.Assert(t, response != nil)
	assert.Equal(t, "rule-2", response.Id)
	assert.Equal(t, "Test Rule Without Config", response.Name)
	assert.Assert(t, response.FirmwareConfig == nil)
}

func TestMacRuleBean_Creation(t *testing.T) {
	bean := &MacRuleBean{
		Id:           "mac-rule-1",
		Name:         "Test MAC Rule",
		MacAddresses: "AA:BB:CC:DD:EE:FF",
	}
	
	assert.Equal(t, "mac-rule-1", bean.Id)
	assert.Equal(t, "Test MAC Rule", bean.Name)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", bean.MacAddresses)
}

func TestMacRuleBean_WithMacListRef(t *testing.T) {
	bean := &MacRuleBean{
		Id:         "mac-rule-2",
		MacListRef: "mac-list-ref-123",
	}
	
	assert.Equal(t, "mac-list-ref-123", bean.MacListRef)
}

func TestMacRuleBean_WithTargetedModelIds(t *testing.T) {
	targetedModels := []string{"MODEL-A", "MODEL-B"}
	bean := &MacRuleBean{
		Id:               "mac-rule-3",
		TargetedModelIds: &targetedModels,
	}
	
	assert.Assert(t, bean.TargetedModelIds != nil)
	assert.Equal(t, 2, len(*bean.TargetedModelIds))
	assert.Equal(t, "MODEL-A", (*bean.TargetedModelIds)[0])
}

func TestMacRuleBean_WithMacList(t *testing.T) {
	macList := []string{"AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66"}
	bean := &MacRuleBean{
		Id:      "mac-rule-4",
		MacList: &macList,
	}
	
	assert.Assert(t, bean.MacList != nil)
	assert.Equal(t, 2, len(*bean.MacList))
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", (*bean.MacList)[0])
}

func TestEnvModelBean_Creation(t *testing.T) {
	bean := &EnvModelBean{
		Id:            "env-model-1",
		Name:          "Test Env Model",
		EnvironmentId: "PROD",
		ModelId:       "MODEL-X",
	}
	
	assert.Equal(t, "env-model-1", bean.Id)
	assert.Equal(t, "Test Env Model", bean.Name)
	assert.Equal(t, "PROD", bean.EnvironmentId)
	assert.Equal(t, "MODEL-X", bean.ModelId)
	assert.Assert(t, !bean.Noop)
}

func TestEnvModelBean_WithFirmwareConfig(t *testing.T) {
	config := &FirmwareConfig{
		ID:          "config-123",
		Description: "Firmware Config",
	}
	
	bean := &EnvModelBean{
		Id:             "env-model-2",
		FirmwareConfig: config,
	}
	
	assert.Assert(t, bean.FirmwareConfig != nil)
	assert.Equal(t, "config-123", bean.FirmwareConfig.ID)
}

func TestEnvModelBean_WithNoop(t *testing.T) {
	bean := &EnvModelBean{
		Id:   "env-model-3",
		Noop: true,
	}
	
	assert.Assert(t, bean.Noop)
}

func TestIpRuleBean_Creation(t *testing.T) {
	bean := &IpRuleBean{
		Id:   "ip-rule-1",
		Name: "Test IP Rule",
	}
	
	assert.Equal(t, "ip-rule-1", bean.Id)
	assert.Equal(t, "Test IP Rule", bean.Name)
	assert.Assert(t, !bean.Noop)
}

func TestIpRuleBean_WithFirmwareConfig(t *testing.T) {
	config := &FirmwareConfig{
		ID:          "fw-1",
		Description: "Firmware",
	}
	
	bean := &IpRuleBean{
		Id:             "ip-rule-2",
		FirmwareConfig: config,
	}
	
	assert.Assert(t, bean.FirmwareConfig != nil)
	assert.Equal(t, "fw-1", bean.FirmwareConfig.ID)
}

func TestIpRuleBean_WithIpAddressGroup(t *testing.T) {
	ipGroup := &shared.IpAddressGroup{
		Id:   "ip-group-1",
		Name: "IP Group",
	}
	
	bean := &IpRuleBean{
		Id:             "ip-rule-3",
		IpAddressGroup: ipGroup,
	}
	
	assert.Assert(t, bean.IpAddressGroup != nil)
	assert.Equal(t, "ip-group-1", bean.IpAddressGroup.Id)
}

func TestIpRuleBean_WithEnvironmentAndModel(t *testing.T) {
	bean := &IpRuleBean{
		Id:            "ip-rule-4",
		EnvironmentId: "PROD",
		ModelId:       "MODEL-X",
	}
	
	assert.Equal(t, "PROD", bean.EnvironmentId)
	assert.Equal(t, "MODEL-X", bean.ModelId)
}

func TestIpRuleBean_WithNoop(t *testing.T) {
	bean := &IpRuleBean{
		Id:   "ip-rule-5",
		Noop: true,
	}
	
	assert.Assert(t, bean.Noop)
}

func TestNewDefaulttFirmwareConfigFacade(t *testing.T) {
	facade := NewDefaulttFirmwareConfigFacade()
	
	assert.Assert(t, facade != nil)
	assert.Assert(t, facade.Properties != nil)
	assert.Equal(t, 0, len(facade.Properties))
}

func TestNewFirmwareConfigFacadeEmptyProperties(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	
	assert.Assert(t, facade != nil)
	assert.Assert(t, facade.Properties != nil)
	assert.Equal(t, 0, len(facade.Properties))
}

func TestFirmwareConfig_ToPropertiesMap(t *testing.T) {
	config := &FirmwareConfig{
		ID:                       "config-1",
		Description:              "Test Config",
		FirmwareFilename:         "firmware.bin",
		FirmwareVersion:          "1.0.0",
		SupportedModelIds:        []string{"MODEL-A"},
		FirmwareDownloadProtocol: "https",
		FirmwareLocation:         "http://example.com",
		Ipv6FirmwareLocation:     "http://[::1]",
		UpgradeDelay:             3600,
		RebootImmediately:        true,
		MandatoryUpdate:          false,
		Updated:                  123456789,
	}
	
	propsMap := config.ToPropertiesMap()
	
	assert.Assert(t, propsMap != nil)
	assert.Assert(t, len(propsMap) > 0)
	assert.Equal(t, "config-1", propsMap["id"])
	assert.Equal(t, "Test Config", propsMap["description"])
	assert.Equal(t, "firmware.bin", propsMap["firmwareFilename"])
	assert.Equal(t, "1.0.0", propsMap["firmwareVersion"])
	assert.Equal(t, "https", propsMap["firmwareDownloadProtocol"])
}

func TestFirmwareConfig_ToFirmwareConfigResponseMap(t *testing.T) {
	config := &FirmwareConfig{
		ID:                "config-2",
		Description:       "Response Config",
		FirmwareFilename:  "firmware-v2.bin",
		FirmwareVersion:   "2.0.0",
		SupportedModelIds: []string{"MODEL-B", "MODEL-C"},
	}
	
	responseMap := config.ToFirmwareConfigResponseMap()
	
	assert.Assert(t, responseMap != nil)
	assert.Equal(t, "config-2", responseMap["id"])
	assert.Equal(t, "Response Config", responseMap["description"])
	assert.Equal(t, "firmware-v2.bin", responseMap["firmwareFilename"])
	assert.Equal(t, "2.0.0", responseMap["firmwareVersion"])
	
	// Should not have other fields like FirmwareLocation
	_, hasLocation := responseMap["firmwareLocation"]
	assert.Assert(t, !hasLocation)
}

func TestEnvModelBean_AllFields(t *testing.T) {
	config := &FirmwareConfig{
		ID:               "fw-2",
		Description:      "Complete Config",
		FirmwareVersion:  "1.0.0",
		FirmwareFilename: "firmware.bin",
	}
	
	bean := &EnvModelBean{
		Id:             "complete-env-model",
		Name:           "Complete Env Model Bean",
		EnvironmentId:  "QA",
		ModelId:        "MODEL-Y",
		FirmwareConfig: config,
		Noop:           false,
	}
	
	assert.Equal(t, "complete-env-model", bean.Id)
	assert.Equal(t, "Complete Env Model Bean", bean.Name)
	assert.Equal(t, "QA", bean.EnvironmentId)
	assert.Equal(t, "MODEL-Y", bean.ModelId)
	assert.Assert(t, bean.FirmwareConfig != nil)
	assert.Equal(t, "1.0.0", bean.FirmwareConfig.FirmwareVersion)
	assert.Assert(t, !bean.Noop)
}

func TestExpression_Creation(t *testing.T) {
	expr := &Expression{
		TargetedModelIds: []string{"MODEL-A", "MODEL-B"},
		EnvironmentId:    "PROD",
		ModelId:          "MODEL-X",
	}
	
	assert.Equal(t, 2, len(expr.TargetedModelIds))
	assert.Equal(t, "PROD", expr.EnvironmentId)
	assert.Equal(t, "MODEL-X", expr.ModelId)
}

func TestExpression_WithIpAddressGroup(t *testing.T) {
	ipGroup := &shared.IpAddressGroup{
		Id:   "ip-group-1",
		Name: "IP Group",
	}
	
	expr := &Expression{
		TargetedModelIds: []string{"MODEL-C"},
		IpAddressGroup:   ipGroup,
	}
	
	assert.Assert(t, expr.IpAddressGroup != nil)
	assert.Equal(t, "ip-group-1", expr.IpAddressGroup.Id)
}

func TestExpression_Empty(t *testing.T) {
	expr := &Expression{}
	
	assert.Assert(t, expr.TargetedModelIds == nil || len(expr.TargetedModelIds) == 0)
	assert.Equal(t, "", expr.EnvironmentId)
	assert.Equal(t, "", expr.ModelId)
	assert.Assert(t, expr.IpAddressGroup == nil)
}

func TestFirmwareConfigForMacRuleBeanResponse_Creation(t *testing.T) {
	response := &FirmwareConfigForMacRuleBeanResponse{
		ID:                "response-1",
		Description:       "Response Config",
		FirmwareFilename:  "firmware.bin",
		FirmwareVersion:   "1.0.0",
		SupportedModelIds: []string{"MODEL-A"},
	}
	
	assert.Equal(t, "response-1", response.ID)
	assert.Equal(t, "Response Config", response.Description)
	assert.Equal(t, "firmware.bin", response.FirmwareFilename)
	assert.Equal(t, "1.0.0", response.FirmwareVersion)
	assert.Equal(t, 1, len(response.SupportedModelIds))
}

// Tests for utility functions
func TestNewEmptyFirmwareConfig(t *testing.T) {
	config := NewEmptyFirmwareConfig()
	
	assert.Assert(t, config != nil)
	assert.Equal(t, "", config.ID)
	assert.Equal(t, "", config.Description)
}

func TestFirmwareConfigFacade_GetSetters(t *testing.T) {
	config := &FirmwareConfig{
		ID:               "config-123",
		FirmwareVersion:  "2.5.0",
		FirmwareFilename: "firmware.bin",
	}
	
	facade := NewFirmwareConfigFacade(config)
	
	// Test getters
	assert.Equal(t, "2.5.0", facade.GetFirmwareVersion())
	assert.Equal(t, "firmware.bin", facade.GetFirmwareFilename())
	
	// Test FirmwareLocation getter/setter
	facade.SetFirmwareLocation("http://example.com/firmware.bin")
	assert.Equal(t, "http://example.com/firmware.bin", facade.GetFirmwareLocation())
	
	// Test RebootImmediately getter/setter
	facade.SetRebootImmediately(true)
	assert.Assert(t, facade.GetRebootImmediately())
	
	facade.SetRebootImmediately(false)
	assert.Assert(t, !facade.GetRebootImmediately())
}

func TestFirmwareConfigFacade_Ipv6FirmwareLocation(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	
	// Initially should be empty
	assert.Equal(t, "", facade.GetIpv6FirmwareLocation())
	
	// Set via properties and verify
	facade.Properties["ipv6FirmwareLocation"] = "http://[::1]/firmware.bin"
	assert.Equal(t, "http://[::1]/firmware.bin", facade.GetIpv6FirmwareLocation())
}

func TestFirmwareConfigFacade_UpgradeDelay(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	
	// Test with upgrade delay set
	facade.Properties["upgradeDelay"] = 300
	assert.Equal(t, 300, facade.GetUpgradeDelay())
	
	// Test with nil value
	facade.Properties["upgradeDelay"] = nil
	assert.Equal(t, 0, facade.GetUpgradeDelay())
	
	// Test with invalid type
	facade.Properties["upgradeDelay"] = "not-an-int"
	assert.Equal(t, 0, facade.GetUpgradeDelay())
}

func TestFirmwareConfigFacade_FirmwareDownloadProtocol(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	
	// Test getter
	assert.Equal(t, "", facade.GetFirmwareDownloadProtocol())
	
	// Test setter
	facade.SetFirmwareDownloadProtocol("https")
	assert.Equal(t, "https", facade.GetFirmwareDownloadProtocol())
}

func TestFirmwareConfigFacade_Creation(t *testing.T) {
	config := &FirmwareConfig{
		ID:               "test-id",
		Description:      "Test Config",
		FirmwareVersion:  "1.0.0",
		FirmwareFilename: "test.bin",
	}
	
	facade := NewFirmwareConfigFacade(config)
	
	assert.Assert(t, facade != nil)
	assert.Assert(t, facade.Properties != nil)
	assert.Equal(t, "1.0.0", facade.GetFirmwareVersion())
	assert.Equal(t, "test.bin", facade.GetFirmwareFilename())
}

func TestFirmwareConfigFacade_EmptyProperties(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	
	assert.Assert(t, facade != nil)
	assert.Assert(t, facade.Properties != nil)
	assert.Equal(t, 0, len(facade.Properties))
}

func TestFirmwareConfigFacade_Default(t *testing.T) {
	facade := NewDefaulttFirmwareConfigFacade()
	
	assert.Assert(t, facade != nil)
	assert.Assert(t, facade.Properties != nil)
}

func TestFirmwareConfigFacade_PutIfPresent(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	
	// Test with non-empty string
	facade.PutIfPresent("key1", "value1")
	assert.Equal(t, "value1", facade.Properties["key1"])
	
	// Test with empty string - should not be added
	facade.PutIfPresent("key2", "")
	_, exists := facade.Properties["key2"]
	assert.Assert(t, !exists)
}

func TestModelFirmwareConfiguration_Creation(t *testing.T) {
	mfc := NewModelFirmwareConfiguration("MODEL-A", "firmware.bin", "1.0.0")
	
	assert.Assert(t, mfc != nil)
	assert.Equal(t, "MODEL-A", mfc.Model)
	assert.Equal(t, "firmware.bin", mfc.FirmwareFilename)
	assert.Equal(t, "1.0.0", mfc.FirmwareVersion)
}

func TestModelFirmwareConfiguration_ToString(t *testing.T) {
	mfc := &ModelFirmwareConfiguration{
		Model:            "MODEL-X",
		FirmwareFilename: "test.bin",
		FirmwareVersion:  "2.0.0",
	}
	
	result := mfc.ToString()
	
	assert.Assert(t, result != "")
	// Should contain model, filename, and version
	assert.Assert(t, len(result) > 0)
}
