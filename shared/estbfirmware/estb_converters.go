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
	"math"
	"strconv"
	"strings"

	"xconfwebconfig/common"
	re "xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	"xconfwebconfig/shared/firmware"
	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

const (
	MAC_RULE                  = "MAC_RULE"
	IP_RULE                   = "IP_RULE"
	ENV_MODEL_RULE            = "ENV_MODEL_RULE"
	IP_FILTER                 = "IP_FILTER"
	TIME_FILTER               = "TIME_FILTER"
	REBOOT_IMMEDIATELY_FILTER = "REBOOT_IMMEDIATELY_FILTER"
	DOWNLOAD_LOCATION_FILTER  = "DOWNLOAD_LOCATION_FILTER"
	IV_RULE                   = "IV_RULE"
	MIN_CHECK_RULE            = "MIN_CHECK_RULE"
	MIN_CHECK_RI              = "MIN_CHECK_RI"
	GLOBAL_PERCENT            = "GLOBAL_PERCENT"
	ACTIVATION_VERSION        = "ACTIVATION_VERSION"
	HTTP_SUFFIX               = "_http"
	TFTP_SUFFIX               = "_tftp"
)

func ConvertFirmwareRuleToIpFilter(firmwareRule *firmware.FirmwareRule) *IpFilter {
	filter := NewEmptyIpFilter()

	filter.Name = firmwareRule.Name
	filter.Id = firmwareRule.ID
	conds := re.ToConditions(&firmwareRule.Rule)
	for _, cond := range conds {
		if RuleFactoryIP.Equals(cond.GetFreeArg()) {
			filter.IpAddressGroup = GetIpAddressGroup(cond)
		}
	}
	return filter
}

func ConvertIpFilterToFirmwareRule(ipFilter *IpFilter) *firmware.FirmwareRule {
	rule := &firmware.FirmwareRule{}
	iprule := NewRuleFactory().NewIpFilter(ipFilter.IpAddressGroup.Name)
	rule.Rule = *iprule
	rule.Type = IP_FILTER
	rule.ApplicableAction = firmware.NewApplicableActionAndType(firmware.BlockingFilterActionClass, firmware.BLOCKING_FILTER, "")
	rule.Name = ipFilter.Name
	rule.ID = ipFilter.Id
	return rule
}

func GetIpAddressGroup(cond *re.Condition) *shared.IpAddressGroup {
	operation := cond.GetOperation()

	if RuleFactoryIN_LIST == operation {
		listId := cond.GetFixedArg().GetValue().(string)
		list, _ := shared.GetGenericNamedListOneDB(listId)
		if list != nil {
			return shared.ConvertToIpAddressGroup(list)
		} else {
			return makeIpAddressGroup(listId)
		}
	} else if re.StandardOperationIn == operation {
		ipgrpstrs := cond.GetFixedArg().GetValue().([]string)
		// convertto shared.IpAddressGroup use field of this  cond as id, name
		return shared.NewIpAddressGroupWithAddrStrings(cond.GetOperation(), cond.String(), ipgrpstrs)
	} else {
		log.Warn(fmt.Sprintf("Unknown operation for IP freeArg: %v", operation))
		return &shared.IpAddressGroup{}
	}
}

// makeIpAddressGroup ...
func makeIpAddressGroup(id string) *shared.IpAddressGroup {
	group := shared.IpAddressGroup{}
	group.Id = id
	group.Name = id
	return &group
}

// isLegacyIpCondition ...
func IsLegacyIpCondition(condition re.Condition) bool {
	if IsLegacyIpFreeArg(condition.GetFreeArg()) && re.StandardOperationIn == condition.GetOperation() {
		if _, ok := condition.GetFixedArg().GetValue().(shared.IpAddressGroup); ok {
			return true
		}
	}
	return false
}

func IsLegacyIpFreeArg(freeArg *re.FreeArg) bool {
	return re.AuxFreeArgTypeIpAddress == freeArg.GetType() && common.IP_ADDRESS == freeArg.GetName()
}

func IsLegacyMacFreeArg(freeArg re.FreeArg) bool {
	return re.AuxFreeArgTypeMacAddress == freeArg.GetType() && common.ESTB_MAC == freeArg.GetName()
}

func IsLegacyLocalTimeFreeArg(freeArg re.FreeArg) bool {
	return re.AuxFreeArgTypeTime == freeArg.GetType() && common.TIME == freeArg.GetName()
}

// convertFirmwareRuleToIpRuleBean ...
func ConvertFirmwareRuleToIpRuleBean(firmwareRule *firmware.FirmwareRule) *IpRuleBean {
	bean := IpRuleBean{}
	bean.Name = firmwareRule.Name
	bean.Id = firmwareRule.ID

	rules := firmwareRule.Rule.GetCompoundParts()
	for _, r := range rules {
		cond := r.GetCondition()
		if IsLegacyIpFreeArg(cond.GetFreeArg()) || RuleFactoryIP.Equals(cond.GetFreeArg()) {
			bean.IpAddressGroup = GetIpAddressGroup(cond)
		} else if RuleFactoryENV.Equals(cond.GetFreeArg()) {
			bean.EnvironmentId = cond.GetFixedArg().GetValue().(string)
		} else if RuleFactoryMODEL.Equals(cond.GetFreeArg()) {
			bean.ModelId = cond.GetFixedArg().GetValue().(string)
		}
		bean.ModelId = cond.GetFixedArg().GetValue().(string)
	}

	return &bean
}

// convertIpRuleBeanToFirmwareRule ...
func ConvertIpRuleBeanToFirmwareRule(bean *IpRuleBean) *firmware.FirmwareRule {
	ipRule := firmware.FirmwareRule{}
	ipRule.Rule = NewRuleFactory().NewIpRule(bean.IpAddressGroup.Name, bean.EnvironmentId, bean.ModelId)
	ipRule.Type = IP_RULE
	ipRule.ID = bean.Id
	ipRule.Name = bean.Name
	action := &firmware.ApplicableAction{}
	action.ActionType = firmware.RULE
	action.Type = firmware.RuleActionClass
	if bean.FirmwareConfig != nil {
		action.ConfigId = bean.FirmwareConfig.ID
	}
	ipRule.ApplicableAction = action
	return &ipRule
}

// convertIntoRule
func ConvertPercentageBeanToFirmwareRule(bean PercentageBean) *firmware.FirmwareRule {
	envModelRule := re.Rule{}
	if bean.Environment != "" {
		envModelRule = NewRuleFactory().NewEnvModelRule(bean.Environment, bean.Model)
	} else {
		envModelRule = NewRuleFactory().NewModelRule(bean.Model)
	}
	if containsAnyCondition(bean.OptionalConditions) {
		envModelRule = re.AndRules(envModelRule, *bean.OptionalConditions)
	}
	distributions := []firmware.ConfigEntry{}
	for _, distribution := range bean.Distributions {
		distributions = append(distributions, *distribution)
	}
	configEntries := []firmware.ConfigEntry{}
	for _, configEntry := range ConvertIntoPercentRange(distributions) {
		configEntries = append(configEntries, *configEntry)
	}

	action := &firmware.ApplicableAction{
		Type:                  firmware.RuleActionClass,
		ActionType:            firmware.RULE,
		Active:                bean.Active,
		FirmwareCheckRequired: bean.FirmwareCheckRequired,
		RebootImmediately:     bean.RebootImmediately,
		Whitelist:             bean.Whitelist,
		IntermediateVersion:   bean.IntermediateVersion,
		ConfigId:              bean.LastKnownGood,
		ConfigEntries:         configEntries,
		FirmwareVersions:      bean.FirmwareVersions,
		UseAccountPercentage:  bean.UseAccountIdPercentage,
	}
	firmwareRule := firmware.NewFirmwareRule(bean.ID, bean.Name, "ENV_MODEL_RULE", &envModelRule, action, true)
	firmwareRule.ApplicationType = bean.ApplicationType
	return firmwareRule
}

func containsAnyCondition(rule *re.Rule) bool {
	return (rule != nil && rule.Condition != nil) || (rule != nil && rule.CompoundParts != nil && len(rule.CompoundParts) > 0)
}

func ConvertFirmwareRuleToIpRuleBeanAddFirmareConfig(firmwareRule *firmware.FirmwareRule) (bean *IpRuleBean, err error) {
	bean = ConvertFirmwareRuleToIpRuleBean(firmwareRule)
	action := firmwareRule.ApplicableAction
	if action == nil || action.ConfigId == "" {
		err = fmt.Errorf("FirmwareRule [id=%s, name=%s] is corrupted: ApplicableAction is missing", firmwareRule.ID, firmwareRule.Name)
		log.Error(err)
	} else {
		if cfg, e := GetFirmwareConfigOneDB(action.ConfigId); e == nil {
			bean.FirmwareConfig = cfg
		}
	}
	return bean, err
}

func ConvertFirmwareRuleToPercentageBean(firmwareRule *firmware.FirmwareRule) *PercentageBean {
	bean := NewPercentageBean()
	ParseEnvModelRule(bean, firmwareRule)
	if firmwareRule.ApplicableAction != nil {
		parseRuleAction(bean, firmwareRule.ApplicableAction)
	}

	return bean
}

func ParseEnvModelRule(bean *PercentageBean, envModelRule *firmware.FirmwareRule) {
	bean.ID = envModelRule.ID
	bean.Name = envModelRule.Name
	bean.ApplicationType = envModelRule.ApplicationType

	var model, environment string
	var optionalRule *re.Rule

	rules := re.FlattenRule(envModelRule.Rule)
	for _, rule := range rules {
		condition := rule.Condition
		if condition != nil && condition.Operation == re.StandardOperationIs && RuleFactoryMODEL.Equals(condition.FreeArg) {
			model = condition.GetFixedArg().GetValue().(string)
		} else if condition != nil && condition.Operation == re.StandardOperationIs && RuleFactoryENV.Equals(condition.FreeArg) {
			environment = condition.GetFixedArg().GetValue().(string)
		} else {
			optionalRule = getOptionalRule(optionalRule, &rule)
		}
	}

	bean.Model = model
	bean.Environment = environment
	bean.OptionalConditions = optionalRule
}

func getOptionalRule(optionalRule *re.Rule, ruleToAdd *re.Rule) *re.Rule {
	if ruleToAdd == nil || ruleToAdd.Condition == nil {
		return optionalRule
	}

	if optionalRule == nil {
		result := re.Copy(*ruleToAdd)
		if result.IsCompoundPartsEmpty() {
			result.SetRelation("")
		}
		return &result
	}

	var compoundRule re.Rule
	if ruleToAdd.Relation == re.RelationAnd {
		compoundRule = re.And(*ruleToAdd)
	} else {
		compoundRule = re.Or(*ruleToAdd)
	}

	if optionalRule.IsCompound() {
		optionalRule.AddCompoundPart(compoundRule)
	} else {
		result := re.Rule{}
		result.SetCompoundParts(
			[]re.Rule{
				*optionalRule,
				compoundRule,
			},
		)
		return &result
	}

	return optionalRule
}

func parseRuleAction(bean *PercentageBean, action *firmware.ApplicableAction) {
	bean.Active = action.Active
	bean.Whitelist = action.Whitelist
	bean.RebootImmediately = action.RebootImmediately
	bean.FirmwareCheckRequired = action.FirmwareCheckRequired
	if len(action.FirmwareVersions) == 0 {
		bean.FirmwareVersions = make([]string, 0)
	} else {
		bean.FirmwareVersions = action.FirmwareVersions
	}
	bean.IntermediateVersion = action.IntermediateVersion
	bean.Distributions = ConvertIntoPercentRange(action.ConfigEntries)
	bean.LastKnownGood = action.ConfigId
	bean.UseAccountIdPercentage = action.UseAccountPercentage
}

func ConvertIntoPercentRange(configEntries []firmware.ConfigEntry) []*firmware.ConfigEntry {
	var result = make([]*firmware.ConfigEntry, 0)
	var prevPercentEnd float64 = 0

	for _, configEntry := range configEntries {
		inst := firmware.ConfigEntry{
			ConfigId:          configEntry.ConfigId,
			Percentage:        configEntry.Percentage,
			StartPercentRange: configEntry.StartPercentRange,
			EndPercentRange:   configEntry.EndPercentRange,
			IsCanaryDisabled:  configEntry.IsCanaryDisabled,
			IsPaused:          configEntry.IsPaused,
		}

		if inst.StartPercentRange == 0 || inst.EndPercentRange == 0 {
			inst.StartPercentRange = math.Round(prevPercentEnd*1000) / 1000
			inst.EndPercentRange = math.Round((prevPercentEnd+inst.Percentage)*1000) / 1000
		}

		prevPercentEnd += math.Round((inst.EndPercentRange-inst.StartPercentRange)*1000) / 1000

		result = append(result, &inst)
	}

	return result
}

func ReplaceConfigIdWithFirmwareVersion(bean *PercentageBean) *PercentageBean {
	if len(bean.Distributions) > 0 {
		for _, configEntry := range bean.Distributions {
			firmwareVersion := GetFirmwareVersion(configEntry.ConfigId)
			if firmwareVersion != "" {
				configEntry.ConfigId = firmwareVersion
			}
		}
	}

	if bean.LastKnownGood != "" {
		firmwareVersion := GetFirmwareVersion(bean.LastKnownGood)
		if firmwareVersion != "" {
			bean.LastKnownGood = firmwareVersion
		}
	}

	if bean.IntermediateVersion != "" {
		firmwareVersion := GetFirmwareVersion(bean.IntermediateVersion)
		if firmwareVersion != "" {
			bean.IntermediateVersion = firmwareVersion
		}
	}

	return bean
}

func ConvertFirmwareRuleToMacRuleBeanWrapper(firmwareRule *firmware.FirmwareRule) *MacRuleBean {
	macRuleBean := MacRuleBean{}
	macRuleBean.Name = firmwareRule.Name
	macRuleBean.Id = firmwareRule.ID
	macRuleBean.MacList = &[]string{}
	for _, condition := range re.ToConditions(&firmwareRule.Rule) {
		if condition.GetFreeArg().Equals(RuleFactoryMAC) {
			if condition.GetOperation() == RuleFactoryIN_LIST {
				macRuleBean.MacListRef = condition.GetFixedArg().GetValue().(string)
			} else if re.StandardOperationIn == condition.GetOperation() && condition.GetFixedArg().IsCollectionValue() {
				value := condition.GetFixedArg().GetValue().([]string)
				macRuleBean.MacList = &value
			} else if re.StandardOperationIs == condition.GetOperation() && condition.GetFixedArg().IsStringValue() {
				value := condition.GetFixedArg().GetValue().(string)
				if util.IsValidMacAddressForAdminService(value) {
					ar := []string{value}
					macRuleBean.MacList = &ar
				}
			}
		}
	}
	return &macRuleBean
}

func ConvertFirmwareRuleToEnvModelRuleBean(firmwareRule *firmware.FirmwareRule) *EnvModelBean {
	envModelRuleBean := EnvModelBean{}
	envModelRuleBean.Name = firmwareRule.Name
	envModelRuleBean.Id = firmwareRule.ID
	for _, condition := range re.ToConditions(&firmwareRule.Rule) {
		if condition.FreeArg != nil {
			if condition.FreeArg.Equals(RuleFactoryENV) {
				envModelRuleBean.EnvironmentId = condition.GetFixedArg().GetValue().(string)
			} else if condition.FreeArg.Equals(RuleFactoryMODEL) {
				envModelRuleBean.ModelId = condition.GetFixedArg().GetValue().(string)
			}
		}
	}
	action := firmwareRule.ApplicableAction
	if action != nil && action.ConfigId != "" {
		config, err := GetFirmwareConfigOneDB(action.ConfigId)
		if err != nil {
			log.Error(fmt.Sprintf("GetFirmwareConfigOneDB: %v", err))
		}
		envModelRuleBean.FirmwareConfig = config
	}
	return &envModelRuleBean
}

func GetWhitelistName(ipAddressGroup *shared.IpAddressGroup) string {
	if ipAddressGroup != nil {
		return ipAddressGroup.Name
	}
	return ""
}

func getGlobalPercentId(applicationType string) string {
	if "stb" == applicationType {
		return "GLOBAL_PERCENT"
	}
	return fmt.Sprintf("%s_%s", strings.ToUpper(applicationType), "GLOBAL_PERCENT")
}

func NewGlobalPercentFilter(rule *re.Rule) *firmware.FirmwareRule {
	firmwareRule := firmware.NewEmptyFirmwareRule()
	firmwareRule.ID = "GLOBAL_PERCENT"
	firmwareRule.Type = "GLOBAL_PERCENT"
	firmwareRule.Name = "GLOBAL_PERCENT"
	firmwareRule.Rule = *rule
	firmwareRule.ApplicableAction = firmware.NewApplicableActionAndType(firmware.BlockingFilterActionClass, firmware.BLOCKING_FILTER, "")
	return firmwareRule
}

func ConvertIntoGlobalPercentage(percentFilterValue *PercentFilterValue, applicationType string) *firmware.FirmwareRule {
	percentage := float64(percentFilterValue.Percentage)
	//BigDecimal hundredPercentage = new BigDecimal(100);
	var hundredPercentage float64 = 100.0
	whitelistName := GetWhitelistName(percentFilterValue.Whitelist)
	if whitelistName == "" && percentage == 100.0 {
		return nil
	}

	globalPercentFirmwareRule := NewGlobalPercentFilter(NewRuleFactory().NewGlobalPercentFilter(hundredPercentage-percentage, whitelistName))
	globalPercentFirmwareRule.ID = getGlobalPercentId(applicationType)
	globalPercentFirmwareRule.ApplicationType = applicationType
	return globalPercentFirmwareRule
}

func ConvertIntoGlobalPercentageFirmwareRule(firmwareRule *firmware.FirmwareRule) *GlobalPercentage {
	result := NewGlobalPercentage()

	conditions := re.ToConditions(&firmwareRule.Rule)
	for _, condition := range conditions {
		if condition.GetOperation() == re.StandardOperationPercent {
			var value float64
			if condition.FixedArg.IsStringValue() {
				fixedArgValue := *condition.FixedArg.Bean.Value.JLString
				value, _ = strconv.ParseFloat(fixedArgValue, 64)
			} else {
				value = condition.FixedArg.GetValue().(float64)
			}
			result.Percentage = 100 - float32(value)
		} else if condition.GetOperation() == re.StandardOperationInList && RuleFactoryIP.Equals(condition.FreeArg) {
			groupId := condition.FixedArg.GetValue().(string)
			result.Whitelist = groupId
		}
	}

	result.ApplicationType = firmwareRule.ApplicationType

	return result
}

func getMolelfromfirmwarerule(firmwareRule *firmware.FirmwareRule) string {
	var model string
	rules := firmwareRule.Rule.GetCompoundParts()
	if len(rules) > 0 {
		for _, r := range rules {
			cond := r.GetCondition()
			if RuleFactoryMODEL.Equals(cond.GetFreeArg()) {
				model = cond.GetFixedArg().GetValue().(string)
			}
		}
	} else {
		cond := firmwareRule.Rule.GetCondition()
		if RuleFactoryMODEL.Equals(cond.GetFreeArg()) {
			model = cond.GetFixedArg().GetValue().(string)
		}
	}
	return model
}

func getPartneridFromFirmwarerule(firmwareRule *firmware.FirmwareRule) string {
	var partnerid string
	rules := firmwareRule.Rule.GetCompoundParts()
	if len(rules) > 0 {
		for _, r := range rules {
			cond := r.GetCondition()
			if RuleFactoryPARTNER_ID.Equals(cond.GetFreeArg()) {
				partnerid = cond.GetFixedArg().GetValue().(string)
			}
		}
	} else {
		cond := firmwareRule.Rule.GetCondition()
		if RuleFactoryPARTNER_ID.Equals(cond.GetFreeArg()) {
			partnerid = cond.GetFixedArg().GetValue().(string)
		}
	}

	return partnerid
}

func ConvertIntoActivationVersion(fwRule *firmware.FirmwareRule) *firmware.ActivationVersion {
	amv := firmware.ActivationVersion{}
	amv.ID = fwRule.ID
	amv.Description = fwRule.Name
	amv.ApplicationType = fwRule.ApplicationType
	amv.Model = getMolelfromfirmwarerule(fwRule)
	amv.PartnerId = getPartneridFromFirmwarerule(fwRule)
	action := fwRule.ApplicableAction
	amv.FirmwareVersions = action.GetFirmwareVersions()
	amv.RegularExpressions = action.GetFirmwareVersionRegExs()
	return &amv
}

func convertDefineProperties(activationVersion *firmware.ActivationVersion) *firmware.ApplicableAction {
	//action := firmware.DefinePropertiesAction()
	//Properties  := make(map[string]firmware.PropertyValue)
	Properties := make(map[string]string)
	Properties[common.REBOOT_IMMEDIATELY] = "FALSE"
	action := &firmware.ApplicableAction{
		Type:       firmware.DefinePropertiesActionClass,
		ActionType: firmware.DEFINE_PROPERTIES,
		Properties: Properties,
	}
	//action.Properties = Properties
	activationFirmwareVersions := make(map[string][]string)
	activationFirmwareVersions[common.FIRMWARE_VERSIONS] = activationVersion.FirmwareVersions
	activationFirmwareVersions[common.REGULAR_EXPRESSIONS] = activationVersion.RegularExpressions
	action.ActivationFirmwareVersions = activationFirmwareVersions
	return action
}

// This goes in Rule_factory _template.go
func newActivationRule(partnerId string, model string) re.Rule {

	rule1 := re.Rule{}
	rule1.SetCondition(re.NewCondition(RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg(model)))
	if partnerId != "" {
		rule2 := re.Rule{}
		rule2.SetCondition(re.NewCondition(RuleFactoryPARTNER_ID, re.StandardOperationIs, re.NewFixedArg(partnerId)))

		rule1 = re.AndRules(rule1, rule2)
	}
	return rule1
}

func ConvertIntoRule(activationVersion *firmware.ActivationVersion) *firmware.FirmwareRule {
	firmwareRule := firmware.FirmwareRule{}

	firmwareRule.ID = activationVersion.ID
	firmwareRule.Name = activationVersion.Description
	firmwareRule.ApplicationType = activationVersion.ApplicationType
	firmwareRule.ApplicableAction = convertDefineProperties(activationVersion)
	firmwareRule.Active = true
	firmwareRule.Type = ACTIVATION_VERSION
	firmwareRule.Rule = newActivationRule(activationVersion.PartnerId, activationVersion.Model)
	return &firmwareRule

}
