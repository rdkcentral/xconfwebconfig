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
package settings

import (
	"sort"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"

	log "github.com/sirupsen/logrus"
)

func GetSettingRuleAllAsList(tenantId string) ([]*logupload.SettingRule, error) {
	cm := db.GetCacheManager()
	cacheKey := "SettingRuleList"
	cacheInst := cm.ApplicationCacheGet(tenantId, db.TABLE_SETTING_RULES, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*logupload.SettingRule), nil
	}

	list, err := db.GetCachedSimpleDao().GetAllAsList(tenantId, db.TABLE_SETTING_RULES, 0)
	if err != nil {
		return nil, err
	}

	settingRules := make([]*logupload.SettingRule, 0, len(list))

	for _, v := range list {
		rule := v.(*logupload.SettingRule)
		settingRules = append(settingRules, rule)
	}

	if len(settingRules) > 0 {
		cm.ApplicationCacheSet(tenantId, db.TABLE_SETTING_RULES, cacheKey, settingRules)
	}

	return settingRules, nil
}

func GetSettingRulesBySettingType(tenantId string, settingType string) []*logupload.SettingRule {
	settingTypeEnum := logupload.SettingTypeEnum(settingType)
	var settingRules []*logupload.SettingRule
	list, err := GetSettingRuleAllAsList(tenantId)
	if err == nil {
		for _, rule := range list {
			if profile := GetSettingProfileBySettingRule(tenantId, rule); profile != nil {
				profileSettingType := logupload.SettingTypeEnum(profile.SettingType)
				if profileSettingType == settingTypeEnum {
					settingRules = append(settingRules, rule)
				}
			}
		}
	}
	return settingRules
}

func GetSettingProfileBySettingRule(tenantId string, settingRule *logupload.SettingRule) *logupload.SettingProfiles {
	var settingProfile *logupload.SettingProfiles
	if settingRule != nil && settingRule.BoundSettingID != "" {
		profileData, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_SETTING_PROFILES, settingRule.BoundSettingID)
		if err == nil {
			settingProfile = profileData.(*logupload.SettingProfiles)
		}
	}
	return settingProfile
}

func GetMaxRule(tenantId string, settingsRules []logupload.SettingRule) *logupload.SettingRule {
	if len(settingsRules) > 0 {
		sort.Slice(settingsRules, func(i, j int) bool { return re.CompareRules(settingsRules[i].Rule, settingsRules[j].Rule) > 0 })
		return &settingsRules[0]
	}
	return nil
}

func GetSettingsRuleByTypeForContext(settingType string, contextMap map[string]string) *logupload.SettingRule {
	tenantId := contextMap[common.TENANT_ID]
	settingRules := GetSettingRulesBySettingType(tenantId, settingType)
	var rules []logupload.SettingRule
	for _, rule := range settingRules {
		ruleProcessorFactory := re.NewRuleProcessorFactory()
		// TODO: please add log.Fields to this method
		if contextMap[common.APPLICATION_TYPE] == rule.GetApplicationType() && ruleProcessorFactory.Processor.Evaluate(&rule.Rule, contextMap, log.Fields{}) {
			rules = append(rules, *rule)
		}
	}
	return GetMaxRule(tenantId, rules)
}
