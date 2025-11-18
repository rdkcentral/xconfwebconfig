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
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Test SatTokenMgr getter/setter functions
func TestSatTokenMgr_TestOnly_Default(t *testing.T) {
	mgr := NewSatTokenMgr()

	result := mgr.TestOnly()

	assert.False(t, result)
}

func TestSatTokenMgr_TestOnly_True(t *testing.T) {
	mgr := NewSatTokenMgr(true)

	result := mgr.TestOnly()

	assert.True(t, result)
}

func TestSatTokenMgr_TestOnly_False(t *testing.T) {
	mgr := NewSatTokenMgr(false)

	result := mgr.TestOnly()

	assert.False(t, result)
}

func TestSatTokenMgr_SetTestOnly_True(t *testing.T) {
	mgr := NewSatTokenMgr(false)

	mgr.SetTestOnly(true)

	assert.True(t, mgr.testOnly)
	assert.True(t, mgr.TestOnly())
}

func TestSatTokenMgr_SetTestOnly_False(t *testing.T) {
	mgr := NewSatTokenMgr(true)

	mgr.SetTestOnly(false)

	assert.False(t, mgr.testOnly)
	assert.False(t, mgr.TestOnly())
}

func TestSatTokenMgr_SetTestOnly_Toggle(t *testing.T) {
	mgr := NewSatTokenMgr()

	// Start false (default)
	assert.False(t, mgr.TestOnly())

	// Toggle to true
	mgr.SetTestOnly(true)
	assert.True(t, mgr.TestOnly())

	// Toggle back to false
	mgr.SetTestOnly(false)
	assert.False(t, mgr.TestOnly())
}

// Test SatToken structure
func TestSatToken_AllFields(t *testing.T) {
	token := &SatToken{
		Token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		Source:   "sat-service-prod",
		KeyName:  "sat_token_prod",
		Expiry:   "2025-12-31 23:59:59",
		TokenTTL: 3600,
	}

	assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", token.Token)
	assert.Equal(t, "sat-service-prod", token.Source)
	assert.Equal(t, "sat_token_prod", token.KeyName)
	assert.Equal(t, "2025-12-31 23:59:59", token.Expiry)
	assert.Equal(t, 3600, token.TokenTTL)
}

func TestSatToken_EmptyFields(t *testing.T) {
	token := &SatToken{}

	assert.Equal(t, "", token.Token)
	assert.Equal(t, "", token.Source)
	assert.Equal(t, "", token.KeyName)
	assert.Equal(t, "", token.Expiry)
	assert.Equal(t, 0, token.TokenTTL)
}

func TestSatToken_PartialFields(t *testing.T) {
	token := &SatToken{
		Token:  "token123",
		Source: "test-source",
	}

	assert.Equal(t, "token123", token.Token)
	assert.Equal(t, "test-source", token.Source)
	assert.Equal(t, "", token.KeyName)
	assert.Equal(t, "", token.Expiry)
	assert.Equal(t, 0, token.TokenTTL)
}

func TestNewSatTokenMgr_NoArgs(t *testing.T) {
	mgr := NewSatTokenMgr()

	assert.NotNil(t, mgr)
	assert.NotNil(t, mgr.SatToken)
	assert.False(t, mgr.testOnly)
}

func TestNewSatTokenMgr_WithTrueArg(t *testing.T) {
	mgr := NewSatTokenMgr(true)

	assert.NotNil(t, mgr)
	assert.NotNil(t, mgr.SatToken)
	assert.True(t, mgr.testOnly)
}

func TestNewSatTokenMgr_WithFalseArg(t *testing.T) {
	mgr := NewSatTokenMgr(false)

	assert.NotNil(t, mgr)
	assert.NotNil(t, mgr.SatToken)
	assert.False(t, mgr.testOnly)
}

func TestSatToken_LongExpiry(t *testing.T) {
	token := &SatToken{
		Token:    "long-lived-token",
		Source:   "sat-prod",
		KeyName:  "sat_token_key",
		Expiry:   "2030-01-01 00:00:00",
		TokenTTL: 86400, // 24 hours
	}

	assert.Equal(t, "long-lived-token", token.Token)
	assert.Equal(t, 86400, token.TokenTTL)
}

func TestSatToken_ShortExpiry(t *testing.T) {
	token := &SatToken{
		Token:    "short-lived-token",
		Source:   "sat-dev",
		KeyName:  "sat_token_dev",
		Expiry:   "2025-11-01 00:01:00",
		TokenTTL: 60, // 1 minute
	}

	assert.Equal(t, "short-lived-token", token.Token)
	assert.Equal(t, 60, token.TokenTTL)
}

// Test functions using existing mock infrastructure for 0% coverage functions

func TestGetSatTokenManager_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	// Test GetSatTokenManager function
	manager := GetSatTokenManager()
	assert.NotNil(t, manager)
}

func TestGetSatToken_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure
	InitSatTokenManager(server)

	manager := GetSatTokenManager()
	fields := log.Fields{"test": "GetSatToken"}

	// Test GetSatToken function with proper fields parameter
	token, err := manager.GetSatToken(fields)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGetLocalSatToken_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure
	InitSatTokenManager(server)

	fields := log.Fields{"test": "GetLocalSatToken"}

	// Test GetLocalSatToken function with proper fields parameter
	token, err := GetLocalSatToken(fields)
	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestSetLocalSatToken_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure
	InitSatTokenManager(server)

	fields := log.Fields{"test": "SetLocalSatToken"}

	// Test SetLocalSatToken function with proper fields parameter
	err := SetLocalSatToken(fields)
	assert.NoError(t, err)

	// Verify the token was set by getting it back
	retrievedToken, err := GetLocalSatToken(fields)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedToken)
}

func TestGetSatTokenFromSatService_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure
	InitSatTokenManager(server)

	fields := log.Fields{"test": "GetSatTokenFromSatService"}

	// Test GetSatTokenFromSatService function using existing Ws mock
	token, err := GetSatTokenFromSatService(fields)
	// Since we're using mocks, we expect successful execution
	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestIsTokenExpired_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure
	InitSatTokenManager(server)

	fields := log.Fields{"test": "IsTokenExpired"}

	// Test IsTokenExpired with an expired token
	expiredToken := &SatToken{
		Token:    "expired-token",
		Expiry:   "2020-01-01 00:00:00", // Past date
		TokenTTL: 3600,
	}

	isExpired := expiredToken.IsTokenExpired(fields)
	assert.True(t, isExpired)

	// Test IsTokenExpired with a valid token
	futureToken := &SatToken{
		Token:    "future-token",
		Expiry:   "2030-12-31 23:59:59", // Future date
		TokenTTL: 3600,
	}

	isExpired = futureToken.IsTokenExpired(fields)
	assert.False(t, isExpired)
}

func TestGetTokenExpiryTime_WithMockInfrastructure(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure
	InitSatTokenManager(server)

	// Test GetTokenExpiryTime function
	expiryTime := GetTokenExpiryTime()
	assert.NotEmpty(t, expiryTime)

	// Verify it's a valid timestamp format
	_, err := time.Parse("2006-01-02 15:04:05", expiryTime)
	assert.NoError(t, err)
}
