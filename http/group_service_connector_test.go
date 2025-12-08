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
	conversion "github.com/rdkcentral/xconfwebconfig/protobuf"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// Test DefaultGroupService getter/setter functions
func TestDefaultGroupService_GroupServiceHost(t *testing.T) {
	service := &DefaultGroupService{
		host: "https://group-service.example.com",
	}

	result := service.GroupServiceHost()

	assert.Equal(t, "https://group-service.example.com", result)
}

func TestDefaultGroupService_SetGroupServiceHost(t *testing.T) {
	service := &DefaultGroupService{
		host: "https://old-group-host.com",
	}

	service.SetGroupServiceHost("https://new-group-host.com")

	assert.Equal(t, "https://new-group-host.com", service.host)
}

func TestDefaultGroupService_GroupServiceHost_Empty(t *testing.T) {
	service := &DefaultGroupService{
		host: "",
	}

	result := service.GroupServiceHost()

	assert.Equal(t, "", result)
}

func TestDefaultGroupService_GroupPrefix(t *testing.T) {
	service := &DefaultGroupService{
		groupPrefix: "prod_",
	}

	result := service.GroupPrefix()

	assert.Equal(t, "prod_", result)
}

func TestDefaultGroupService_SetGroupPrefix(t *testing.T) {
	service := &DefaultGroupService{
		groupPrefix: "old_prefix",
	}

	service.SetGroupPrefix("new_prefix")

	assert.Equal(t, "new_prefix", service.groupPrefix)
}

func TestDefaultGroupService_GroupPrefix_Empty(t *testing.T) {
	service := &DefaultGroupService{
		groupPrefix: "",
	}

	result := service.GroupPrefix()

	assert.Equal(t, "", result)
}

func TestDefaultGroupService_SetGroupPrefix_EmptyString(t *testing.T) {
	service := &DefaultGroupService{
		groupPrefix: "existing_prefix",
	}

	service.SetGroupPrefix("")

	assert.Equal(t, "", service.groupPrefix)
}

func TestDefaultGroupService_BothGetters(t *testing.T) {
	service := &DefaultGroupService{
		host:        "https://group-api.example.com",
		groupPrefix: "staging_",
	}

	assert.Equal(t, "https://group-api.example.com", service.GroupServiceHost())
	assert.Equal(t, "staging_", service.GroupPrefix())
}

func TestDefaultGroupService_BothSetters(t *testing.T) {
	service := &DefaultGroupService{
		host:        "https://old-host.com",
		groupPrefix: "old_",
	}

	service.SetGroupServiceHost("https://new-host.com")
	service.SetGroupPrefix("new_")

	assert.Equal(t, "https://new-host.com", service.host)
	assert.Equal(t, "new_", service.groupPrefix)
}

func TestDefaultGroupService_MultipleUpdates(t *testing.T) {
	service := &DefaultGroupService{}

	// First update
	service.SetGroupServiceHost("https://host1.com")
	service.SetGroupPrefix("prefix1_")
	assert.Equal(t, "https://host1.com", service.host)
	assert.Equal(t, "prefix1_", service.groupPrefix)

	// Second update
	service.SetGroupServiceHost("https://host2.com")
	service.SetGroupPrefix("prefix2_")
	assert.Equal(t, "https://host2.com", service.host)
	assert.Equal(t, "prefix2_", service.groupPrefix)
}

// Test GetFeatureTagsHashedItems function with mocked HTTP responses
func TestDefaultGroupService_GetFeatureTagsHashedItems_Success(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "test-feature")

		// Create a test XdasHashes protobuf message
		testHashes := &conversion.XdasHashes{
			Fields: map[string]string{
				"hash1": "value1",
				"hash2": "value2",
				"hash3": "value3",
			},
		}
		data, err := proto.Marshal(testHashes)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer mockServer.Close()

	// Create test configuration
	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_service_name = "group-service"
			}
			group-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceConnector(conf, nil, nil).(*DefaultGroupService)
	fields := log.Fields{"test": "value"}

	result, err := service.GetFeatureTagsHashedItems("test-feature", fields)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "value1", result["hash1"])
	assert.Equal(t, "value2", result["hash2"])
	assert.Equal(t, "value3", result["hash3"])
}

func TestDefaultGroupService_GetFeatureTagsHashedItems_EmptyResponse(t *testing.T) {
	// Create a mock HTTP server returning empty hashes
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		// Create empty XdasHashes protobuf message
		testHashes := &conversion.XdasHashes{
			Fields: map[string]string{},
		}
		data, err := proto.Marshal(testHashes)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer mockServer.Close()

	// Create test configuration
	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_service_name = "group-service"
			}
			group-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceConnector(conf, nil, nil).(*DefaultGroupService)
	fields := log.Fields{"test": "value"}

	result, err := service.GetFeatureTagsHashedItems("empty-feature", fields)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestDefaultGroupService_GetFeatureTagsHashedItems_ServerError(t *testing.T) {
	// Create a mock HTTP server that returns 500 error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	// Create test configuration
	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_service_name = "group-service"
			}
			group-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceConnector(conf, nil, nil).(*DefaultGroupService)
	fields := log.Fields{"test": "value"}

	_, err := service.GetFeatureTagsHashedItems("test-feature", fields)

	// Should return an error due to HTTP 500
	assert.Error(t, err)
}

// Test GetSecurityTokenInfo function with mocked HTTP responses
func TestDefaultGroupService_GetSecurityTokenInfo_Success(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "security-123")

		// Create a test XdasHashes protobuf message for security token info
		testHashes := &conversion.XdasHashes{
			Fields: map[string]string{
				"token":      "abc123token",
				"expires_at": "2025-12-31",
				"scope":      "read",
			},
		}
		data, err := proto.Marshal(testHashes)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer mockServer.Close()

	// Create test configuration
	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_service_name = "group-service"
			}
			group-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceConnector(conf, nil, nil).(*DefaultGroupService)
	fields := log.Fields{"security_test": "value"}

	result, err := service.GetSecurityTokenInfo("security-123", fields)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "abc123token", result["token"])
	assert.Equal(t, "2025-12-31", result["expires_at"])
	assert.Equal(t, "read", result["scope"])
}

func TestDefaultGroupService_GetSecurityTokenInfo_InvalidIdentifier(t *testing.T) {
	// Create a mock HTTP server returning empty response for invalid identifier
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "invalid-id")

		// Create empty XdasHashes protobuf message
		testHashes := &conversion.XdasHashes{
			Fields: map[string]string{},
		}
		data, err := proto.Marshal(testHashes)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer mockServer.Close()

	// Create test configuration
	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_service_name = "group-service"
			}
			group-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceConnector(conf, nil, nil).(*DefaultGroupService)
	fields := log.Fields{"test": "value"}

	result, err := service.GetSecurityTokenInfo("invalid-id", fields)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestDefaultGroupService_GetSecurityTokenInfo_MultipleFields(t *testing.T) {
	// Test with comprehensive security token info
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testHashes := &conversion.XdasHashes{
			Fields: map[string]string{
				"access_token":  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
				"refresh_token": "refresh_abc123",
				"token_type":    "Bearer",
				"expires_in":    "3600",
				"scope":         "read write",
				"client_id":     "client-456",
				"issued_at":     "2024-11-12T10:00:00Z",
			},
		}
		data, err := proto.Marshal(testHashes)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_service_name = "group-service"
			}
			group-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceConnector(conf, nil, nil).(*DefaultGroupService)
	fields := log.Fields{"comprehensive_test": "value"}

	result, err := service.GetSecurityTokenInfo("comprehensive-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 7, len(result))
	assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", result["access_token"])
	assert.Equal(t, "refresh_abc123", result["refresh_token"])
	assert.Equal(t, "Bearer", result["token_type"])
	assert.Equal(t, "3600", result["expires_in"])
	assert.Equal(t, "read write", result["scope"])
	assert.Equal(t, "client-456", result["client_id"])
	assert.Equal(t, "2024-11-12T10:00:00Z", result["issued_at"])
}

// Test CreateListFromGroupServiceProto function (non-HTTP, direct testing)
func TestDefaultGroupService_CreateListFromGroupServiceProto_AllTrue(t *testing.T) {
	service := &DefaultGroupService{
		groupPrefix: "prod_",
	}

	cpeGroup := &conversion.CpeGroup{
		StormReadyFw: true,
		Wanfailover:  true,
		Gwfailover:   true,
	}

	result := service.CreateListFromGroupServiceProto(cpeGroup)

	assert.Equal(t, 3, len(result))
	assert.Contains(t, result, "prod_StormReadyFw")
	assert.Contains(t, result, "prod_Wanfailover")
	assert.Contains(t, result, "prod_Gwfailover")
}

func TestDefaultGroupService_CreateListFromGroupServiceProto_PartialTrue(t *testing.T) {
	service := &DefaultGroupService{
		groupPrefix: "test_",
	}

	cpeGroup := &conversion.CpeGroup{
		StormReadyFw: true,
		Wanfailover:  false,
		Gwfailover:   true,
	}

	result := service.CreateListFromGroupServiceProto(cpeGroup)

	assert.Equal(t, 2, len(result))
	assert.Contains(t, result, "test_StormReadyFw")
	assert.Contains(t, result, "test_Gwfailover")
	assert.NotContains(t, result, "test_Wanfailover")
}

func TestDefaultGroupService_CreateListFromGroupServiceProto_AllFalse(t *testing.T) {
	service := &DefaultGroupService{
		groupPrefix: "empty_",
	}

	cpeGroup := &conversion.CpeGroup{
		StormReadyFw: false,
		Wanfailover:  false,
		Gwfailover:   false,
	}

	result := service.CreateListFromGroupServiceProto(cpeGroup)

	assert.Equal(t, 0, len(result))
}

func TestDefaultGroupService_CreateListFromGroupServiceProto_NoPrefix(t *testing.T) {
	service := &DefaultGroupService{
		groupPrefix: "",
	}

	cpeGroup := &conversion.CpeGroup{
		StormReadyFw: true,
		Wanfailover:  true,
		Gwfailover:   false,
	}

	result := service.CreateListFromGroupServiceProto(cpeGroup)

	assert.Equal(t, 2, len(result))
	assert.Contains(t, result, "StormReadyFw")
	assert.Contains(t, result, "Wanfailover")
	assert.NotContains(t, result, "Gwfailover")
}

// Test GetCpeGroups function with mocked HTTP responses
func TestDefaultGroupService_GetCpeGroups_Success(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "AA:BB:CC:DD:EE:FF")

		// Create a test CpeGroup protobuf message
		testGroup := &conversion.CpeGroup{
			StormReadyFw: true,
			Wanfailover:  true,
			Gwfailover:   false,
		}
		data, err := proto.Marshal(testGroup)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer mockServer.Close()

	// Create test configuration
	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_service_name = "group-service"
			}
			group-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceConnector(conf, nil, nil).(*DefaultGroupService)
	service.SetGroupPrefix("test_")
	fields := log.Fields{"test": "value"}

	result, err := service.GetCpeGroups("AA:BB:CC:DD:EE:FF", fields)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result))
	assert.Contains(t, result, "test_StormReadyFw")
	assert.Contains(t, result, "test_Wanfailover")
	assert.NotContains(t, result, "test_Gwfailover")
}

func TestDefaultGroupService_GetCpeGroups_AllGroupsEnabled(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testGroup := &conversion.CpeGroup{
			StormReadyFw: true,
			Wanfailover:  true,
			Gwfailover:   true,
		}
		data, err := proto.Marshal(testGroup)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_service_name = "group-service"
			}
			group-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceConnector(conf, nil, nil).(*DefaultGroupService)
	service.SetGroupPrefix("prod_")
	fields := log.Fields{"test": "all_groups"}

	result, err := service.GetCpeGroups("11:22:33:44:55:66", fields)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(result))
	assert.Contains(t, result, "prod_StormReadyFw")
	assert.Contains(t, result, "prod_Wanfailover")
	assert.Contains(t, result, "prod_Gwfailover")
}

func TestDefaultGroupService_GetCpeGroups_NoGroupsEnabled(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testGroup := &conversion.CpeGroup{
			StormReadyFw: false,
			Wanfailover:  false,
			Gwfailover:   false,
		}
		data, err := proto.Marshal(testGroup)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_service_name = "group-service"
			}
			group-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceConnector(conf, nil, nil).(*DefaultGroupService)
	service.SetGroupPrefix("none_")
	fields := log.Fields{"test": "no_groups"}

	result, err := service.GetCpeGroups("00:00:00:00:00:00", fields)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

func TestDefaultGroupService_GetCpeGroups_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_service_name = "group-service"
			}
			group-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceConnector(conf, nil, nil).(*DefaultGroupService)
	fields := log.Fields{"test": "error"}

	_, err := service.GetCpeGroups("error-mac", fields)

	assert.Error(t, err)
}

// Test GetRfcPrecookDetails function with mocked HTTP responses
func TestDefaultGroupService_GetRfcPrecookDetails_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "AA:BB:CC:DD:EE:FF")

		// Create a test XconfDevice protobuf message
		testDevice := &conversion.XconfDevice{
			AccountId:        "acc-123",
			Partner:          "comcast",
			Model:            "TG1682G",
			ApplicationType:  "stb",
			Env:              "PROD",
			FwVersion:        "1.0.0",
			Experience:       "X1",
			IsAtWarehouse:    false,
			OfferedFwVersion: "2.0.0",
		}
		data, err := proto.Marshal(testDevice)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_service_name = "group-service"
			}
			group-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceConnector(conf, nil, nil).(*DefaultGroupService)
	fields := log.Fields{"test": "rfc_precook"}

	result, err := service.GetRfcPrecookDetails("AA:BB:CC:DD:EE:FF", fields)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "acc-123", result.AccountId)
	assert.Equal(t, "comcast", result.Partner)
	assert.Equal(t, "TG1682G", result.Model)
	assert.Equal(t, "stb", result.ApplicationType)
	assert.Equal(t, "PROD", result.Env)
	assert.Equal(t, "1.0.0", result.FwVersion)
	assert.Equal(t, "X1", result.Experience)
	assert.Equal(t, false, result.IsAtWarehouse)
	assert.Equal(t, "2.0.0", result.OfferedFwVersion)
}

func TestDefaultGroupService_GetRfcPrecookDetails_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				group_service_name = "group-service"
			}
			group-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewGroupServiceConnector(conf, nil, nil).(*DefaultGroupService)
	fields := log.Fields{"test": "error"}

	_, err := service.GetRfcPrecookDetails("error-mac", fields)

	assert.Error(t, err)
}
