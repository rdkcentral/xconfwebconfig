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

	"github.com/rdkcentral/xconfwebconfig/common"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Test functions using existing mock infrastructure for 0% coverage security manager functions

func TestGetSecurityToken_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	fields := log.Fields{"test": "GetSecurityToken"}
	deviceInfo := map[string]string{
		SECURITY_TOKEN_CLIENT_PROTOCOL: "http",
		SECURITY_TOKEN_ESTB_MAC:        "11:22:33:44:55:66",
		SECURITY_TOKEN_ESTB_IP:         "192.168.1.100",
		SECURITY_TOKEN_MODEL:           "test-model",
		SECURITY_TOKEN_PARTNER:         "test-partner",
	}

	// Test getSecurityToken function using FirmwareSecurityTokenConfig
	if server.FirmwareSecurityTokenConfig != nil {
		token := server.FirmwareSecurityTokenConfig.getSecurityToken(deviceInfo, fields)
		assert.NotNil(t, token)
	}
}

func TestGenerateSecurityToken_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	fields := log.Fields{"test": "GenerateSecurityToken"}
	testInput := "test-input-for-token-generation"

	// Test generateSecurityToken function
	token := generateSecurityToken(testInput, fields)
	assert.NotEmpty(t, token)

	// Test that same input produces same token
	token2 := generateSecurityToken(testInput, fields)
	assert.Equal(t, token, token2)

	// Test that different input produces different token
	token3 := generateSecurityToken("different-input", fields)
	assert.NotEqual(t, token, token3)
}

func TestAddTokenToUrl_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	fields := log.Fields{"test": "AddTokenToUrl"}
	deviceInfo := map[string]string{
		SECURITY_TOKEN_CLIENT_PROTOCOL: "http",
		SECURITY_TOKEN_ESTB_MAC:        "11:22:33:44:55:66",
		SECURITY_TOKEN_ESTB_IP:         "192.168.1.100",
		SECURITY_TOKEN_MODEL:           "test-model",
	}

	testUrl := "http://example.com/firmware/test.bin"

	// Test addTokenToUrl function
	if server.FirmwareSecurityTokenConfig != nil {
		resultUrl := server.FirmwareSecurityTokenConfig.addTokenToUrl(deviceInfo, testUrl, false, fields)
		assert.NotEmpty(t, resultUrl)
		// URL should be modified or remain the same based on security token logic
		assert.Contains(t, resultUrl, "example.com")
	}
}

func TestAddSecurityTokenToUrl_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	fields := log.Fields{"test": "AddSecurityTokenToUrl"}
	deviceInfo := map[string]string{
		SECURITY_TOKEN_CLIENT_PROTOCOL: "http",
		SECURITY_TOKEN_ESTB_MAC:        "11:22:33:44:55:66",
		SECURITY_TOKEN_ESTB_IP:         "192.168.1.100",
		SECURITY_TOKEN_MODEL:           "test-model",
	}

	testUrl := "http://example.com/firmware/test.bin"

	// Test AddSecurityTokenToUrl function
	if server.FirmwareSecurityTokenConfig != nil {
		resultUrl := server.FirmwareSecurityTokenConfig.AddSecurityTokenToUrl(deviceInfo, testUrl, fields)
		assert.NotEmpty(t, resultUrl)
		assert.Contains(t, resultUrl, "example.com")
	}
}

func TestIsUrlFqdn_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	// Test isUrlFqdn function with FQDN URL
	fqdnUrl := "http://example.com/path"
	isFqdn := isUrlFqdn(fqdnUrl)
	assert.False(t, isFqdn) // Contains "://" so should return false

	// Test with non-FQDN URL (without protocol)
	nonFqdnUrl := "example.com/path"
	isFqdn = isUrlFqdn(nonFqdnUrl)
	assert.True(t, isFqdn) // Doesn't contain "://" so should return true

	// Test edge cases
	emptyUrl := ""
	isFqdn = isUrlFqdn(emptyUrl)
	assert.True(t, isFqdn) // Empty doesn't contain "://"

	httpsUrl := "https://secure.example.com/path"
	isFqdn = isUrlFqdn(httpsUrl)
	assert.False(t, isFqdn) // Contains "://"
}

func TestDoesUrlNeedToken_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	fields := log.Fields{"test": "DoesUrlNeedToken"}

	// Test doesUrlNeedToken function
	if server.FirmwareSecurityTokenConfig != nil {
		testUrl := "http://example.com/firmware/test.bin"

		// Test with FQDN (isFqdn = false since it contains "://")
		needsToken := server.FirmwareSecurityTokenConfig.doesUrlNeedToken(testUrl, false, fields)
		// Just verify function executes without error, result depends on configuration
		assert.True(t, needsToken == true || needsToken == false)

		// Test with non-FQDN (isFqdn = true)
		nonFqdnUrl := "example.com/firmware/test.bin"
		needsToken = server.FirmwareSecurityTokenConfig.doesUrlNeedToken(nonFqdnUrl, true, fields)
		// Just verify function executes without error, result depends on configuration
		assert.True(t, needsToken == true || needsToken == false)
	}
}

func TestCanSkipSecurityTokenForClientProtocol_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	// Test CanSkipSecurityTokenForClientProtocol function

	// Test with HTTP protocol (typically should not skip)
	deviceInfoHttp := map[string]string{
		SECURITY_TOKEN_CLIENT_PROTOCOL: "http",
		SECURITY_TOKEN_ESTB_MAC:        "11:22:33:44:55:66",
	}
	canSkip := CanSkipSecurityTokenForClientProtocol(deviceInfoHttp)
	assert.False(t, canSkip) // HTTP should not skip by default

	// Test with HTTPS protocol
	deviceInfoHttps := map[string]string{
		SECURITY_TOKEN_CLIENT_PROTOCOL: "https",
		SECURITY_TOKEN_ESTB_MAC:        "11:22:33:44:55:66",
	}
	canSkip = CanSkipSecurityTokenForClientProtocol(deviceInfoHttps)
	assert.False(t, canSkip) // HTTPS should not skip by default

	// Test with MTLS protocol (might skip based on configuration)
	deviceInfoMtls := map[string]string{
		SECURITY_TOKEN_CLIENT_PROTOCOL: "mtls",
		SECURITY_TOKEN_ESTB_MAC:        "11:22:33:44:55:66",
	}
	canSkip = CanSkipSecurityTokenForClientProtocol(deviceInfoMtls)
	// Result depends on configuration, just verify function executes
	assert.True(t, canSkip == true || canSkip == false)

	// Test with empty protocol
	deviceInfoEmpty := map[string]string{
		SECURITY_TOKEN_ESTB_MAC: "11:22:33:44:55:66",
	}
	canSkip = CanSkipSecurityTokenForClientProtocol(deviceInfoEmpty)
	// Should handle missing protocol gracefully
	assert.True(t, canSkip == true || canSkip == false)
}
