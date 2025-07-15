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
	"unicode"

	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
)

type IpFilter struct {
	Id             string                 `json:"id" xml:"id"`
	Name           string                 `json:"name" xml:"name"`
	IpAddressGroup *shared.IpAddressGroup `json:"ipAddressGroup" xml:"ipAddressGroup"`
	Warehouse      bool                   `json:"warehouse" xml:"warehouse"`
}

func NewEmptyIpFilter() *IpFilter {
	return &IpFilter{}
}

/**
 * Quick and dirty way to tell if this filter is tied to a warehouse or not.
 * If it is we don't want to allow editing/deleting
 */
func (i *IpFilter) IsWarehouse() bool {
	return IsLetter(i.Id) && IsLower(i.Id)
}

func IsLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func IsLower(s string) bool {
	for _, r := range s {
		if !unicode.IsLower(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func IpFiltersByApplicationType(applicationType string) ([]*IpFilter, error) {
	rulelst, err := firmware.GetFirmwareRuleAllAsListDB()
	if err != nil {
		return nil, err
	}

	filtedRules := make([]*IpFilter, 0)
	for _, frule := range rulelst {
		if frule.ApplicationType != applicationType {
			continue
		}
		if frule.GetTemplateId() != IP_FILTER {
			continue
		}
		fr := &IpFilter{
			Id:             frule.ID,
			Name:           frule.Name,
			IpAddressGroup: GetIpAddressGroup(frule.Rule.Condition),
		}
		fr.Warehouse = fr.IsWarehouse()
		filtedRules = append(filtedRules, fr)
	}

	return filtedRules, nil
}

func IpFilterByName(name string, applicationType string) (*IpFilter, error) {
	rulelst, err := firmware.GetFirmwareRuleAllAsListDB()
	if err != nil {
		return nil, err
	}

	for _, frule := range rulelst {
		if frule.ApplicationType != applicationType {
			continue
		}
		if frule.GetTemplateId() != IP_FILTER {
			continue
		}
		if frule.Name == name {
			fr := &IpFilter{
				Id:             frule.ID,
				Name:           frule.Name,
				IpAddressGroup: GetIpAddressGroup(frule.Rule.Condition),
			}
			fr.Warehouse = fr.IsWarehouse()
			return fr, nil
		}
	}
	return nil, nil
}
