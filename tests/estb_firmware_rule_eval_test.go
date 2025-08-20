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
	"fmt"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

func TestExtractConfigFromAction(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	rulelst, _ := setUpRules(t)
	assert.Assert(t, rulelst != nil)

	contextMap := map[string]string{}
	contextMap["eStbMac"] = mac1
	contextMap["eCMMac"] = mac3
	contextMap["partnerId"] = "comcast"
	contextMap["ipAddress"] = IpAddress3
	contextMap["bypassFilters"] = "someFilter,bypassFilters,PercentFilter"
	contextMap["time"] = "time"
	contextMap["applicationType"] = "stb"
	e := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	convertedContext := coreef.GetContextConverted(contextMap)
	applyversions := map[string]string{}
	// result is in applyversion
	cfgId := e.ExtractConfigFromAction(convertedContext, rulelst[1].ApplicableAction, applyversions)
	assert.Assert(t, cfgId != "")
	assert.Equal(t, cfgId, FirmwareConfigId2)
	assert.Assert(t, len(applyversions) != 0)
	assert.Assert(t, applyversions[estbfirmware.FIRMWARE_SOURCE] != "")
	assert.Equal(t, applyversions[estbfirmware.FIRMWARE_SOURCE], "LKG,doesntMeetMinCheck")
	assert.Assert(t, len(convertedContext.ForceFilters) != 0)
	_, ok := convertedContext.ForceFilters[firmware.REBOOT_IMMEDIATELY_FILTER]
	assert.Assert(t, ok)

}

func TestIsInWhitelist(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	rulelst, _ := setUpRules(t)
	assert.Assert(t, rulelst != nil)

	contextMap := map[string]string{}
	contextMap["eStbMac"] = mac1
	contextMap["eCMMac"] = mac3
	contextMap["partnerId"] = "comcast"
	contextMap["ipAddress"] = IpAddress3
	contextMap["bypassFilters"] = "someFilter,bypassFilters,PercentFilter"
	contextMap["time"] = "time"
	contextMap["applicationType"] = "stb"
	e := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	convertedContext := coreef.GetContextConverted(contextMap)
	flag := e.IsInWhitelist(convertedContext, NamespaceIPListKey)
	assert.Assert(t, flag)

	flag = e.IsInWhitelist(convertedContext, "NotExistInDBNamespaceIPListKey")
	assert.Assert(t, !flag)
}

func TestFindMatchedRules(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	rulelst, _ := setUpRules(t)
	assert.Assert(t, rulelst != nil)

	contextMap := map[string]string{}
	contextMap["eStbMac"] = mac1
	contextMap["eCMMac"] = mac2
	contextMap["partnerId"] = "comcast"
	contextMap["ipAddress"] = IpAddress3
	contextMap["bypassFilters"] = "someFilter,bypassFilters,PercentFilter"
	contextMap["time"] = "time"
	contextMap["applicationType"] = "stb"
	e := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	convertedContext := coreef.GetContextConverted(contextMap)
	bypassFilters := convertedContext.GetBypassFiltersConverted()

	rules := e.FilterByAppType(rulelst, "stb")
	assert.Assert(t, rules != nil)

	copyrules := map[string][]*corefw.FirmwareRule{}

	for k, v := range rules {
		copyrules[k] = v
	}

	templates := e.GetSortedTemplate(corefw.RULE_TEMPLATE, false, log.Fields{})
	for _, template := range templates {
		ruleType := template.ID
		if _, ok := bypassFilters[ruleType]; ok {
			continue
		}

		// Java code based on google collection Collection<FirmwareRule> firmwareRules = rules.get(ruleType);
		firmwareRules := corefw.GetRulesByRuleTypes(rules, ruleType)
		if firmwareRules == nil || len(firmwareRules) == 0 {
			continue
		}

		if ruleType == "IP_RULE" {
			assert.Assert(t, len(firmwareRules) == 2)
		} else if ruleType == "IP_RULE" {
			assert.Assert(t, len(firmwareRules) == 1)
		} else if ruleType == "REMCTR_XR15-20_ENV_MODEL_RULE" {
			assert.Assert(t, len(firmwareRules) == 0)
		}
	}
	matchedRules := e.FindMatchedRules(rules, corefw.RULE_TEMPLATE, convertedContext.GetProperties(), bypassFilters, true, false, log.Fields{})
	assert.Assert(t, matchedRules != nil)

	assert.Assert(t, len(matchedRules) != 0)

	matchedRule := e.FindMatchedRule(copyrules, corefw.RULE_TEMPLATE, convertedContext.GetProperties(), bypassFilters, log.Fields{})
	assert.Assert(t, matchedRule != nil)

	matchedActivationVersionRule := e.FindMatchedRuleByRules(copyrules, corefw.DEFINE_PROPERTIES_TEMPLATE, firmware.ACTIVATION_VERSION, convertedContext.GetProperties(), convertedContext.GetBypassFiltersConverted())
	assert.Assert(t, matchedActivationVersionRule == nil)
}

func TestDoFilters(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	rulelst, _ := setUpRules(t)

	contextMap := map[string]string{}
	contextMap["eStbMac"] = "00:0a:95:9d:68:16"
	contextMap["eCMMac"] = "00:0a:95:9d:68:17"
	contextMap["partnerId"] = "comcast"
	contextMap["ipAddress"] = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	contextMap["bypassFilters"] = "someFilter,bypassFilters,PercentFilter"
	contextMap["time"] = "time"
	contextMap["applicationType"] = "stb"
	e := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	convertedContext := coreef.GetContextConverted(contextMap)
	result := estbfirmware.NewEvaluationResult()
	result.FirmwareConfig = coreef.NewDefaulttFirmwareConfigFacade()
	result.MatchedRule = corefw.NewEmptyFirmwareRule()
	rules := e.FilterByAppType(rulelst, "stb")
	blocked := e.DoFilters(contextMap, convertedContext, "stb", rules, result, log.Fields{})

	assert.Assert(t, blocked == false)
}

func TestFilterByTemplate(t *testing.T) {
	// t.Skip()
	// setup e
	firmwareRule1 := GetFirmwareRule1()
	assert.Assert(t, firmwareRule1.ID != "")

	firmwareRule2 := GetFirmwareRule2()
	assert.Assert(t, firmwareRule2.ID != "")

	rules := make([]*corefw.FirmwareRule, 2)
	rules[0] = firmwareRule1
	rules[1] = firmwareRule2

	// todo e := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	e := &estbfirmware.EstbFirmwareRuleBase{}
	filter_rules := e.FilterByTemplate(rules, "MAC_RULE")
	assert.Assert(t, filter_rules != nil)
	assert.Equal(t, len(filter_rules), 1)
	assert.Equal(t, filter_rules[0].ID, firmwareRuleId2)
}

func TestApplyMatchedFilters(t *testing.T) {
	// t.Skip()
	// setup e
	firmwareRule1 := GetFirmwareRule1()
	assert.Assert(t, firmwareRule1.ID != "")

	firmwareRule2 := GetFirmwareRule2()
	assert.Assert(t, firmwareRule2.ID != "")

	rules := map[string][]*corefw.FirmwareRule{}
	rules[firmwareRule1.Type] = []*corefw.FirmwareRule{firmwareRule1}
	rules[firmwareRule2.Type] = []*corefw.FirmwareRule{firmwareRule2}

	context := map[string]string{}
	bypassFilters := map[string]struct{}{}
	context[common.FIRMWARE_VERSION] = "DPC3939B_3.9p32s1_PROD_sey"

	// ses the default method to create EstbFirmwareRuleBase
	es := estbfirmware.NewEvaluationResult()
	e := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	//e := &estbfirmware.EstbFirmwareRuleBase{}
	res := e.ApplyMatchedFilters(rules, corefw.DEFINE_PROPERTIES_TEMPLATE, context, bypassFilters, es)
	assert.Assert(t, res != nil)
	t.Log(fmt.Sprintf("TestApplyMatchedFilters result of applyMarchedFilter %v", res))

}

func TestEvalEmpty(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	genlist, err1 := GetSetDaoGenericNamespacedList()
	assert.NilError(t, err1)
	assert.Assert(t, genlist != nil)

	contextMap := map[string]string{}
	contextMap["eStbMac"] = "00:0a:95:9d:68:16"
	contextMap["eCMMac"] = "00:0a:95:9d:68:17"
	contextMap["partnerId"] = "comcast"
	contextMap["ipAddress"] = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	contextMap["bypassFilters"] = "someFilter,bypassFilters,PercentFilter"
	contextMap["time"] = "time"
	contextMap["applicationType"] = "stb"

	e := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	convertedContext := coreef.GetContextConverted(contextMap)
	res, err := e.Eval(contextMap, convertedContext, "stb", log.Fields{})
	assert.Assert(t, err == nil)
	assert.Assert(t, res != nil)
}

func TestEval(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	setUpRules(t)

	// check the rule templte

	templates, errdb := corefw.GetFirmwareRuleTemplateAllAsListByActionType(corefw.DEFINE_PROPERTIES_TEMPLATE)
	assert.NilError(t, errdb)
	assert.Assert(t, templates != nil)

	templates, errdb = corefw.GetFirmwareRuleTemplateAllAsListByActionType(corefw.BLOCKING_FILTER_TEMPLATE)
	assert.NilError(t, errdb)
	assert.Assert(t, templates != nil)
	assert.Equal(t, len(templates), 1)

	templates, errdb = corefw.GetFirmwareRuleTemplateAllAsListByActionType(corefw.RULE_TEMPLATE)
	assert.NilError(t, errdb)
	assert.Assert(t, templates != nil)
	assert.Assert(t, len(templates) >= 2)

	contextMap := map[string]string{}
	contextMap["eStbMac"] = "00:0a:95:9d:68:16"
	contextMap["eCMMac"] = "00:0a:95:9d:68:17"
	contextMap["partnerId"] = "comcast"
	contextMap["ipAddress"] = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	contextMap["bypassFilters"] = "someFilter,bypassFilters,PercentFilter"
	contextMap["time"] = "time"
	contextMap["applicationType"] = "stb"
	e := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	//e := &estbfirmware.EstbFirmwareRuleBase{}
	convertedContext := coreef.GetContextConverted(contextMap)
	res, err := e.Eval(contextMap, convertedContext, "stb", log.Fields{})
	assert.Assert(t, err == nil)
	assert.Assert(t, res != nil)
}

func setUpRules(t *testing.T) ([]*corefw.FirmwareRule, []*coreef.FirmwareConfig) {
	//Prepare db data
	firmwareRule1 := GetFirmwareRule1()
	assert.Assert(t, firmwareRule1.ID != "")
	firmwareRule2 := GetFirmwareRule2()
	assert.Assert(t, firmwareRule2.ID != "")
	firmwareRule3 := GetFirmwareRule3()
	assert.Assert(t, firmwareRule3.ID != "")
	firmwareRule4 := GetFirmwareRule4()
	assert.Assert(t, firmwareRule4.ID != "")

	firmwareConfig1 := GetFirmwareConfig1()
	assert.Assert(t, firmwareConfig1.ID != "")
	firmwareConfig2 := GetFirmwareConfig2()
	assert.Assert(t, firmwareConfig2.ID != "")

	firmwareRuleTemplate1 := GetFirmwareRuleTemplate(1)
	assert.Assert(t, firmwareRuleTemplate1.ID != "")
	firmwareRuleTemplate2 := GetFirmwareRuleTemplate(2)
	assert.Assert(t, firmwareRuleTemplate1.ID != "")
	firmwareRuleTemplate3 := GetFirmwareRuleTemplate(3)
	assert.Assert(t, firmwareRuleTemplate1.ID != "")
	firmwareRuleTemplate4 := GetFirmwareRuleTemplate(4)
	assert.Assert(t, firmwareRuleTemplate1.ID != "")

	// store into DB
	err := corefw.CreateFirmwareRuleOneDB(firmwareRule1)
	assert.NilError(t, err)
	err = corefw.CreateFirmwareRuleOneDB(firmwareRule2)
	assert.NilError(t, err)
	err = corefw.CreateFirmwareRuleOneDB(firmwareRule3)
	assert.NilError(t, err)
	err = corefw.CreateFirmwareRuleOneDB(firmwareRule4)
	assert.NilError(t, err)

	err = corefw.CreateFirmwareRuleTemplateOneDB(firmwareRuleTemplate1)
	assert.NilError(t, err)
	err = corefw.CreateFirmwareRuleTemplateOneDB(firmwareRuleTemplate2)
	assert.NilError(t, err)
	err = corefw.CreateFirmwareRuleTemplateOneDB(firmwareRuleTemplate3)
	assert.NilError(t, err)
	err = corefw.CreateFirmwareRuleTemplateOneDB(firmwareRuleTemplate4)
	assert.NilError(t, err)

	// read from DB
	dbrules, dberr := corefw.GetFirmwareRuleAllAsListDB()
	assert.NilError(t, dberr)
	assert.Assert(t, dbrules != nil)
	assert.Assert(t, len(dbrules) != 0)
	assert.Assert(t, len(dbrules) >= 4)

	err = coreef.CreateFirmwareConfigOneDB(firmwareConfig1)
	assert.NilError(t, err)
	err = coreef.CreateFirmwareConfigOneDB(firmwareConfig2)
	assert.NilError(t, err)

	// read from DB
	dbcfs, dberr1 := coreef.GetFirmwareConfigAsListDB()
	assert.NilError(t, dberr1)
	assert.Assert(t, dbcfs != nil)
	assert.Assert(t, len(dbcfs) != 0)
	//assert.Equal(t, len(dbcfs), 2)

	// name list to create IP groups in the DB
	genlist, err1 := GetSetDaoGenericNamespacedList()
	assert.NilError(t, err1)
	assert.Assert(t, genlist != nil)

	dbgenlist, err2 := shared.GetGenericNamedListOneDB(genlist.ID)
	assert.NilError(t, err2)
	assert.Assert(t, dbgenlist != nil)

	dbgenlists, err3 := shared.GetGenericNamedListListsDB()
	assert.NilError(t, err3)
	assert.Assert(t, dbgenlists != nil)

	rules := make([]*corefw.FirmwareRule, 4)
	rules[0] = firmwareRule1
	rules[1] = firmwareRule2
	rules[2] = firmwareRule3
	rules[3] = firmwareRule4

	cfgs := make([]*coreef.FirmwareConfig, 2)
	cfgs[0] = firmwareConfig1
	cfgs[1] = firmwareConfig2

	return rules, cfgs
}
func TestGetBseConfiguration(t *testing.T) {
	//t.Skip()

	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	//clean up rules
	//truncateTable(ds.TABLE_FIRMWARE_RULE)

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

	rulelst, err := corefw.GetFirmwareRuleAllAsListDB()
	assert.NilError(t, err)
	assert.Assert(t, rulelst != nil)

	sortedrulelst, err := corefw.GetFirmwareSortedRuleAllAsListDB()
	assert.NilError(t, err)
	assert.Assert(t, sortedrulelst != nil)
	assert.Equal(t, len(rulelst), len(sortedrulelst))

	// Dowmload data based on
	// https://github.com/rdkcentral/xconfserver/blob/main/xconf-dataservice/src/test/java/com/comcast/xconf/estbfirmware/evaluation/DownloadLocationRoundRobinFilterTest.java
	dw := coreef.NewEmptyDownloadLocationRoundRobinFilterValue()
	dw.Locations = []coreef.Location{}
	dw.Ipv6locations = []coreef.Location{}
	dw.HttpLocation = DownloadLocationRoundRobinFilterHTTPLOCATION
	dw.HttpFullUrlLocation = DownloadLocationRoundRobinFilterHTTPFULLURLLOCATION
	loc := coreef.Location{
		LocationIp: DownloadLocationRoundRobinFilterIPADDRESS,
		Percentage: 100,
	}
	dw.Locations = append(dw.Ipv6locations, loc)
	dw.Ipv6locations = append(dw.Ipv6locations, loc)
	// store into DB
	err = coreef.CreateDownloadLocationRoundRobinFilterValOneDB(dw)
	assert.NilError(t, err)

	// check this Rules convension
	ipaddress := shared.NewIpAddress(IpAddress4)
	ipRuleBean, err := coreef.ConvertFirmwareRuleToIpRuleBeanAddFirmareConfig(firmwareRule4)
	assert.NilError(t, err)
	firmwareConfig := ipRuleBean.FirmwareConfig
	assert.Equal(t, firmware.IP_RULE, firmwareRule4.Type)
	assert.Assert(t, firmware.IP_RULE == firmwareRule4.Type)
	assert.Assert(t, !firmwareRule4.IsNoop())
	assert.Equal(t, firmwareRule4.ApplicationType, shared.STB)
	assert.Assert(t, firmwareRule4.ApplicationType == shared.STB)

	assert.Assert(t, firmwareConfig != nil)
	assert.Assert(t, len(ipRuleBean.IpAddressGroup.IpAddresses) == 3)
	assert.Assert(t, ipRuleBean.IpAddressGroup.IsInRange(IpAddress4))
	// failed assert.Assert(t, ipRuleBean.IpAddressGroup.IsInRange(ipaddress))
	assert.Assert(t, ipRuleBean.IpAddressGroup.IsInRange(ipaddress.GetAddress()))
	//assert.Assert(t, ipRuleBean.IpAddressGroup.IsInRange(ipaddress))

	e := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	config, err := e.GetBseConfiguration(ipaddress)
	assert.NilError(t, err)
	assert.Assert(t, config != nil)
}

func TestGetBseConfigurationSecondVersions(t *testing.T) {
	//t.Skip()

	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	setUpRules(t)

	dw := GetGetRDKCDownloadLocationROUNDROBINFILTERVALUE()
	// store into DB
	err := coreef.CreateDownloadLocationRoundRobinFilterValOneDB(dw)
	assert.NilError(t, err)

	// check this Rules convension
	ipaddress := shared.NewIpAddress(IpAddress4)

	e := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	config, err := e.GetBseConfiguration(ipaddress)
	assert.NilError(t, err)
	assert.Assert(t, config != nil)
}

func TestHasMinimumFirmware(t *testing.T) {
	//t.Skip()

	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	setUpRules(t)

	contextMap := map[string]string{}
	contextMap["eStbMac"] = "00:0a:95:9d:68:16"
	contextMap["eCMMac"] = "00:0a:95:9d:68:17"
	contextMap["partnerId"] = "comcast"
	contextMap["ipAddress"] = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	contextMap["bypassFilters"] = "someFilter,bypassFilters,PercentFilter"
	contextMap["time"] = "time"
	contextMap["applicationType"] = "stb"
	e := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	flag := e.HasMinimumFirmware(contextMap)
	assert.Assert(t, flag)
	convertedContext := coreef.GetContextConverted(contextMap)
	eval, err := e.Eval(contextMap, convertedContext, shared.STB, log.Fields{})
	assert.NilError(t, err)
	assert.Assert(t, eval != nil)
}

func TestGetBoundConfigId(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	rules, _ := setUpRules(t)

	contextMap := map[string]string{}
	contextMap["eStbMac"] = "00:0a:95:9d:68:16"
	contextMap["eCMMac"] = "00:0a:95:9d:68:17"
	contextMap["partnerId"] = "comcast"
	contextMap["ipAddress"] = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	contextMap["bypassFilters"] = "someFilter,bypassFilters,PercentFilter"
	contextMap["time"] = "time"
	contextMap["applicationType"] = "stb"
	convertedContext := coreef.GetContextConverted(contextMap)
	e := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	applyversions := map[string]string{}
	cfgId := e.GetBoundConfigId(contextMap, convertedContext, rules[0], applyversions)
	assert.Assert(t, cfgId != "")

}

func TestGetSortedTemplate(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	setUpRules(t)

	e := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	ruletemplates := e.GetSortedTemplate(corefw.RULE, false, log.Fields{})
	assert.Assert(t, ruletemplates == nil)
	ruletemplates = e.GetSortedTemplate(corefw.RULE_TEMPLATE, true, log.Fields{})
	assert.Assert(t, ruletemplates != nil)
	assert.Assert(t, len(ruletemplates) >= 2)
	// reverse is true, so should get the lowest priority rule template
	assert.Assert(t, ruletemplates[0].ID == "ENV_MODEL_RULE" || ruletemplates[0].ID == "IP_RULE")

	ruletemplates = e.GetSortedTemplate(corefw.RULE_TEMPLATE, false, log.Fields{})
	assert.Assert(t, ruletemplates != nil)
	assert.Assert(t, len(ruletemplates) >= 2)
	// reverse is false, so should get the highest priority rule template
	assert.Assert(t, ruletemplates[0].ID == "ENV_MODEL_RULE" || ruletemplates[0].ID == "MAC_RULE")
}
