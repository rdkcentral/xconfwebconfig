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

// Test AccountServiceDevices
func TestAccountServiceDevices_IsEmpty_True(t *testing.T) {
	devices := &AccountServiceDevices{}

	result := devices.IsEmpty()

	assert.True(t, result)
}

func TestAccountServiceDevices_IsEmpty_WithID(t *testing.T) {
	devices := &AccountServiceDevices{
		Id: "DEV12345",
	}

	result := devices.IsEmpty()

	assert.False(t, result)
}

func TestAccountServiceDevices_IsEmpty_WithPartner(t *testing.T) {
	devices := &AccountServiceDevices{
		DeviceData: DeviceData{
			Partner: "PARTNER1",
		},
	}

	result := devices.IsEmpty()

	assert.False(t, result)
}

func TestAccountServiceDevices_IsEmpty_WithServiceAccountUri(t *testing.T) {
	devices := &AccountServiceDevices{
		DeviceData: DeviceData{
			ServiceAccountUri: "/account/123",
		},
	}

	result := devices.IsEmpty()

	assert.False(t, result)
}

func TestAccountServiceDevices_IsEmpty_AllFields(t *testing.T) {
	devices := &AccountServiceDevices{
		Id: "DEV123",
		DeviceData: DeviceData{
			Partner:           "PARTNER",
			ServiceAccountUri: "/account/123",
		},
	}

	result := devices.IsEmpty()

	assert.False(t, result)
}

// Test DeviceServiceData
func TestDeviceServiceData_Structure(t *testing.T) {
	data := DeviceServiceData{
		AccountId: "ACC123",
		CpeMac:    "00:11:22:33:44:55",
		TimeZone:  "America/New_York",
		PartnerId: "PARTNER1",
	}

	assert.Equal(t, "ACC123", data.AccountId)
	assert.Equal(t, "00:11:22:33:44:55", data.CpeMac)
	assert.Equal(t, "America/New_York", data.TimeZone)
	assert.Equal(t, "PARTNER1", data.PartnerId)
}

func TestDeviceServiceObject_Structure(t *testing.T) {
	obj := DeviceServiceObject{
		Status:  200,
		Message: "Success",
		DeviceServiceData: &DeviceServiceData{
			AccountId: "ACC456",
			CpeMac:    "AA:BB:CC:DD:EE:FF",
		},
	}

	assert.Equal(t, 200, obj.Status)
	assert.Equal(t, "Success", obj.Message)
	assert.NotNil(t, obj.DeviceServiceData)
	assert.Equal(t, "ACC456", obj.DeviceServiceData.AccountId)
}

// Test Account struct
func TestAccount_Structure(t *testing.T) {
	account := Account{
		Id: "account-123",
		AccountData: AccountData{
			AccountAttributes: AccountAttributes{
				TimeZone:    "America/New_York",
				CountryCode: "US",
			},
		},
	}

	assert.Equal(t, "account-123", account.Id)
	assert.Equal(t, "America/New_York", account.AccountData.AccountAttributes.TimeZone)
	assert.Equal(t, "US", account.AccountData.AccountAttributes.CountryCode)
}

func TestAccountAttributes_Structure(t *testing.T) {
	attrs := AccountAttributes{
		TimeZone:    "Europe/London",
		CountryCode: "GB",
	}

	assert.Equal(t, "Europe/London", attrs.TimeZone)
	assert.Equal(t, "GB", attrs.CountryCode)
}

// Test DeviceData struct
func TestDeviceData_Structure(t *testing.T) {
	data := DeviceData{
		Partner:           "Comcast",
		ServiceAccountUri: "/accounts/12345",
	}

	assert.Equal(t, "Comcast", data.Partner)
	assert.Equal(t, "/accounts/12345", data.ServiceAccountUri)
}

// Test DefaultAccountService getter/setter functions
func TestDefaultAccountService_AccountServiceHost(t *testing.T) {
	service := &DefaultAccountService{
		host: "https://account-service.example.com",
	}

	result := service.AccountServiceHost()

	assert.Equal(t, "https://account-service.example.com", result)
}

func TestDefaultAccountService_SetAccountServiceHost(t *testing.T) {
	service := &DefaultAccountService{
		host: "https://old-host.example.com",
	}

	service.SetAccountServiceHost("https://new-host.example.com")

	assert.Equal(t, "https://new-host.example.com", service.host)
}

func TestDefaultAccountService_AccountServiceHost_Empty(t *testing.T) {
	service := &DefaultAccountService{
		host: "",
	}

	result := service.AccountServiceHost()

	assert.Equal(t, "", result)
}

func TestDefaultAccountService_SetAccountServiceHost_EmptyString(t *testing.T) {
	service := &DefaultAccountService{
		host: "https://existing-host.com",
	}

	service.SetAccountServiceHost("")

	assert.Equal(t, "", service.host)
}
