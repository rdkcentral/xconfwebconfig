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
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	"gotest.tools/assert"
)

func TestNewLocation(t *testing.T) {
	loc := NewLocation("192.168.1.1", 50.5)
	
	assert.Assert(t, loc != nil)
	assert.Equal(t, "192.168.1.1", loc.LocationIp)
	assert.Equal(t, 50.5, loc.Percentage)
}

func TestLocation_Creation(t *testing.T) {
	loc := Location{
		LocationIp: "10.0.0.1",
		Percentage: 75.0,
	}
	
	assert.Equal(t, "10.0.0.1", loc.LocationIp)
	assert.Equal(t, 75.0, loc.Percentage)
}

func TestLocation_ZeroPercentage(t *testing.T) {
	loc := NewLocation("172.16.0.1", 0.0)
	
	assert.Equal(t, "172.16.0.1", loc.LocationIp)
	assert.Equal(t, 0.0, loc.Percentage)
}

func TestLocation_FullPercentage(t *testing.T) {
	loc := NewLocation("192.168.100.1", 100.0)
	
	assert.Equal(t, "192.168.100.1", loc.LocationIp)
	assert.Equal(t, 100.0, loc.Percentage)
}

func TestNewDownloadLocationRoundRobinFilterValue(t *testing.T) {
	obj := NewDownloadLocationRoundRobinFilterValue()
	
	assert.Assert(t, obj != nil)
	dlrrfv, ok := obj.(*DownloadLocationRoundRobinFilterValue)
	assert.Assert(t, ok)
	assert.Equal(t, ROUND_ROBIN_FILTER_SINGLETON_ID, dlrrfv.ID)
	assert.Equal(t, RoundRobinFilterClass, dlrrfv.Type)
	assert.Equal(t, shared.STB, dlrrfv.ApplicationType)
}

func TestNewEmptyDownloadLocationRoundRobinFilterValue(t *testing.T) {
	dlrrfv := NewEmptyDownloadLocationRoundRobinFilterValue()
	
	assert.Assert(t, dlrrfv != nil)
	assert.Equal(t, ROUND_ROBIN_FILTER_SINGLETON_ID, dlrrfv.ID)
	assert.Equal(t, RoundRobinFilterClass, dlrrfv.Type)
	assert.Equal(t, shared.STB, dlrrfv.ApplicationType)
	assert.Assert(t, dlrrfv.Locations != nil)
	assert.Equal(t, 0, len(dlrrfv.Locations))
	assert.Assert(t, dlrrfv.Ipv6locations != nil)
	assert.Equal(t, 0, len(dlrrfv.Ipv6locations))
}

func TestDownloadLocationRoundRobinFilterValue_WithLocations(t *testing.T) {
	locations := []Location{
		{LocationIp: "192.168.1.1", Percentage: 50.0},
		{LocationIp: "192.168.1.2", Percentage: 50.0},
	}
	
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:              ROUND_ROBIN_FILTER_SINGLETON_ID,
		ApplicationType: shared.STB,
		Locations:       locations,
	}
	
	assert.Equal(t, 2, len(dlrrfv.Locations))
	assert.Equal(t, "192.168.1.1", dlrrfv.Locations[0].LocationIp)
	assert.Equal(t, 50.0, dlrrfv.Locations[0].Percentage)
	assert.Equal(t, "192.168.1.2", dlrrfv.Locations[1].LocationIp)
}

func TestDownloadLocationRoundRobinFilterValue_WithIpv6Locations(t *testing.T) {
	ipv6locations := []Location{
		{LocationIp: "2001:db8::1", Percentage: 33.33},
		{LocationIp: "2001:db8::2", Percentage: 33.33},
		{LocationIp: "2001:db8::3", Percentage: 33.34},
	}
	
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:              ROUND_ROBIN_FILTER_SINGLETON_ID,
		ApplicationType: shared.STB,
		Ipv6locations:   ipv6locations,
	}
	
	assert.Equal(t, 3, len(dlrrfv.Ipv6locations))
	assert.Equal(t, "2001:db8::1", dlrrfv.Ipv6locations[0].LocationIp)
	assert.Equal(t, 33.33, dlrrfv.Ipv6locations[0].Percentage)
}

func TestDownloadLocationRoundRobinFilterValue_WithHttpLocation(t *testing.T) {
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:              ROUND_ROBIN_FILTER_SINGLETON_ID,
		ApplicationType: shared.STB,
		HttpLocation:    "http://example.com/firmware",
	}
	
	assert.Equal(t, "http://example.com/firmware", dlrrfv.HttpLocation)
}

func TestDownloadLocationRoundRobinFilterValue_WithHttpFullUrlLocation(t *testing.T) {
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:                  ROUND_ROBIN_FILTER_SINGLETON_ID,
		ApplicationType:     shared.STB,
		HttpFullUrlLocation: "http://example.com/firmware/v1.2.3/image.bin",
	}
	
	assert.Equal(t, "http://example.com/firmware/v1.2.3/image.bin", dlrrfv.HttpFullUrlLocation)
}

func TestDownloadLocationFilter_Creation(t *testing.T) {
	filter := &DownloadLocationFilter{
		Id:   "filter-1",
		Name: "Test Download Location Filter",
	}
	
	assert.Equal(t, "filter-1", filter.Id)
	assert.Equal(t, "Test Download Location Filter", filter.Name)
	assert.Assert(t, !filter.ForceHttp)
}

func TestDownloadLocationFilter_WithIpAddressGroup(t *testing.T) {
	ipGroup := &shared.IpAddressGroup{Id: "ip-1", Name: "IP Group"}
	filter := &DownloadLocationFilter{
		Id:             "filter-2",
		Name:           "Filter with IP Group",
		IpAddressGroup: ipGroup,
	}
	
	assert.Assert(t, filter.IpAddressGroup != nil)
	assert.Equal(t, "ip-1", filter.IpAddressGroup.Id)
	assert.Equal(t, "IP Group", filter.IpAddressGroup.Name)
}

func TestDownloadLocationFilter_WithEnvironments(t *testing.T) {
	filter := &DownloadLocationFilter{
		Id:           "filter-3",
		Name:         "Filter with Environments",
		Environments: []string{"DEV", "QA", "PROD"},
	}
	
	assert.Equal(t, 3, len(filter.Environments))
	assert.Equal(t, "DEV", filter.Environments[0])
	assert.Equal(t, "QA", filter.Environments[1])
	assert.Equal(t, "PROD", filter.Environments[2])
}

func TestDownloadLocationFilter_WithModels(t *testing.T) {
	filter := &DownloadLocationFilter{
		Id:     "filter-4",
		Name:   "Filter with Models",
		Models: []string{"MODEL-A", "MODEL-B", "MODEL-C"},
	}
	
	assert.Equal(t, 3, len(filter.Models))
	assert.Equal(t, "MODEL-A", filter.Models[0])
	assert.Equal(t, "MODEL-B", filter.Models[1])
	assert.Equal(t, "MODEL-C", filter.Models[2])
}

func TestDownloadLocationFilter_WithFirmwareLocations(t *testing.T) {
	ipv4Location := &shared.IpAddress{Address: "192.168.1.100"}
	ipv6Location := &shared.IpAddress{Address: "2001:db8::100"}
	
	filter := &DownloadLocationFilter{
		Id:                   "filter-5",
		Name:                 "Filter with Firmware Locations",
		FirmwareLocation:     ipv4Location,
		Ipv6FirmwareLocation: ipv6Location,
	}
	
	assert.Assert(t, filter.FirmwareLocation != nil)
	assert.Equal(t, "192.168.1.100", filter.FirmwareLocation.Address)
	assert.Assert(t, filter.Ipv6FirmwareLocation != nil)
	assert.Equal(t, "2001:db8::100", filter.Ipv6FirmwareLocation.Address)
}

func TestDownloadLocationFilter_WithHttpLocation(t *testing.T) {
	filter := &DownloadLocationFilter{
		Id:           "filter-6",
		Name:         "Filter with HTTP Location",
		HttpLocation: "http://cdn.example.com/firmware",
	}
	
	assert.Equal(t, "http://cdn.example.com/firmware", filter.HttpLocation)
}

func TestDownloadLocationFilter_WithForceHttp(t *testing.T) {
	filter := &DownloadLocationFilter{
		Id:        "filter-7",
		Name:      "Filter with Force HTTP",
		ForceHttp: true,
	}
	
	assert.Assert(t, filter.ForceHttp)
}

func TestDownloadLocationFilter_WithDownloadProtocol(t *testing.T) {
	filter := &DownloadLocationFilter{
		Id:                       "filter-8",
		Name:                     "Filter with Protocol",
		FirmwareDownloadProtocol: "https",
	}
	
	assert.Equal(t, "https", filter.FirmwareDownloadProtocol)
}

func TestDownloadLocationFilter_WithBoundConfigId(t *testing.T) {
	filter := &DownloadLocationFilter{
		Id:            "filter-9",
		Name:          "Filter with Bound Config",
		BoundConfigId: "config-123",
	}
	
	assert.Equal(t, "config-123", filter.BoundConfigId)
}

func TestDownloadLocationFilter_CompleteFilter(t *testing.T) {
	ipGroup := &shared.IpAddressGroup{Id: "ip-group", Name: "IP Group"}
	ipv4Location := &shared.IpAddress{Address: "10.1.2.3"}
	ipv6Location := &shared.IpAddress{Address: "fe80::1"}
	
	filter := &DownloadLocationFilter{
		Id:                       "complete-filter",
		Name:                     "Complete Download Location Filter",
		IpAddressGroup:           ipGroup,
		Environments:             []string{"PROD", "STAGE"},
		Models:                   []string{"MODEL-X", "MODEL-Y"},
		FirmwareDownloadProtocol: "https",
		FirmwareLocation:         ipv4Location,
		Ipv6FirmwareLocation:     ipv6Location,
		HttpLocation:             "http://example.com",
		ForceHttp:                true,
		BoundConfigId:            "bound-123",
	}
	
	// Verify all fields
	assert.Equal(t, "complete-filter", filter.Id)
	assert.Equal(t, "Complete Download Location Filter", filter.Name)
	assert.Assert(t, filter.IpAddressGroup != nil)
	assert.Equal(t, 2, len(filter.Environments))
	assert.Equal(t, 2, len(filter.Models))
	assert.Equal(t, "https", filter.FirmwareDownloadProtocol)
	assert.Assert(t, filter.FirmwareLocation != nil)
	assert.Assert(t, filter.Ipv6FirmwareLocation != nil)
	assert.Equal(t, "http://example.com", filter.HttpLocation)
	assert.Assert(t, filter.ForceHttp)
	assert.Equal(t, "bound-123", filter.BoundConfigId)
}

func TestDownloadLocationRoundRobinFilterValue_CompleteValue(t *testing.T) {
	locations := []Location{
		{LocationIp: "192.168.1.1", Percentage: 60.0},
		{LocationIp: "192.168.1.2", Percentage: 40.0},
	}
	ipv6locations := []Location{
		{LocationIp: "2001:db8::1", Percentage: 100.0},
	}
	
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:                  "XHOME_" + ROUND_ROBIN_FILTER_SINGLETON_ID,
		Updated:             1234567890,
		Type:                RoundRobinFilterClass,
		ApplicationType:     "xhome",
		Locations:           locations,
		Ipv6locations:       ipv6locations,
		HttpLocation:        "http://cdn.example.com",
		HttpFullUrlLocation: "http://cdn.example.com/firmware/image.bin",
	}
	
	assert.Equal(t, "XHOME_"+ROUND_ROBIN_FILTER_SINGLETON_ID, dlrrfv.ID)
	assert.Equal(t, int64(1234567890), dlrrfv.Updated)
	assert.Equal(t, RoundRobinFilterClass, dlrrfv.Type)
	assert.Equal(t, "xhome", dlrrfv.ApplicationType)
	assert.Equal(t, 2, len(dlrrfv.Locations))
	assert.Equal(t, 1, len(dlrrfv.Ipv6locations))
	assert.Equal(t, "http://cdn.example.com", dlrrfv.HttpLocation)
	assert.Equal(t, "http://cdn.example.com/firmware/image.bin", dlrrfv.HttpFullUrlLocation)
}

func TestDownloadLocationRoundRobinFilterValue_SetApplicationType(t *testing.T) {
	dlrrfv := NewEmptyDownloadLocationRoundRobinFilterValue()
	dlrrfv.SetApplicationType("xhome")
	
	assert.Equal(t, "xhome", dlrrfv.ApplicationType)
}

func TestDownloadLocationRoundRobinFilterValue_GetApplicationType(t *testing.T) {
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ApplicationType: "rdkcloud",
	}
	
	appType := dlrrfv.GetApplicationType()
	assert.Equal(t, "rdkcloud", appType)
}

func TestDownloadLocationRoundRobinFilterValue_Validate_Valid(t *testing.T) {
	locations := []Location{
		{LocationIp: "192.168.1.1", Percentage: 50.0},
		{LocationIp: "192.168.1.2", Percentage: 50.0},
	}
	
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:                  ROUND_ROBIN_FILTER_SINGLETON_ID,
		Type:                RoundRobinFilterClass,
		ApplicationType:     shared.STB,
		Locations:           locations,
		HttpFullUrlLocation: "http://example.com/firmware.bin",
	}
	
	err := dlrrfv.Validate()
	assert.NilError(t, err)
}

func TestDownloadLocationRoundRobinFilterValue_Validate_InvalidType(t *testing.T) {
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:                  ROUND_ROBIN_FILTER_SINGLETON_ID,
		Type:                PercentFilterClass, // Wrong type
		ApplicationType:     shared.STB,
		HttpFullUrlLocation: "http://example.com",
	}
	
	err := dlrrfv.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Type is invalid")
}

func TestDownloadLocationRoundRobinFilterValue_Validate_InvalidId(t *testing.T) {
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:                  "INVALID_ID",
		Type:                RoundRobinFilterClass,
		ApplicationType:     shared.STB,
		HttpFullUrlLocation: "http://example.com",
	}
	
	err := dlrrfv.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Id is invalid")
}

func TestDownloadLocationRoundRobinFilterValue_Validate_InvalidUrl(t *testing.T) {
	locations := []Location{
		{LocationIp: "192.168.1.1", Percentage: 100.0},
	}
	
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:                  ROUND_ROBIN_FILTER_SINGLETON_ID,
		Type:                RoundRobinFilterClass,
		ApplicationType:     shared.STB,
		Locations:           locations,
		HttpFullUrlLocation: "not-a-valid-url",
	}
	
	err := dlrrfv.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Location URL is not valid")
}

func TestDownloadLocationRoundRobinFilterValue_Validate_NegativePercentage(t *testing.T) {
	locations := []Location{
		{LocationIp: "192.168.1.1", Percentage: -10.0},
	}
	
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:                  ROUND_ROBIN_FILTER_SINGLETON_ID,
		Type:                RoundRobinFilterClass,
		ApplicationType:     shared.STB,
		Locations:           locations,
		HttpFullUrlLocation: "http://example.com",
	}
	
	err := dlrrfv.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Percentage cannot be negative")
}

func TestDownloadLocationRoundRobinFilterValue_Validate_InvalidPercentageSum(t *testing.T) {
	locations := []Location{
		{LocationIp: "192.168.1.1", Percentage: 60.0},
		{LocationIp: "192.168.1.2", Percentage: 30.0}, // Only 90% total
	}
	
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:                  ROUND_ROBIN_FILTER_SINGLETON_ID,
		Type:                RoundRobinFilterClass,
		ApplicationType:     shared.STB,
		Locations:           locations,
		HttpFullUrlLocation: "http://example.com",
	}
	
	err := dlrrfv.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Summary IPv4 percentage should be 100")
}

func TestDownloadLocationRoundRobinFilterValue_Validate_DuplicateLocations(t *testing.T) {
	locations := []Location{
		{LocationIp: "192.168.1.1", Percentage: 50.0},
		{LocationIp: "192.168.1.1", Percentage: 50.0}, // Duplicate
	}
	
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:                  ROUND_ROBIN_FILTER_SINGLETON_ID,
		Type:                RoundRobinFilterClass,
		ApplicationType:     shared.STB,
		Locations:           locations,
		HttpFullUrlLocation: "http://example.com",
	}
	
	err := dlrrfv.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Locations are duplicated")
}

func TestDownloadLocationRoundRobinFilterValue_Validate_Ipv6InIpv4List(t *testing.T) {
	locations := []Location{
		{LocationIp: "192.168.1.1", Percentage: 50.0},
		{LocationIp: "2001:db8::1", Percentage: 50.0}, // IPv6 in IPv4 list
	}
	
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:                  ROUND_ROBIN_FILTER_SINGLETON_ID,
		Type:                RoundRobinFilterClass,
		ApplicationType:     shared.STB,
		Locations:           locations,
		HttpFullUrlLocation: "http://example.com",
	}
	
	err := dlrrfv.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "IP address has an invalid version")
}

func TestDownloadLocationRoundRobinFilterValue_Validate_Ipv4InIpv6List(t *testing.T) {
	ipv6locations := []Location{
		{LocationIp: "2001:db8::1", Percentage: 50.0},
		{LocationIp: "192.168.1.1", Percentage: 50.0}, // IPv4 in IPv6 list
	}
	
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:                  ROUND_ROBIN_FILTER_SINGLETON_ID,
		Type:                RoundRobinFilterClass,
		ApplicationType:     shared.STB,
		Locations:           []Location{{LocationIp: "10.0.0.1", Percentage: 100.0}},
		Ipv6locations:       ipv6locations,
		HttpFullUrlLocation: "http://example.com",
	}
	
	err := dlrrfv.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "IP address has an invalid version")
}

func TestDownloadLocationRoundRobinFilterValue_Validate_InvalidIpv6Percentage(t *testing.T) {
	ipv6locations := []Location{
		{LocationIp: "2001:db8::1", Percentage: 60.0}, // Not 100%
	}
	
	dlrrfv := &DownloadLocationRoundRobinFilterValue{
		ID:                  ROUND_ROBIN_FILTER_SINGLETON_ID,
		Type:                RoundRobinFilterClass,
		ApplicationType:     shared.STB,
		Locations:           []Location{{LocationIp: "192.168.1.1", Percentage: 100.0}},
		Ipv6locations:       ipv6locations,
		HttpFullUrlLocation: "http://example.com",
	}
	
	err := dlrrfv.Validate()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "Summary IPv6 percentage should be 100")
}

func TestDownloadLocationFilter_NewEmptyDownloadLocationFilter(t *testing.T) {
	filter := NewEmptyDownloadLocationFilter()
	
	assert.Assert(t, filter != nil)
	assert.Equal(t, "", filter.Id)
	assert.Equal(t, "", filter.Name)
	assert.Assert(t, !filter.ForceHttp)
}
