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
package firmware

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"xconfwebconfig/common"
	"xconfwebconfig/db"
	re "xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	"xconfwebconfig/util"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	MAC_RULE       = "MAC_RULE"
	IP_RULE        = "IP_RULE"
	ENV_MODEL_RULE = "ENV_MODEL_RULE"

	IP_FILTER   = "IP_FILTER"
	TIME_FILTER = "TIME_FILTER"

	REBOOT_IMMEDIATELY_FILTER = "REBOOT_IMMEDIATELY_FILTER"
	DOWNLOAD_LOCATION_FILTER  = "DOWNLOAD_LOCATION_FILTER"

	IV_RULE        = "IV_RULE"
	MIN_CHECK_RULE = "MIN_CHECK_RULE"
	MIN_CHECK_RI   = "MIN_CHECK_RI"

	GLOBAL_PERCENT = "GLOBAL_PERCENT"

	ACTIVATION_VERSION = "ACTIVATION_VERSION"
	KEY                = "FREE_ARG"
	VALUE              = "FIXED_ARG"
)

var (
	TemplateNames = []string{MAC_RULE, IP_RULE, ENV_MODEL_RULE, IP_FILTER, TIME_FILTER, REBOOT_IMMEDIATELY_FILTER,
		DOWNLOAD_LOCATION_FILTER, IV_RULE, MIN_CHECK_RULE, MIN_CHECK_RI, GLOBAL_PERCENT, ACTIVATION_VERSION}

	PercentFilterTemplateNames = []string{ENV_MODEL_RULE, MIN_CHECK_RULE, IV_RULE}
)

// FirmwareRule FirmwareRule4 table
type ApplicableAction struct {
	Type                       string               `json:"type"`
	ActionType                 ApplicableActionType `json:"actionType,omitempty"`
	ConfigId                   string               `json:"configId,omitempty"`
	ConfigEntries              []ConfigEntry        `json:"configEntries"` // RuleAction
	Active                     bool                 `json:"active"`
	UseAccountPercentage       bool                 `json:"useAccountPercentage"`
	FirmwareCheckRequired      bool                 `json:"firmwareCheckRequired"`
	RebootImmediately          bool                 `json:"rebootImmediately"`
	Whitelist                  string               `json:"whitelist,omitempty"`
	IntermediateVersion        string               `json:"intermediateVersion,omitempty"`
	FirmwareVersions           []string             `json:"firmwareVersions,omitempty"`
	Properties                 map[string]string    `json:"properties,omitempty"` // DefinePropertiesAction
	ByPassFilters              []string             `json:"byPassFilters,omitempty"`
	ActivationFirmwareVersions map[string][]string  `json:"activationFirmwareVersions,omitempty"`
}

type TemplateApplicableAction struct {
	Type                       string                   `json:"type"`
	ActionType                 ApplicableActionType     `json:"actionType"`
	ConfigId                   string                   `json:"configI,omitempty"`
	ConfigEntries              []ConfigEntry            `json:"configEntries,omitempty"` // RuleAction
	Active                     bool                     `json:"active"`
	UseAccountPercentage       bool                     `json:"useAccountPercentage"`
	FirmwareCheckRequired      bool                     `json:"firmwareCheckRequired"`
	RebootImmediately          bool                     `json:"rebootImmediately"`
	Whitelist                  string                   `json:"whitelist,omitempty"`
	IntermediateVersion        string                   `json:"intermediateVersion,omitempty"`
	FirmwareVersions           []string                 `json:"firmwareVersions,omitempty"`
	Properties                 map[string]PropertyValue `json:"properties,omitempty"` // DefinePropertiesAction
	ByPassFilters              []string                 `json:"byPassFilters,omitempty"`
	ActivationFirmwareVersions map[string][]string      `json:"activationFirmwareVersions,omitempty"`
}

func NewTemplateApplicableActionAndType(typ string, actionType ApplicableActionType, configId string) *TemplateApplicableAction {
	action := &TemplateApplicableAction{
		Type:                       typ,
		ActionType:                 actionType,
		ConfigId:                   configId,
		Active:                     true,
		UseAccountPercentage:       false,
		FirmwareCheckRequired:      false,
		RebootImmediately:          false,
		ConfigEntries:              []ConfigEntry{},
		FirmwareVersions:           []string{},
		ByPassFilters:              []string{},
		ActivationFirmwareVersions: map[string][]string{},
		Properties:                 map[string]PropertyValue{},
	}
	return action
}

func NewTemplateApplicableAction(typ string, configId string) *TemplateApplicableAction {
	action := &TemplateApplicableAction{
		Type:                       typ,
		ConfigId:                   configId,
		Active:                     true,
		UseAccountPercentage:       false,
		FirmwareCheckRequired:      false,
		RebootImmediately:          false,
		ConfigEntries:              []ConfigEntry{},
		FirmwareVersions:           []string{},
		ByPassFilters:              []string{},
		ActivationFirmwareVersions: map[string][]string{},
		Properties:                 map[string]PropertyValue{},
	}
	return action
}

func NewApplicableActionAndType(typ string, actionType ApplicableActionType, configId string) *ApplicableAction {
	action := &ApplicableAction{
		Type:                       typ,
		ActionType:                 actionType,
		ConfigId:                   configId,
		Active:                     true,
		UseAccountPercentage:       false,
		FirmwareCheckRequired:      false,
		RebootImmediately:          false,
		ConfigEntries:              []ConfigEntry{},
		FirmwareVersions:           []string{},
		ByPassFilters:              []string{},
		ActivationFirmwareVersions: map[string][]string{},
	}
	return action
}

func NewApplicableAction(typ string, configId string) *ApplicableAction {
	action := &ApplicableAction{
		Type:                       typ,
		ConfigId:                   configId,
		Active:                     true,
		UseAccountPercentage:       false,
		FirmwareCheckRequired:      false,
		RebootImmediately:          false,
		ConfigEntries:              []ConfigEntry{},
		FirmwareVersions:           []string{},
		ByPassFilters:              []string{},
		ActivationFirmwareVersions: map[string][]string{},
	}
	return action
}

func (a ApplicableAction) IsValid() bool {
	return true
}

func (a ApplicableAction) String() string {
	return fmt.Sprintf(`ApplicableAction(
	Type=%v,
	ActionType=%v,
	ConfigId=%v,
	ConfigEntries=%v,
	Active=%v,
	UseAccountPercentage=%v,
	FirmwareCheckRequired=%v,
	RebootImmediately=%v,
	Whitelist=%v,
	IntermediateVersion=%v,
	FirmwareVersions=%v,
	Properties=%v,
	ByPassFilters=%v,
	ActivationFirmwareVersions=%v,
  )`,
		a.Type,
		a.ActionType,
		a.ConfigId,
		a.ConfigEntries,
		a.Active,
		a.UseAccountPercentage,
		a.FirmwareCheckRequired,
		a.RebootImmediately,
		a.Whitelist,
		a.IntermediateVersion,
		a.FirmwareVersions,
		a.Properties,
		a.ByPassFilters,
		a.ActivationFirmwareVersions,
	)
}

// FirmwareRule FirmwareRule4 table
type FirmwareRule struct {
	ID               string            `json:"id"`
	Updated          int64             `json:"updated"`
	Name             string            `json:"name"`
	ApplicableAction *ApplicableAction `json:"applicableAction"`
	Rule             re.Rule           `json:"rule"`
	Type             string            `json:"type"`
	Active           bool              `json:"active"`
	ApplicationType  string            `json:"applicationType,omitempty"`
}

func (obj *FirmwareRule) SetApplicationType(appType string) {
	obj.ApplicationType = appType
}

func (obj *FirmwareRule) GetApplicationType() string {
	return obj.ApplicationType
}

func (obj *FirmwareRule) Clone() (*FirmwareRule, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*FirmwareRule), nil
}

func NewEmptyFirmwareRule() *FirmwareRule {
	return &FirmwareRule{
		Active:           true,
		ApplicationType:  shared.STB,
		ApplicableAction: NewApplicableAction("", ""),
	}
}

func NewFirmwareRuleInf() interface{} {
	return &FirmwareRule{
		Active:          true,
		ApplicationType: shared.STB,
	}
}

func NewFirmwareRule(id string, name string, ruleType string, rule *re.Rule, action *ApplicableAction, active bool) *FirmwareRule {
	firmwareRule := &FirmwareRule{
		Rule:             *rule,
		ID:               id,
		Active:           active,
		Name:             name,
		Type:             ruleType,
		ApplicableAction: action,
	}
	return firmwareRule
}

func (r *FirmwareRule) Validate() error {
	if r.Type == "" {
		return fmt.Errorf("FirmwareRule's Type is is not present")
	}

	if r.ApplicableAction == nil {
		return fmt.Errorf("FirmwareRule's ApplicableAction is not present: %s", r.ID)
	}

	if !IsValidApplicableActionType(r.ApplicableAction.ActionType) {
		return fmt.Errorf("ApplicableAction's ActionType is invalid: %s", r.ApplicableAction.ActionType)
	}

	if !isValidApplicableClass(r.ApplicableAction.Type) {
		return fmt.Errorf("ApplicableAction's Type is invalid: %s", r.ApplicableAction.Type)
	}

	return nil
}

func (r *FirmwareRule) Equals(f *FirmwareRule) bool {
	if r.ID != f.ID {
		return false
	}

	if r.Name != f.Name {
		return false
	}

	r1 := &r.Rule
	r2 := &f.Rule
	if !r1.Equals(r2) {
		return false
	}
	if r.Type != f.Type {
		return false
	}
	if r.Active != f.Active {
		return false
	}
	if r.ApplicationType != f.ApplicationType {
		return false
	}
	return true
}

func (r *FirmwareRule) String() string {
	return fmt.Sprintf(`FirmwareRule(
  Id=%v,
  Updated=%v,
  Name=%v,
  Type=%v,
  Active=%v,
  ApplicationType=%v,
  Rule=%v,
  ApplicableAction=%v,
)`,
		r.ID,
		r.Updated,
		r.Name,
		r.Type,
		r.Active,
		r.ApplicationType,
		r.Rule.String(),
		r.ApplicableAction.String(),
	)
}

func (r *FirmwareRule) ConfigId() string {
	if len(r.ApplicableAction.ConfigId) > 0 {
		return r.ApplicableAction.ConfigId
	}
	if len(r.ApplicableAction.ConfigEntries) > 0 {
		return r.ApplicableAction.ConfigEntries[0].ConfigId
	}
	return ""
}

// GetId XRule interface
func (r *FirmwareRule) GetId() string {
	return r.ID
}

// GetRule XRule interface
func (r *FirmwareRule) GetRule() *re.Rule {
	return &r.Rule
}

// GetName XRule interface
func (r *FirmwareRule) GetName() string {
	return r.Name
}

// GetTemplateId XRule interface
func (r *FirmwareRule) GetTemplateId() string {
	return r.Type
}

// GetRuleType XRule interface
func (r *FirmwareRule) GetRuleType() string {
	return "FirmwareRule"
}

// IsNoop ...
func (r *FirmwareRule) IsNoop() bool {
	if r.ApplicableAction != nil {
		if len(r.ApplicableAction.ConfigId) > 0 {
			return false
		}

		if r.ApplicableAction.ConfigEntries != nil {
			for _, entry := range r.ApplicableAction.ConfigEntries {
				if len(entry.ConfigId) > 0 {
					return false
				}
			}
		}
	}
	return true
}

// FirmwareRuleTemplate table
type FirmwareRuleTemplate struct {
	ID                   string                    `json:"id"`
	Updated              int64                     `json:"updated,omitempty"`
	Rule                 re.Rule                   `json:"rule"`
	ApplicableAction     *TemplateApplicableAction `json:"applicableAction"`
	Priority             int32                     `json:"priority"`
	RequiredFields       []string                  `json:"requiredFields,omitempty"`
	ByPassFilters        []string                  `json:"byPassFilters,omitempty"`
	ValidationExpression string                    `json:"validationExpression,omitempty"`
	Editable             bool                      `json:"editable"`
}

func (obj *FirmwareRuleTemplate) Clone() (*FirmwareRuleTemplate, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*FirmwareRuleTemplate), nil
}

func NewEmptyFirmwareRuleTemplate() *FirmwareRuleTemplate {
	return &FirmwareRuleTemplate{
		Editable:       true,
		RequiredFields: []string{},
		ByPassFilters:  []string{},
	}
}

func NewFirmwareRuleTemplateInf() interface{} {
	return &FirmwareRuleTemplate{
		Editable: true,
	}
}

// GetId XRule interface
func (r *FirmwareRuleTemplate) GetId() string {
	return r.ID
}

func (r *FirmwareRuleTemplate) GetID() string {
	return r.ID
}
func (obj *FirmwareRuleTemplate) GetPriority() int {
	return int(obj.Priority)
}

func (obj *FirmwareRuleTemplate) SetPriority(priority int) {
	obj.Priority = int32(priority)
}

func (obj *FirmwareRuleTemplate) Validate() error {
	if obj.ApplicableAction == nil {
		return fmt.Errorf("FirmwareRuleTemplate's TemplateApplicableAction is not present: %s", obj.ID)
	}

	if !IsValidApplicableActionType(obj.ApplicableAction.ActionType) {
		return fmt.Errorf("TemplateApplicableAction's ActionType is invalid: %s", obj.ApplicableAction.ActionType)
	}

	if !isValidApplicableClass(obj.ApplicableAction.Type) {
		return fmt.Errorf("TemplateApplicableAction's Type is invalid: %s", obj.ApplicableAction.Type)
	}

	return nil
}

// GetRule XRule interface
func (r *FirmwareRuleTemplate) GetRule() *re.Rule {
	return &r.Rule
}

// GetTemplateId XRule interface
func (f *FirmwareRuleTemplate) GetTemplateId() string {
	return f.ID
}

// GetRuleType XRule interface
func (f *FirmwareRuleTemplate) GetRuleType() string {
	return "FirmwareRuleTemplate"
}

// GetName XRule interface
func (f *FirmwareRuleTemplate) GetName() string {
	return f.ID
}

// GetRulesByRuleTypes ...
func GetRulesByRuleTypes(rules map[string][]*FirmwareRule, ruleType string) []*FirmwareRule {
	// typedRules, ok := rules[ruleType]
	// if ok {
	// 	return typedRules
	// }
	// return []*FirmwareRule{}
	return rules[ruleType]
}

func RemoveAllByRuleTypes(rules map[string][]*FirmwareRule, ruleType string) {
	delete(rules, ruleType)
}

func GetFirmwareRuleOneDB(id string) (*FirmwareRule, error) {
	inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_FIRMWARE_RULE, id)
	if err != nil {
		return nil, err
	}
	frule := inst.(*FirmwareRule)
	return frule, nil
}

func GetFirmwareRuleTemplateOneDBWithId(id string) (*FirmwareRuleTemplate, error) {
	dbinst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_FIRMWARE_RULE_TEMPLATE, id)
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareRuleTemplateOneDBWithId: %v", err))
		return nil, err
	}
	t := dbinst.(*FirmwareRuleTemplate)
	return t, nil
}

func GetFirmwareRuleTemplateOneDB(ruleType string) (*FirmwareRuleTemplate, error) {
	dbinst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_FIRMWARE_RULE_TEMPLATE, ruleType)
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareRuleTemplateOneDB: %v", err))
		return nil, err
	}
	t := dbinst.(*FirmwareRuleTemplate)
	return t, nil
}

func GetFirmwareRuleAllAsListDB() ([]*FirmwareRule, error) {
	cm := db.GetCacheManager()
	cacheKey := "FirmwareRuleList"
	cacheInst := cm.ApplicationCacheGet(db.TABLE_FIRMWARE_RULE, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*FirmwareRule), nil
	}

	// pass 0 or -1 as unlimit
	rulelst, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_FIRMWARE_RULE, 0)
	if err != nil {
		return nil, err
	}

	if len(rulelst) == 0 {
		return nil, common.NotFound
	}

	//var rulereflst = make([]*FirmwareRule, 0, len(rulelst))
	var rulereflst []*FirmwareRule

	for _, r := range rulelst {
		frule := r.(*FirmwareRule)
		rulereflst = append(rulereflst, frule)
	}

	//cm.ApplicationCacheSet(db.TABLE_FIRMWARE_RULE, cacheKey, rulereflst)

	return rulereflst, nil
}

func GetFirmwareRuleAllAsListDBForAdmin() ([]*FirmwareRule, error) {
	// pass 0 or -1 as unlimit
	rulelst, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_FIRMWARE_RULE, 0)
	if err != nil {
		return nil, err
	}

	var rulereflst []*FirmwareRule

	for _, r := range rulelst {
		frule := r.(*FirmwareRule)
		rulereflst = append(rulereflst, frule)
	}
	return rulereflst, nil
}

func GetFirmwareRulesByApplicationType(applicationType string) ([]*FirmwareRule, error) {
	cm := db.GetCacheManager()
	cacheKey := fmt.Sprintf("%s_%s", "FirmwareRuleList", applicationType)
	cacheInst := cm.ApplicationCacheGet(db.TABLE_FIRMWARE_RULE, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*FirmwareRule), nil
	}

	rulelst, err := GetFirmwareRuleAllAsListDB()
	if err != nil {
		return nil, err
	}

	if len(rulelst) == 0 {
		return nil, common.NotFound
	}

	filtereddRules := make([]*FirmwareRule, 0, len(rulelst))

	for _, rule := range rulelst {
		if rule.ApplicationType == applicationType {
			filtereddRules = append(filtereddRules, rule)
		}
	}

	cm.ApplicationCacheSet(db.TABLE_FIRMWARE_RULE, cacheKey, filtereddRules)

	return filtereddRules, nil
}

func GetEnvModelFirmwareRules(applicationType string) ([]*FirmwareRule, error) {
	cm := db.GetCacheManager()
	cacheKey := fmt.Sprintf("%s_%s", "EnvModelFirmwareRuleList", applicationType)
	cacheInst := cm.ApplicationCacheGet(db.TABLE_FIRMWARE_RULE, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*FirmwareRule), nil
	}

	rules, err := GetFirmwareRulesByApplicationType(applicationType)
	if err != nil {
		return rules, err
	}

	filtereddRules := make([]*FirmwareRule, 0, len(rules))

	for _, rule := range rules {
		if ENV_MODEL_RULE != rule.GetTemplateId() {
			continue
		}
		filtereddRules = append(filtereddRules, rule)
	}

	if len(filtereddRules) > 0 {
		cm.ApplicationCacheSet(db.TABLE_FIRMWARE_RULE, cacheKey, filtereddRules)
	}

	return filtereddRules, nil
}

func GetEnvModelFirmwareRulesForAS(applicationType string) ([]*FirmwareRule, error) {
	rules, err := GetFirmwareRuleAllAsListDBForAdmin()
	if err != nil {
		return rules, err
	}

	var filteredRules []*FirmwareRule

	for _, rule := range rules {
		if rule.ApplicationType != applicationType {
			continue
		}
		if "ENV_MODEL_RULE" != rule.GetTemplateId() {
			continue
		}
		filteredRules = append(filteredRules, rule)
	}
	return filteredRules, nil
}

// GetFirmwareSortedRuleAllAsListDB returns all FirmwareRule sorted by Name
func GetFirmwareSortedRuleAllAsListDB() ([]*FirmwareRule, error) {
	cm := db.GetCacheManager()
	cacheKey := "FirmwareSortedRuleList"
	cacheInst := cm.ApplicationCacheGet(db.TABLE_FIRMWARE_RULE, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*FirmwareRule), nil
	}

	rulelst, err := GetFirmwareRuleAllAsListDBForAdmin()
	if err != nil {
		return nil, err
	}

	// sort rules based on rule.Name
	var sortedList []*FirmwareRule
	sortedList = append(sortedList, rulelst...)

	sort.Slice(sortedList, func(i, j int) bool {
		return strings.Compare(sortedList[i].Name, sortedList[j].Name) < 0
	})

	cm.ApplicationCacheSet(db.TABLE_FIRMWARE_RULE, cacheKey, sortedList)

	return sortedList, nil
}

func GetFirmwareRuleAllAsListByApplicationType(applicationType string) (map[string][]*FirmwareRule, error) {
	rulelst, err := GetFirmwareRulesByApplicationType(applicationType)
	if err != nil {
		return nil, err
	}

	result := map[string][]*FirmwareRule{}

	// categorize rules by rule.Type
	for _, rule := range rulelst {
		list, ok := result[rule.Type]
		if !ok {
			list = make([]*FirmwareRule, 0, 100)
		}
		list = append(list, rule)
		result[rule.Type] = list
	}

	return result, nil
}

func GetFirmwareRuleAllAsListByApplicationTypeForAS(applicationType string) (map[string][]*FirmwareRule, error) {
	log.Debug("GetFirmwareRuleAllAsListByApplicationType starts...")
	// pass 0 or -1 as unlimit
	rulemap, err := db.GetCachedSimpleDao().GetAllAsMap(db.TABLE_FIRMWARE_RULE)
	if err != nil {
		return nil, err
	}

	if len(rulemap) == 0 {
		return nil, common.NotFound
	}

	result := map[string][]*FirmwareRule{}

	for _, v := range rulemap {
		rule := v.(*FirmwareRule)
		if rule.ApplicationType == applicationType {
			r, ok := result[rule.Type]
			if !ok {
				r = []*FirmwareRule{}
			}
			r = append(r, rule)
			result[rule.Type] = r
		}
	}

	log.Debug("GetFirmwareRuleAllAsListByApplicationType ends...")
	return result, nil
}

func GetFirmwareRuleTemplateAllAsListDB(actionType ApplicableActionType) ([]*FirmwareRuleTemplate, error) {
	cm := db.GetCacheManager()
	cacheKey := "FirmwareRuleTemplateList"
	cacheInst := cm.ApplicationCacheGet(db.TABLE_FIRMWARE_RULE_TEMPLATE, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*FirmwareRuleTemplate), nil
	}

	// pass 0 or -1 as unlimit
	rulelst, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_FIRMWARE_RULE_TEMPLATE, 0)
	if err != nil {
		log.Error(fmt.Sprintf("Error load all template rules %v", err))
		return nil, err
	}

	result := make([]*FirmwareRuleTemplate, 0, len(rulelst))

	for _, rule := range rulelst {
		trule := rule.(*FirmwareRuleTemplate)
		if actionType == "" || trule.ApplicableAction.ActionType == actionType {
			result = append(result, trule)
		}
	}

	if len(result) > 0 {
		cm.ApplicationCacheSet(db.TABLE_FIRMWARE_RULE_TEMPLATE, cacheKey, result)
	}

	return result, nil
}

func GetFirmwareRuleTemplateAllAsListDBForAS(actionType ApplicableActionType) ([]*FirmwareRuleTemplate, error) {
	tmprulelst, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_FIRMWARE_RULE_TEMPLATE, 0)
	if err != nil {
		log.Error(fmt.Sprintf("Error load all template rules %v", err))
		return nil, err
	}

	if tmprulelst == nil {
		log.Error("Error load all template rules == failed to load , nil result")
		return nil, err
	}

	var rulereflst []*FirmwareRuleTemplate
	for _, tr := range tmprulelst {
		tmprule := tr.(*FirmwareRuleTemplate)
		if actionType == "" || tmprule.ApplicableAction.ActionType == actionType {
			rulereflst = append(rulereflst, tmprule)
		}
	}
	return rulereflst, nil
}

func GetFirmwareRuleTemplateAllAsListByActionType(actionType ApplicableActionType) ([]*FirmwareRuleTemplate, error) {
	cm := db.GetCacheManager()
	cacheKey := fmt.Sprintf("%s_%s", "FirmwareRuleTemplateList", actionType)
	cacheInst := cm.ApplicationCacheGet(db.TABLE_FIRMWARE_RULE_TEMPLATE, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*FirmwareRuleTemplate), nil
	}

	// pass 0 or -1 as unlimit
	tmprulelst, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_FIRMWARE_RULE_TEMPLATE, 0)
	if err != nil {
		log.Error(fmt.Sprintf("Error load all template rules %v", err))
		return nil, err
	}

	if len(tmprulelst) == 0 {
		log.Error("Error load all template rules empty of result")
		return nil, err
	}

	rulereflst := make([]*FirmwareRuleTemplate, 0, len(tmprulelst))

	for _, tr := range tmprulelst {
		tmprule := tr.(*FirmwareRuleTemplate)
		if tmprule.ApplicableAction.ActionType == actionType {
			rulereflst = append(rulereflst, tmprule)
		}
	}

	if len(rulereflst) == 0 {
		return nil, common.NotFound
	}

	cm.ApplicationCacheSet(db.TABLE_FIRMWARE_RULE_TEMPLATE, cacheKey, rulereflst)

	return rulereflst, nil
}

func CreateFirmwareRuleOneDB(fr *FirmwareRule) error {
	if err := fr.Validate(); err != nil {
		return err
	}

	if util.IsBlank(fr.ID) {
		fr.ID = uuid.New().String()
	}
	fr.Updated = util.GetTimestamp()

	return db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, fr.ID, fr)
}

func DeleteOneFirmwareRule(id string) error {
	err := db.GetCachedSimpleDao().DeleteOne(db.TABLE_FIRMWARE_RULE, id)
	if err != nil {
		return err
	}
	return nil
}

func CreateFirmwareRuleTemplateOneDB(ft *FirmwareRuleTemplate) error {
	if err := ft.Validate(); err != nil {
		return err
	}
	return db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE_TEMPLATE, ft.ID, ft)
}

func ValidateRuleName(id string, name string, applicationType string) error {
	list, err := GetFirmwareRuleAllAsListDBForAdmin()
	if err != nil {
		return err
	}

	for _, rule := range list {
		if rule.ID != id && rule.Name == name && rule.ApplicationType == applicationType {
			return errors.New("Name is already used")
		}
	}

	return nil
}
