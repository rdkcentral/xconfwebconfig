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
package tests

import (
	"strings"
	"testing"

	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	"gotest.tools/assert"
)

const (
	estbMacValue = "AA:AA:AA:AA:AA:AA"
	ipListName   = "ipList"
	ipAddress    = "10.10.10.10"
)

func TestConvertFirmwareRuleToIpRuleBeanAddFirmareConfig(t *testing.T) {
	firmwareRule4 := GetFirmwareRule4()
	assert.Assert(t, firmwareRule4.ID != "")

	firmwareConfig1 := GetFirmwareConfig1()
	assert.Assert(t, firmwareConfig1.ID != "")

	// store into DB
	err := corefw.CreateFirmwareRuleOneDB(firmwareRule4)
	assert.NilError(t, err)

	err = coreef.CreateFirmwareConfigOneDB(firmwareConfig1)
	assert.NilError(t, err)

	// name list to create IP groups in the DB
	genlist, err1 := GetSetDaoGenericNamespacedList()
	assert.NilError(t, err1)
	assert.Assert(t, genlist != nil)

	ipRuleBean := coreef.ConvertFirmwareRuleToIpRuleBean(firmwareRule4)
	assert.Assert(t, ipRuleBean.IpAddressGroup != nil)
	assert.Equal(t, ipRuleBean.IpAddressGroup.Id, NamespaceIPListKey)
	assert.Assert(t, len(ipRuleBean.IpAddressGroup.IpAddresses) == 3)
	assert.Assert(t, strings.EqualFold(ipRuleBean.IpAddressGroup.IpAddresses[0].GetAddress(), IpAddress4))
	assert.Assert(t, ipRuleBean.IpAddressGroup.IsInRange(IpAddress4))

	ipRuleBean, err = coreef.ConvertFirmwareRuleToIpRuleBeanAddFirmareConfig(firmwareRule4)
	assert.NilError(t, err1)
	firmwareConfig := ipRuleBean.FirmwareConfig
	assert.Assert(t, ipRuleBean.IpAddressGroup != nil)
	assert.Assert(t, firmwareConfig != nil)
	assert.Assert(t, len(ipRuleBean.IpAddressGroup.IpAddresses) == 3)
	assert.Assert(t, strings.EqualFold(ipRuleBean.IpAddressGroup.IpAddresses[0].GetAddress(), IpAddress4))
	assert.Assert(t, ipRuleBean.IpAddressGroup.IsInRange(IpAddress4))
}

func TestConvertToIpAddressGroup(t *testing.T) {
	genlist, err1 := GetSetDaoGenericNamespacedList()
	assert.NilError(t, err1)
	assert.Assert(t, genlist != nil)

	ipAddrGrp := shared.ConvertToIpAddressGroup(genlist)
	assert.Assert(t, ipAddrGrp != nil)
	assert.Assert(t, ipAddrGrp.IpAddresses != nil)
	assert.Equal(t, len(ipAddrGrp.IpAddresses), 3)
}

func TestConvertFirmwareRuleToIpFilter(t *testing.T) {
	firmwareRule := GetFirmwareRule1()
	assert.Assert(t, firmwareRule.ID != "")
	ipFilter := coreef.ConvertFirmwareRuleToIpFilter(firmwareRule)
	assert.Assert(t, ipFilter != nil)
	assert.Assert(t, ipFilter.Id != "")
	assert.Assert(t, ipFilter.Name != "")
	assert.Assert(t, ipFilter.IpAddressGroup == nil)

	firmwareRule = GetFirmwareRule2()
	assert.Assert(t, firmwareRule.ID != "")
	ipFilter = coreef.ConvertFirmwareRuleToIpFilter(firmwareRule)
	assert.Assert(t, ipFilter != nil)
	assert.Assert(t, ipFilter.Id != "")
	assert.Assert(t, ipFilter.Name != "")
	assert.Assert(t, ipFilter.IpAddressGroup == nil)

	firmwareRule = GetFirmwareRule3()
	assert.Assert(t, firmwareRule.ID != "")
	ipFilter = coreef.ConvertFirmwareRuleToIpFilter(firmwareRule)
	assert.Assert(t, ipFilter != nil)
	assert.Assert(t, ipFilter.Id != "")
	assert.Assert(t, ipFilter.Name != "")
	assert.Assert(t, ipFilter.IpAddressGroup != nil)
}

func createDownloadLocationFilter() *coreef.DownloadLocationFilter {
	filter := coreef.NewEmptyDownloadLocationFilter()
	filter.Id = "filterID"
	filter.Name = "filterName"
	filter.BoundConfigId = "configID"
	filter.ForceHttp = true
	filter.FirmwareLocation = shared.NewIpAddress("1.1.1")
	filter.Ipv6FirmwareLocation = shared.NewIpAddress("::1")
	filter.HttpLocation = "http://comcast.com"
	return filter
}

// based on Java DownloadLocationFilterConverterTest
func TestDownloadLocationFilterConverterConvertFilterWithTftpConditions(t *testing.T) {
	downloadLocFilter := GetGetRDKCDownloadLocationROUNDROBINFILTERVALUE()
	assert.Assert(t, downloadLocFilter != nil)

	filter := createDownloadLocationFilter()
	filter.HttpLocation = ""
	//assertConvertedEquals(filter);
	// todo
	/**
			rule = converter.convert(filter)
	        converted = converter.convert(rule)
	        assert.AssertEquals(filter, converted)
			**/
}

func TestIpFilterConverterConvertFirmwareRuleToIpFilterByMultipleRuleConditions(t *testing.T) {
	ipListPtr := CreateGenericNamespacedList(ipListName, shared.IpList, ipAddress)
	err := shared.CreateGenericNamedListOneDB(ipListPtr)
	assert.NilError(t, err)

	firmwareRule := createIpRule()
	rules := []*re.Rule{&firmwareRule.Rule}
	assert.Assert(t, rules[0].GetCompoundParts() != nil)
	assert.Equal(t, len(rules[0].GetCompoundParts()), 2)
	assert.Assert(t, rules[0].GetCompoundParts()[0].GetCondition().GetFreeArg().Equals(coreef.RuleFactoryIP))
	assert.Assert(t, rules[0].GetCompoundParts()[1].GetCondition().GetFreeArg().Equals(coreef.RuleFactoryMAC))

	// validat the ToConditions
	conds := re.ToConditions(&firmwareRule.Rule)
	assert.Assert(t, conds != nil)
	assert.Equal(t, len(conds), 2)
	assert.Assert(t, conds[0].GetFreeArg().Equals(coreef.RuleFactoryIP))
	assert.Assert(t, conds[1].GetFreeArg().Equals(coreef.RuleFactoryMAC))

	// convert to IPFilter
	ipFilter := coreef.ConvertFirmwareRuleToIpFilter(firmwareRule)

	assert.Equal(t, firmwareRule.ID, ipFilter.Id)
	assert.Equal(t, firmwareRule.Name, ipFilter.Name)
	assert.Assert(t, ipFilter.IpAddressGroup != nil)
	assert.Assert(t, ipFilter.IpAddressGroup.IpAddresses != nil)
	assert.Assert(t, len(ipFilter.IpAddressGroup.IpAddresses) == 1)
	ipaddrs1 := ipFilter.IpAddressGroup.IpAddresses
	ipaddrs2 := []shared.IpAddress{*shared.NewIpAddress(ipAddress)}
	assert.Assert(t, ipaddrs1[0].Equals(ipaddrs2[0]))

	// convertAndVerify
	firmwareRuleConverted := coreef.ConvertIpFilterToFirmwareRule(ipFilter)
	assert.Equal(t, firmwareRuleConverted.ID, firmwareRule.ID)
	assert.Equal(t, firmwareRuleConverted.Name, firmwareRule.Name)
	//assert.Assert(t, firmwareRuleConverted.Rule.Equals(&firmwareRule.Rule))
	// assert.Assert(t, firmwareRuleConverted.Equals(firmwareRule))
	// compare the firmareRule vs firmwareRuleConverted
	converted := coreef.ConvertFirmwareRuleToIpFilter(firmwareRuleConverted)

	//todo
	if 1 == 0 {
		assert.Equal(t, ipFilter, converted)
	}
}

func createIpRule() *corefw.FirmwareRule {
	firmwareRule := corefw.NewEmptyFirmwareRule()
	firmwareRule.Rule = re.Rule{}
	firmwareRule.ID = "firmwareRuleId"
	firmwareRule.Name = "firmwareRuleName"
	firmwareRule.Type = firmware.IP_RULE
	firmwareRule.Rule.SetCompoundParts([]re.Rule{})
	firmwareRule.Rule.AddCompoundPart(*CreateRule("", *coreef.RuleFactoryIP, coreef.RuleFactoryIN_LIST, ipListName))
	firmwareRule.Rule.AddCompoundPart(*CreateRule(re.RelationAnd, *coreef.RuleFactoryMAC, coreef.RuleFactoryIN_LIST, estbMacValue))
	return firmwareRule
}
