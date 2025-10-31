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
