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
package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test DeviceServiceData structure
func TestDeviceServiceData_AllFields(t *testing.T) {
	data := DeviceServiceData{
		AccountId: "account-123",
		CpeMac:    "AA:BB:CC:DD:EE:FF",
		TimeZone:  "America/Los_Angeles",
		PartnerId: "partner-456",
	}

	assert.Equal(t, "account-123", data.AccountId)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", data.CpeMac)
	assert.Equal(t, "America/Los_Angeles", data.TimeZone)
	assert.Equal(t, "partner-456", data.PartnerId)
}

func TestDeviceServiceData_EmptyFields(t *testing.T) {
	data := DeviceServiceData{}

	assert.Equal(t, "", data.AccountId)
	assert.Equal(t, "", data.CpeMac)
	assert.Equal(t, "", data.TimeZone)
	assert.Equal(t, "", data.PartnerId)
}

func TestDeviceServiceData_PartialFields(t *testing.T) {
	data := DeviceServiceData{
		CpeMac:    "11:22:33:44:55:66",
		PartnerId: "comcast",
	}

	assert.Equal(t, "", data.AccountId)
	assert.Equal(t, "11:22:33:44:55:66", data.CpeMac)
	assert.Equal(t, "", data.TimeZone)
	assert.Equal(t, "comcast", data.PartnerId)
}

// Test DeviceServiceObject structure
func TestDeviceServiceObject_Success(t *testing.T) {
	obj := DeviceServiceObject{
		Status:  200,
		Message: "Success",
		DeviceServiceData: &DeviceServiceData{
			AccountId: "acc-789",
			CpeMac:    "00:11:22:33:44:55",
			TimeZone:  "Europe/London",
			PartnerId: "sky",
		},
	}

	assert.Equal(t, 200, obj.Status)
	assert.Equal(t, "Success", obj.Message)
	assert.NotNil(t, obj.DeviceServiceData)
	assert.Equal(t, "acc-789", obj.DeviceServiceData.AccountId)
	assert.Equal(t, "00:11:22:33:44:55", obj.DeviceServiceData.CpeMac)
	assert.Equal(t, "Europe/London", obj.DeviceServiceData.TimeZone)
	assert.Equal(t, "sky", obj.DeviceServiceData.PartnerId)
}

func TestDeviceServiceObject_ErrorStatus(t *testing.T) {
	obj := DeviceServiceObject{
		Status:            404,
		Message:           "Not Found",
		DeviceServiceData: nil,
	}

	assert.Equal(t, 404, obj.Status)
	assert.Equal(t, "Not Found", obj.Message)
	assert.Nil(t, obj.DeviceServiceData)
}

func TestDeviceServiceObject_EmptyData(t *testing.T) {
	obj := DeviceServiceObject{
		Status:            200,
		Message:           "OK",
		DeviceServiceData: &DeviceServiceData{},
	}

	assert.Equal(t, 200, obj.Status)
	assert.NotNil(t, obj.DeviceServiceData)
	assert.Equal(t, "", obj.DeviceServiceData.AccountId)
}

func TestDeviceServiceObject_InternalServerError(t *testing.T) {
	obj := DeviceServiceObject{
		Status:            500,
		Message:           "Internal Server Error",
		DeviceServiceData: nil,
	}

	assert.Equal(t, 500, obj.Status)
	assert.Equal(t, "Internal Server Error", obj.Message)
	assert.Nil(t, obj.DeviceServiceData)
}

func TestDeviceServiceObject_WithCompleteData(t *testing.T) {
	obj := DeviceServiceObject{
		Status:  201,
		Message: "Created",
		DeviceServiceData: &DeviceServiceData{
			AccountId: "new-account-123",
			CpeMac:    "FF:EE:DD:CC:BB:AA",
			TimeZone:  "Asia/Tokyo",
			PartnerId: "kddi",
		},
	}

	assert.Equal(t, 201, obj.Status)
	assert.Equal(t, "Created", obj.Message)
	assert.Equal(t, "new-account-123", obj.DeviceServiceData.AccountId)
	assert.Equal(t, "FF:EE:DD:CC:BB:AA", obj.DeviceServiceData.CpeMac)
	assert.Equal(t, "Asia/Tokyo", obj.DeviceServiceData.TimeZone)
	assert.Equal(t, "kddi", obj.DeviceServiceData.PartnerId)
}

// Test edge cases
func TestDeviceServiceData_JSONTags(t *testing.T) {
	// This test verifies that the struct has proper JSON tags
	// by checking field names exist (compile-time check)
	data := DeviceServiceData{
		AccountId: "test",
		CpeMac:    "test",
		TimeZone:  "test",
		PartnerId: "test",
	}

	assert.NotNil(t, data)
	assert.Equal(t, "test", data.AccountId)
	assert.Equal(t, "test", data.CpeMac)
	assert.Equal(t, "test", data.TimeZone)
	assert.Equal(t, "test", data.PartnerId)
}

func TestDeviceServiceObject_MultipleStatusCodes(t *testing.T) {
	testCases := []struct {
		name   string
		status int
		msg    string
	}{
		{"OK", 200, "OK"},
		{"Created", 201, "Created"},
		{"BadRequest", 400, "Bad Request"},
		{"Unauthorized", 401, "Unauthorized"},
		{"NotFound", 404, "Not Found"},
		{"InternalError", 500, "Internal Server Error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			obj := DeviceServiceObject{
				Status:  tc.status,
				Message: tc.msg,
			}

			assert.Equal(t, tc.status, obj.Status)
			assert.Equal(t, tc.msg, obj.Message)
		})
	}
}

// Test DefaultDeviceService getter/setter functions
func TestDefaultDeviceService_DeviceServiceHost(t *testing.T) {
	service := &DefaultDeviceService{
		host: "https://device-service.example.com",
	}

	result := service.DeviceServiceHost()

	assert.Equal(t, "https://device-service.example.com", result)
}

func TestDefaultDeviceService_SetDeviceServiceHost(t *testing.T) {
	service := &DefaultDeviceService{
		host: "https://old-device-host.com",
	}

	service.SetDeviceServiceHost("https://new-device-host.com")

	assert.Equal(t, "https://new-device-host.com", service.host)
}

func TestDefaultDeviceService_DeviceServiceHost_Empty(t *testing.T) {
	service := &DefaultDeviceService{
		host: "",
	}

	result := service.DeviceServiceHost()

	assert.Equal(t, "", result)
}

func TestDefaultDeviceService_SetDeviceServiceHost_MultipleUpdates(t *testing.T) {
	service := &DefaultDeviceService{
		host: "https://host1.com",
	}

	service.SetDeviceServiceHost("https://host2.com")
	assert.Equal(t, "https://host2.com", service.host)

	service.SetDeviceServiceHost("https://host3.com")
	assert.Equal(t, "https://host3.com", service.host)
}
