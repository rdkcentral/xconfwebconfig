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
package shared

import (
	"testing"

	"gotest.tools/assert"
)

func TestValidateApplicationType(t *testing.T) {
	err := ValidateApplicationType("")
	assert.Assert(t, err != nil)

	err = ValidateApplicationType(STB)
	assert.Assert(t, err != nil)
}

// Test Environment struct
func TestEnvironment_NewEnvironment(t *testing.T) {
	env := NewEnvironment("test-id", "Test Environment")

	assert.Equal(t, "TEST-ID", env.ID) // Should be uppercase
	assert.Equal(t, "Test Environment", env.Description)
}

func TestEnvironment_NewEnvironmentWithWhitespace(t *testing.T) {
	env := NewEnvironment("  test-id  ", "Test Description")

	assert.Equal(t, "TEST-ID", env.ID) // Should trim and uppercase
	assert.Equal(t, "Test Description", env.Description)
}

func TestEnvironment_NewEnvironmentEmpty(t *testing.T) {
	env := NewEnvironment("", "Description")

	assert.Equal(t, "", env.ID)
	assert.Equal(t, "Description", env.Description)
}

func TestEnvironment_Clone(t *testing.T) {
	original := NewEnvironment("ENV-1", "Original Environment")
	original.Updated = 12345

	cloned, err := original.Clone()

	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Description, cloned.Description)
	assert.Equal(t, original.Updated, cloned.Updated)

	// Verify it's a deep copy
	cloned.ID = "MODIFIED"
	assert.Assert(t, original.ID != cloned.ID)
}

func TestEnvironment_Validate_Valid(t *testing.T) {
	env := NewEnvironment("VALID-ENV_123", "Valid")
	err := env.Validate()
	assert.NilError(t, err)
}

func TestEnvironment_Validate_ValidWithSpaces(t *testing.T) {
	env := NewEnvironment("ENV 123", "Valid with spaces")
	err := env.Validate()
	assert.NilError(t, err)
}

func TestEnvironment_Validate_ValidWithDots(t *testing.T) {
	env := NewEnvironment("ENV.123.TEST", "Valid with dots")
	err := env.Validate()
	assert.NilError(t, err)
}

func TestEnvironment_Validate_ValidWithApostrophe(t *testing.T) {
	env := NewEnvironment("ENV'S_TEST", "Valid with apostrophe")
	err := env.Validate()
	assert.NilError(t, err)
}

func TestEnvironment_Validate_Invalid(t *testing.T) {
	env := &Environment{ID: "invalid@env#123", Description: "Invalid"}
	err := env.Validate()
	assert.Error(t, err, "Id is invalid")
}

func TestEnvironment_Validate_Empty(t *testing.T) {
	env := &Environment{ID: "", Description: "Empty"}
	err := env.Validate()
	assert.Error(t, err, "Id is invalid")
}

func TestEnvironment_Validate_OnlyWhitespace(t *testing.T) {
	env := &Environment{ID: "   ", Description: "Whitespace"}
	err := env.Validate()
	assert.Error(t, err, "Id is invalid")
}

func TestEnvironment_CreateEnvironmentResponse(t *testing.T) {
	env := NewEnvironment("ENV-1", "Test Environment")

	response := env.CreateEnvironmentResponse()

	assert.Assert(t, response != nil)
	assert.Equal(t, env.ID, response.ID)
	assert.Equal(t, env.Description, response.Description)
}

func TestNewEnvironmentInf(t *testing.T) {
	obj := NewEnvironmentInf()

	assert.Assert(t, obj != nil)
	env, ok := obj.(*Environment)
	assert.Assert(t, ok)
	assert.Assert(t, env != nil)
}

// Test Model struct
func TestModel_NewModel(t *testing.T) {
	model := NewModel("test-model", "Test Model")

	assert.Equal(t, "TEST-MODEL", model.ID) // Should be uppercase
	assert.Equal(t, "Test Model", model.Description)
}

func TestModel_Clone(t *testing.T) {
	original := NewModel("MODEL-1", "Original Model")
	original.Updated = 12345

	cloned, err := original.Clone()

	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Description, cloned.Description)
	assert.Equal(t, original.Updated, cloned.Updated)

	// Verify it's a deep copy
	cloned.ID = "MODIFIED"
	assert.Assert(t, original.ID != cloned.ID)
}

func TestModel_Validate_Valid(t *testing.T) {
	model := NewModel("VALID-MODEL_123", "Valid")
	err := model.Validate()
	assert.NilError(t, err)
}

func TestModel_Validate_ValidWithSpaces(t *testing.T) {
	model := NewModel("MODEL 123", "Valid with spaces")
	err := model.Validate()
	assert.NilError(t, err)
}

func TestModel_Validate_ValidWithDots(t *testing.T) {
	model := NewModel("MODEL.123.TEST", "Valid with dots")
	err := model.Validate()
	assert.NilError(t, err)
}

func TestModel_Validate_ValidWithApostrophe(t *testing.T) {
	model := NewModel("MODEL'S_TEST", "Valid with apostrophe")
	err := model.Validate()
	assert.NilError(t, err)
}

func TestModel_Validate_Invalid(t *testing.T) {
	model := &Model{ID: "invalid@model#123", Description: "Invalid"}
	err := model.Validate()
	assert.Error(t, err, "Id is invalid. Valid Characters: alphanumeric _ . -")
}

func TestModel_Validate_Empty(t *testing.T) {
	model := &Model{ID: "", Description: "Empty"}
	err := model.Validate()
	assert.Error(t, err, "Id is invalid. Valid Characters: alphanumeric _ . -")
}

func TestModel_CreateModelResponse(t *testing.T) {
	model := NewModel("MODEL-1", "Test Model")

	response := model.CreateModelResponse()

	assert.Assert(t, response != nil)
	assert.Equal(t, model.ID, response.ID)
	assert.Equal(t, model.Description, response.Description)
}

func TestNewModelInf(t *testing.T) {
	obj := NewModelInf()

	assert.Assert(t, obj != nil)
	model, ok := obj.(*Model)
	assert.Assert(t, ok)
	assert.Assert(t, model != nil)
}

// Test StringListWrapper
func TestNewStringListWrapper(t *testing.T) {
	list := []string{"item1", "item2", "item3"}
	wrapper := NewStringListWrapper(list)

	assert.Assert(t, wrapper != nil)
	assert.Equal(t, 3, len(wrapper.List))
	assert.Equal(t, "item1", wrapper.List[0])
}

func TestNewStringListWrapper_Empty(t *testing.T) {
	list := []string{}
	wrapper := NewStringListWrapper(list)

	assert.Assert(t, wrapper != nil)
	assert.Equal(t, 0, len(wrapper.List))
}

func TestNewStringListWrapper_Nil(t *testing.T) {
	wrapper := NewStringListWrapper(nil)

	assert.Assert(t, wrapper != nil)
	assert.Assert(t, wrapper.List == nil)
}

// Test AppSetting struct
func TestAppSetting_Clone(t *testing.T) {
	original := &AppSetting{
		ID:      "test-key",
		Updated: 12345,
		Value:   "test-value",
	}

	cloned, err := original.Clone()

	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, original.Value, cloned.Value)

	// Verify it's a deep copy
	cloned.ID = "modified"
	assert.Assert(t, original.ID != cloned.ID)
}

func TestNewAppSettingInf(t *testing.T) {
	obj := NewAppSettingInf()

	assert.Assert(t, obj != nil)
	setting, ok := obj.(*AppSetting)
	assert.Assert(t, ok)
	assert.Assert(t, setting != nil)
}

// Test constants
func TestApplicationTypeConstants(t *testing.T) {
	assert.Equal(t, "stb", STB)
	assert.Equal(t, "xhome", XHOME)
	assert.Equal(t, "rdkcloud", RDKCLOUD)
	assert.Equal(t, "sky", SKY)
	assert.Equal(t, "all", ALL)
}

func TestProtocolConstants(t *testing.T) {
	assert.Equal(t, "tftp", Tftp)
	assert.Equal(t, "http", Http)
	assert.Equal(t, "https", Https)
}

func TestTableConstants(t *testing.T) {
	assert.Equal(t, "column1", TABLE_LOGS_KEY2_FIELD_NAME)
	assert.Equal(t, "0", LAST_CONFIG_LOG_ID)
	assert.Equal(t, "time", StbContextTime)
	assert.Equal(t, "model", StbContextModel)
	assert.Equal(t, "MAC_LIST", MacList)
	assert.Equal(t, "IP_LIST", IpList)
	assert.Equal(t, "GenericXconfNamedList", TableGenericNSList)
	assert.Equal(t, "FirmwareConfig", TableFirmwareConfig)
	assert.Equal(t, "FirmwareRule4", TableFirmwareRule)
}
