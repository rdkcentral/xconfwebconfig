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
package estbfirmware

import (
	"encoding/json"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	"gotest.tools/assert"
)

func TestNewSingletonFilterValueInf(t *testing.T) {
	obj := NewSingletonFilterValueInf()
	assert.Assert(t, obj != nil)
	
	sfv, ok := obj.(*SingletonFilterValue)
	assert.Assert(t, ok)
	assert.Equal(t, "", sfv.ID)
	assert.Assert(t, sfv.PercentFilterValue == nil)
	assert.Assert(t, sfv.DownloadLocationRoundRobinFilterValue == nil)
}

func TestSingletonFilterValue_IsPercentFilterValue(t *testing.T) {
	// Test with PercentFilter ID
	sfv1 := &SingletonFilterValue{ID: "PERCENT_FILTER_VALUE"}
	assert.Assert(t, sfv1.IsPercentFilterValue())
	
	sfv2 := &SingletonFilterValue{ID: "TEST_PERCENT_FILTER_VALUE"}
	assert.Assert(t, sfv2.IsPercentFilterValue())
	
	// Test with non-PercentFilter ID
	sfv3 := &SingletonFilterValue{ID: "DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE"}
	assert.Assert(t, !sfv3.IsPercentFilterValue())
	
	sfv4 := &SingletonFilterValue{ID: "SOME_OTHER_VALUE"}
	assert.Assert(t, !sfv4.IsPercentFilterValue())
	
	// Test with empty ID
	sfv5 := &SingletonFilterValue{ID: ""}
	assert.Assert(t, !sfv5.IsPercentFilterValue())
}

func TestSingletonFilterValue_IsDownloadLocationRoundRobinFilterValue(t *testing.T) {
	// Test with RoundRobin ID
	sfv1 := &SingletonFilterValue{ID: "DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE"}
	assert.Assert(t, sfv1.IsDownloadLocationRoundRobinFilterValue())
	
	sfv2 := &SingletonFilterValue{ID: "TEST_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE"}
	assert.Assert(t, sfv2.IsDownloadLocationRoundRobinFilterValue())
	
	// Test with non-RoundRobin ID
	sfv3 := &SingletonFilterValue{ID: "PERCENT_FILTER_VALUE"}
	assert.Assert(t, !sfv3.IsDownloadLocationRoundRobinFilterValue())
	
	sfv4 := &SingletonFilterValue{ID: "SOME_OTHER_VALUE"}
	assert.Assert(t, !sfv4.IsDownloadLocationRoundRobinFilterValue())
	
	// Test with empty ID
	sfv5 := &SingletonFilterValue{ID: ""}
	assert.Assert(t, !sfv5.IsDownloadLocationRoundRobinFilterValue())
}

func TestGetRoundRobinIdByApplication_STB(t *testing.T) {
	id := GetRoundRobinIdByApplication(shared.STB)
	assert.Equal(t, "DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE", id)
}

func TestGetRoundRobinIdByApplication_Other(t *testing.T) {
	id1 := GetRoundRobinIdByApplication("xhome")
	assert.Equal(t, "XHOME_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE", id1)
	
	id2 := GetRoundRobinIdByApplication("rdkcloud")
	assert.Equal(t, "RDKCLOUD_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE", id2)
	
	id3 := GetRoundRobinIdByApplication("sky")
	assert.Equal(t, "SKY_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE", id3)
}

func TestGetRoundRobinIdByApplication_Empty(t *testing.T) {
	id := GetRoundRobinIdByApplication("")
	assert.Equal(t, "_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE", id)
}

func TestSingletonFilterValue_Clone_PercentFilter(t *testing.T) {
	whitelist := &shared.IpAddressGroup{Id: "whitelist-1"}
	original := &SingletonFilterValue{
		ID: "TEST_PERCENT_FILTER_VALUE",
		PercentFilterValue: &PercentFilterValue{
			ID:         "TEST_PERCENT_FILTER_VALUE",
			Percentage: 50.0,
			Whitelist:  whitelist,
		},
	}
	
	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Assert(t, cloned.PercentFilterValue != nil)
	
	// Verify deep copy
	assert.Assert(t, original.PercentFilterValue != cloned.PercentFilterValue)
}

func TestSingletonFilterValue_Clone_RoundRobin(t *testing.T) {
	original := &SingletonFilterValue{
		ID: "TEST_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE",
		DownloadLocationRoundRobinFilterValue: &DownloadLocationRoundRobinFilterValue{
			ID: "TEST_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE",
		},
	}
	
	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Assert(t, cloned.DownloadLocationRoundRobinFilterValue != nil)
}

func TestSingletonFilterValue_MarshalJSON_PercentFilter(t *testing.T) {
	sfv := &SingletonFilterValue{
		ID: "TEST_PERCENT_FILTER_VALUE",
		PercentFilterValue: &PercentFilterValue{
			ID:         "TEST_PERCENT_FILTER_VALUE",
			Percentage: 75.5,
		},
	}
	
	bytes, err := json.Marshal(sfv)
	assert.NilError(t, err)
	assert.Assert(t, len(bytes) > 0)
	
	// Verify the JSON contains percentage
	var result map[string]interface{}
	err = json.Unmarshal(bytes, &result)
	assert.NilError(t, err)
	assert.Equal(t, "TEST_PERCENT_FILTER_VALUE", result["id"])
}

func TestSingletonFilterValue_MarshalJSON_RoundRobin(t *testing.T) {
	locations := []Location{
		{LocationIp: "192.168.1.1", Percentage: 50.0},
		{LocationIp: "192.168.1.2", Percentage: 50.0},
	}
	
	sfv := &SingletonFilterValue{
		ID: "TEST_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE",
		DownloadLocationRoundRobinFilterValue: &DownloadLocationRoundRobinFilterValue{
			ID:        "TEST_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE",
			Locations: locations,
		},
	}
	
	bytes, err := json.Marshal(sfv)
	assert.NilError(t, err)
	assert.Assert(t, len(bytes) > 0)
	
	// Verify the JSON contains locations
	var result map[string]interface{}
	err = json.Unmarshal(bytes, &result)
	assert.NilError(t, err)
	assert.Equal(t, "TEST_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE", result["id"])
}

func TestSingletonFilterValue_MarshalJSON_Invalid(t *testing.T) {
	sfv := &SingletonFilterValue{
		ID: "TEST_ID",
		// Both subtype fields are nil or empty
	}
	
	_, err := json.Marshal(sfv)
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Invalid SingletonFilterValue")
}

func TestSingletonFilterValue_UnmarshalJSON_PercentFilter(t *testing.T) {
	jsonData := `{
		"id": "TEST_PERCENT_FILTER_VALUE",
		"percentage": 60.0
	}`
	
	var sfv SingletonFilterValue
	err := json.Unmarshal([]byte(jsonData), &sfv)
	assert.NilError(t, err)
	assert.Equal(t, "TEST_PERCENT_FILTER_VALUE", sfv.ID)
	assert.Assert(t, sfv.IsPercentFilterValue())
	assert.Assert(t, sfv.PercentFilterValue != nil)
	assert.Equal(t, float32(60.0), sfv.PercentFilterValue.Percentage)
}

func TestSingletonFilterValue_UnmarshalJSON_RoundRobin(t *testing.T) {
	jsonData := `{
		"id": "TEST_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE",
		"locations": []
	}`
	
	var sfv SingletonFilterValue
	err := json.Unmarshal([]byte(jsonData), &sfv)
	assert.NilError(t, err)
	assert.Equal(t, "TEST_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE", sfv.ID)
	assert.Assert(t, sfv.IsDownloadLocationRoundRobinFilterValue())
	assert.Assert(t, sfv.DownloadLocationRoundRobinFilterValue != nil)
}

func TestSingletonFilterValue_UnmarshalJSON_InvalidID(t *testing.T) {
	jsonData := `{
		"id": "INVALID_ID_TYPE"
	}`
	
	var sfv SingletonFilterValue
	err := json.Unmarshal([]byte(jsonData), &sfv)
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Invalid ID for SingletonFilterValue")
}

func TestSingletonFilterValue_UnmarshalJSON_InvalidJSON(t *testing.T) {
	jsonData := `{invalid json}`
	
	var sfv SingletonFilterValue
	err := json.Unmarshal([]byte(jsonData), &sfv)
	assert.Assert(t, err != nil)
}

func TestSingletonFilterValue_RoundTripSerialization_PercentFilter(t *testing.T) {
	original := &SingletonFilterValue{
		ID: "APP_PERCENT_FILTER_VALUE",
		PercentFilterValue: &PercentFilterValue{
			ID:         "APP_PERCENT_FILTER_VALUE",
			Percentage: 45.7,
		},
	}
	
	// Marshal
	bytes, err := json.Marshal(original)
	assert.NilError(t, err)
	
	// Unmarshal
	var restored SingletonFilterValue
	err = json.Unmarshal(bytes, &restored)
	assert.NilError(t, err)
	
	// Verify
	assert.Equal(t, original.ID, restored.ID)
	assert.Assert(t, restored.PercentFilterValue != nil)
	assert.Equal(t, original.PercentFilterValue.Percentage, restored.PercentFilterValue.Percentage)
}
