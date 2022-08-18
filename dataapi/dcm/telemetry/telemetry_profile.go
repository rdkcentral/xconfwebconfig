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
package telemetry

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	common "xconfwebconfig/common"
	re "xconfwebconfig/rulesengine"
	"xconfwebconfig/shared/logupload"

	scheduler "github.com/carlescere/scheduler"
	log "github.com/sirupsen/logrus"
)

type TelemetryProfileService struct {
	//RuleProcessorFactory		ev.RuleProcessorFactory
	CacheUpdateWindowSize int64
}

var NewRuleProcessorFactoryFunc = re.NewRuleProcessorFactory

func NewTelemetryProfileService() *TelemetryProfileService {
	tps := TelemetryProfileService{}
	//tps.RuleProcessorFactory = NewRuleProcessorFactoryFunc
	return &tps
}

func (t *TelemetryProfileService) ConvertToDescriptor(rule logupload.TelemetryRule) *logupload.PermanentTelemetryRuleDescriptor {
	ruleDescriptor := logupload.NewPermanentTelemetryRuleDescriptor()
	ruleDescriptor.RuleId = rule.ID
	ruleDescriptor.RuleName = rule.Name
	return ruleDescriptor
}

func (t *TelemetryProfileService) ConvertToProfileDescriptor(profile logupload.TelemetryProfile) *logupload.TelemetryProfileDescriptor {
	profileDescriptor := logupload.NewTelemetryProfileDescriptor()
	profileDescriptor.ID = profile.ID
	profileDescriptor.Name = profile.Name
	return profileDescriptor
}

func (t *TelemetryProfileService) ExpireTemporaryTelemetryRules() {
	job := func() {
		logupload.DeleteExpiredTelemetryProfile(common.CacheUpdateWindowSize)
	}
	windowSize := common.CacheUpdateWindowSize / 60000
	scheduler.Every(int(windowSize)).Minutes().Run(job)
}

func (t *TelemetryProfileService) CreateRuleForAttribute(contextAttribute string, expectedValue string) *logupload.TimestampedRule {
	freeArg := re.NewFreeArg("STRING", contextAttribute)
	fixedArg := re.NewFixedArg(expectedValue)
	condition := re.NewCondition(freeArg, "IS", fixedArg)
	timestampedRule := logupload.NewTimestampedRule()
	timestampedRule.Rule.Condition = condition
	return timestampedRule
}

var SetTelemetryProfileFunc = logupload.SetTelemetryProfile

func (t *TelemetryProfileService) CreateTelemetryProfile(contextAttribute string, expectedValue string, telemetry logupload.TelemetryProfile) *logupload.TimestampedRule {
	telemetryRule := t.CreateRuleForAttribute(contextAttribute, expectedValue)
	telemetryRuleBytes, _ := json.Marshal(telemetryRule)
	SetTelemetryProfileFunc(string(telemetryRuleBytes), telemetry)
	return telemetryRule
}

var GetTelemetryProfileMapFunc = logupload.GetTelemetryProfileMap

var DeleteTelemetryProfileFunc = logupload.DeleteTelemetryProfile

func (t *TelemetryProfileService) DropTelemetryFor(contextAttribute string, expectedValue string) *[]logupload.TelemetryProfile {
	context := make(map[string]string)
	context[contextAttribute] = expectedValue
	//telemetryProfileMap type of *map[string]TelemetryProfile
	telemetryProfileMap := GetTelemetryProfileMapFunc()
	if telemetryProfileMap == nil {
		log.Warn(fmt.Sprintf("no TimestampedRule found"))
		return nil
	}
	rules := []re.Rule{}
	telemetryListAll := []logupload.TelemetryProfile{}
	for k, v := range *telemetryProfileMap {
		bytes := []byte(k)
		var timestampedRule logupload.TimestampedRule
		json.Unmarshal(bytes, &timestampedRule)
		rules = append(rules, timestampedRule.Rule)
		telemetryListAll = append(telemetryListAll, v)
	}
	//matchedRules := t.RuleProcessorFactory.Processor.Filter(rules, context)
	ruleProcessorFactory := NewRuleProcessorFactoryFunc()
	matchedRules := ruleProcessorFactory.Processor.Filter(rules, context)
	telemetryList := make([]logupload.TelemetryProfile, 0)
	for _, matchedRule := range matchedRules {
		for j, rule := range rules {
			if matchedRule.Equals(&rule) {
				telemetryProfile := telemetryListAll[j]
				telemetryList = append(telemetryList, telemetryProfile)
				log.Debug("removing temporary rule: {}", rule)
				DeleteTelemetryProfileFunc(rule.String())
			}
		}
	}
	return &telemetryList
}

var GetOneTelemetryProfileFunc = logupload.GetOneTelemetryProfile
var GetTimestampedRulesFunc = logupload.GetTimestampedRules

func (t *TelemetryProfileService) GetTemporaryProfileForContext(context map[string]string) *logupload.TelemetryProfile {
	//tRules type of []logupload.TimestampedRule{}
	tRules := GetTimestampedRulesFunc()
	sort.Slice(tRules, func(i, j int) bool { return tRules[i].Timestamp < tRules[j].Timestamp })
	ruleProcessorFactory := NewRuleProcessorFactoryFunc()
	processor := ruleProcessorFactory.Processor
	//Filter
	matched := []logupload.TimestampedRule{}
	for _, tRule := range tRules {
		// TODO: please add log.Fields to this method
		if processor.Evaluate(&tRule.Rule, context, log.Fields{}) {
			matched = append(matched, tRule)
		}
	}
	if len(matched) < 1 {
		return nil
	}
	var telemetry *logupload.TelemetryProfile
	for _, tRule := range matched {
		telemetry = GetOneTelemetryProfileFunc(tRule.ToString())
		if telemetry == nil {
			continue
		}
		if (telemetry.Expires + common.CacheUpdateWindowSize) > time.Now().UTC().Unix()*1000 {
			break
		}
	}
	for _, tRule := range matched {
		DeleteTelemetryProfileFunc(tRule.ToString())
	}
	return telemetry
}

func (t *TelemetryProfileService) GetTelemetryRuleForContext(context map[string]string) *logupload.TelemetryRule {
	//all type of []*TelemetryRule
	all := logupload.GetTelemetryRuleList()
	rules := t.ProcessEntityRules(all, context)
	return t.GetMaxRule(rules)
}

func (t *TelemetryProfileService) ProcessEntityRules(telemetryRuleList []*logupload.TelemetryRule, context map[string]string) []*logupload.TelemetryRule {
	var newTelemetryRuleList []*logupload.TelemetryRule
	ruleProcessorFactory := NewRuleProcessorFactoryFunc()
	for _, rule := range telemetryRuleList {

		// TODO: please add log.Fields to this method
		if context["applicationType"] == rule.GetApplicationType() && ruleProcessorFactory.Processor.Evaluate(&rule.Rule, context, log.Fields{}) {
			newTelemetryRuleList = append(newTelemetryRuleList, rule)
		}
	}
	return newTelemetryRuleList
}

func (t *TelemetryProfileService) GetMaxRule(tRules []*logupload.TelemetryRule) *logupload.TelemetryRule {
	if len(tRules) < 1 {
		return nil
	}
	sort.Slice(tRules, func(i, j int) bool { return re.CompareRules(tRules[i].Rule, tRules[j].Rule) > 0 })
	return tRules[0]
}

func (t *TelemetryProfileService) GetPermanentProfileByTelemetryRule(telemetryRule *logupload.TelemetryRule) *logupload.PermanentTelemetryProfile {
	if telemetryRule != nil && len(telemetryRule.BoundTelemetryID) > 0 {
		telemetry := logupload.GetOnePermanentTelemetryProfile(telemetryRule.BoundTelemetryID)
		return telemetry
	}
	return nil
}

func (t *TelemetryProfileService) GetPermanentProfileForContext(context map[string]string) *logupload.PermanentTelemetryProfile {
	telemetryRule := t.GetTelemetryRuleForContext(context)
	return t.GetPermanentProfileByTelemetryRule(telemetryRule)
}

func (t *TelemetryProfileService) GetTelemetryForContext(context map[string]string) *logupload.TelemetryProfile {
	telemetryProfile := t.GetTemporaryProfileForContext(context)
	return telemetryProfile
}

func (t *TelemetryProfileService) GetAvailableDescriptors(applicationType string) []*logupload.PermanentTelemetryRuleDescriptor {
	var descriptors []*logupload.PermanentTelemetryRuleDescriptor
	//all type of []*TelemetryRule
	all := logupload.GetTelemetryRuleList()
	for _, rule := range all {
		if rule.ApplicationType == applicationType {
			telemetryRuleDescriptor := logupload.NewPermanentTelemetryRuleDescriptor()
			telemetryRuleDescriptor.RuleId = rule.ID
			telemetryRuleDescriptor.RuleName = rule.Name
			descriptors = append(descriptors, telemetryRuleDescriptor)
		}
	}
	return descriptors
}

func (t *TelemetryProfileService) GetAvailableProfileDescriptors(applicationType string) []*logupload.TelemetryProfileDescriptor {
	var descriptors []*logupload.TelemetryProfileDescriptor
	//all type of []*PermanentTelemetryProfile
	all := logupload.GetPermanentTelemetryProfileList()
	for _, profile := range all {
		if profile.ApplicationType == applicationType {
			telemetryProfileDescriptor := logupload.NewTelemetryProfileDescriptor()
			telemetryProfileDescriptor.ID = profile.ID
			telemetryProfileDescriptor.Name = profile.Name
			descriptors = append(descriptors, telemetryProfileDescriptor)
		}
	}
	return descriptors
}

func (t *TelemetryProfileService) ProcessTelemetryTwoRules(context map[string]string) []*logupload.TelemetryTwoRule {
	//all type of []*TelemetryTwoRule
	all := logupload.GetTelemetryTwoRuleList()
	if all == nil {
		return nil
	}
	ruleProcessorFactory := NewRuleProcessorFactoryFunc()
	processor := ruleProcessorFactory.Processor
	matched := []*logupload.TelemetryTwoRule{}
	for _, tRule := range all {
		// TODO: please add log.Fields to this method
		if processor.Evaluate(&tRule.Rule, context, log.Fields{}) {
			matched = append(matched, tRule)
		}
	}
	return matched
}

var GetOneTelemetryTwoProfileFunc = logupload.GetOneTelemetryTwoProfile

func (t *TelemetryProfileService) GetTelemetryTwoProfileByTelemetryRules(telemetryTwoRules []*logupload.TelemetryTwoRule) []*logupload.TelemetryTwoProfile {
	var telemetryTwoProfiles []*logupload.TelemetryTwoProfile
	telemetryTwoProfiles = make([]*logupload.TelemetryTwoProfile, 0)
	for _, telemetryTwoRule := range telemetryTwoRules {
		if telemetryTwoRule != nil && len(telemetryTwoRule.BoundTelemetryIDs) > 0 {
			for _, boundTelemetryId := range telemetryTwoRule.BoundTelemetryIDs {
				if len(boundTelemetryId) < 1 {
					continue
				}
				//telemetryTwoProfile type of *TelemetryTwoProfile
				telemetryTwoProfile := GetOneTelemetryTwoProfileFunc(boundTelemetryId)
				if telemetryTwoProfile != nil {
					telemetryTwoProfiles = append(telemetryTwoProfiles, telemetryTwoProfile)
				}
			}
		}
	}
	nameMap := make(map[string]bool)
	uniqueTelemetryTwoProfiles := make([]*logupload.TelemetryTwoProfile, 0)
	for _, telemetryTwoProfile := range telemetryTwoProfiles {
		if _, ok := nameMap[telemetryTwoProfile.Name]; ok {
			continue
		} else {
			nameMap[telemetryTwoProfile.Name] = true
			uniqueTelemetryTwoProfiles = append(uniqueTelemetryTwoProfiles, telemetryTwoProfile)
		}
	}
	return uniqueTelemetryTwoProfiles
}
