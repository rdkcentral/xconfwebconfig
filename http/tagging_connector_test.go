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

// MockTaggingServiceConnector is a test mock that implements TaggingConnector
type MockTaggingServiceConnector struct {
	host        string
	tags        []string
	shouldError bool
}

func (m *MockTaggingServiceConnector) MakeGetTagsRequest(url string, token string, vargs ...log.Fields) ([]string, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.tags, nil
}

func (m *MockTaggingServiceConnector) GetTagsForContext(contextMap map[string]string, token string, fields log.Fields) ([]string, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.tags, nil
}

func (m *MockTaggingServiceConnector) TaggingHost() string {
	return m.host
}

func (m *MockTaggingServiceConnector) SetTaggingHost(host string) {
	m.host = host
}

func (m *MockTaggingServiceConnector) GetTagsForMacAddress(macAddress string, token string, fields log.Fields) ([]string, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.tags, nil
}

func (m *MockTaggingServiceConnector) GetTagsForPartner(partnerId string, token string, fields log.Fields) ([]string, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.tags, nil
}

func (m *MockTaggingServiceConnector) GetTagsForPartnerAndMacAddress(partnerId string, macAddress string, token string, fields log.Fields) ([]string, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.tags, nil
}

func (m *MockTaggingServiceConnector) GetTagsForMacAddressAndAccount(macAddress string, accountId string, token string, fields log.Fields) ([]string, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.tags, nil
}

func (m *MockTaggingServiceConnector) GetTagsForAccount(accountId string, token string, fields log.Fields) ([]string, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.tags, nil
}

func (m *MockTaggingServiceConnector) GetTagsForPartnerAndMacAddressAndAccount(partnerId string, macAddress string, accountId string, token string, fields log.Fields) ([]string, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.tags, nil
}

func (m *MockTaggingServiceConnector) GetTagsForPartnerAndAccount(partnerId string, accountId string, token string, fields log.Fields) ([]string, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.tags, nil
}

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
	// Create mock service with test tags
	mockService := &MockTaggingServiceConnector{
		host: "https://test-tagging-service.example.com",
		tags: []string{"tag1", "tag2", "tag3"},
	}

	fields := log.Fields{"test": "mac_tags"}

	result, err := mockService.GetTagsForMacAddress("AA:BB:CC:DD:EE:FF", "test-token", fields)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, len(result))
	assert.Contains(t, result, "tag1")
	assert.Contains(t, result, "tag2")
	assert.Contains(t, result, "tag3")
}

func TestDefaultTaggingService_GetTagsForPartner_Success(t *testing.T) {
	// Create mock service with test tags
	mockService := &MockTaggingServiceConnector{
		host: "https://test-tagging-service.example.com",
		tags: []string{"partner-tag1", "partner-tag2"},
	}

	fields := log.Fields{"test": "partner_tags"}

	result, err := mockService.GetTagsForPartner("comcast", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Contains(t, result, "partner-tag1")
	assert.Contains(t, result, "partner-tag2")
}

func TestDefaultTaggingService_GetTagsForPartnerAndMacAddress_Success(t *testing.T) {
	// Create mock service with test tags
	mockService := &MockTaggingServiceConnector{
		host: "https://test-tagging-service.example.com",
		tags: []string{"combined-tag1", "combined-tag2", "combined-tag3"},
	}

	fields := log.Fields{"test": "combined_tags"}

	result, err := mockService.GetTagsForPartnerAndMacAddress("comcast", "AA:BB:CC:DD:EE:FF", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(result))
	assert.Contains(t, result, "combined-tag1")
	assert.Contains(t, result, "combined-tag2")
	assert.Contains(t, result, "combined-tag3")
}

func TestDefaultTaggingService_GetTagsForMacAddressAndAccount_Success(t *testing.T) {
	// Create mock service with test tags
	mockService := &MockTaggingServiceConnector{
		host: "https://test-tagging-service.example.com",
		tags: []string{"account-tag1", "account-tag2"},
	}

	fields := log.Fields{"test": "account_tags"}

	result, err := mockService.GetTagsForMacAddressAndAccount("AA:BB:CC:DD:EE:FF", "account-123", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Contains(t, result, "account-tag1")
	assert.Contains(t, result, "account-tag2")
}

func TestDefaultTaggingService_GetTagsForAccount_Success(t *testing.T) {
	// Create mock service with test tag
	mockService := &MockTaggingServiceConnector{
		host: "https://test-tagging-service.example.com",
		tags: []string{"account-only-tag"},
	}

	fields := log.Fields{"test": "account_only"}

	result, err := mockService.GetTagsForAccount("account-456", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Contains(t, result, "account-only-tag")
}

func TestDefaultTaggingService_GetTagsForPartnerAndMacAddressAndAccount_Success(t *testing.T) {
	// Create mock service with test tags
	mockService := &MockTaggingServiceConnector{
		host: "https://test-tagging-service.example.com",
		tags: []string{"full-context-tag1", "full-context-tag2"},
	}

	fields := log.Fields{"test": "full_context"}

	result, err := mockService.GetTagsForPartnerAndMacAddressAndAccount("comcast", "AA:BB:CC:DD:EE:FF", "account-789", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Contains(t, result, "full-context-tag1")
	assert.Contains(t, result, "full-context-tag2")
}

func TestDefaultTaggingService_GetTagsForPartnerAndAccount_Success(t *testing.T) {
	// Create mock service with test tag
	mockService := &MockTaggingServiceConnector{
		host: "https://test-tagging-service.example.com",
		tags: []string{"partner-account-tag"},
	}

	fields := log.Fields{"test": "partner_account"}

	result, err := mockService.GetTagsForPartnerAndAccount("sky", "account-999", "test-token", fields)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Contains(t, result, "partner-account-tag")
}

func TestDefaultTaggingService_GetTagsForContext_Success(t *testing.T) {
	// Create mock service with test tags
	mockService := &MockTaggingServiceConnector{
		host: "https://test-tagging-service.example.com",
		tags: []string{"context-tag1", "context-tag2"},
	}

	fields := log.Fields{"test": "context"}

	contextMap := map[string]string{
		"partner":    "comcast",
		"macAddress": "AA:BB:CC:DD:EE:FF",
		"accountId":  "account-123",
	}

	result, err := mockService.GetTagsForContext(contextMap, "test-token", fields)

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
				tagging_service_name = "tagging_service"
			}
			tagging_service {
				host = "%s"
				tags_mac_address_template = ""
				tags_partner_template = ""
				tags_partner_and_mac_address_template = ""
				tags_partner_and_mac_address_template = ""
				tags_account_template = ""
				tags_partner_and_mac_address_and_account_template = ""
				tags_partner_and_account_template = ""
			}
		}
	`, mockServer.URL))

	service := NewTaggingConnector(conf, nil, nil).(*DefaultTaggingService)
	fields := log.Fields{"test": "error"}

	_, err := service.GetTagsForMacAddress("error-mac", "test-token", fields)

	assert.Error(t, err)
}

func TestDefaultTaggingService_GetTagsForPartner_EmptyResponse(t *testing.T) {
	// Create mock service with empty tags
	mockService := &MockTaggingServiceConnector{
		host: "https://test-tagging-service.example.com",
		tags: []string{},
	}

	fields := log.Fields{"test": "empty"}

	result, err := mockService.GetTagsForPartner("no-tags-partner", "test-token", fields)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}
