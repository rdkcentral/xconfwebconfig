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
	"math"
	"sort"
	"strings"

	"xconfwebconfig/common"
)

type Void struct{}

type ApplicableActionType string

const (
	RULE                       ApplicableActionType = "RULE"
	DEFINE_PROPERTIES          ApplicableActionType = "DEFINE_PROPERTIES"
	BLOCKING_FILTER            ApplicableActionType = "BLOCKING_FILTER"
	RULE_TEMPLATE              ApplicableActionType = "RULE_TEMPLATE"
	DEFINE_PROPERTIES_TEMPLATE ApplicableActionType = "DEFINE_PROPERTIES_TEMPLATE"
	BLOCKING_FILTER_TEMPLATE   ApplicableActionType = "BLOCKING_FILTER_TEMPLATE"
)

func IsValidApplicableActionType(t ApplicableActionType) bool {
	return t == RULE || t == DEFINE_PROPERTIES || t == BLOCKING_FILTER ||
		t == RULE_TEMPLATE || t == DEFINE_PROPERTIES_TEMPLATE || t == BLOCKING_FILTER_TEMPLATE
}

func ApplicableActionTypeToString(actionType ApplicableActionType) string {
	switch actionType {
	case RULE:
		return "RULE"
	case DEFINE_PROPERTIES:
		return "DEFINE_PROPERTIES"
	case BLOCKING_FILTER:
		return "BLOCKING_FILTER"
	case RULE_TEMPLATE:
		return "RULE_TEMPLATE"
	case DEFINE_PROPERTIES_TEMPLATE:
		return "DEFINE_PROPERTIES_TEMPLATE"
	case BLOCKING_FILTER_TEMPLATE:
		return "BLOCKING_FILTER_TEMPLATE"
	default:
		return ""
	}
}

func (this *ApplicableActionType) CaseIgnoreEquals(that ApplicableActionType) bool {
	return strings.ToLower(string(*this)) == strings.ToLower(string(that))
}

func (this *ApplicableActionType) IsSuperSetOf(that *ApplicableActionType) bool {
	baseName := strings.ToLower(string(*this))
	givenName := strings.ToLower(string(*that))
	return strings.Contains(baseName, givenName)
}

// ApplicableAction's Type is the Java class name and has to be one of these
const (
	ApplicableActionClass               = ".ApplicableAction"
	RuleActionClass                     = ".RuleAction"
	DefinePropertiesActionClass         = ".DefinePropertiesAction"
	DefinePropertiesTemplateActionClass = ".DefinePropertiesTemplateAction"
	BlockingFilterActionClass           = ".BlockingFilterAction"
)

func isValidApplicableClass(class string) bool {
	return class == ApplicableActionClass || class == RuleActionClass ||
		class == DefinePropertiesActionClass || class == DefinePropertiesTemplateActionClass ||
		class == BlockingFilterActionClass
}

// RuleAction ...
type RuleAction struct {
	ApplicableAction
	ConfigId              string        `json:"configId"`
	ConfigEntries         []ConfigEntry `json:"configEntries"`
	Active                bool          `json:"active"`
	UseAccountPercentage  bool          `json:"useAccountPercentage"`
	FirmwareCheckRequired bool          `json:"firmwareCheckRequired"`
	RebootImmediately     bool          `json:"rebootImmediately"`
	Whitelist             string        `json:"whitelist"`
	IntermediateVersion   string        `json:"intermediateVersion"`
	FirmwareVersions      []string      `json:"firmwareVersions"`
}

func NewRuleAction() *RuleAction {
	return &RuleAction{
		Active:                true,
		UseAccountPercentage:  false,
		FirmwareCheckRequired: false,
		RebootImmediately:     false,
	}
}

type ConfigEntry struct {
	ConfigId          string  `json:"configId"`
	Percentage        float64 `json:"percentage"`
	StartPercentRange float64 `json:"startPercentRange"`
	EndPercentRange   float64 `json:"endPercentRange"`
}

func NewConfigEntry(configId string, startPercentRange float64, endPercentRange float64) *ConfigEntry {
	inst := &ConfigEntry{
		ConfigId:          configId,
		StartPercentRange: startPercentRange,
		EndPercentRange:   endPercentRange,
	}

	inst.Percentage = math.Round((endPercentRange-startPercentRange)*1000) / 1000

	return inst
}

func (c *ConfigEntry) Equals(configEntry *ConfigEntry) bool {
	return configEntry != nil &&
		c.ConfigId == configEntry.ConfigId &&
		c.Percentage == configEntry.Percentage &&
		c.StartPercentRange == configEntry.StartPercentRange &&
		c.EndPercentRange == configEntry.EndPercentRange
}

func (c *ConfigEntry) CompareTo(configEntry *ConfigEntry) int {
	if configEntry == nil || configEntry.StartPercentRange == 0 || c.StartPercentRange == 0 {
		return 0
	}
	if c.StartPercentRange > configEntry.StartPercentRange {
		return 1
	} else if c.StartPercentRange < configEntry.StartPercentRange {
		return -1
	}
	return 0
}

func (d *ApplicableAction) GetFirmwareVersions() []string {
	if values, ok := d.ActivationFirmwareVersions[common.FIRMWARE_VERSIONS]; ok && len(values) > 0 {
		return values
	}
	return []string{}
}

func (d *ApplicableAction) GetFirmwareVersionRegExs() []string {
	if values, ok := d.ActivationFirmwareVersions[common.REGULAR_EXPRESSIONS]; ok && len(values) > 0 {
		return values
	}
	return []string{}
}

// BlockingFilterAction ...
type BlockingFilterAction struct {
	ApplicableAction
}

func NewBlockingFilterAction() interface{} {
	return &BlockingFilterAction{
		ApplicableAction: ApplicableAction{Type: BlockingFilterActionClass},
	}
}

// DefinePropertiesTemplateAction ...
type DefinePropertiesTemplateAction struct {
	ApplicableAction
	Properties                 map[string]PropertyValue `json:"properties"`
	ByPassFilters              []string                 `json:"byPassFilters"`
	ActivationFirmwareVersions map[string][]string      `json:"activationFirmwareVersions"`
}

func NewDefinePropertiesTemplateAction() interface{} {
	return &DefinePropertiesTemplateAction{
		ApplicableAction: ApplicableAction{Type: DefinePropertiesTemplateActionClass},
	}
}

// DefinePropertiesAction ...
type DefinePropertiesAction struct {
	ApplicableAction
	Properties                 map[string]PropertyValue `json:"properties"`
	ByPassFilters              []string                 `json:"byPassFilters"`
	ActivationFirmwareVersions map[string][]string      `json:"activationFirmwareVersions"`
}

func NewDefinePropertiesAction() interface{} {
	return &DefinePropertiesAction{
		ApplicableAction: ApplicableAction{Type: DefinePropertiesActionClass},
	}
}

type ValidationType string

const (
	STRING  ValidationType = "STRING"
	BOOLEAN                = "BOOLEAN"
	NUMBER                 = "NUMBER"
	PERCENT                = "PERCENT"
	PORT                   = "PORT"
	URL                    = "URL"
	IPV4                   = "IPV4"
	IPV6                   = "IPV6"
)

type PropertyValue struct {
	Value           string           `json:"value"`
	Optional        bool             `json:"optional"`
	ValidationTypes []ValidationType `json:"validationTypes"`
}

func NewPropertyValue(value string, optional bool, vtype ValidationType) *PropertyValue {
	propertyValue := &PropertyValue{
		Value:           value,
		Optional:        optional,
		ValidationTypes: []ValidationType{vtype},
	}
	return propertyValue
}

func HasFirmwareVersion(firmwareVersions []string, version string) bool {
	if len(firmwareVersions) > 0 {
		for _, v := range firmwareVersions {
			if v == version {
				return true
			}
		}
	}
	return false
}

func SortConfigEntry(entries []*ConfigEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CompareTo(entries[j]) < 0
	})
}
