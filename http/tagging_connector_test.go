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

// Test TaggingHost getter/setter functions
func TestDefaultTaggingService_TaggingHost(t *testing.T) {
	service := &DefaultTaggingService{
		host: "https://tagging-service.example.com",
	}

	result := service.TaggingHost()

	assert.Equal(t, "https://tagging-service.example.com", result)
}

func TestDefaultTaggingService_SetTaggingHost(t *testing.T) {
	service := &DefaultTaggingService{
		host: "https://old-tagging-host.com",
	}

	service.SetTaggingHost("https://new-tagging-host.com")

	assert.Equal(t, "https://new-tagging-host.com", service.host)
}

// Test GetTagsForMacAddress function with mocked HTTP responses
func TestDefaultTaggingService_GetTagsForMacAddress_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "AA:BB:CC:DD:EE:FF")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["tag1", "tag2", "tag3"]`))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				tagging_service_name = "tagging-service"
			}
			tagging-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewTaggingConnector(conf, nil, nil).(*DefaultTaggingService)
	fields := log.Fields{"test": "mac_tags"}

	result, err := service.GetTagsForMacAddress("AA:BB:CC:DD:EE:FF", "test-token", fields)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, len(result))
	assert.Contains(t, result, "tag1")
	assert.Contains(t, result, "tag2")
	assert.Contains(t, result, "tag3")
}

func TestDefaultTaggingService_GetTagsForPartner_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "comcast")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["partner-tag1", "partner-tag2"]`))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				tagging_service_name = "tagging-service"
			}
			tagging-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewTaggingConnector(conf, nil, nil).(*DefaultTaggingService)
	fields := log.Fields{"test": "partner_tags"}

	result, err := service.GetTagsForPartner("comcast", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Contains(t, result, "partner-tag1")
	assert.Contains(t, result, "partner-tag2")
}

func TestDefaultTaggingService_GetTagsForPartnerAndMacAddress_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "comcast")
		assert.Contains(t, r.URL.Path, "AA:BB:CC:DD:EE:FF")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["combined-tag1", "combined-tag2", "combined-tag3"]`))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				tagging_service_name = "tagging-service"
			}
			tagging-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewTaggingConnector(conf, nil, nil).(*DefaultTaggingService)
	fields := log.Fields{"test": "combined_tags"}

	result, err := service.GetTagsForPartnerAndMacAddress("comcast", "AA:BB:CC:DD:EE:FF", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(result))
	assert.Contains(t, result, "combined-tag1")
	assert.Contains(t, result, "combined-tag2")
	assert.Contains(t, result, "combined-tag3")
}

func TestDefaultTaggingService_GetTagsForMacAddressAndAccount_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "AA:BB:CC:DD:EE:FF")
		assert.Contains(t, r.URL.Path, "account-123")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["account-tag1", "account-tag2"]`))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				tagging_service_name = "tagging-service"
			}
			tagging-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewTaggingConnector(conf, nil, nil).(*DefaultTaggingService)
	fields := log.Fields{"test": "account_tags"}

	result, err := service.GetTagsForMacAddressAndAccount("AA:BB:CC:DD:EE:FF", "account-123", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Contains(t, result, "account-tag1")
	assert.Contains(t, result, "account-tag2")
}

func TestDefaultTaggingService_GetTagsForAccount_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "account-456")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["account-only-tag"]`))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				tagging_service_name = "tagging-service"
			}
			tagging-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewTaggingConnector(conf, nil, nil).(*DefaultTaggingService)
	fields := log.Fields{"test": "account_only"}

	result, err := service.GetTagsForAccount("account-456", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Contains(t, result, "account-only-tag")
}

func TestDefaultTaggingService_GetTagsForPartnerAndMacAddressAndAccount_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "comcast")
		assert.Contains(t, r.URL.Path, "AA:BB:CC:DD:EE:FF")
		assert.Contains(t, r.URL.Path, "account-789")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["full-context-tag1", "full-context-tag2"]`))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				tagging_service_name = "tagging-service"
			}
			tagging-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewTaggingConnector(conf, nil, nil).(*DefaultTaggingService)
	fields := log.Fields{"test": "full_context"}

	result, err := service.GetTagsForPartnerAndMacAddressAndAccount("comcast", "AA:BB:CC:DD:EE:FF", "account-789", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Contains(t, result, "full-context-tag1")
	assert.Contains(t, result, "full-context-tag2")
}

func TestDefaultTaggingService_GetTagsForPartnerAndAccount_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "sky")
		assert.Contains(t, r.URL.Path, "account-999")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["partner-account-tag"]`))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				tagging_service_name = "tagging-service"
			}
			tagging-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewTaggingConnector(conf, nil, nil).(*DefaultTaggingService)
	fields := log.Fields{"test": "partner_account"}

	result, err := service.GetTagsForPartnerAndAccount("sky", "account-999", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Contains(t, result, "partner-account-tag")
}

func TestDefaultTaggingService_GetTagsForContext_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["context-tag1", "context-tag2"]`))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				tagging_service_name = "tagging-service"
			}
			tagging-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewTaggingConnector(conf, nil, nil).(*DefaultTaggingService)
	fields := log.Fields{"test": "context"}

	contextMap := map[string]string{
		"partner":    "comcast",
		"macAddress": "AA:BB:CC:DD:EE:FF",
		"accountId":  "account-123",
	}

	result, err := service.GetTagsForContext(contextMap, "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Contains(t, result, "context-tag1")
	assert.Contains(t, result, "context-tag2")
}

// Test error scenarios
func TestDefaultTaggingService_GetTagsForMacAddress_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				tagging_service_name = "tagging-service"
			}
			tagging-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewTaggingConnector(conf, nil, nil).(*DefaultTaggingService)
	fields := log.Fields{"test": "error"}

	_, err := service.GetTagsForMacAddress("error-mac", "test-token", fields)

	assert.Error(t, err)
}

func TestDefaultTaggingService_GetTagsForPartner_EmptyResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer mockServer.Close()

	conf := configuration.ParseString(fmt.Sprintf(`
		xconfwebconfig {
			xconf {
				tagging_service_name = "tagging-service"
			}
			tagging-service {
				host = "%s"
			}
		}
	`, mockServer.URL))

	service := NewTaggingConnector(conf, nil, nil).(*DefaultTaggingService)
	fields := log.Fields{"test": "empty"}

	result, err := service.GetTagsForPartner("no-tags-partner", "test-token", fields)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}
