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
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
)

var (
	RuleFactoryMAC                  = re.NewFreeArg("STRING", common.ESTB_MAC)
	RuleFactoryIP                   = re.NewFreeArg("STRING", common.IP_ADDRESS)
	RuleFactoryVERSION              = re.NewFreeArg("STRING", common.FIRMWARE_VERSION)
	RuleFactoryENV                  = re.NewFreeArg("STRING", common.ENV)
	RuleFactoryMODEL                = re.NewFreeArg("STRING", common.MODEL)
	RuleFactoryFIRMWARE_VERSION     = re.NewFreeArg("STRING", common.FIRMWARE_VERSION)
	RuleFactoryREGEX                = re.NewFreeArg("STRING", common.FIRMWARE_VERSION)
	RuleFactoryMATCHED_RULE_TYPE    = re.NewFreeArg("STRING", common.MATCHED_RULE_TYPE)
	RuleFactoryTIME_ZONE            = re.NewFreeArg("STRING", common.TIME_ZONE) // may be "UTC"
	RuleFactoryTIME                 = re.NewFreeArg("TIME", common.TIME)
	RuleFactoryLOCAL_TIME           = re.NewFreeArg("STRING", common.TIME)
	RuleFactoryCERT_EXPIRY_DURATION = re.NewFreeArg("LONG", common.CERT_EXPIRY_DURATION)
	// required for TimeFilter. it must be added after rules matching
	RuleFactoryFIRMWARE_DOWNLOAD_PROTOCOL = re.NewFreeArg("STRING", common.DOWNLOAD_PROTOCOL) // tftp or http
	RuleFactoryREBOOT_DECOUPLED           = re.NewFreeArg("ANY", common.REBOOT_DECOUPLED)
	RuleFactoryPARTNER_ID                 = re.NewFreeArg("STRING", common.PARTNER_ID)
	RuleFactoryHTTP                       = re.NewFixedArg("http")
	RuleFactoryUTC                        = re.NewFixedArg("UTC")
	RuleFactoryIN_LIST                    = "IN_LIST"
	/**
	    MATCH = re.Operation.forName("MATCH")
	    RANGE = re.Operation.forName("RANGE")
	**/
	GlbRuleFactory *RuleFactory
)

type RuleFactory struct {
}

// NewRuleFactory ...
func NewRuleFactory() *RuleFactory {
	if GlbRuleFactory == nil {
		GlbRuleFactory = &RuleFactory{}
	}
	return GlbRuleFactory
}

func NewMacRule(macListName string) re.Rule {
	rule := re.NewEmptyRule()
	rule.SetCondition(re.NewCondition(RuleFactoryMAC, RuleFactoryIN_LIST, re.NewFixedArg(macListName)))
	return *rule
}

// NewRiFilter
func (f *RuleFactory) NewRiFilter(ipAddressGroups []string, macAddresses []string, environments []string, models []string) *re.Rule {
	var rule *re.Rule

	if len(ipAddressGroups) > 0 {
		for _, ipAddressGroup := range ipAddressGroups {
			ipRule := re.Rule{}
			ipRule.SetCondition(re.NewCondition(RuleFactoryIP, RuleFactoryIN_LIST, re.NewFixedArg(ipAddressGroup)))
			if rule == nil {
				rule = &ipRule
			} else {
				newRule := re.OrRules(*rule, ipRule)
				rule = &newRule
			}
		}
	}

	if len(macAddresses) > 0 {
		macRule := re.Rule{}
		macRule.SetCondition(re.NewCondition(RuleFactoryMAC, re.StandardOperationIn, re.NewFixedArg(macAddresses)))
		if rule == nil {
			rule = &macRule
		} else {
			newRule := re.AndRules(*rule, macRule)
			rule = &newRule
		}
	}

	if len(environments) > 0 {
		envRule := re.Rule{}
		envRule.SetCondition(re.NewCondition(RuleFactoryENV, re.StandardOperationIn, re.NewFixedArg(environments)))
		if rule == nil {
			rule = &envRule
		} else {
			newRule := re.AndRules(*rule, envRule)
			rule = &newRule
		}
	}

	if len(models) > 0 {
		modelRule := re.Rule{}
		modelRule.SetCondition(re.NewCondition(RuleFactoryMODEL, re.StandardOperationIn, re.NewFixedArg(models)))
		if rule == nil {
			rule = &modelRule
		} else {
			newRule := re.AndRules(*rule, modelRule)
			rule = &newRule
		}
	}

	return rule
}

func (f *RuleFactory) NewTimeFilterTemplate(neverBlockRebootDecoupled bool, neverBlockHttpDownload bool, isLocalTime bool,
	environment string, model string, ipWhiteList string, start string, end string) *re.Rule {
	return createTimeFilter(neverBlockRebootDecoupled, neverBlockHttpDownload, isLocalTime,
		environment, model, ipWhiteList, start, end, true)
}

// NewDownloadLocationFilter
func (f *RuleFactory) NewDownloadLocationFilter(ipList string, downloadProtocol string) *re.Rule {
	rule := re.Rule{}
	rule.SetCondition(re.NewCondition(RuleFactoryIP, RuleFactoryIN_LIST, re.NewFixedArg(ipList)))

	if downloadProtocol != "" {
		rule1 := re.Rule{}
		rule1.SetCondition(re.NewCondition(RuleFactoryFIRMWARE_DOWNLOAD_PROTOCOL, re.StandardOperationIs, re.NewFixedArg(downloadProtocol)))
		rule = re.AndRules(rule, rule1)
	}

	return &rule
}

func (f *RuleFactory) NewGlobalPercentFilterTemplate(percent float64, ipList string) *re.Rule {
	return createGlobalPercentFilter(percent, ipList, true)
}

func createGlobalPercentFilter(percent float64, ipList string, isTemplate bool) *re.Rule {
	excludedRules := []string{ENV_MODEL_RULE, MIN_CHECK_RULE, IV_RULE}
	rule := re.Rule{}
	rule.SetCondition(re.NewCondition(RuleFactoryMATCHED_RULE_TYPE, re.StandardOperationIn, re.NewFixedArg(excludedRules)))
	rule = re.Not(rule)
	rule1 := re.Rule{}
	rule1.SetCondition(re.NewCondition(RuleFactoryMAC, re.StandardOperationPercent, re.NewFixedArg(percent)))
	rule = re.AndRules(rule, rule1)

	// For template, we are allow creation of rule with empty ipList
	if ipList != "" || isTemplate {
		rule2 := re.Rule{}
		rule2.SetCondition(re.NewCondition(RuleFactoryIP, RuleFactoryIN_LIST, re.NewFixedArg(ipList)))
		rule = re.AndRules(rule, rule2)
	}
	return &rule
}

// NewEnvModelRule
func (f *RuleFactory) NewEnvModelRule(env string, model string) re.Rule {
	//envModelRule := re.NewEmptyRule()

	envRule := re.Rule{}
	envRule.SetCondition(re.NewCondition(RuleFactoryENV, re.StandardOperationIs, re.NewFixedArg(env)))

	modelRule := re.Rule{}
	modelRule.SetCondition(re.NewCondition(RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg(model)))

	// envModelRule.AddCompoundPart(envRule)
	// envModelRule.AddCompoundPart(modelRule)

	// return *envModelRule
	envModelRule := re.AndRules(envRule, modelRule)
	return envModelRule
}

// NewModelRule
func (f *RuleFactory) NewModelRule(model string) re.Rule {
	modelRule := re.NewEmptyRule()
	modelRule.SetCondition(re.NewCondition(RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg(model)))
	return *modelRule
}

// NewTimeFilter
func (f *RuleFactory) NewTimeFilter(neverBlockRebootDecoupled bool, neverBlockHttpDownload bool, isLocalTime bool,
	environment string, model string, ipWhiteList string, start string, end string) *re.Rule {
	return createTimeFilter(neverBlockRebootDecoupled, neverBlockHttpDownload, isLocalTime,
		environment, model, ipWhiteList, start, end, false)
}

func createTimeFilter(neverBlockRebootDecoupled bool, neverBlockHttpDownload bool, isLocalTime bool,
	environment string, model string, ipWhiteList string, start string, end string, isTemplate bool) *re.Rule {

	sTime, _ := time.Parse("15:04", start)
	eTime, _ := time.Parse("15:04", end)

	rule := re.Rule{}
	rule.SetCondition(re.NewCondition(RuleFactoryTIME_ZONE, re.StandardOperationIs, RuleFactoryUTC))
	if isLocalTime {
		rule = re.Not(rule)
	}

	timeRule := re.Rule{}
	timeRule.SetCondition(re.NewCondition(RuleFactoryLOCAL_TIME, re.StandardOperationGte, re.NewFixedArg(sTime.Format("15:04:00"))))
	rule = re.AndRules(rule, timeRule)

	timeRule = re.Rule{}
	timeRule.SetCondition(re.NewCondition(RuleFactoryLOCAL_TIME, re.StandardOperationLte, re.NewFixedArg(eTime.Format("15:04:00"))))
	if sTime.Before(eTime) {
		rule = re.AndRules(rule, timeRule)
	} else {
		rule = re.OrRules(rule, timeRule)
	}

	if neverBlockRebootDecoupled {
		r := re.Rule{}
		r.SetCondition(re.NewCondition(RuleFactoryREBOOT_DECOUPLED, re.StandardOperationExists, nil))
		rule = re.AndRules(rule, re.Not(r))
	}

	if neverBlockHttpDownload {
		r := re.Rule{}
		r.SetCondition(re.NewCondition(RuleFactoryFIRMWARE_DOWNLOAD_PROTOCOL, re.StandardOperationIs, RuleFactoryHTTP))
		rule = re.AndRules(rule, re.Not(r))
	}

	if ipWhiteList != "" || isTemplate {
		r := re.Rule{}
		r.SetCondition(re.NewCondition(RuleFactoryIP, RuleFactoryIN_LIST, re.NewFixedArg(ipWhiteList)))
		rule = re.AndRules(rule, re.Not(r))
	}

	if (environment != "" && model != "") || isTemplate {
		envRule := re.Rule{}
		envRule.SetCondition(re.NewCondition(RuleFactoryENV, re.StandardOperationIs, re.NewFixedArg(environment)))
		rule = re.AndRules(rule, re.Not(envRule))

		modelRule := re.Rule{}
		modelRule.SetCondition(re.NewCondition(RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg(model)))
		rule = re.OrRules(rule, re.Not(modelRule))
	}

	return &rule
}

// NewIntermediateVersionRule
func (f *RuleFactory) NewIntermediateVersionRule(env string, model string, version string) re.Rule {
	rule1 := re.Rule{}
	rule1.SetCondition(re.NewCondition(RuleFactoryENV, re.StandardOperationIs, re.NewFixedArg(env)))

	rule2 := re.Rule{}
	rule2.SetCondition(re.NewCondition(RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg(model)))

	rule3 := re.Rule{}
	rule3.SetCondition(re.NewCondition(RuleFactoryVERSION, re.StandardOperationIs, re.NewFixedArg(version)))

	rule1 = re.AndRules(rule1, rule2)
	return re.AndRules(rule1, rule3)
}

func (f *RuleFactory) NewActivationVersionRule(model string, partnerId string) re.Rule {
	rule1 := re.Rule{}
	rule1.SetCondition(re.NewCondition(RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg(model)))

	rule2 := re.Rule{}
	rule2.SetCondition(re.NewCondition(RuleFactoryPARTNER_ID, re.StandardOperationIs, re.NewFixedArg(partnerId)))

	activationRule := re.AndRules(rule1, rule2)
	return activationRule
}

// NewRiFilter
func (f *RuleFactory) NewRiFilterTemplate() *re.Rule {
	var rule *re.Rule

	ipRule := re.Rule{}
	ipRule.SetCondition(re.NewCondition(RuleFactoryIP, RuleFactoryIN_LIST, re.NewFixedArg(EMPTY_NAME)))
	rule = &ipRule

	macRule := re.Rule{}
	macRule.SetCondition(re.NewCondition(RuleFactoryMAC, re.StandardOperationIn, re.NewFixedArg(EMPTY_LIST)))
	newRule := re.AndRules(*rule, macRule)
	rule = &newRule

	envRule := re.Rule{}
	envRule.SetCondition(re.NewCondition(RuleFactoryENV, re.StandardOperationIn, re.NewFixedArg(EMPTY_LIST)))
	newRule = re.AndRules(*rule, envRule)
	rule = &newRule

	modelRule := re.Rule{}
	modelRule.SetCondition(re.NewCondition(RuleFactoryMODEL, re.StandardOperationIn, re.NewFixedArg(EMPTY_LIST)))
	newRule = re.AndRules(*rule, modelRule)
	rule = &newRule

	return rule
}

// NewMinVersionCheckRule
func (f *RuleFactory) NewMinVersionCheckRule(env string, model string, versions []string) re.Rule {
	rule1 := re.Rule{}
	rule1.SetCondition(re.NewCondition(RuleFactoryENV, re.StandardOperationIs, re.NewFixedArg(env)))

	rule2 := re.Rule{}
	rule2.SetCondition(re.NewCondition(RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg(model)))

	rule3 := re.Rule{}
	rule3.SetCondition(re.NewCondition(RuleFactoryVERSION, re.StandardOperationIn, re.NewFixedArg(versions)))

	rule1 = re.AndRules(rule1, rule2)
	return re.AndRules(rule1, rule3)
}

// NEWIpRule ...
func (f *RuleFactory) NewIpRule(listName string, environment string, model string) re.Rule {
	rule1 := re.Rule{}
	rule1.SetCondition(re.NewCondition(RuleFactoryIP, RuleFactoryIN_LIST, re.NewFixedArg(listName)))

	rule2 := re.Rule{}
	rule2.SetCondition(re.NewCondition(RuleFactoryENV, re.StandardOperationIs, re.NewFixedArg(environment)))

	rule3 := re.Rule{}
	rule3.SetCondition(re.NewCondition(RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg(model)))

	rule1 = re.AndRules(rule1, rule2)
	return re.AndRules(rule1, rule3)
}

// NEWIpFilter ...
func (f *RuleFactory) NewIpFilter(listName string) *re.Rule {
	rule := re.Rule{}
	rule.SetCondition(re.NewCondition(RuleFactoryIP, RuleFactoryIN_LIST, re.NewFixedArg(listName)))
	return &rule
}

// NewGlobalPercentFilter
func (f *RuleFactory) NewGlobalPercentFilter(percent float64, ipList string) *re.Rule {
	excludedRules := []string{"ENV_MODEL_RULE", "MIN_CHECK_RULE", "IV_RULE"}
	rule := re.Rule{}
	rule.SetCondition(re.NewCondition(RuleFactoryMATCHED_RULE_TYPE, re.StandardOperationIn, re.NewFixedArg(excludedRules)))
	rule = re.Not(rule)
	rule1 := re.Rule{}
	rule1.SetCondition(re.NewCondition(RuleFactoryMAC, re.StandardOperationPercent, re.NewFixedArg(percent)))
	rule = re.AndRules(rule, rule1)

	if ipList != "" {
		rule2 := re.Rule{}
		rule2.SetCondition(re.NewCondition(RuleFactoryIP, RuleFactoryIN_LIST, re.NewFixedArg(ipList)))
		rule = re.AndRules(rule, rule2)

	}
	return &rule
}

const (
	EMPTY_NAME      = ""
	DEFAULT_PERCENT = 0
	TRUE            = "true"
)

var (
	EMPTY_LIST = []string{}
	EMPTY_SET  = map[string]struct{}{}

	GlbTemplateFactory *TemplateFactory
)

type TemplateFactory struct {
}

// NEWTemplateFactory ...
func NEWTemplateFactory() *TemplateFactory {
	if GlbTemplateFactory == nil {
		GlbTemplateFactory = &TemplateFactory{}
	}
	return GlbTemplateFactory
}
