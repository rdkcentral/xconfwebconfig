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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/stretchr/testify/assert"
)

// Test GetContextMapAndSettingTypes function

func TestGetContextMapAndSettingTypes_BasicRequest(t *testing.T) {
	// Setup
	req := httptest.NewRequest(http.MethodGet, "/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&env=PROD", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, shared.STB, contextMap[common.APPLICATION_TYPE])
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Equal(t, "PROD", contextMap["env"])
	assert.Empty(t, settingTypes)
}

func TestGetContextMapAndSettingTypes_WithSettingType(t *testing.T) {
	// Setup
	req := httptest.NewRequest(http.MethodGet, "/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&settingType=partnersettings", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, shared.STB, contextMap[common.APPLICATION_TYPE])
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Len(t, settingTypes, 1)
	assert.Equal(t, "partnersettings", settingTypes[0])
}

func TestGetContextMapAndSettingTypes_WithMultipleSettingTypes(t *testing.T) {
	// Setup
	req := httptest.NewRequest(http.MethodGet, "/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&settingType=partnersettings&settingType=epon", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, shared.STB, contextMap[common.APPLICATION_TYPE])
	assert.Len(t, settingTypes, 2)
	assert.Contains(t, settingTypes, "partnersettings")
	assert.Contains(t, settingTypes, "epon")
}

func TestGetContextMapAndSettingTypes_NoApplicationType(t *testing.T) {
	// Setup - no APPLICATION_TYPE in vars
	req := httptest.NewRequest(http.MethodGet, "/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF", nil)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert - should default to STB
	assert.Equal(t, shared.STB, contextMap[common.APPLICATION_TYPE])
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Empty(t, settingTypes)
}

func TestGetContextMapAndSettingTypes_WithXhomeApplicationType(t *testing.T) {
	// Setup
	req := httptest.NewRequest(http.MethodGet, "/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.XHOME,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, shared.XHOME, contextMap[common.APPLICATION_TYPE])
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Empty(t, settingTypes)
}

func TestGetContextMapAndSettingTypes_EmptyQueryParams(t *testing.T) {
	// Setup
	req := httptest.NewRequest(http.MethodGet, "/loguploader/getSettings", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, shared.STB, contextMap[common.APPLICATION_TYPE])
	assert.Empty(t, settingTypes)
	// Only APPLICATION_TYPE should be in the map
	assert.Len(t, contextMap, 1)
}

func TestGetContextMapAndSettingTypes_WithAllCommonParams(t *testing.T) {
	// Setup
	req := httptest.NewRequest(http.MethodGet,
		"/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&env=PROD&model=TestModel&firmwareVersion=1.0.0&partnerId=partner1", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, shared.STB, contextMap[common.APPLICATION_TYPE])
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Equal(t, "PROD", contextMap["env"])
	assert.Equal(t, "TestModel", contextMap["model"])
	assert.Equal(t, "1.0.0", contextMap["firmwareVersion"])
	assert.Equal(t, "partner1", contextMap["partnerId"])
	assert.Empty(t, settingTypes)
}

func TestGetContextMapAndSettingTypes_WithMultipleValuesPerParam(t *testing.T) {
	// Setup - when a param has multiple values, only first is taken (except settingType)
	req := httptest.NewRequest(http.MethodGet,
		"/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&eStbMac=11:22:33:44:55:66&env=PROD&env=DEV", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert - should take first value for each param
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Equal(t, "PROD", contextMap["env"])
	assert.Empty(t, settingTypes)
}

func TestGetContextMapAndSettingTypes_WithSpecialCharacters(t *testing.T) {
	// Setup
	req := httptest.NewRequest(http.MethodGet,
		"/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&partnerId=partner-1_test&model=Model%20123", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Equal(t, "partner-1_test", contextMap["partnerId"])
	assert.Equal(t, "Model 123", contextMap["model"]) // URL decoded
	assert.Empty(t, settingTypes)
}

func TestGetContextMapAndSettingTypes_WithCheckNow(t *testing.T) {
	// Setup
	req := httptest.NewRequest(http.MethodGet,
		"/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&checkNow=true", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Equal(t, "true", contextMap["checkNow"])
	assert.Empty(t, settingTypes)
}

func TestGetContextMapAndSettingTypes_WithUploadImmediately(t *testing.T) {
	// Setup
	req := httptest.NewRequest(http.MethodGet,
		"/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&uploadImmediately=true", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Equal(t, "true", contextMap["uploadImmediately"])
	assert.Empty(t, settingTypes)
}

func TestGetContextMapAndSettingTypes_WithVersion(t *testing.T) {
	// Setup
	req := httptest.NewRequest(http.MethodGet,
		"/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&version=2.1", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Equal(t, "2.1", contextMap["version"])
	assert.Empty(t, settingTypes)
}

func TestGetContextMapAndSettingTypes_OnlySettingTypes(t *testing.T) {
	// Setup - only settingType params
	req := httptest.NewRequest(http.MethodGet,
		"/loguploader/getSettings?settingType=type1&settingType=type2&settingType=type3", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, shared.STB, contextMap[common.APPLICATION_TYPE])
	assert.Len(t, settingTypes, 3)
	assert.Contains(t, settingTypes, "type1")
	assert.Contains(t, settingTypes, "type2")
	assert.Contains(t, settingTypes, "type3")
}

func TestGetContextMapAndSettingTypes_MixedParamsWithSettingTypes(t *testing.T) {
	// Setup
	req := httptest.NewRequest(http.MethodGet,
		"/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&settingType=type1&model=Model123&settingType=type2&env=PROD", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, shared.STB, contextMap[common.APPLICATION_TYPE])
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Equal(t, "Model123", contextMap["model"])
	assert.Equal(t, "PROD", contextMap["env"])
	assert.Len(t, settingTypes, 2)
	assert.Contains(t, settingTypes, "type1")
	assert.Contains(t, settingTypes, "type2")
}

func TestGetContextMapAndSettingTypes_CasePreservation(t *testing.T) {
	// Setup - test that parameter names are case-sensitive
	req := httptest.NewRequest(http.MethodGet,
		"/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&Model=TestModel&ENV=PROD", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert - parameter names should be preserved as-is
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Equal(t, "TestModel", contextMap["Model"])
	assert.Equal(t, "PROD", contextMap["ENV"])
	assert.Empty(t, settingTypes)
}

func TestGetContextMapAndSettingTypes_WithIPAddress(t *testing.T) {
	// Setup
	req := httptest.NewRequest(http.MethodGet,
		"/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&ipAddress=192.168.1.100", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Equal(t, "192.168.1.100", contextMap["ipAddress"])
	assert.Empty(t, settingTypes)
}

func TestGetContextMapAndSettingTypes_WithAccountID(t *testing.T) {
	// Setup
	req := httptest.NewRequest(http.MethodGet,
		"/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&accountId=account123", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Equal(t, "account123", contextMap["accountId"])
	assert.Empty(t, settingTypes)
}

func TestGetContextMapAndSettingTypes_EmptySettingTypeValue(t *testing.T) {
	// Setup - settingType with empty value
	req := httptest.NewRequest(http.MethodGet,
		"/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&settingType=", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	// Empty settingType value should still be included
	assert.Len(t, settingTypes, 1)
	assert.Equal(t, "", settingTypes[0])
}

func TestGetContextMapAndSettingTypes_LongQueryString(t *testing.T) {
	// Setup - test with many parameters
	req := httptest.NewRequest(http.MethodGet,
		"/loguploader/getSettings?eStbMac=AA:BB:CC:DD:EE:FF&env=PROD&model=Model123&"+
			"firmwareVersion=1.0.0&partnerId=partner1&accountId=account123&ipAddress=192.168.1.1&"+
			"controllerId=controller1&channelMapId=channel1&vodId=vod1&settingType=type1&settingType=type2", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, shared.STB, contextMap[common.APPLICATION_TYPE])
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap["eStbMac"])
	assert.Equal(t, "PROD", contextMap["env"])
	assert.Equal(t, "Model123", contextMap["model"])
	assert.Equal(t, "1.0.0", contextMap["firmwareVersion"])
	assert.Equal(t, "partner1", contextMap["partnerId"])
	assert.Equal(t, "account123", contextMap["accountId"])
	assert.Equal(t, "192.168.1.1", contextMap["ipAddress"])
	assert.Equal(t, "controller1", contextMap["controllerId"])
	assert.Equal(t, "channel1", contextMap["channelMapId"])
	assert.Equal(t, "vod1", contextMap["vodId"])
	assert.Len(t, settingTypes, 2)
}

func TestGetContextMapAndSettingTypes_NilQueryParams(t *testing.T) {
	// Setup - URL with ? but no params
	req := httptest.NewRequest(http.MethodGet, "/loguploader/getSettings?", nil)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	// Execute
	contextMap, settingTypes := GetContextMapAndSettingTypes(req)

	// Assert
	assert.Equal(t, shared.STB, contextMap[common.APPLICATION_TYPE])
	assert.Empty(t, settingTypes)
	assert.Len(t, contextMap, 1) // Only APPLICATION_TYPE
}

// Tests for HTTP Handler Functions

func TestGetLogUploaderSettingsHandler(t *testing.T) {
	t.Run("HandlerWrapperExistsAndIsCallable", func(t *testing.T) {
		// Test that the handler wrapper exists and can be referenced
		// The actual execution requires complex setup (DB, SAT tokens, services)
		// which is better tested in integration tests

		assert.NotNil(t, GetLogUploaderSettingsHandler)

		// Verify it's a handler function with correct signature
		var handler func(http.ResponseWriter, *http.Request) = GetLogUploaderSettingsHandler
		assert.NotNil(t, handler)
	})
}

func TestGetLogUploaderT2SettingsHandler(t *testing.T) {
	t.Run("HandlerWrapperExistsAndIsCallable", func(t *testing.T) {
		assert.NotNil(t, GetLogUploaderT2SettingsHandler)

		// Verify it's a handler function with correct signature
		var handler func(http.ResponseWriter, *http.Request) = GetLogUploaderT2SettingsHandler
		assert.NotNil(t, handler)
	})
}

func TestGetLogUploaderTelemetryProfilesHandler(t *testing.T) {
	t.Run("HandlerReturnsErrorWithNonXResponseWriter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/loguploader/getTelemetryProfiles", nil)
		w := httptest.NewRecorder()

		// Call with regular ResponseWriter (not XResponseWriter)
		GetLogUploaderTelemetryProfilesHandler(w, req)

		// Should return 500 error because writer is not XResponseWriter
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestGetLogUploaderSettings(t *testing.T) {
	t.Run("HandlerReturnsErrorWithNonXResponseWriter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/loguploader/getSettings", nil)
		w := httptest.NewRecorder()

		// Call with regular ResponseWriter (not XResponseWriter)
		GetLogUploaderSettings(w, req, false)

		// Should return 500 error because writer is not XResponseWriter
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("HandlerReturnsErrorWithNonXResponseWriterForT2", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/loguploader/getSettings", nil)
		w := httptest.NewRecorder()

		// Call with regular ResponseWriter (not XResponseWriter) and T2 flag
		GetLogUploaderSettings(w, req, true)

		// Should return 500 error because writer is not XResponseWriter
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
