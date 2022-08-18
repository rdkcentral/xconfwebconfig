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
 */ /**
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
	"math"

	"xconfwebconfig/db"
	re "xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	"xconfwebconfig/shared/firmware"
	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

// PercentFilterValue a subtype in SingletonFilterValue table
type PercentFilterValue struct {
	ID                  string                        `json:"id"`
	Updated             int64                         `json:"updated"`
	Type                SingletonFilterClass          `json:"type"`
	Whitelist           *shared.IpAddressGroup        `json:"whitelist"`
	Percentage          float32                       `json:"percentage"`
	Percent             int                           `json:"percent"`
	EnvModelPercentages map[string]EnvModelPercentage `json:"envModelPercentages"`
}

type EnvModelPercentage struct {
	Percentage            float32                `json:"percentage"`
	Active                bool                   `json:"active"`
	FirmwareCheckRequired bool                   `json:"firmwareCheckRequired"`
	RebootImmediately     bool                   `json:"rebootImmediately"`
	LastKnownGood         string                 `json:"lastKnownGood"`
	IntermediateVersion   string                 `json:"intermediateVersion"`
	Whitelist             *shared.IpAddressGroup `json:"whitelist"`
	FirmwareVersions      []string               `json:"firmwareVersions"`
}

func NewEnvModelPercentage() *EnvModelPercentage {
	return &EnvModelPercentage{
		RebootImmediately: false,
	}
}

func NewEmptyPercentFilterValue() *PercentFilterValue {
	return &PercentFilterValue{}
}

func NewPercentFilterValue(whiteList *shared.IpAddressGroup, percentage float32, envModelPercentages map[string]EnvModelPercentage) *PercentFilterValue {
	return &PercentFilterValue{
		ID:                  PERCENT_FILTER_SINGLETON_ID,
		Type:                PercentFilterClass,
		Whitelist:           whiteList,
		Percentage:          percentage,
		EnvModelPercentages: envModelPercentages,
	}
}

func (p *PercentFilterValue) SetId(id string) error {
	if PERCENT_FILTER_SINGLETON_ID != id {
		return errors.New("PercentFilterValue id is PERCENT_FILTER_VALUE")
	}
	p.ID = id
	return nil
}

func (p *PercentFilterValue) GetId() string {
	return p.ID
}

type GlobalPercentage struct {
	Whitelist       string
	Percentage      float32
	ApplicationType string
}

func NewGlobalPercentage() *GlobalPercentage {
	return &GlobalPercentage{
		Percentage:      100.0,
		ApplicationType: shared.STB,
	}
}

type PercentFilterVo struct {
	GlobalPercentage GlobalPercentage
	PercentageBeans  []PercentageBean
}

func NewDefaultPercentFilterVo() *PercentFilterVo {
	return &PercentFilterVo{}
}

func NewPercentFilterVo(globalPercentage GlobalPercentage, percentageBeans []PercentageBean) *PercentFilterVo {
	return &PercentFilterVo{
		GlobalPercentage: globalPercentage,
		PercentageBeans:  percentageBeans,
	}
}

type PercentageBean struct {
	ID                     string                  `json:"id"`
	Name                   string                  `json:"name,omitempty"`
	Whitelist              string                  `json:"whitelist,omitempty"`
	Active                 bool                    `json:"active"`
	FirmwareCheckRequired  bool                    `json:"firmwareCheckRequired"`
	RebootImmediately      bool                    `json:"rebootImmediately"`
	LastKnownGood          string                  `json:"lastKnownGood,omitempty"`
	IntermediateVersion    string                  `json:"intermediateVersion,omitempty"`
	FirmwareVersions       []string                `json:"firmwareVersions"`
	Distributions          []*firmware.ConfigEntry `json:"distributions"`
	ApplicationType        string                  `json:"applicationType,omitempty"`
	Environment            string                  `json:"environment,omitempty"`
	Model                  string                  `json:"model,omitempty"`
	OptionalConditions     *re.Rule                `json:"optionalConditions,omitempty"`
	UseAccountIdPercentage bool                    `json:"useAccountIdPercentage"`
}

func NewPercentageBean() *PercentageBean {
	return &PercentageBean{
		ApplicationType: shared.STB,
	}
}

func (p *PercentageBean) Validate() error {
	if util.IsBlank(p.Name) {
		return errors.New("Name could not be blank")
	}
	if util.IsBlank(p.Model) {
		return errors.New("Model could not be blank")
	}
	if p.OptionalConditions != nil {
		conditions := re.ToConditions(p.OptionalConditions)
		for _, condition := range conditions {
			if RuleFactoryENV.Equals(condition.FreeArg) || RuleFactoryMODEL.Equals(condition.FreeArg) {
				return fmt.Errorf("Optional condition should not contain %s", condition.FreeArg.Name)
			}
		}
	}
	if err := shared.ValidateApplicationType(p.ApplicationType); err != nil {
		return err
	}
	if p.FirmwareCheckRequired && len(p.FirmwareVersions) == 0 {
		return errors.New("Please select at least one version or disable firmware check")
	}
	if err := validateDistributionDuplicates(p.Distributions); err != nil {
		return err
	}

	var totalPercentage float64 = 0
	for _, entry := range p.Distributions {
		if entry != nil {
			if err := validatePercentageRange(entry.Percentage, "Percentage"); err != nil {
				return err
			}
			if err := validatePercentageRange(entry.StartPercentRange, "StartPercentRange"); err != nil {
				return err
			}
			if err := validatePercentageRange(entry.EndPercentRange, "EndPercentRange"); err != nil {
				return err
			}
			if err := validateDistributionOverlapping(entry, p.Distributions); err != nil {
				return err
			}
			if entry.StartPercentRange > 0 && entry.EndPercentRange > 0 && entry.StartPercentRange >= entry.EndPercentRange {
				return errors.New("StartPercentRange should be less than EndPercentRange")
			}

			config, err := GetFirmwareConfigOneDB(entry.ConfigId)
			if err != nil {
				return fmt.Errorf("FirmwareConfig with id %s does not exist", entry.ConfigId)
			}
			if p.FirmwareCheckRequired && !util.Contains(p.FirmwareVersions, config.FirmwareVersion) {
				return errors.New("Distribution version should be selected in MinCheck list")
			}
			if p.ApplicationType != config.ApplicationType {
				return errors.New("ApplicationTypes of FirmwareConfig and PercentageBean do not match")
			}

			totalPercentage += entry.Percentage
		}
	}
	if totalPercentage > 100 {
		return errors.New("Distribution total percentage > 100")
	}

	if !util.IsBlank(p.LastKnownGood) {
		lkgConfig, err := GetFirmwareConfigOneDB(p.LastKnownGood)
		if err != nil {
			return fmt.Errorf("LastKnownGood: config with id %s does not exist", p.LastKnownGood)
		}
		if p.ApplicationType != lkgConfig.ApplicationType {
			return errors.New("ApplicationTypes of FirmwareConfig and PercentageBean do not match")
		}
		if !util.Contains(p.FirmwareVersions, lkgConfig.FirmwareVersion) {
			return errors.New("LastKnownGood should be selected in min check list")
		}
		if math.Abs(totalPercentage-100.0) < 1.0e-8 {
			return errors.New("Can't set LastKnownGood when percentage=100")
		}
	}

	if p.Active && len(p.Distributions) > 0 && totalPercentage < 100 && util.IsBlank(p.LastKnownGood) {
		return errors.New("LastKnownGood is required when percentage < 100")
	}

	if !util.IsBlank(p.IntermediateVersion) {
		if !p.FirmwareCheckRequired {
			return errors.New("Can't set IntermediateVersion when firmware check is disabled")
		}
		intermediateConfig, err := GetFirmwareConfigOneDB(p.IntermediateVersion)
		if err != nil {
			return fmt.Errorf("IntermediateVersion: config with id %s does not exist", p.LastKnownGood)
		}
		if p.ApplicationType != intermediateConfig.ApplicationType {
			return errors.New("ApplicationTypes of FirmwareConfig and PercentageBean do not match")
		}
	}

	return nil
}

func (p *PercentageBean) GetTemplateId() string {
	return "ENV_MODEL_RULE"
}

func (p *PercentageBean) GetRuleType() string {
	return "PercentFilter"
}

//  GetPercentFilterValueOneDB ...  Java getRaw() in dataapi
func GetDefaultPercentFilterValueOneDB() (*PercentFilterValue, error) {
	dbinst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_SINGLETON_FILTER_VALUE, PERCENT_FILTER_SINGLETON_ID)

	if err != nil {
		log.Error(fmt.Sprintf("GetDefaultPercentFilterValueOneDB %v", err))
		return nil, err
	}
	filter := dbinst.(PercentFilterValue)
	return &filter, nil
}

func validateDistributionDuplicates(configEntries []*firmware.ConfigEntry) error {
	newDistributions := make(map[firmware.ConfigEntry]bool)
	for _, configEntry := range configEntries {
		if configEntry != nil {
			if newDistributions[*configEntry] {
				errors.New("Distributions contain duplicates")
			} else {
				newDistributions[*configEntry] = true
			}
		}
	}

	return nil
}

func validatePercentageRange(value float64, name string) error {
	if value < 0 {
		fmt.Errorf("%s filter contains negative value", name)
	}
	if value > 100 {
		fmt.Errorf("%s should be within [0, 100]", name)
	}

	return nil
}

func validateDistributionOverlapping(distributionToCheck *firmware.ConfigEntry, configEntries []*firmware.ConfigEntry) error {
	if distributionToCheck == nil {
		return nil
	}
	for _, distribution := range configEntries {
		if distribution != nil && !distribution.Equals(distributionToCheck) &&
			distribution.StartPercentRange > 0 && distribution.EndPercentRange > 0 &&
			distributionToCheck.StartPercentRange > 0 && distributionToCheck.EndPercentRange > 0 &&
			distributionToCheck.StartPercentRange <= distribution.StartPercentRange &&
			distribution.StartPercentRange < distributionToCheck.EndPercentRange {
			return errors.New("Distributions overlap each other")
		}
	}

	return nil
}
