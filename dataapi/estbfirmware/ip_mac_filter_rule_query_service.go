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
	"fmt"
	"strings"

	"github.com/google/uuid"

	"xconfwebconfig/common"
	"xconfwebconfig/db"
	re "xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	sharedef "xconfwebconfig/shared/estbfirmware"
	sharedfw "xconfwebconfig/shared/firmware"

	log "github.com/sirupsen/logrus"
)

type IpRuleService struct {
	//FirmwareRulePredicates  *FirmwareRulePredicates
}

// GetByApplicationTyp ...
func (i *IpRuleService) GetByApplicationType(applicationType string) []*sharedef.IpRuleBean {
	insts, err := sharedfw.GetFirmwareRuleAllAsListDB()
	if err != nil {
		log.Error(fmt.Sprintf("GetByApplicationType: %v", err))
		return []*sharedef.IpRuleBean{}
	}
	result := []*sharedef.IpRuleBean{}
	resultNameSet := map[string]struct{}{} // as Set
	for _, frule := range insts {
		if frule.ApplicationType != applicationType && frule.ApplicationType != shared.ALL && applicationType != "" {
			continue
		}
		if frule.Type != sharedfw.IP_RULE {
			continue
		}
		ipfilter := i.ConvertToIpRuleOrReturnNull(frule)
		resultNameSet[ipfilter.Name] = struct{}{}
		result = append(result, ipfilter)
	}
	return result
}

// ConvertToIpRuleOrReturnNull ...
func (i *IpRuleService) ConvertToIpRuleOrReturnNull(firmwareRule *sharedfw.FirmwareRule) *sharedef.IpRuleBean {
	bean := sharedef.ConvertFirmwareRuleToIpRuleBeanAddFirmareConfig(firmwareRule)
	if bean == nil {
		log.Error("Could not convert: ")
		//return &sharedef.IpRuleBean{} // or return nil
		return nil
	}
	return bean
}

// Save ...
func (i *IpRuleService) Save(bean *sharedef.IpRuleBean, applicationType string) {
	if len(bean.Id) == 0 {
		bean.Id = uuid.New().String()
	}

	ipRule := sharedef.ConvertIpRuleBeanToFirmwareRule(bean)
	if len(applicationType) != 0 {
		ipRule.ApplicationType = applicationType
	}

	db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, ipRule.ID, ipRule)
}

// Delete ...
func (i *IpRuleService) Delete(id string) {
	db.GetCachedSimpleDao().DeleteOne(db.TABLE_FIRMWARE_RULE, id)
}

func (i *IpRuleService) getOne(id string) *sharedef.IpRuleBean {
	frule, err := sharedfw.GetFirmwareRuleOneDB(id)
	if err != nil {
		return nil
	}
	return sharedef.ConvertFirmwareRuleToIpRuleBeanAddFirmareConfig(frule)
}

func NullifyUnwantedFields(config *sharedef.FirmwareConfig) *sharedef.FirmwareConfig {
	if config != nil {
		config.Updated = 0
		config.FirmwareDownloadProtocol = ""
		config.RebootImmediately = false
	}
	return config
}

type IpFilterService struct {
}

// private FirmwareRulePredicates firmwareRulePredicates;

func (i *IpFilterService) getOneIpFilterFromDB(id string) *sharedef.IpFilter {
	frule, err := sharedfw.GetFirmwareRuleOneDB(id)
	if err == nil {
		return sharedef.ConvertFirmwareRuleToIpFilter(frule)
	}
	return nil
}

func (i *IpFilterService) getIpFilterByName(name string, applicationType string) *sharedef.IpFilter {
	for _, ipFilter := range i.getByApplicationType(applicationType) {
		if strings.ToUpper(ipFilter.Name) == strings.ToUpper(name) {
			return ipFilter
		}
	}
	return nil
}

func (i *IpFilterService) getByApplicationType(applicationType string) []*sharedef.IpFilter {
	insts, err := sharedfw.GetFirmwareRuleAllAsListDB()
	if err != nil {
		log.Error(fmt.Sprintf("getByApplicationType: %v", err))
		return []*sharedef.IpFilter{}
	}

	result := []*sharedef.IpFilter{}
	resultNameSet := map[string]struct{}{} // as Set
	for _, frule := range insts {
		if string(frule.ApplicableAction.ActionType) != applicationType {
			continue
		}
		if frule.Type != sharedfw.IP_RULE {
			continue
		}

		ipfilter := sharedef.ConvertFirmwareRuleToIpFilter(frule)
		// Avoid Dup based on Name
		if ipfilter == nil {
			continue
		}
		if _, ok := resultNameSet[ipfilter.Name]; ok {
			continue
		}
		resultNameSet[ipfilter.Name] = struct{}{}
		result = append(result, ipfilter)
	}
	return result
}

func (i *IpFilterService) save(filter *sharedef.IpFilter, applicationType string) {
	if len(filter.Id) == 0 {
		filter.Id = uuid.New().String()
	}

	rule := sharedef.ConvertIpFilterToFirmwareRule(filter)
	if rule == nil {
		return
	}
	if len(applicationType) != 0 {
		rule.ApplicationType = applicationType
	}
	db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, rule.ID, rule)
}

func (i *IpFilterService) delete(id string) {
	db.GetCachedSimpleDao().DeleteOne(db.TABLE_FIRMWARE_RULE, id)
}

type PercentFilterService struct {
}

func NewPercentFilterService() *PercentFilterService {
	return &PercentFilterService{}
}

func (p *PercentFilterService) Save(filter *sharedef.PercentFilterValue, applicationType string) {
	globalPercentage := sharedef.ConvertIntoGlobalPercentage(filter, applicationType)
	if globalPercentage != nil {
		globalPercentage.ApplicationType = applicationType
		sharedfw.CreateFirmwareRuleOneDB(globalPercentage)
	}

	rules, err := sharedfw.GetEnvModelFirmwareRules(applicationType)
	if err != nil {
		log.Error(fmt.Sprintf("PercentFilterService.Save : %v %v", rules, err))
		return
	}

	//todo migration rules
}

func getEnvModelPercentage(filter sharedef.PercentFilterValue, name string) *sharedef.EnvModelPercentage {
	if filter.EnvModelPercentages != nil {
		envModelPercentage, ok := filter.EnvModelPercentages[name]
		if ok {
			return &envModelPercentage
		}
	}
	return nil
}

type MacRuleService struct {
}

func (m *MacRuleService) GetRulesWithMacCondition(applicationType string) []*sharedef.MacRuleBean {
	insts, err := sharedfw.GetFirmwareRuleAllAsListDB()
	if err != nil {
		log.Error(fmt.Sprintf("GetRulesWithMacCondition: %v", err))
		return []*sharedef.MacRuleBean{}
	}
	macRuleBeanIdSet := make(map[string]bool)
	macRuleBeanNameSet := make(map[string]bool)
	result := []*sharedef.MacRuleBean{}
	for _, frule := range insts {
		if frule.ApplicationType != applicationType && frule.ApplicationType != shared.ALL && applicationType != "" {
			continue
		}
		if frule.ApplicableAction == nil || frule.ApplicableAction.ActionType != sharedfw.RULE {
			continue
		}
		if !re.IsExistConditionByFreeArgName(frule.Rule, common.ESTB_MAC) {
			continue
		}
		macRuleBean := convertFirmwareRuleToMacRuleBean(frule)
		_, idExists := macRuleBeanIdSet[macRuleBean.Id]
		if !idExists {
			macRuleBeanIdSet[macRuleBean.Id] = true
		}
		_, nameExists := macRuleBeanNameSet[strings.ToLower(macRuleBean.Name)]
		if !nameExists {
			macRuleBeanNameSet[strings.ToLower(macRuleBean.Name)] = true
		}
		if !idExists && !nameExists {
			result = append(result, macRuleBean)
		}
	}
	return result
}

func convertFirmwareRuleToMacRuleBean(firmwareRule *sharedfw.FirmwareRule) *sharedef.MacRuleBean {
	macRuleBean := sharedef.ConvertFirmwareRuleToMacRuleBeanWrapper(firmwareRule)
	action := firmwareRule.ApplicableAction
	if action != nil && action.ConfigId != "" {
		config, err := sharedef.GetFirmwareConfigOneDB(action.ConfigId)
		if err != nil {
			log.Error(fmt.Sprintf("GetFirmwareConfigOneDB: %v", err))
		}
		macRuleBean.FirmwareConfig = config
		if config != nil {
			macRuleBean.TargetedModelIds = &config.SupportedModelIds
		} else {
			macRuleBean.TargetedModelIds = &[]string{}
		}
	}
	return macRuleBean
}

type EnvModelRuleService struct {
}

func (em *EnvModelRuleService) GetByApplicationType(applicationType string) []*sharedef.EnvModelBean {
	insts, err := sharedfw.GetFirmwareRuleAllAsListDB()
	if err != nil {
		log.Error(fmt.Sprintf("GetByApplicationType: %v", err))
		return []*sharedef.EnvModelBean{}
	}
	macRuleBeanIdSet := make(map[string]bool)
	macRuleBeanNameSet := make(map[string]bool)
	result := []*sharedef.EnvModelBean{}
	for _, frule := range insts {
		if frule.Type != sharedfw.ENV_MODEL_RULE {
			continue
		}
		if frule.ApplicationType != applicationType && frule.ApplicationType != shared.ALL && applicationType != "" {
			continue
		}
		emRuleBean := sharedef.ConvertFirmwareRuleToEnvModelRuleBean(frule)
		_, idExists := macRuleBeanIdSet[emRuleBean.Id]
		if !idExists {
			macRuleBeanIdSet[emRuleBean.Id] = true
		}
		_, nameExists := macRuleBeanNameSet[strings.ToLower(emRuleBean.Name)]
		if !nameExists {
			macRuleBeanNameSet[strings.ToLower(emRuleBean.Name)] = true
		}
		if !idExists && !nameExists {
			result = append(result, emRuleBean)
		}
	}
	return result
}
