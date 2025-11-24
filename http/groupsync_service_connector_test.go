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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Test protobufHeaders function
func TestProtobufHeaders_ContentType(t *testing.T) {
	headers := protobufHeaders()

	assert.NotNil(t, headers)
	assert.Equal(t, "application/x-protobuf", headers["Content-Type"])
}

func TestProtobufHeaders_Accept(t *testing.T) {
	headers := protobufHeaders()

	assert.NotNil(t, headers)
	assert.Equal(t, "application/x-protobuf", headers["Accept"])
}

func TestProtobufHeaders_BothHeaders(t *testing.T) {
	headers := protobufHeaders()

	assert.NotNil(t, headers)
	assert.Len(t, headers, 2)
	assert.Equal(t, "application/x-protobuf", headers["Content-Type"])
	assert.Equal(t, "application/x-protobuf", headers["Accept"])
}

func TestProtobufHeaders_MapNotEmpty(t *testing.T) {
	headers := protobufHeaders()

	assert.NotNil(t, headers)
	assert.NotEmpty(t, headers)
	assert.Greater(t, len(headers), 0)
}

func TestProtobufHeaders_ConsistentValues(t *testing.T) {
	headers1 := protobufHeaders()
	headers2 := protobufHeaders()

	assert.Equal(t, headers1["Content-Type"], headers2["Content-Type"])
	assert.Equal(t, headers1["Accept"], headers2["Accept"])
}

// Test DefaultGroupServiceSync getter/setter functions
func TestDefaultGroupServiceSync_GroupServiceSyncHost(t *testing.T) {
	service := &DefaultGroupServiceSync{
		host: "https://groupsync-service.example.com/api/v1",
	}

	result := service.GroupServiceSyncHost()

	assert.Equal(t, "https://groupsync-service.example.com/api/v1", result)
}

func TestDefaultGroupServiceSync_SetGroupServiceSyncHost(t *testing.T) {
	service := &DefaultGroupServiceSync{
		host: "https://old-groupsync-host.com",
	}

	service.SetGroupServiceSyncHost("https://new-groupsync-host.com")

	assert.Equal(t, "https://new-groupsync-host.com", service.host)
}

func TestDefaultGroupServiceSync_GroupServiceSyncHost_Empty(t *testing.T) {
	service := &DefaultGroupServiceSync{
		host: "",
	}

	result := service.GroupServiceSyncHost()

	assert.Equal(t, "", result)
}

func TestDefaultGroupServiceSync_SetGroupServiceSyncHost_WithPath(t *testing.T) {
	service := &DefaultGroupServiceSync{
		host: "https://host1.com/v1",
	}

	service.SetGroupServiceSyncHost("https://host2.com/v2")

	assert.Equal(t, "https://host2.com/v2", service.host)
}

func TestDefaultGroupServiceSync_MultipleUpdates(t *testing.T) {
	service := &DefaultGroupServiceSync{}

	service.SetGroupServiceSyncHost("https://host1.com")
	assert.Equal(t, "https://host1.com", service.host)

	service.SetGroupServiceSyncHost("https://host2.com")
	assert.Equal(t, "https://host2.com", service.host)

	service.SetGroupServiceSyncHost("https://host3.com")
	assert.Equal(t, "https://host3.com", service.host)
}

// Test AddSecurityTokenInfo function with mocked HTTP responses
func TestDefaultGroupServiceSync_AddSecurityTokenInfo_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "security-123")
		assert.Equal(t, "application/x-protobuf", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/x-protobuf", r.Header.Get("Accept"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_sync_service_name = "groupsync-service"
			}
			groupsync-service {
				host = "%s"
				path = "/api/v1"
				security_token_path = "/security"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceSyncConnector(conf, nil, nil).(*DefaultGroupServiceSync)
	fields := log.Fields{"test": "security_token"}

	// Create test security token info
	tokenInfo := map[string]string{
		"token":      "test-token-123",
		"expires_at": "2025-12-31",
		"scope":      "read write",
	}

	err := service.AddSecurityTokenInfo("security-123", tokenInfo, fields)

	assert.NoError(t, err)
}

func TestDefaultGroupServiceSync_AddSecurityTokenInfo_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_sync_service_name = "groupsync-service"
			}
			groupsync-service {
				host = "%s"
				path = "/api/v1"
				security_token_path = "/security"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceSyncConnector(conf, nil, nil).(*DefaultGroupServiceSync)
	fields := log.Fields{"test": "error"}

	tokenInfo := map[string]string{
		"token": "test-token",
	}

	err := service.AddSecurityTokenInfo("error-id", tokenInfo, fields)

	assert.Error(t, err)
}

func TestDefaultGroupServiceSync_AddSecurityTokenInfo_EmptyTokenInfo(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_sync_service_name = "groupsync-service"
			}
			groupsync-service {
				host = "%s"
				path = "/api/v1"
				security_token_path = "/security"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceSyncConnector(conf, nil, nil).(*DefaultGroupServiceSync)
	fields := log.Fields{"test": "empty"}

	tokenInfo := map[string]string{}

	err := service.AddSecurityTokenInfo("empty-id", tokenInfo, fields)

	assert.NoError(t, err)
}
