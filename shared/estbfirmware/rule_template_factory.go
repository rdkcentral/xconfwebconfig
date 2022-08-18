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
	"xconfwebconfig/common"
	re "xconfwebconfig/rulesengine"
)

// based on RuleFactory.java

var (
	RuleFactoryMAC               = re.NewFreeArg("STRING", common.ESTB_MAC)
	RuleFactoryIP                = re.NewFreeArg("STRING", common.IP_ADDRESS)
	RuleFactoryVERSION           = re.NewFreeArg("STRING", common.FIRMWARE_VERSION)
	RuleFactoryENV               = re.NewFreeArg("STRING", common.ENV)
	RuleFactoryMODEL             = re.NewFreeArg("STRING", common.MODEL)
	RuleFactoryFIRMWARE_VERSION  = re.NewFreeArg("STRING", common.FIRMWARE_VERSION)
	RuleFactoryREGEX             = re.NewFreeArg("STRING", common.FIRMWARE_VERSION)
	RuleFactoryMATCHED_RULE_TYPE = re.NewFreeArg("STRING", common.MATCHED_RULE_TYPE)
	RuleFactoryTIME_ZONE         = re.NewFreeArg("STRING", common.TIME_ZONE) // may be "UTC"
	RuleFactoryTIME              = re.NewFreeArg("TIME", common.TIME)
	RuleFactoryLOCAL_TIME        = re.NewFreeArg("STRING", common.TIME)

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

// NewEnvModelRule
func (f *RuleFactory) NewEnvModelRule(env string, model string) re.Rule {
	envModelRule := re.NewEmptyRule()

	envRule := re.Rule{}
	envRule.SetCondition(re.NewCondition(RuleFactoryENV, re.StandardOperationIs, re.NewFixedArg(env)))

	modelRule := re.Rule{}
	modelRule.SetCondition(re.NewCondition(RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg(model)))

	envModelRule.AddCompoundPart(envRule)
	envModelRule.AddCompoundPart(modelRule)

	return *envModelRule
}

// NewModelRule
func (f *RuleFactory) NewModelRule(model string) re.Rule {
	modelRule := re.NewEmptyRule()
	modelRule.SetCondition(re.NewCondition(RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg(model)))
	return *modelRule
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
func (f *RuleFactory) NewGlobalPercentFilter(percent float32, ipList string) *re.Rule {
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
		rule = re.AndRules(rule, rule1)

	}
	return &rule
}

const (
	EMPTY_NAME      = ""
	DEFAULT_PERCENT = 0
	TRUE            = "true"
)

var (
	EMPTY_LIST = []interface{}{}
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
