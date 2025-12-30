/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
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
	"testing"

	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/stretchr/testify/assert"
)

func TestNewEvaluationResult(t *testing.T) {
	result := NewEvaluationResult()

	assert.NotNil(t, result)
	assert.Nil(t, result.MatchedRule)
	assert.NotNil(t, result.AppliedFilters)
	assert.Empty(t, result.AppliedFilters)
	assert.Nil(t, result.FirmwareConfig)
	assert.Equal(t, "", result.Description)
	assert.False(t, result.Blocked)
	assert.NotNil(t, result.AppliedVersionInfo)
	assert.Empty(t, result.AppliedVersionInfo)
}

func TestAddAppliedFilters(t *testing.T) {
	t.Run("AddFilterToExistingSlice", func(t *testing.T) {
		result := NewEvaluationResult()

		filter1 := "filter1"
		filter2 := &firmware.FirmwareRule{ID: "rule1"}

		result.AddAppliedFilters(filter1)
		result.AddAppliedFilters(filter2)

		assert.Len(t, result.AppliedFilters, 2)
		assert.Equal(t, filter1, result.AppliedFilters[0])
		assert.Equal(t, filter2, result.AppliedFilters[1])
	})

	t.Run("AddFilterToNilSlice", func(t *testing.T) {
		result := &EvaluationResult{
			AppliedFilters: nil,
		}

		filter := "testFilter"
		result.AddAppliedFilters(filter)

		assert.NotNil(t, result.AppliedFilters)
		assert.Len(t, result.AppliedFilters, 1)
		assert.Equal(t, filter, result.AppliedFilters[0])
	})

	t.Run("AddMultipleFilterTypes", func(t *testing.T) {
		result := NewEvaluationResult()

		stringFilter := "stringFilter"
		ruleFilter := &firmware.FirmwareRule{ID: "rule1", Name: "TestRule"}
		percentFilter := coreef.PercentFilterValue{Percentage: 50.0}

		result.AddAppliedFilters(stringFilter)
		result.AddAppliedFilters(ruleFilter)
		result.AddAppliedFilters(percentFilter)

		assert.Len(t, result.AppliedFilters, 3)
		assert.Equal(t, stringFilter, result.AppliedFilters[0])
		assert.Equal(t, ruleFilter, result.AppliedFilters[1])
		assert.Equal(t, percentFilter, result.AppliedFilters[2])
	})
}

func TestDownloadLocationRoundRobinFilterContainsVersion(t *testing.T) {
	t.Run("VersionPresentInList", func(t *testing.T) {
		firmwareVersions := "1.0.0 2.0.0 3.0.0"
		contextVersion := "2.0.0"

		result := DownloadLocationRoundRobinFilterContainsVersion(firmwareVersions, contextVersion)

		assert.True(t, result)
	})

	t.Run("VersionNotPresentInList", func(t *testing.T) {
		firmwareVersions := "1.0.0 2.0.0 3.0.0"
		contextVersion := "4.0.0"

		result := DownloadLocationRoundRobinFilterContainsVersion(firmwareVersions, contextVersion)

		assert.False(t, result)
	})

	t.Run("EmptyFirmwareVersionsList", func(t *testing.T) {
		firmwareVersions := ""
		contextVersion := "1.0.0"

		result := DownloadLocationRoundRobinFilterContainsVersion(firmwareVersions, contextVersion)

		assert.False(t, result)
	})

	t.Run("EmptyContextVersion", func(t *testing.T) {
		firmwareVersions := "1.0.0 2.0.0 3.0.0"
		contextVersion := ""

		result := DownloadLocationRoundRobinFilterContainsVersion(firmwareVersions, contextVersion)

		assert.False(t, result)
	})

	t.Run("SingleVersion", func(t *testing.T) {
		firmwareVersions := "1.0.0"
		contextVersion := "1.0.0"

		result := DownloadLocationRoundRobinFilterContainsVersion(firmwareVersions, contextVersion)

		assert.True(t, result)
	})

	t.Run("VersionAtStart", func(t *testing.T) {
		firmwareVersions := "1.0.0 2.0.0 3.0.0"
		contextVersion := "1.0.0"

		result := DownloadLocationRoundRobinFilterContainsVersion(firmwareVersions, contextVersion)

		assert.True(t, result)
	})

	t.Run("VersionAtEnd", func(t *testing.T) {
		firmwareVersions := "1.0.0 2.0.0 3.0.0"
		contextVersion := "3.0.0"

		result := DownloadLocationRoundRobinFilterContainsVersion(firmwareVersions, contextVersion)

		assert.True(t, result)
	})

	t.Run("PartialMatchNotFound", func(t *testing.T) {
		firmwareVersions := "1.0.0 2.0.0 3.0.0"
		contextVersion := "2.0"

		result := DownloadLocationRoundRobinFilterContainsVersion(firmwareVersions, contextVersion)

		assert.False(t, result)
	})

	t.Run("MultipleSpacesSeparator", func(t *testing.T) {
		firmwareVersions := "1.0.0  2.0.0   3.0.0"
		contextVersion := "2.0.0"

		result := DownloadLocationRoundRobinFilterContainsVersion(firmwareVersions, contextVersion)

		assert.True(t, result)
	})
}

func TestFitsPercentByAccountId(t *testing.T) {
	t.Run("ZeroPercent", func(t *testing.T) {
		accountId := "test-account-123"
		percent := 0.0

		result := FitsPercentByAccountId(accountId, percent)

		assert.False(t, result)
	})

	t.Run("HundredPercent", func(t *testing.T) {
		accountId := "test-account-123"
		percent := 100.0

		result := FitsPercentByAccountId(accountId, percent)

		assert.True(t, result)
	})

	t.Run("FiftyPercent", func(t *testing.T) {
		accountId := "test-account-123"
		percent := 50.0

		result := FitsPercentByAccountId(accountId, percent)

		assert.True(t, result)
	})

	t.Run("VerySmallPercent", func(t *testing.T) {
		accountId := "test-account-123"
		percent := 0.0001

		result := FitsPercentByAccountId(accountId, percent)

		assert.True(t, result)
	})

	t.Run("DifferentAccountIds", func(t *testing.T) {
		accountId1 := "account-1"
		accountId2 := "account-2"
		percent := 50.0

		result1 := FitsPercentByAccountId(accountId1, percent)
		result2 := FitsPercentByAccountId(accountId2, percent)

		assert.Equal(t, result1, result2)
	})

	t.Run("EmptyAccountId", func(t *testing.T) {
		accountId := ""
		percent := 50.0

		result := FitsPercentByAccountId(accountId, percent)

		assert.True(t, result)
	})

	t.Run("NegativePercent", func(t *testing.T) {
		accountId := "test-account"
		percent := -10.0

		result := FitsPercentByAccountId(accountId, percent)

		assert.False(t, result)
	})

	t.Run("PercentGreaterThan100", func(t *testing.T) {
		accountId := "test-account"
		percent := 150.0

		result := FitsPercentByAccountId(accountId, percent)

		assert.True(t, result)
	})
}

func TestDownloadLocationRoundRobinFilterSetLocationByConnectionType(t *testing.T) {
	t.Run("SecureConnectionWithHTTP", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		fullHttpLocation := "http://example.com/firmware.bin"

		DownloadLocationRoundRobinFilterSetLocationByConnectionType(true, firmwareConfig, fullHttpLocation)

		location := firmwareConfig.Properties[coreef.FIRMWARE_LOCATION]
		assert.Equal(t, "https://example.com/firmware.bin", location)
	})

	t.Run("SecureConnectionWithHTTPS", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		fullHttpLocation := "https://example.com/firmware.bin"

		DownloadLocationRoundRobinFilterSetLocationByConnectionType(true, firmwareConfig, fullHttpLocation)

		location := firmwareConfig.Properties[coreef.FIRMWARE_LOCATION]
		assert.Equal(t, "https://example.com/firmware.bin", location)
	})

	t.Run("NonSecureConnectionWithHTTPS", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		fullHttpLocation := "https://example.com/firmware.bin"

		DownloadLocationRoundRobinFilterSetLocationByConnectionType(false, firmwareConfig, fullHttpLocation)

		location := firmwareConfig.Properties[coreef.FIRMWARE_LOCATION]
		assert.Equal(t, "http://example.com/firmware.bin", location)
	})

	t.Run("NonSecureConnectionWithHTTP", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		fullHttpLocation := "http://example.com/firmware.bin"

		DownloadLocationRoundRobinFilterSetLocationByConnectionType(false, firmwareConfig, fullHttpLocation)

		location := firmwareConfig.Properties[coreef.FIRMWARE_LOCATION]
		assert.Equal(t, "http://example.com/firmware.bin", location)
	})

	t.Run("MultipleHTTPSInURL", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		fullHttpLocation := "https://https-server.com/https/firmware.bin"

		DownloadLocationRoundRobinFilterSetLocationByConnectionType(false, firmwareConfig, fullHttpLocation)

		location := firmwareConfig.Properties[coreef.FIRMWARE_LOCATION]
		assert.Contains(t, location, "http")
	})
}

func TestDownloadLocationRoundRobinFilterSetupIPv4Location(t *testing.T) {
	t.Run("EmptyLocations", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		filterValue := &coreef.DownloadLocationRoundRobinFilterValue{
			Locations: []coreef.Location{},
		}

		result := DownloadLocationRoundRobinFilterSetupIPv4Location(firmwareConfig, filterValue)

		assert.False(t, result)
	})

	t.Run("NilLocations", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		filterValue := &coreef.DownloadLocationRoundRobinFilterValue{
			Locations: nil,
		}

		result := DownloadLocationRoundRobinFilterSetupIPv4Location(firmwareConfig, filterValue)

		assert.False(t, result)
	})

	t.Run("SingleLocationWith100Percent", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		filterValue := &coreef.DownloadLocationRoundRobinFilterValue{
			Locations: []coreef.Location{
				{LocationIp: "192.168.1.1", Percentage: 100.0},
			},
		}

		result := DownloadLocationRoundRobinFilterSetupIPv4Location(firmwareConfig, filterValue)

		assert.True(t, result)
		assert.NotNil(t, firmwareConfig.Properties[coreef.FIRMWARE_LOCATION])
	})

	t.Run("MultipleLocationsWithVariousPercentages", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		filterValue := &coreef.DownloadLocationRoundRobinFilterValue{
			Locations: []coreef.Location{
				{LocationIp: "192.168.1.1", Percentage: 30.0},
				{LocationIp: "192.168.1.2", Percentage: 30.0},
				{LocationIp: "192.168.1.3", Percentage: 40.0},
			},
		}

		result := DownloadLocationRoundRobinFilterSetupIPv4Location(firmwareConfig, filterValue)

		assert.True(t, result)
		location := firmwareConfig.Properties[coreef.FIRMWARE_LOCATION]
		assert.NotEmpty(t, location)
	})
}

func TestDownloadLocationRoundRobinFilterSetupIPv6Location(t *testing.T) {
	t.Run("EmptyIPv6Locations", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		filterValue := &coreef.DownloadLocationRoundRobinFilterValue{
			Ipv6locations: []coreef.Location{},
		}

		result := DownloadLocationRoundRobinFilterSetupIPv6Location(firmwareConfig, filterValue)

		assert.False(t, result)
	})

	t.Run("NilIPv6Locations", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		filterValue := &coreef.DownloadLocationRoundRobinFilterValue{
			Ipv6locations: nil,
		}

		result := DownloadLocationRoundRobinFilterSetupIPv6Location(firmwareConfig, filterValue)

		assert.False(t, result)
	})

	t.Run("SingleIPv6LocationWith100Percent", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		filterValue := &coreef.DownloadLocationRoundRobinFilterValue{
			Ipv6locations: []coreef.Location{
				{LocationIp: "2001:0db8::1", Percentage: 100.0},
			},
		}

		result := DownloadLocationRoundRobinFilterSetupIPv6Location(firmwareConfig, filterValue)

		assert.True(t, result)
		assert.NotNil(t, firmwareConfig.Properties[coreef.IPV6_FIRMWARE_LOCATION])
	})

	t.Run("MultipleIPv6LocationsWithVariousPercentages", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		filterValue := &coreef.DownloadLocationRoundRobinFilterValue{
			Ipv6locations: []coreef.Location{
				{LocationIp: "2001:0db8::1", Percentage: 25.0},
				{LocationIp: "2001:0db8::2", Percentage: 25.0},
				{LocationIp: "2001:0db8::3", Percentage: 50.0},
			},
		}

		result := DownloadLocationRoundRobinFilterSetupIPv6Location(firmwareConfig, filterValue)

		assert.True(t, result)
		location := firmwareConfig.Properties[coreef.IPV6_FIRMWARE_LOCATION]
		assert.NotEmpty(t, location)
	})
}

func TestDownloadLocationRoundRobinFilterFilter(t *testing.T) {
	t.Run("WithHttpLocationAndFullUrlLocation", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		filterValue := &coreef.DownloadLocationRoundRobinFilterValue{
			HttpLocation:        "192.168.1.1",
			HttpFullUrlLocation: "http://example.com/firmware.bin",
		}
		context := &coreef.ConvertedContext{
			SupportsFullHttpUrl: true,
			XconfHttpHeader:     "http",
		}

		result := DownloadLocationRoundRobinFilterFilter(firmwareConfig, filterValue, context)

		assert.True(t, result)
		assert.Equal(t, "http", firmwareConfig.Properties[coreef.FIRMWARE_DOWNLOAD_PROTOCOL])
	})

	t.Run("WithSecureConnectionAndFullUrl", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		filterValue := &coreef.DownloadLocationRoundRobinFilterValue{
			HttpLocation:        "192.168.1.1",
			HttpFullUrlLocation: "http://example.com/firmware.bin",
		}
		context := &coreef.ConvertedContext{
			SupportsFullHttpUrl: true,
			XconfHttpHeader:     "https",
		}

		result := DownloadLocationRoundRobinFilterFilter(firmwareConfig, filterValue, context)

		assert.True(t, result)
		assert.Equal(t, "http", firmwareConfig.Properties[coreef.FIRMWARE_DOWNLOAD_PROTOCOL])
		location := firmwareConfig.Properties[coreef.FIRMWARE_LOCATION]
		assert.NotNil(t, location)
	})

	t.Run("WithoutFullUrlSupport", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		filterValue := &coreef.DownloadLocationRoundRobinFilterValue{
			HttpLocation:        "192.168.1.1",
			HttpFullUrlLocation: "http://example.com/firmware.bin",
		}
		context := &coreef.ConvertedContext{
			SupportsFullHttpUrl: false,
			XconfHttpHeader:     "http",
		}

		result := DownloadLocationRoundRobinFilterFilter(firmwareConfig, filterValue, context)

		assert.True(t, result)
		location := firmwareConfig.Properties[coreef.FIRMWARE_LOCATION]
		assert.Equal(t, "192.168.1.1", location)
	})

	t.Run("WithIPv4AndIPv6Locations", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		filterValue := &coreef.DownloadLocationRoundRobinFilterValue{
			HttpLocation:        "",
			HttpFullUrlLocation: "",
			Locations: []coreef.Location{
				{LocationIp: "192.168.1.1", Percentage: 100.0},
			},
			Ipv6locations: []coreef.Location{
				{LocationIp: "2001:0db8::1", Percentage: 100.0},
			},
		}
		context := &coreef.ConvertedContext{
			SupportsFullHttpUrl: false,
			XconfHttpHeader:     "http",
		}

		result := DownloadLocationRoundRobinFilterFilter(firmwareConfig, filterValue, context)

		assert.True(t, result)
	})

	t.Run("WithEmptyLocations", func(t *testing.T) {
		firmwareConfig := &coreef.FirmwareConfigFacade{
			Properties: make(map[string]interface{}),
		}
		filterValue := &coreef.DownloadLocationRoundRobinFilterValue{
			HttpLocation:        "",
			HttpFullUrlLocation: "",
			Locations:           []coreef.Location{},
			Ipv6locations:       []coreef.Location{},
		}
		context := &coreef.ConvertedContext{
			SupportsFullHttpUrl: false,
			XconfHttpHeader:     "http",
		}

		result := DownloadLocationRoundRobinFilterFilter(firmwareConfig, filterValue, context)

		assert.False(t, result)
	})
}
