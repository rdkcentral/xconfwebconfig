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
package firmware

import (
	"testing"

	"gotest.tools/assert"
)

// Test ActivationVersion constructor
func TestNewActivationVersion(t *testing.T) {
	av := NewActivationVersion()

	assert.Assert(t, av != nil)
	assert.Assert(t, av.RegularExpressions != nil)
	assert.Equal(t, 0, len(av.RegularExpressions))
	assert.Assert(t, av.FirmwareVersions != nil)
	assert.Equal(t, 0, len(av.FirmwareVersions))
}

// Test ActivationVersion getters and setters
func TestActivationVersion_ApplicationType(t *testing.T) {
	av := NewActivationVersion()

	// Initially empty
	assert.Equal(t, "", av.GetApplicationType())

	// Set and get
	av.SetApplicationType("stb")
	assert.Equal(t, "stb", av.GetApplicationType())

	av.SetApplicationType("xhome")
	assert.Equal(t, "xhome", av.GetApplicationType())
}

// Test ActivationVersion fields
func TestActivationVersion_Fields(t *testing.T) {
	av := &ActivationVersion{
		ID:                 "test-id",
		ApplicationType:    "stb",
		Description:        "Test Activation",
		Model:              "MODEL-123",
		PartnerId:          "PARTNER-1",
		RegularExpressions: []string{"^1\\.2\\..*", "^2\\.0\\..*"},
		FirmwareVersions:   []string{"1.2.3", "2.0.1"},
	}

	assert.Equal(t, "test-id", av.ID)
	assert.Equal(t, "stb", av.ApplicationType)
	assert.Equal(t, "Test Activation", av.Description)
	assert.Equal(t, "MODEL-123", av.Model)
	assert.Equal(t, "PARTNER-1", av.PartnerId)
	assert.Equal(t, 2, len(av.RegularExpressions))
	assert.Equal(t, 2, len(av.FirmwareVersions))
}

// Test ActivationVersion with empty arrays
func TestActivationVersion_EmptyArrays(t *testing.T) {
	av := &ActivationVersion{
		ID:                 "test",
		RegularExpressions: []string{},
		FirmwareVersions:   []string{},
	}

	assert.Equal(t, 0, len(av.RegularExpressions))
	assert.Equal(t, 0, len(av.FirmwareVersions))
}

// Test ActivationVersion constants
func TestActivationVersion_Constants(t *testing.T) {
	assert.Equal(t, "DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE", SINGLETON_ID)
}
