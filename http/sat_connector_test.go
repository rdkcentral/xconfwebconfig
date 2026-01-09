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

// Test DefaultSatService getter/setter functions
func TestDefaultSatService_SatServiceName(t *testing.T) {
	service := &DefaultSatService{
		name: "sat-service-prod",
	}

	result := service.SatServiceName()

	assert.Equal(t, "sat-service-prod", result)
}

func TestDefaultSatService_SetSatServiceName(t *testing.T) {
	service := &DefaultSatService{
		name: "old-service-name",
	}

	service.SetSatServiceName("new-service-name")

	assert.Equal(t, "new-service-name", service.name)
}

func TestDefaultSatService_SatServiceHost(t *testing.T) {
	service := &DefaultSatService{
		host: "https://sat-service.example.com",
	}

	result := service.SatServiceHost()

	assert.Equal(t, "https://sat-service.example.com", result)
}

func TestDefaultSatService_SetSatServiceHost(t *testing.T) {
	service := &DefaultSatService{
		host: "https://old-sat-host.com",
	}

	service.SetSatServiceHost("https://new-sat-host.com")

	assert.Equal(t, "https://new-sat-host.com", service.host)
}

func TestDefaultSatService_ConsumerHost(t *testing.T) {
	service := &DefaultSatService{
		consumerHost: "https://consumer-host.example.com",
	}

	result := service.ConsumerHost()

	assert.Equal(t, "https://consumer-host.example.com", result)
}

func TestDefaultSatService_SatServiceName_Empty(t *testing.T) {
	service := &DefaultSatService{
		name: "",
	}

	result := service.SatServiceName()

	assert.Equal(t, "", result)
}

func TestDefaultSatService_SatServiceHost_Empty(t *testing.T) {
	service := &DefaultSatService{
		host: "",
	}

	result := service.SatServiceHost()

	assert.Equal(t, "", result)
}

func TestDefaultSatService_ConsumerHost_Empty(t *testing.T) {
	service := &DefaultSatService{
		consumerHost: "",
	}

	result := service.ConsumerHost()

	assert.Equal(t, "", result)
}

func TestDefaultSatService_AllGetters(t *testing.T) {
	service := &DefaultSatService{
		name:         "sat-prod",
		host:         "https://sat.example.com",
		consumerHost: "https://consumer.example.com",
	}

	assert.Equal(t, "sat-prod", service.SatServiceName())
	assert.Equal(t, "https://sat.example.com", service.SatServiceHost())
	assert.Equal(t, "https://consumer.example.com", service.ConsumerHost())
}

func TestDefaultSatService_MultipleUpdates(t *testing.T) {
	service := &DefaultSatService{}

	// First update
	service.SetSatServiceName("service1")
	service.SetSatServiceHost("https://host1.com")
	assert.Equal(t, "service1", service.name)
	assert.Equal(t, "https://host1.com", service.host)

	// Second update
	service.SetSatServiceName("service2")
	service.SetSatServiceHost("https://host2.com")
	assert.Equal(t, "service2", service.name)
	assert.Equal(t, "https://host2.com", service.host)
}

// Test SatServiceResponse structure
func TestSatServiceResponse_AllFields(t *testing.T) {
	response := SatServiceResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", // Mock JWT token for testing only - not a real credential
		ExpiresIn:    3600,
		Scope:        "read write",
		TokenType:    "Bearer",
		ResponseCode: 200,
		Description:  "Success",
	}

	assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", response.AccessToken)
	assert.Equal(t, 3600, response.ExpiresIn)
	assert.Equal(t, "read write", response.Scope)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, 200, response.ResponseCode)
	assert.Equal(t, "Success", response.Description)
}

func TestSatServiceResponse_EmptyFields(t *testing.T) {
	response := SatServiceResponse{}

	assert.Equal(t, "", response.AccessToken)
	assert.Equal(t, 0, response.ExpiresIn)
	assert.Equal(t, "", response.Scope)
	assert.Equal(t, "", response.TokenType)
	assert.Equal(t, 0, response.ResponseCode)
	assert.Equal(t, "", response.Description)
}

func TestSatServiceResponse_ErrorResponse(t *testing.T) {
	response := SatServiceResponse{
		ResponseCode: 401,
		Description:  "Unauthorized",
	}

	assert.Equal(t, "", response.AccessToken)
	assert.Equal(t, 401, response.ResponseCode)
	assert.Equal(t, "Unauthorized", response.Description)
}
