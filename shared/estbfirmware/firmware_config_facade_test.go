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

	"gotest.tools/assert"
)

func TestFirmwareConfigFacade_MarshalJSON(t *testing.T) {
	config := &FirmwareConfig{
		ID:               "config-123",
		Description:      "Test Config",
		FirmwareVersion:  "2.0.0",
		FirmwareFilename: "firmware.bin",
		FirmwareLocation: "http://example.com/firmware.bin",
	}
	
	facade := NewFirmwareConfigFacade(config)
	
	jsonBytes, err := json.Marshal(facade)
	
	assert.NilError(t, err)
	assert.Assert(t, len(jsonBytes) > 0)
	
	// Unmarshal to verify structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NilError(t, err)
	
	// Should contain firmware fields
	assert.Equal(t, "2.0.0", result["firmwareVersion"])
	assert.Equal(t, "firmware.bin", result["firmwareFilename"])
}

func TestFirmwareConfigFacade_MarshalJSON_ExcludesEmptyValues(t *testing.T) {
	config := &FirmwareConfig{
		FirmwareVersion:  "1.0.0",
		FirmwareFilename: "test.bin",
		// Other fields left empty
	}
	
	facade := NewFirmwareConfigFacade(config)
	
	jsonBytes, err := json.Marshal(facade)
	
	assert.NilError(t, err)
	
	// Unmarshal to verify structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NilError(t, err)
	
	// Should only contain non-empty fields
	assert.Equal(t, "1.0.0", result["firmwareVersion"])
	assert.Equal(t, "test.bin", result["firmwareFilename"])
	
	// Empty fields should not be present
	_, hasLocation := result["firmwareLocation"]
	assert.Assert(t, !hasLocation)
}

func TestFirmwareConfigFacade_MarshalJSON_WithUpgradeDelay(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	facade.Properties["upgradeDelay"] = int64(300)
	facade.Properties["firmwareVersion"] = "1.0.0"
	
	jsonBytes, err := json.Marshal(facade)
	
	assert.NilError(t, err)
	
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NilError(t, err)
	
	// upgradeDelay should be included when non-zero
	assert.Assert(t, result["upgradeDelay"] != nil)
}

func TestFirmwareConfigFacade_MarshalJSON_SkipsZeroUpgradeDelay(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	facade.Properties["upgradeDelay"] = int64(0)
	facade.Properties["firmwareVersion"] = "1.0.0"
	
	jsonBytes, err := json.Marshal(facade)
	
	assert.NilError(t, err)
	
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NilError(t, err)
	
	// upgradeDelay should not be included when zero
	_, hasUpgradeDelay := result["upgradeDelay"]
	assert.Assert(t, !hasUpgradeDelay)
}

func TestFirmwareConfigFacade_UnmarshalJSON(t *testing.T) {
	jsonStr := `{
		"firmwareVersion": "3.0.0",
		"firmwareFilename": "new-firmware.bin",
		"firmwareLocation": "http://example.com/new.bin"
	}`
	
	facade := NewFirmwareConfigFacadeEmptyProperties()
	err := json.Unmarshal([]byte(jsonStr), facade)
	
	assert.NilError(t, err)
	assert.Assert(t, facade.Properties != nil)
	assert.Equal(t, "3.0.0", facade.GetFirmwareVersion())
	assert.Equal(t, "new-firmware.bin", facade.GetFirmwareFilename())
	assert.Equal(t, "http://example.com/new.bin", facade.GetFirmwareLocation())
}

func TestFirmwareConfigFacade_UnmarshalJSON_Invalid(t *testing.T) {
	jsonStr := `{invalid json`
	
	facade := NewFirmwareConfigFacadeEmptyProperties()
	err := json.Unmarshal([]byte(jsonStr), facade)
	
	assert.Assert(t, err != nil)
}

func TestFirmwareConfigFacade_PutIfPresent_WithEmptyString(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	
	// Empty string should not be added
	facade.PutIfPresent("emptyKey", "")
	_, exists := facade.Properties["emptyKey"]
	assert.Assert(t, !exists)
}

func TestFirmwareConfigFacade_PutIfPresent_WithNil(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	
	// Nil should not be added
	facade.PutIfPresent("nilKey", nil)
	_, exists := facade.Properties["nilKey"]
	assert.Assert(t, !exists)
}

func TestFirmwareConfigFacade_PutIfPresent_WithValidValue(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	
	// Valid value should be added
	facade.PutIfPresent("validKey", "validValue")
	assert.Equal(t, "validValue", facade.Properties["validKey"])
}

func TestFirmwareConfigFacade_GetStringValue(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	facade.Properties["testKey"] = "testValue"
	
	value := facade.GetStringValue("testKey")
	assert.Equal(t, "testValue", value)
	
	// Non-existent key should return empty string
	value2 := facade.GetStringValue("nonExistent")
	assert.Equal(t, "", value2)
}

func TestFirmwareConfigFacade_SetStringValue(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	
	facade.SetStringValue("newKey", "newValue")
	assert.Equal(t, "newValue", facade.Properties["newKey"])
}

func TestFirmwareConfigFacade_GetValue(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	facade.Properties["intKey"] = 42
	facade.Properties["boolKey"] = true
	
	intVal := facade.GetValue("intKey")
	assert.Equal(t, 42, intVal)
	
	boolVal := facade.GetValue("boolKey")
	assert.Equal(t, true, boolVal)
	
	// Non-existent key should return nil
	nilVal := facade.GetValue("nonExistent")
	assert.Assert(t, nilVal == nil)
}

func TestFirmwareConfigFacade_PutAll(t *testing.T) {
	facade := NewFirmwareConfigFacadeEmptyProperties()
	
	newProps := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": 123,
	}
	
	facade.PutAll(newProps)
	
	assert.Equal(t, "value1", facade.Properties["key1"])
	assert.Equal(t, "value2", facade.Properties["key2"])
	assert.Equal(t, 123, facade.Properties["key3"])
}

func TestBseConfiguration_ToString(t *testing.T) {
	bse := &BseConfiguration{
		Location:     "http://example.com",
		Ipv6Location: "http://[::1]",
		Protocol:     "https",
	}
	
	result := bse.ToString()
	
	assert.Assert(t, result != "")
	assert.Assert(t, len(result) > 0)
}

func TestMinimumFirmwareCheckBean_Creation(t *testing.T) {
	bean := &MinimumFirmwareCheckBean{
		HasMinimumFirmware: true,
	}
	
	assert.Assert(t, bean.HasMinimumFirmware)
	
	bean2 := &MinimumFirmwareCheckBean{
		HasMinimumFirmware: false,
	}
	
	assert.Assert(t, !bean2.HasMinimumFirmware)
}
