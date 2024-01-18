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
	"net/http"
	"net/http/httptest"
)

var (
	mockEmptyResponse = []byte(`{}`)
)

// setupMocks sets up mock servers that return the same predefined response for any call to the server
// mock servers are set up for all external services - SatService, DeviceService, TaggingService, AccountService
// If a different mock response is desired for a test, use the same template below, but just define a different mockResponse
// An example for a different mock response can be seen in http/supplementary_handler_test.go
func (server *XconfServer) SetupMocks() {
	server.mockSatService()
	server.mockDeviceService()
	server.mockTagging()
	server.mockAccountService()
	server.mockGroupService()
}

func (server *XconfServer) mockSatService() {
	mockResponse := []byte(`{"access_token":"one_mock_token","expires_in":86400,"scope":"scope1 scope2 scope3","token_type":"Bearer"}`)

	// SatService mock server
	mockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(mockResponse)
		}))
	server.SetSatServiceHost(mockServer.URL)
}

func (server *XconfServer) mockDeviceService() {
	mockResponse := []byte(`{"status":200,"data":{"account_id":"testAccountId", "cpe_mac":"testCpeMac"}}`)

	// DeviceService mock server
	mockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(mockResponse)
		}))
	server.SetDeviceServiceHost(mockServer.URL)
}

func (server *XconfServer) mockTagging() {
	mockResponse := []byte(`["value1", "value2", "value3"]`)
	// tagging mock server
	mockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(mockResponse)
		}))
	server.SetTaggingHost(mockServer.URL)
}

func (server *XconfServer) mockAccountService() {
	// AccountService mock server
	mockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockEmptyResponse))
		}))
	server.AccountServiceConnector.SetAccountServiceHost(mockServer.URL)
}

func (server *XconfServer) mockGroupService() {
	mockResponse := []byte(`{"hasAccountServiceData": true,"serviceAccountUri": "123456789012345","partnerId": "unittesting"}`)

	// SatService mock server
	mockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(mockResponse)
		}))
	server.SetSatServiceHost(mockServer.URL)
}
