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
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"xconfwebconfig/common"
	"xconfwebconfig/db"
	re "xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	coreef "xconfwebconfig/shared/estbfirmware"
	"xconfwebconfig/shared/firmware"
	corefw "xconfwebconfig/shared/firmware"

	log "github.com/sirupsen/logrus"
)

type RunningVersionInfo struct {
	HasActivationMinFW bool `json:"hasActivationMinFW"`
	HasMinimumFW       bool `json:"hasMinimumFW"`
}

// EstbFirmwareRuleBase ...
type EstbFirmwareRuleBase struct {
	ruleProcessorFactory *re.RuleProcessorFactory
	driAlwaysReply       bool
	driStateIdentifiers  string
}

// NewEstbFirmwareRuleBaseDefault ...
func NewEstbFirmwareRuleBaseDefault() *EstbFirmwareRuleBase {
	return NewEstbFirmwareRuleBase(true, "P-DRI,B-DRI")
}

// NewEstbFirmwareRuleBase ...
func NewEstbFirmwareRuleBase(driAlwaysReply bool, driStateIdentifiers string) *EstbFirmwareRuleBase {
	return &EstbFirmwareRuleBase{
		ruleProcessorFactory: re.NewRuleProcessorFactory(),
		driAlwaysReply:       driAlwaysReply,
		driStateIdentifiers:  driStateIdentifiers,
	}
}

// Eval ... the main entry poinit to the module
func (e *EstbFirmwareRuleBase) Eval(ctx map[string]string, convertedContext *coreef.ConvertedContext, applicationType string, fields log.Fields) (*EvaluationResult, error) {
	start := time.Now()
	_ = start
	// log.WithFields(fields).Debugf("EstbFirmwareRuleBase.Eval Start ... : context %v and applicationType %s", ctx, applicationType)
	result := NewEvaluationResult()
	// rulereflst, err := corefw.GetFirmwareSortedRuleAllAsListDB()
	// var rules map[string][]*corefw.FirmwareRule
	// if err == nil {
	// 	rules = e.FilterByAppType(rulereflst, applicationType)
	// } else {
	// 	log.Warn("No rules")
	// }

	funcStartTime := time.Now()
	_ = funcStartTime
	rules, err := corefw.GetFirmwareRuleAllAsListByApplicationType(applicationType)
	if err != nil {
		log.Warn("No rules")
	}
	// log.WithFields(fields).Debugf("EstbFirmwareRuleBase.Eval ... corefw.GetFirmwareRuleAllAsListByApplicationType: finish in %v", time.Since(funcStartTime))

	bypassFilters := convertedContext.GetBypassFiltersConverted()

	funcStartTime = time.Now()
	matchedRule := e.FindMatchedRule(rules, corefw.RULE_TEMPLATE, convertedContext.GetProperties(), bypassFilters, fields)
	if matchedRule == nil {
		fields["context"] = ctx
		log.WithFields(fields).Info("EstbFirmwareRuleBase no rules matched")
		result.Description = "No rules matched"
		// log.WithFields(fields).Debugf("EstbFirmwareRuleBase.Eval ... End %s: context %v and applicationType %s, finish in %v", result.Description, ctx, applicationType, time.Since(start))
		return result, nil
	}
	// log.WithFields(fields).Debugf("EstbFirmwareRuleBase.Eval ... e.FindMatchedRule End: finish in %v", time.Since(funcStartTime))

	result.MatchedRule = matchedRule
	var firmwareConfig *coreef.FirmwareConfigFacade = nil

	funcStartTime = time.Now()
	boundConfigId := e.GetBoundConfigId(ctx, convertedContext, matchedRule, result.AppliedVersionInfo)
	// log.WithFields(fields).Debugf("EstbFirmwareRuleBase.Eval ... e.GetBoundConfigId End: finish in %v", time.Since(funcStartTime))

	if boundConfigId != "" && len(boundConfigId) != 0 { // check for no-op rules
		config, err := coreef.GetFirmwareConfigOneDB(boundConfigId)
		if err != nil {
			log.WithFields(fields).Warn(fmt.Sprintf("EstbFirmwareRuleBase no config found by %v: %v boundConfigId: %v, it was deleted", matchedRule.Type, matchedRule.Name, boundConfigId))
			result.Description = fmt.Sprintf("no config found by id: %s", boundConfigId)
			// log.WithFields(fields).Debugf("EstbFirmwareRuleBase.Eval ... End %s: context %v and applicationType %s, finish in %v", result.Description, ctx, applicationType, time.Since(start))
			return result, nil
		} else if !strings.EqualFold(config.ApplicationType, matchedRule.ApplicationType) {
			log.WithFields(fields).Error(fmt.Sprintf("EstbFirmwareRuleBase ApplicationTypeMatchingException: Application types of FirmwareConfig %s and FirmwareRule %v do not match", config.Description, matchedRule))
			result.Description = fmt.Sprintf("no config found by id: %s", boundConfigId)
			// log.WithFields(fields).Debugf("EstbFirmwareRuleBase.Eval ... End %s: context %v and applicationType %s, finish in %v", result.Description, ctx, applicationType, time.Since(start))
			return result, nil
		} else {
			firmwareConfig = coreef.NewFirmwareConfigFacade(config)
			result.AppliedVersionInfo[FIRMWARE_SOURCE] = matchedRule.Type
		}
	} else if !matchedRule.IsNoop() {
		log.WithFields(fields).Info(fmt.Sprintf("EstbFirmwareRuleBase output is blocked by distribution percent in  %v", matchedRule.ApplicableAction))
		result.Blocked = true
		result.AddAppliedFilters(matchedRule.ApplicableAction)
		result.Description = "output is blocked by distribution percent in rule action"
		// log.WithFields(fields).Debugf("EstbFirmwareRuleBase.Eval ... End %s: context %v and applicationType %s, finish in %v", result.Description, ctx, applicationType, time.Since(start))
		return result, nil
	} else {
		log.WithFields(fields).Info(fmt.Sprintf("EstbFirmwareRuleBase rule %s : %s is noop: %s ", matchedRule.Type, matchedRule.Name, matchedRule.ID))
		result.Description = fmt.Sprintf("rule is noop: %s " + matchedRule.ID)
		// log.WithFields(fields).Debugf("EstbFirmwareRuleBase.Eval ... End %s: context %v and applicationType %s, finish in %v", result.Description, ctx, applicationType, time.Since(start))
		return result, nil
	}

	result.FirmwareConfig = firmwareConfig
	funcStartTime = time.Now()
	blocked := e.DoFilters(ctx, convertedContext, applicationType, rules, result, fields)
	// log.WithFields(fields).Debugf("EstbFirmwareRuleBase.Eval ... e.DoFilters End: finish in %v", time.Since(funcStartTime))

	if e.driAlwaysReply {
		funcStartTime = time.Now()
		blocked = e.checkForDRIState(ctx, firmwareConfig, blocked)
		// log.WithFields(fields).Debugf("EstbFirmwareRuleBase.Eval ... e.checkForDRIState End: finish in %v", time.Since(funcStartTime))
	}

	result.Blocked = blocked
	if blocked {
		result.Description = "output is blocked by filter"
	}
	// log.WithFields(fields).Debugf("EstbFirmwareRuleBase.Eval ... End Succesful : context %v and applicationType %s, finish in %v", ctx, applicationType, time.Since(start))
	return result, nil
}

func (e *EstbFirmwareRuleBase) FindMatchedRule(rules map[string][]*corefw.FirmwareRule, templateType corefw.ApplicableActionType,
	contextMap map[string]string, bypassFilters map[string]struct{}, fields log.Fields) *corefw.FirmwareRule {
	start := time.Now()
	log.WithFields(fields).Debugf("FindMatchedRule starts for templateType %s", templateType)
	matchedRules := e.FindMatchedRules(rules, templateType, contextMap, bypassFilters, true, false, fields)
	if len(matchedRules) == 0 {
		log.WithFields(fields).Debugf("===FindMatchedRule no match finished in %v", time.Since(start))
		return nil
	}
	log.WithFields(fields).Debugf("===FindMatchedRule for templateType %s finished in %v", templateType, time.Since(start))
	return matchedRules[0]
}

// looking rule template/rule --> fiind matched rule with the context
func (e *EstbFirmwareRuleBase) FindMatchedRules(rules map[string][]*corefw.FirmwareRule, templateType corefw.ApplicableActionType,
	contextMap map[string]string, bypassFilters map[string]struct{}, isSingle bool, reverse bool, fields log.Fields) []*corefw.FirmwareRule {
	start := time.Now()
	log.WithFields(fields).Debugf("FindMatchedRules starts for templateType %s", templateType)

	results := make([]*corefw.FirmwareRule, 0, 100)
	templates := e.GetSortedTemplate(templateType, reverse, fields)
	log.WithFields(fields).Debugf("GetSortedTemplate for templateType %s ends at %v", templateType, time.Since(start))
	for _, template := range templates {
		ruleType := template.ID
		// tbeforebypassfiltr := time.Now()
		//this template has been added as filter
		if len(template.ByPassFilters) > 0 {
			for _, filter := range template.ByPassFilters {
				bypassFilters[filter] = struct{}{}
			}
		}

		// tbeforeruletypefilter := time.Now()
		if _, ok := bypassFilters[ruleType]; ok {
			continue
		}
		// log.Debugf("check of by pass finished in %v", time.Since(loopstart))

		firmwareRules := corefw.GetRulesByRuleTypes(rules, ruleType)
		if len(firmwareRules) == 0 {
			continue
		}

		if firmware.ENV_MODEL_RULE == ruleType || firmware.ACTIVATION_VERSION == ruleType {
			// sortedRules = sortByConditionsSize(firmwareRules)
			sortByConditionsSize(firmwareRules)
		}

		// log.Debugf("length of ruletype %s: %d ", ruleType, len(firmwareRules))
		log.WithFields(fields).Debugf("templates loop.... templateType %s for ruleType %s time passed %v", templateType, ruleType, time.Since(start))
		fmwareloopstart := time.Now()
		for _, firmwareRule := range firmwareRules {
			frEvalstart := time.Now()

			// evaluate rule only active
			if firmwareRule.Active {
				if firmwareRule != nil {
					fields["firmware_id"] = firmwareRule.ID
					fields["firmware_name"] = firmwareRule.Name
				}
				t0 := time.Now()
				isEvaluate := e.ruleProcessorFactory.RuleProcessor().Evaluate(&firmwareRule.Rule, contextMap, fields)
				diff1 := time.Now().Sub(t0).Milliseconds()
				if diff1 > 10 {
					fields["eval_duration"] = diff1
					log.WithFields(fields).Debugf("rainbow templateType %s for ruleType %s for id %s name as %s", templateType, ruleType, firmwareRule.ID, firmwareRule.Name)
				}

				if time.Since(frEvalstart) > time.Duration(50*time.Microsecond) {
					log.WithFields(fields).Debugf("Evaluate ends.... templateType %s for ruleType %s for id %s name as %s time passed %v", templateType, ruleType, firmwareRule.ID, firmwareRule.Name, time.Since(frEvalstart))
				}

				if isEvaluate {
					log.WithFields(fields).Debugf("===>>> completed if condition ends. templateType %s finished in %v ===>>> to get into %v from start of method %v", templateType, time.Since(frEvalstart), time.Since(fmwareloopstart), time.Since(start))
					results = append(results, firmwareRule)
					log.WithFields(fields).Debugf("===>>> append completed. templateType %s finished in %v ===>>> to get into %v from start of method %v", templateType, time.Since(frEvalstart), time.Since(fmwareloopstart), time.Since(start))

					if isSingle {
						log.WithFields(fields).Debugf("FindMatchedRules for templateType %s ends for isSingle. finished in %v", templateType, time.Since(start))
						return results
					}
				}
			}
		}
		log.WithFields(fields).Debugf("FindMatchedRules for templateType %s  ends. finished in %v", templateType, time.Since(start))
	}
	return results
}

// GetSortedTemplate returns FirmwareRuleTemplate sorted by Priority
func (e *EstbFirmwareRuleBase) GetSortedTemplate(actionType corefw.ApplicableActionType, reverse bool, fields log.Fields) []*corefw.FirmwareRuleTemplate {
	var cacheKey string
	if reverse {
		cacheKey = fmt.Sprintf("%s_%s_%s", "FirmwareRuleTemplateSortedList", actionType, "desc")
	} else {
		cacheKey = fmt.Sprintf("%s_%s_%s", "FirmwareRuleTemplateSortedList", actionType, "asc")
	}

	cm := db.GetCacheManager()
	cacheInst := cm.ApplicationCacheGet(db.TABLE_FIRMWARE_RULE_TEMPLATE, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*corefw.FirmwareRuleTemplate)
	}

	all, err := corefw.GetFirmwareRuleTemplateAllAsListByActionType(actionType)
	if err != nil {
		log.WithFields(fields).Warn(fmt.Sprintf("No rule templates from DB with action Type %v", actionType))
		return all
	}

	if len(all) <= 1 {
		return all
	}

	var sortedList []*corefw.FirmwareRuleTemplate
	sortedList = append(sortedList, all...)

	if reverse { // sort by priority desc
		sort.Slice(sortedList, func(i, j int) bool {
			return sortedList[i].Priority > sortedList[j].Priority
		})
	} else { // sort by priority asc
		sort.Slice(sortedList, func(i, j int) bool {
			return sortedList[i].Priority < sortedList[j].Priority
		})
	}

	cm.ApplicationCacheSet(db.TABLE_FIRMWARE_RULE_TEMPLATE, cacheKey, sortedList)

	return sortedList
}

func (e *EstbFirmwareRuleBase) GetBoundConfigId(ctx map[string]string, convertedContext *coreef.ConvertedContext, firmwareRule *corefw.FirmwareRule, appliedVersionInfo map[string]string) string {
	if firmwareRule.ApplicableAction == nil {
		return ""
	}

	ruleAction := firmwareRule.ApplicableAction

	if (firmware.ENV_MODEL_RULE != firmwareRule.GetTemplateId()) || !ruleAction.Active || e.IsInWhitelist(convertedContext, ruleAction.Whitelist) {
		return e.extractAnyPresentConfig(ruleAction)
	}

	return e.ExtractConfigFromAction(convertedContext, ruleAction, appliedVersionInfo)
}

func (e *EstbFirmwareRuleBase) IsInWhitelist(convertedContext *coreef.ConvertedContext, whitelist string) bool {
	if len(whitelist) == 0 {
		return false
	}

	return e.ruleProcessorFactory.RuleProcessor().Evaluate(coreef.NewRuleFactory().NewIpFilter(whitelist), convertedContext.GetProperties(), log.Fields{})
}

func (e *EstbFirmwareRuleBase) extractAnyPresentConfig(ruleAction *corefw.ApplicableAction) string {
	if ruleAction.ConfigEntries != nil && len(ruleAction.ConfigEntries) > 0 {
		for _, configEntry := range ruleAction.ConfigEntries {
			return configEntry.ConfigId
		}
	}
	return ruleAction.ConfigId
}

func sortByConditionsSize(rules []*corefw.FirmwareRule) {
	// firmwareRules := make([]*corefw.FirmwareRule, len(rules))
	// copy(firmwareRules, rules)
	// assume need to match one with more conditions
	start := time.Now()
	log.Debug(">>>>>> sortByConditionsSize starts...")
	sort.Slice(rules, func(i, j int) bool {
		return getConditionsSize(rules[i].Rule) > getConditionsSize(rules[j].Rule)
	})
	log.Debugf(">>>>>> sortByConditionsSize finished in %v", time.Since(start))
	// return rules
}

func (e *EstbFirmwareRuleBase) ExtractConfigFromAction(context *coreef.ConvertedContext, ruleAction *corefw.ApplicableAction, appliedVersionInfo map[string]string) string {
	firmwareVersionIsAbsentInFilter := false
	if len(ruleAction.FirmwareVersions) == 0 || len(context.GetFirmwareVersionConverted()) == 0 {
		firmwareVersionIsAbsentInFilter = true
	} else {
		if !corefw.HasFirmwareVersion(ruleAction.FirmwareVersions, context.GetFirmwareVersionConverted()) {
			firmwareVersionIsAbsentInFilter = true
		}
	}

	if ruleAction.FirmwareCheckRequired && firmwareVersionIsAbsentInFilter {
		if ruleAction.RebootImmediately {
			context.AddForceFiltersConverted(firmware.REBOOT_IMMEDIATELY_FILTER)
		}

		context.AddBypassFiltersConverted(firmware.TIME_FILTER)

		config, _ := coreef.GetFirmwareConfigOneDB(ruleAction.IntermediateVersion)
		if config != nil && !strings.EqualFold(context.GetFirmwareVersionConverted(), config.FirmwareVersion) {
			// return IntermediateVersion firmware config
			appliedVersionInfo[FIRMWARE_SOURCE] = "IV,doesntMeetMinCheck"
			return ruleAction.IntermediateVersion
		} else {
			config, _ = coreef.GetFirmwareConfigOneDB(ruleAction.ConfigId) // lkg config
			if config != nil {
				// return LKG firmware config
				appliedVersionInfo[FIRMWARE_SOURCE] = "LKG,doesntMeetMinCheck"
				return ruleAction.ConfigId
			}
		}
		return e.extractAnyPresentConfig(ruleAction)
	}

	config, _ := coreef.GetFirmwareConfigOneDB(ruleAction.ConfigId)
	if config != nil {
		appliedVersionInfo[FIRMWARE_SOURCE] = "LKG,meetMinCheck"
	}

	if ruleAction.UseAccountPercentage && len(context.GetAccountIdConverted()) == 0 {
		return ruleAction.ConfigId
	}

	if ruleAction.ConfigEntries != nil && len(ruleAction.ConfigEntries) > 0 {
		var currentPercent float64 = 0
		source := e.getSource(context, ruleAction)
		qtdsource := fmt.Sprintf(`"%v"`, source)
		for _, entry := range ruleAction.ConfigEntries {
			percentage := entry.Percentage
			startPercentRange := entry.StartPercentRange
			endPercentRange := entry.EndPercentRange
			if startPercentRange >= 0 && endPercentRange >= 0 {
				if !re.FitsPercent(qtdsource, startPercentRange) && re.FitsPercent(qtdsource, endPercentRange) {
					appliedVersionInfo[FIRMWARE_SOURCE] = "MultipleVersionDistribution"
					return entry.ConfigId
				}
			} else if percentage > 0 {
				currentPercent += percentage
				if re.FitsPercent(qtdsource, currentPercent) {
					appliedVersionInfo[FIRMWARE_SOURCE] = "MultipleVersionDistribution"
					return entry.ConfigId
				}
			}
		}
	}
	return ruleAction.ConfigId
}

func (e *EstbFirmwareRuleBase) getSource(context *coreef.ConvertedContext, ruleAction *corefw.ApplicableAction) interface{} {
	if ruleAction.UseAccountPercentage {
		return context.GetAccountIdConverted()
	}

	if context.GetEstbMacConverted() != "" {
		return context.GetEstbMacConverted()
	}

	return context
}

func (e *EstbFirmwareRuleBase) getFirmwareTemplate(ruleType string, clone bool) *corefw.FirmwareRuleTemplate {
	if len(ruleType) == 0 {
		return nil
	}
	t, err := corefw.GetFirmwareRuleTemplateOneDB(ruleType)
	if err != nil {
		log.Error(fmt.Sprintf("failed to read template rule from db with ruleType %s", ruleType))
		return nil
	}
	return t
}

func (e *EstbFirmwareRuleBase) ApplyMatchedFilters(
	rules map[string][]*corefw.FirmwareRule,
	templateType corefw.ApplicableActionType,
	context map[string]string,
	bypassFilters map[string]struct{},
	evaluationResult *EvaluationResult) map[string]interface{} {
	mapinst := make(map[string]interface{})
	matchedRules := e.FindMatchedRules(rules, templateType, context, bypassFilters, false, true, log.Fields{})
	matchedActivationVersion := false

	for _, firmwareRule := range matchedRules {
		action := firmwareRule.ApplicableAction
		template := e.getFirmwareTemplate(firmwareRule.Type, false)
		firmwareVersion := context[common.FIRMWARE_VERSION]
		if !matchedActivationVersion && (firmware.ACTIVATION_VERSION == template.ID) && len(firmwareVersion) > 0 {
			matchedActivationVersion = true
			isPresent := corefw.HasFirmwareVersion(action.GetFirmwareVersions(), firmwareVersion)
			versionExs := action.GetFirmwareVersionRegExs()
			if isPresent || len(versionExs) > 0 && e.matchFirmwareVersionRegEx(versionExs, firmwareVersion) {
				e.applyDefinePropertiesFilter(template, firmwareRule, action, mapinst, evaluationResult)
			} else {
				mapinst[coreef.REBOOT_IMMEDIATELY] = true
			}
		} else if firmware.ACTIVATION_VERSION != template.ID {
			e.applyDefinePropertiesFilter(template, firmwareRule, action, mapinst, evaluationResult)
			if len(action.ByPassFilters) > 0 {
				for _, f := range action.ByPassFilters {
					bypassFilters[f] = struct{}{}
				}
			}
		}

	}

	return mapinst
}

func (e *EstbFirmwareRuleBase) applyDefinePropertiesFilter(
	template *corefw.FirmwareRuleTemplate,
	firmwareRule *corefw.FirmwareRule,
	action *corefw.ApplicableAction,
	properties map[string]interface{},
	evaluationResult *EvaluationResult) {

	for k, v := range e.convertProperties(template, action.Properties) {
		properties[k] = v
	}

	evaluationResult.AddAppliedFilters(firmwareRule)
}

func (e *EstbFirmwareRuleBase) matchFirmwareVersionRegEx(regExs []string, firmwareVersion string) bool {
	for _, regEx := range regExs {
		matched, err := regexp.MatchString(regEx, firmwareVersion)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// convertProperties ... using the DefinePropertiesTemplateAction value overwrite default config values
func (e *EstbFirmwareRuleBase) convertProperties(template *corefw.FirmwareRuleTemplate, properties map[string]string) map[string]interface{} {
	converted := map[string]interface{}{}
	templateAction := template.ApplicableAction
	templateProperties := templateAction.Properties
	for key, value := range properties {
		if propertyValue, isPresent := templateProperties[key]; isPresent {
			valueObject, _ := e.convertBasedOnValidationType(propertyValue.ValidationTypes, value)
			converted[key] = valueObject
		} else {
			converted[key] = value
		}

	}

	return converted
}

func (e *EstbFirmwareRuleBase) convertBasedOnValidationType(validationTypes []corefw.ValidationType, value string) (interface{}, error) {
	if validationTypes != nil && len(validationTypes) == 1 {
		switch validationTypes[0] {
		case corefw.NUMBER, corefw.PERCENT, corefw.PORT:
			if s, err := strconv.ParseFloat(value, 32); err == nil {
				return s, nil
			} else {
				return value, err
			}
		case corefw.BOOLEAN:
			if s, err := strconv.ParseBool(value); err == nil {
				return s, nil
			} else {
				return value, err
			}
		default:
			return value, errors.New(fmt.Sprintf("Not validationType value %v", validationTypes[0]))
		}
	}
	return value, errors.New(fmt.Sprintf("Not validationTypes %v and len %d", validationTypes, len(validationTypes)))
}

func (e *EstbFirmwareRuleBase) DoFilters(ctx map[string]string, convertedContext *coreef.ConvertedContext, applicationType string, rules map[string][]*corefw.FirmwareRule, evaluationResult *EvaluationResult, fields log.Fields) bool {

	bypassFilters := convertedContext.GetBypassFiltersConverted()
	firmwareConfig := evaluationResult.FirmwareConfig
	filterId := e.getRoundRobinIdByApplication(applicationType)

	log.WithFields(fields).Debug(fmt.Sprintf("debug %v %v %v %v", bypassFilters, evaluationResult.AppliedFilters, firmwareConfig, filterId))
	downloadLocationRoundRobinFilterValue, err := coreef.GetDownloadLocationRoundRobinFilterValOneDB(filterId)
	if err != nil {
		log.WithFields(fields).Error("Failed to get download filter values")
	} else {
		if DownloadLocationRoundRobinFilterFilter(firmwareConfig, downloadLocationRoundRobinFilterValue, convertedContext) {
			evaluationResult.AddAppliedFilters(downloadLocationRoundRobinFilterValue)
		}
	}

	contextProperties := convertedContext.GetProperties()

	// download protocol should be setup in download location filter
	contextProperties[common.DOWNLOAD_PROTOCOL] = string(firmwareConfig.GetFirmwareDownloadProtocol())
	contextProperties[common.MATCHED_RULE_TYPE] = evaluationResult.MatchedRule.Type

	var map1 map[string]interface{}
	if e.isPercentFilter(evaluationResult.MatchedRule) {
		map1 = e.ApplyMatchedFilters(rules, corefw.DEFINE_PROPERTIES_TEMPLATE, contextProperties, bypassFilters, evaluationResult)
		applyMandatoryUpdateFlag(evaluationResult, convertedContext, map1)
	} else {
		corefw.RemoveAllByRuleTypes(rules, firmware.ACTIVATION_VERSION)
		map1 = map[string]interface{}{common.MANDATORY_UPDATE: false}
		map1 = e.ApplyMatchedFilters(rules, corefw.DEFINE_PROPERTIES_TEMPLATE, contextProperties, bypassFilters, evaluationResult)
	}

	firmwareConfig.PutAll(map1)
	downloadProtocol := firmwareConfig.GetFirmwareDownloadProtocol()

	// legacy: if protocol=tftp but ipv6 location is empty then read from round robing filter
	if downloadProtocol == "tftp" && len(firmwareConfig.GetStringValue(common.IPV6_FIRMWARE_LOCATION)) > 0 {
		DownloadLocationRoundRobinFilterSetupIPv6Location(firmwareConfig, downloadLocationRoundRobinFilterValue)
	}

	// legacy: if force reboot immediately then set to true flag
	if _, ok := convertedContext.GetForceFiltersConverted()[firmware.REBOOT_IMMEDIATELY_FILTER]; ok {
		firmwareConfig.SetRebootImmediately(true)
	}

	blockingFilter := e.FindMatchedRule(rules, corefw.BLOCKING_FILTER_TEMPLATE, contextProperties, bypassFilters, fields)
	if blockingFilter != nil {
		evaluationResult.AddAppliedFilters(blockingFilter)
		return true
	}

	return false
}

func (e *EstbFirmwareRuleBase) checkForDRIState(ctx map[string]string, config *coreef.FirmwareConfigFacade, blocked bool) bool {
	if len(e.driStateIdentifiers) == 0 || len(ctx[common.FIRMWARE_VERSION]) == 0 {
		return blocked
	}

	identifiers := strings.Split(e.driStateIdentifiers, ",")

	for _, identifier := range identifiers {
		if strings.Contains(strings.ToUpper(ctx[common.FIRMWARE_VERSION]), strings.ToUpper(identifier)) {
			blocked = false
			if config != nil {
				config.SetRebootImmediately(true)
			}
			return blocked
		}
	}
	return blocked
}

func (e *EstbFirmwareRuleBase) isPercentFilter(firmwareRule *corefw.FirmwareRule) bool {
	if firmwareRule != nil && firmware.ENV_MODEL_RULE == firmwareRule.Type {
		return true
	}
	return false
}

func getConditionsSize(rule re.Rule) int {
	return len(re.ToConditions(&rule))
}

func (e *EstbFirmwareRuleBase) FilterByAppType(rules []*corefw.FirmwareRule, applicationType string) map[string][]*corefw.FirmwareRule {
	log.Debug("FilterByAppType starts...")
	result := map[string][]*corefw.FirmwareRule{}
	for _, rule := range rules {
		if rule.Type == "" {
			log.Error(fmt.Sprintf("ruleType is null: %v", rule))
			continue
		}
		if rule.ApplicationType == applicationType {
			r, ok := result[rule.Type]
			if !ok {
				r = []*corefw.FirmwareRule{}
			}
			r = append(r, rule)
			result[rule.Type] = r
		}
	}
	log.Debug("FilterByAppType ends...")
	return result
}

// HasMinimumFirmware ...
func (e *EstbFirmwareRuleBase) HasMinimumFirmware(ctx map[string]string) bool {
	convertedContext := coreef.GetContextConverted(ctx)

	eval, err := e.Eval(ctx, convertedContext, shared.STB, log.Fields{})
	if err != nil {
		log.Error(fmt.Sprintf("HasMinimumFirmware eval error: %v", err))
		return true
	}

	if eval == nil {
		log.Error("HasMinimumFirmware eval result nil")
		return true
	}

	matchedRule := eval.MatchedRule

	if matchedRule != nil && !eval.Blocked && eval.FirmwareConfig != nil && firmware.ENV_MODEL_RULE == matchedRule.Type {
		if matchedRule.ApplicableAction != nil {
			ruleAction := matchedRule.ApplicableAction
			firmwareVersionIsAbsentInFilter := ruleAction.FirmwareVersions == nil || ctx["firmwareVersion"] == ""

			if !firmwareVersionIsAbsentInFilter && !corefw.HasFirmwareVersion(ruleAction.FirmwareVersions, ctx["firmwareVersion"]) {
				firmwareVersionIsAbsentInFilter = true
			}

			if ruleAction.FirmwareCheckRequired && firmwareVersionIsAbsentInFilter {
				return false
			}
		}

	}

	return true
}

// GetBseConfiguration ...
func (e *EstbFirmwareRuleBase) GetBseConfiguration(address *shared.IpAddress) (*coreef.BseConfiguration, error) {
	modelConfigs := []*coreef.ModelFirmwareConfiguration{}
	rulelst, err := corefw.GetFirmwareSortedRuleAllAsListDB()
	if err != nil {
		log.Error(fmt.Sprintf("GetBseConfiguration DB firmware rule error: %v", err))
		return nil, err
	}

	for _, firmwareRule := range rulelst {
		if firmware.IP_RULE == firmwareRule.Type && !firmwareRule.IsNoop() && firmwareRule.ApplicationType == shared.STB {
			ipRuleBean := coreef.ConvertFirmwareRuleToIpRuleBeanAddFirmareConfig(firmwareRule)
			firmwareConfig := ipRuleBean.FirmwareConfig
			if ipRuleBean.IpAddressGroup != nil && ipRuleBean.IpAddressGroup.IsInRange(address.GetAddress()) && firmwareConfig != nil {
				modelConfig := coreef.NewModelFirmwareConfiguration(ipRuleBean.ModelId, firmwareConfig.FirmwareFilename, firmwareConfig.FirmwareVersion)
				modelConfigs = append(modelConfigs, modelConfig)
			}
		}
	}

	if len(modelConfigs) == 0 {
		return nil, errors.New("Can't find any matched Ip Rule/Config")
	}

	config := &coreef.BseConfiguration{}
	config.ModelConfigurations = modelConfigs

	downloadLocationRoundRobinFilterValue, err := coreef.GetDefaultDownloadLocationRoundRobinFilterValOneDB()
	if err != nil {
		log.Error(fmt.Sprintf("GetBseConfiguration DB  downloadLocationRoundRobinFilterValue error: %v", err))
	} else {
		locations := downloadLocationRoundRobinFilterValue.GetDownloadLocations()
		config.Protocol = "Tftp"
		config.Location = locations[0]
		config.Ipv6Location = locations[1]
	}

	for _, firmwareRule := range rulelst {
		if firmware.DOWNLOAD_LOCATION_FILTER == firmwareRule.Type {
			if e.isIpAddressInRange(firmwareRule.Rule, address) {
				e.setupLocations(firmwareRule, config)
			}
		}
	}
	return config, nil
}

func (e *EstbFirmwareRuleBase) setupLocations(firmwareRule *corefw.FirmwareRule, config *coreef.BseConfiguration) {
	action := firmwareRule.ApplicableAction
	location := action.Properties[common.FIRMWARE_LOCATION]
	ipv6Location := action.Properties[common.IPV6_FIRMWARE_LOCATION]
	protocol := action.Properties[common.FIRMWARE_DOWNLOAD_PROTOCOL]

	useHttp := false

	if len(location) != 0 && "tftp" != protocol {
		useHttp = true
	}

	if useHttp {
		config.Protocol = "http"
		config.Location = location
		if len(ipv6Location) != 0 {
			config.Ipv6Location = ipv6Location
		}
	} else if len(location) != 0 {
		config.Protocol = "tftp"
		config.Location = location
		if len(ipv6Location) != 0 {
			config.Ipv6Location = ipv6Location
		}
	}
}

func (e *EstbFirmwareRuleBase) isIpAddressInRange(rule re.Rule, address *shared.IpAddress) bool {
	for _, condition := range re.ToConditions(&rule) {
		fixedArg := condition.GetFixedArg()
		if fixedArg == nil || fixedArg.GetValue() == nil || !coreef.RuleFactoryIP.Equals(condition.GetFreeArg()) {
			continue
		}
		op := condition.GetOperation()
		value := fixedArg.GetValue()
		if re.StandardOperationIs == op && value != nil {
			return isIpInRange(value.(string), *address)
		} else if re.StandardOperationIn == op {
			for _, ipAddressStr := range value.([]string) {
				if isIpInRange(ipAddressStr, *address) {
					return true
				}
			}
		} else if coreef.RuleFactoryIN_LIST == op {
			ipListId := value.(string)
			ipList, err := shared.GetGenericNamedListOneByType(ipListId, shared.IP_LIST)
			if err != nil {
				log.Error(fmt.Sprintf("Call GetGenericNamedListOneByType error %v", err))
			}
			if ipList != nil {
				for _, ipListItem := range ipList.Data {
					if isIpInRange(ipListItem, *address) {
						return true
					}
				}
			}
		}
	}

	return false
}

// TODO need to adjust
func isIpInRange(ipAddressStr string, addressToCheck shared.IpAddress) bool {
	ipAddress := shared.NewIpAddress(ipAddressStr)
	if ipAddress == nil {
		log.Error(fmt.Sprintf("Exception: addressInWhichNeedToCheck %s addressToCheck %v", ipAddressStr, addressToCheck))
		return false
	}
	return ipAddress.IsInRange(addressToCheck)
}

func (e *EstbFirmwareRuleBase) getRoundRobinIdByApplication(applicationType string) string {
	if shared.STB == applicationType {
		return coreef.ROUND_ROBIN_FILTER_SINGLETON_ID
	}
	return fmt.Sprintf("%s_%s", strings.ToUpper(applicationType), coreef.ROUND_ROBIN_FILTER_SINGLETON_ID)
}

func percentFilterTemplateNames() []string {
	return firmware.PercentFilterTemplateNames
}

func (e *EstbFirmwareRuleBase) GetAppliedActivationVersionType(ctx map[string]string, applicationType string) *RunningVersionInfo {
	runningVersionInfo := &RunningVersionInfo{
		HasActivationMinFW: true,
		HasMinimumFW:       true,
	}
	firmwareVersion := ctx["firmwareVersion"]
	// get all firmreRoles as sorted
	rulereflst, err := corefw.GetFirmwareSortedRuleAllAsListDB()
	firmwareRules := map[string][]*corefw.FirmwareRule{}

	if err == nil {
		firmwareRules = e.FilterByAppType(rulereflst, applicationType)
	}

	convertedContext := coreef.GetContextConverted(ctx)

	// TODO: please update log fields here
	var fields log.Fields = log.Fields{}
	eval, err := e.Eval(ctx, convertedContext, applicationType, fields)
	if err != nil {
		log.Error(fmt.Sprintf("Error GetAppliedActivationVersionType call Eval with error %v", err))
		return runningVersionInfo
	}

	matchedRule := eval.MatchedRule

	isPercentRuleIsMatched := false
	if matchedRule != nil && !eval.Blocked && eval.FirmwareConfig != nil && firmware.ENV_MODEL_RULE == matchedRule.Type {
		isPercentRuleIsMatched = true
	}

	if isPercentRuleIsMatched && matchedRule.ApplicableAction != nil {
		ruleAction := matchedRule.ApplicableAction
		if len(ruleAction.FirmwareVersions) == 0 || len(firmwareVersion) == 0 {
			runningVersionInfo.HasMinimumFW = false
		} else if !corefw.HasFirmwareVersion(ruleAction.FirmwareVersions, firmwareVersion) {
			runningVersionInfo.HasMinimumFW = false
		}

		convertedcontext := coreef.GetContextConverted(ctx)
		matchedActivationVersionRule := e.FindMatchedRuleByRules(firmwareRules, corefw.DEFINE_PROPERTIES_TEMPLATE, firmware.ACTIVATION_VERSION, convertedcontext.GetProperties(), convertedcontext.GetBypassFiltersConverted())

		isAmvRuleMatched := false
		if matchedActivationVersionRule != nil {
			action := matchedActivationVersionRule.ApplicableAction
			if len(firmwareVersion) != 0 && (e.firmwareVersionIsMatched(firmwareVersion, action) || e.firmwareVersionRegExIsMatched(firmwareVersion, action)) {
				isAmvRuleMatched = true
			}

		}
		runningVersionInfo.HasActivationMinFW = isAmvRuleMatched

	}

	return runningVersionInfo
}

func (e *EstbFirmwareRuleBase) firmwareVersionIsMatched(firmwareVersion string, action *corefw.ApplicableAction) bool {
	actionVersions := action.GetFirmwareVersions()
	return corefw.HasFirmwareVersion(actionVersions, firmwareVersion)
}

func (e *EstbFirmwareRuleBase) firmwareVersionRegExIsMatched(firmwareVersion string, action *corefw.ApplicableAction) bool {
	actionRegExs := action.GetFirmwareVersionRegExs()
	if len(actionRegExs) != 0 {
		for _, regEx := range actionRegExs {
			matched, err := regexp.MatchString(regEx, firmwareVersion)
			if err != nil && matched {
				return true
			}
		}
	}
	return false
}

func (e *EstbFirmwareRuleBase) FindMatchedRuleByRules(
	firmwareRules map[string][]*corefw.FirmwareRule,
	actionType corefw.ApplicableActionType,
	template string,
	contextMap map[string]string,
	bypassFilters map[string]struct{}) *corefw.FirmwareRule {

	matchedRules := e.FindMatchedRules(firmwareRules, actionType, contextMap, bypassFilters, false, true, log.Fields{})
	filteredByTemplateRules := e.FilterByTemplate(matchedRules, template)
	if len(filteredByTemplateRules) > 0 {
		return filteredByTemplateRules[0]
	}

	return nil
}

// the following is just simple code try to simulate collection applying filer concept
// it could be done simple Loop of rules, and check ruleType == tempalteName
type FirmwareRuleOfTemplate struct {
	templaleName string
}

func (f *FirmwareRuleOfTemplate) isFirmwareRuleOfTemplate(rule *corefw.FirmwareRule) bool {
	if strings.EqualFold(f.templaleName, rule.Type) {
		return true
	}
	return false
}

func (e *EstbFirmwareRuleBase) FilterByTemplate(firmwareRules []*corefw.FirmwareRule, templateName string) []*corefw.FirmwareRule {
	f := &FirmwareRuleOfTemplate{
		templaleName: templateName,
	}

	result := []*corefw.FirmwareRule{}
	// can't use filter --> look like the author also ask using for loop
	for _, rule := range firmwareRules {
		if f.isFirmwareRuleOfTemplate(rule) {
			result = append(result, rule)
		}
	}
	return result
}

func applyMandatoryUpdateFlag(evaluationResult *EvaluationResult, context *coreef.ConvertedContext, properties map[string]interface{}) {
	currentFirmwareVersion := context.FirmwareVersion
	if currentFirmwareVersion == "" && evaluationResult.MatchedRule.ApplicableAction.ActionType == corefw.RULE {
		return
	}
	matchedRule := evaluationResult.MatchedRule
	ruleAction := matchedRule.ApplicableAction
	if ruleAction.FirmwareCheckRequired && len(ruleAction.FirmwareVersions) > 0 && !contains(ruleAction.FirmwareVersions, currentFirmwareVersion) {
		properties[common.MANDATORY_UPDATE] = true
	} else {
		properties[common.MANDATORY_UPDATE] = false
	}
}

// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
