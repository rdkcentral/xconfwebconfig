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
	"math/rand"
	"strings"

	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

// DownloadLocationRoundRobinFilterValue a subtype in SingletonFilterValue table
type DownloadLocationRoundRobinFilterValue struct {
	ID                  string               `json:"id" xml:"id"`
	Updated             int64                `json:"updated" xml:"updated"`
	Type                SingletonFilterClass `json:"type,omitempty" xml:"type"`
	ApplicationType     string               `json:"applicationType" xml:"applicationType"`
	Locations           []Location           `json:"locations" xml:"locations"`
	Ipv6locations       []Location           `json:"ipv6locations" xml:"ipv6locations"`
	HttpLocation        string               `json:"httpLocation" xml:"httpLocation"` // regexp = "(?=^.{1,254}$)(^(?:(?!\\d+\\.|-)[a-zA-Z0-9_\\-]{1,63}(?<!-)\\.)+(?:[a-zA-Z]{2,})$)"
	HttpFullUrlLocation string               `json:"httpFullUrlLocation" xml:"httpFullUrlLocation"`
}

type DownloadLocationFilter struct {
	IpAddressGroup           *shared.IpAddressGroup `json:"ipAddressGroup" xml:"ipAddressGroup"`
	Environments             []string               `json:"environments" xml:"environments"`
	Models                   []string               `json:"models" xml:"models"`
	FirmwareDownloadProtocol string                 `json:"firmwareDownloadProtocol,omitempty" xml:"firmwareDownloadProtocol,omitempty"`
	// tftp location
	FirmwareLocation     *shared.IpAddress `json:"ipv4FirmwareLocation,omitempty" xml:"ipv4FirmwareLocation,omitempty"`
	Ipv6FirmwareLocation *shared.IpAddress `json:"ipv6FirmwareLocation,omitempty" xml:"ipv6FirmwareLocation,omitempty"`
	HttpLocation         string            `json:"httpLocation,omitempty" xml:"httpLocation,omitempty"`
	ForceHttp            bool              `json:"forceHttp" xml:"forceHttp"`
	Id                   string            `json:"id" xml:"id"`
	Name                 string            `json:"name" xml:"name"`
	BoundConfigId        string            `json:"boundConfigId,omitempty" xml:"boundConfigId,omitempty"`
}

type Location struct {
	LocationIp string  `json:"locationIp" xml:"locationIp"`
	Percentage float64 `json:"percentage" xml:"percentage"`
}

func NewLocation(ip string, perc float64) *Location {
	return &Location{
		LocationIp: ip,
		Percentage: perc,
	}
}

func NewDownloadLocationRoundRobinFilterValue() interface{} {
	return &DownloadLocationRoundRobinFilterValue{
		ID:              ROUND_ROBIN_FILTER_SINGLETON_ID,
		Type:            RoundRobinFilterClass,
		ApplicationType: shared.STB,
	}
}

func NewEmptyDownloadLocationRoundRobinFilterValue() *DownloadLocationRoundRobinFilterValue {
	return &DownloadLocationRoundRobinFilterValue{
		ID:              ROUND_ROBIN_FILTER_SINGLETON_ID,
		Type:            RoundRobinFilterClass,
		ApplicationType: shared.STB,
		Locations:       []Location{},
		Ipv6locations:   []Location{},
	}
}

func (obj *DownloadLocationRoundRobinFilterValue) SetApplicationType(appType string) {
	obj.ApplicationType = appType
}

func (obj *DownloadLocationRoundRobinFilterValue) GetApplicationType() string {
	return obj.ApplicationType
}

func (obj *DownloadLocationRoundRobinFilterValue) Validate() error {
	if obj.Type != RoundRobinFilterClass {
		return errors.New("Type is invalid")
	}

	if !strings.HasSuffix(obj.ID, ROUND_ROBIN_FILTER_SINGLETON_ID) {
		return errors.New("Id is invalid")
	}

	if !logupload.IsValidUrl(obj.HttpFullUrlLocation) {
		return errors.New("Location URL is not valid")
	}

	ipv6InIpv4List := false
	percentage := 0.0
	ipSet := util.Set{}
	for _, location := range obj.Locations {
		ip := shared.NewIpAddress(location.LocationIp)
		if !ipv6InIpv4List && ip != nil && ip.IsIpv6() {
			ipv6InIpv4List = true
		}
		if location.Percentage < 0 {
			return errors.New("Percentage cannot be negative")
		}
		percentage += location.Percentage
		ipSet.Add(location.LocationIp)
	}

	ipv4InIpv6List := false
	ipv6Percentage := 0.0
	ipv6Set := util.Set{}
	for _, location := range obj.Ipv6locations {
		ip := shared.NewIpAddress(location.LocationIp)
		if !ipv4InIpv6List && ip != nil && !ip.IsIpv6() {
			ipv4InIpv6List = true
		}
		if location.Percentage < 0 {
			return errors.New("Percentage cannot be negative")
		}
		ipv6Percentage += location.Percentage
		ipv6Set.Add(location.LocationIp)
	}

	if ipv4InIpv6List || ipv6InIpv4List {
		return errors.New("IP address has an invalid version")
	}

	if len(ipSet) != len(obj.Locations) || len(ipv6Set) != len(obj.Ipv6locations) {
		return errors.New("Locations are duplicated")
	}

	if int(percentage) != 100 {
		return errors.New("Summary IPv4 percentage should be 100")
	}

	if int(ipv6Percentage) != 0 && int(ipv6Percentage) != 100 {
		return errors.New("Summary IPv6 percentage should be 100")
	}

	return nil
}

func (d *DownloadLocationRoundRobinFilterValue) GetDownloadLocations() []string {
	result := make([]string, 2)
	random := rand.Float64()

	if d.Ipv6locations != nil && len(d.Ipv6locations) > 0 {
		scale := 0.0
		for _, location := range d.Ipv6locations {
			scale += location.Percentage / 100.00
			if random < scale {
				result[1] = location.LocationIp
				break
			}
		}
	}

	if d.Locations != nil && len(d.Locations) > 0 {
		scale := 0.0
		for _, location := range d.Locations {
			scale += location.Percentage / 100.00
			if random < scale {
				result[0] = location.LocationIp
				break
			}
		}
	}
	return result
}

func GetDownloadLocationRoundRobinFilterValOneDB(filterId string) (*DownloadLocationRoundRobinFilterValue, error) {
	log.Debug("GetDownloadLocationRoundRobinFilterValOneDB starts...")
	inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_SINGLETON_FILTER_VALUE, filterId)
	if err != nil {
		return nil, err
	}

	switch v := inst.(type) {
	case *DownloadLocationRoundRobinFilterValue:
		rrf := inst.(*DownloadLocationRoundRobinFilterValue)
		return rrf, nil
	default:
		log.Debugf("DownloadLocationRoundRobinFilterValue not DownloadLocationRoundRobinFilterValue %v", v)
		sf := inst.(*SingletonFilterValue)
		if sf.IsDownloadLocationRoundRobinFilterValue() {
			rrf := sf.DownloadLocationRoundRobinFilterValue
			return rrf, nil
		}
	}
	log.Debug("GetDownloadLocationRoundRobinFilterValOneDB ends...")
	return nil, fmt.Errorf("DownloadLocationRoundRobinFilterValue not found for %v", filterId)
}

func GetDefaultDownloadLocationRoundRobinFilterValOneDB() (*DownloadLocationRoundRobinFilterValue, error) {
	return GetDownloadLocationRoundRobinFilterValOneDB(ROUND_ROBIN_FILTER_SINGLETON_ID)
}

func CreateDownloadLocationRoundRobinFilterValOneDB(dl *DownloadLocationRoundRobinFilterValue) error {

	dl.Updated = util.GetTimestamp()

	/*
		DownloadLocationRoundRobinFilterValue is a subtype of SingletonFilterValue and it is the object
		type that is in the cache; therefore we need to wrap it in a SingletonFilterValue object.
	*/
	sfv := &SingletonFilterValue{
		ID:                                    dl.ID,
		DownloadLocationRoundRobinFilterValue: dl,
	}

	// create record in DB
	return db.GetCachedSimpleDao().SetOne(db.TABLE_SINGLETON_FILTER_VALUE, sfv.ID, sfv)
}

func NewEmptyDownloadLocationFilter() *DownloadLocationFilter {
	return &DownloadLocationFilter{}
}

func DownloadLocationFiltersByApplicationType(applicationType string) ([]*DownloadLocationFilter, error) {
	rulelst, err := firmware.GetFirmwareRuleAllAsListDB()
	if err != nil {
		return nil, err
	}

	filtedRules := make([]*DownloadLocationFilter, 0)
	for _, frule := range rulelst {
		if frule.ApplicationType != applicationType {
			continue
		}
		if frule.GetTemplateId() != DOWNLOAD_LOCATION_FILTER {
			continue
		}
		filtedRules = append(filtedRules, setDownloadLocationFilter(frule))
	}
	return filtedRules, nil
}

func DownloadLocationFiltersByName(applicationType string, name string) (*DownloadLocationFilter, error) {
	rulelst, err := firmware.GetFirmwareRuleAllAsListDB()
	if err != nil {
		return nil, err
	}

	for _, frule := range rulelst {
		if frule.ApplicationType != applicationType {
			continue
		}
		if frule.GetTemplateId() != DOWNLOAD_LOCATION_FILTER {
			continue
		}
		if frule.Name == name {
			return setDownloadLocationFilter(frule), nil
		}
	}
	return nil, nil
}

func setDownloadLocationFilter(rule *firmware.FirmwareRule) *DownloadLocationFilter {
	dlf := &DownloadLocationFilter{
		Id:   rule.ID,
		Name: rule.Name,
	}
	if rule.Rule.Condition != nil {
		dlf.IpAddressGroup = GetIpAddressGroup(rule.Rule.Condition)
	}
	// } else {
	// 	listId := getListRef(rule)
	// 	log.Infof("===========>>>>>>>>>>>> %s", listId)
	// 	list, _ := shared.GetGenericNamedListOneDB(listId)
	// 	if list != nil {
	// 		dlf.IpAddressGroup = covt.ConvertToIpAddressGroup(list)
	// 	}
	// }

	protocal := rule.ApplicableAction.Properties["firmwareDownloadProtocol"]
	httplocation := rule.ApplicableAction.Properties["firmwareLocation"]
	ipv6location := rule.ApplicableAction.Properties["ipv6FirmwareLocation"]

	httplocation = strings.ReplaceAll(httplocation, "\"", "")

	if protocal == "tftp" {
		dlf.ForceHttp = false
		dlf.FirmwareLocation = &shared.IpAddress{Address: httplocation}
		if ipv6location != "" {
			dlf.Ipv6FirmwareLocation = &shared.IpAddress{Address: ipv6location}
		}
	} else {
		dlf.HttpLocation = httplocation
		dlf.ForceHttp = true
	}
	return dlf
}

func getListRef(rule *firmware.FirmwareRule) string {
	var rulesearch []rulesengine.Rule
	rulesearch = append(rulesearch, rule.Rule)

	for len(rulesearch) != 0 {
		r := rulesearch[0]
		rulesearch = rulesearch[1:]
		listRef := findListRef(r.Condition)
		if listRef != "" {
			return listRef
		}
		if len(r.GetCompoundParts()) != 0 {
			rulesearch = append(rulesearch, r.GetCompoundParts()...)
		}
	}
	return ""
}

func findListRef(cond *rulesengine.Condition) string {
	if cond == nil || cond.GetFixedArg() == nil {
		return ""
	}
	if IsLegacyIpCondition(*cond) {
		return cond.GetFixedArg().String()
	} else if (RuleFactoryIN_LIST == cond.GetOperation()) && RuleFactoryIP.Name == cond.GetFreeArg().Name {
		return cond.GetFixedArg().String()
	}
	return ""
}
