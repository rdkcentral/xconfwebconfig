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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-akka/configuration"
	"github.com/rdkcentral/xconfwebconfig/common"
	log "github.com/sirupsen/logrus"
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

// Test GetAccountData with mocked HTTP server (0% coverage function)
func TestDefaultAccountService_GetAccountData_Success(t *testing.T) {
	// Create a FAKE/MOCK HTTP server (not real!)
	mockAccountServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer test-token")
		assert.Equal(t, common.HeaderXconfDataService, r.Header.Get("User-Agent"))

		// Return FAKE account data (no database involved!)
		mockAccount := Account{
			Id: "test-account-id",
			AccountData: AccountData{
				AccountAttributes: AccountAttributes{
					TimeZone:    "America/New_York",
					CountryCode: "US",
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockAccount)
	}))
	defer mockAccountServer.Close()

	// Create service pointing to MOCK server (not real service!)
	config := configuration.ParseString("") // Empty config for testing
	service := &DefaultAccountService{
		HttpClient: NewHttpClient(config, "test-service", nil),
		host:       mockAccountServer.URL, // Points to our FAKE server
	}

	// Test GetAccountData with our MOCK server
	fields := log.Fields{"test": "field"}
	account, err := service.GetAccountData("test-service-account", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, "test-account-id", account.Id)
	assert.Equal(t, "America/New_York", account.AccountData.AccountAttributes.TimeZone)
}

func TestDefaultAccountService_GetAccountData_HTTPError(t *testing.T) {
	// Create MOCK server that simulates HTTP error (not real service!)
	mockErrorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockErrorServer.Close()

	config := configuration.ParseString("")
	service := &DefaultAccountService{
		HttpClient: NewHttpClient(config, "test-service", nil),
		host:       mockErrorServer.URL, // Points to our FAKE error server
	}

	fields := log.Fields{"test": "field"}
	account, err := service.GetAccountData("test-service-account", "test-token", fields)

	assert.Error(t, err)
	assert.Equal(t, "", account.Id)
}

func TestDefaultAccountService_GetAccountData_InvalidJSON(t *testing.T) {
	// Create MOCK server that returns invalid JSON (simulates bad response)
	mockInvalidJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer mockInvalidJSONServer.Close()

	config := configuration.ParseString("")
	service := &DefaultAccountService{
		HttpClient: NewHttpClient(config, "test-service", nil),
		host:       mockInvalidJSONServer.URL, // Points to our FAKE server with bad JSON
	}

	fields := log.Fields{"test": "field"}
	account, err := service.GetAccountData("test-service-account", "test-token", fields)

	assert.Error(t, err)
	assert.Equal(t, "", account.Id)
}

// Test GetDevices with mocked HTTP server (0% coverage function)
func TestDefaultAccountService_GetDevices_Success(t *testing.T) {
	// Create FAKE/MOCK server that returns device data (no real database!)
	mockDevicesServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer test-token")
		assert.Equal(t, common.HeaderXconfDataService, r.Header.Get("User-Agent"))

		// Return FAKE devices data (hardcoded, no database query!)
		mockDevices := []AccountServiceDevices{
			{
				Id: "device-123",
				DeviceData: DeviceData{
					Partner:           "Comcast",
					ServiceAccountUri: "/accounts/12345",
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockDevices)
	}))
	defer mockDevicesServer.Close()

	config := configuration.ParseString("")
	service := &DefaultAccountService{
		HttpClient: NewHttpClient(config, "test-service", nil),
		host:       mockDevicesServer.URL, // Points to our FAKE devices server
	}

	fields := log.Fields{"test": "field"}
	devices, err := service.GetDevices("macKey", "macValue", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, "device-123", devices.Id)
	assert.Equal(t, "Comcast", devices.DeviceData.Partner)
}

func TestDefaultAccountService_GetDevices_EmptyArray(t *testing.T) {
	// Create MOCK server that returns empty devices array (simulates no devices found)
	mockEmptyDevicesServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]AccountServiceDevices{})
	}))
	defer mockEmptyDevicesServer.Close()

	config := configuration.ParseString("")
	service := &DefaultAccountService{
		HttpClient: NewHttpClient(config, "test-service", nil),
		host:       mockEmptyDevicesServer.URL, // Points to our FAKE empty devices server
	}

	fields := log.Fields{"test": "field"}
	devices, err := service.GetDevices("macKey", "macValue", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, "", devices.Id) // Should be empty when array is empty
}

func TestDefaultAccountService_GetDevices_HTTPError(t *testing.T) {
	// Create MOCK server that simulates HTTP 404 error (device not found scenario)
	mockNotFoundServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer mockNotFoundServer.Close()

	config := configuration.ParseString("")
	service := &DefaultAccountService{
		HttpClient: NewHttpClient(config, "test-service", nil),
		host:       mockNotFoundServer.URL, // Points to our FAKE 404 server
	}

	fields := log.Fields{"test": "field"}
	devices, err := service.GetDevices("macKey", "macValue", "test-token", fields)

	assert.Error(t, err)
	assert.Equal(t, "", devices.Id)
}
