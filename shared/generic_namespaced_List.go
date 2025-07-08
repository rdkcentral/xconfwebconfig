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
package shared

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"
	"xconfwebconfig/db"
	"xconfwebconfig/util"
)

// GenericNamespacedListType
type GenericNamespacedListType string

var namespacedListUpdateMutex sync.Mutex

const (
	STRING      = "STRING"
	MAC_LIST    = "MAC_LIST"
	IP_LIST     = "IP_LIST"
	RI_MAC_LIST = "RI_MAC_LIST"
)

func IsValidType(stype string) bool {
	return STRING == stype || MAC_LIST == stype || IP_LIST == stype || RI_MAC_LIST == stype
}

// NamespacedList XconfNamedList table
type NamespacedList struct {
	ID       string   `json:"id"`
	Updated  int64    `json:"updated"`
	Data     []string `json:"data"`
	TypeName string   `json:"typeName"`
}

func (obj *NamespacedList) Clone() (*NamespacedList, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*NamespacedList), nil
}

// NewGenericNamespacedListInf constructor
func NewNamespacedListInf() interface{} {
	return &NamespacedList{}
}

// GenericNamespacedList GenericXconfNamedList table
type GenericNamespacedList struct {
	ID       string   `json:"id"`
	Updated  int64    `json:"updated,omitempty"`
	Data     []string `json:"data"`
	TypeName string   `json:"typeName,omitempty"`
}

func (obj *GenericNamespacedList) Clone() (*GenericNamespacedList, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*GenericNamespacedList), nil
}

func (obj *GenericNamespacedList) Validate() error {
	matched, _ := regexp.MatchString("^[-a-zA-Z0-9_.' ]+$", obj.ID)
	if !matched {
		return errors.New("name is invalid")
	}

	if !IsValidType(obj.TypeName) {
		return fmt.Errorf("type %s is invalid", obj.TypeName)
	}
	itemsSet := util.Set{}
	itemsSet.Add(obj.Data...)
	obj.Data = itemsSet.ToSlice()

	if err := ValidateListData(obj.TypeName, obj.Data); err != nil {
		return err
	}

	if err := obj.ValidateDataIntersection(); err != nil {
		return err
	}

	return nil
}

func (obj *GenericNamespacedList) ValidateForAdminService() error {
	matched, _ := regexp.MatchString("^[-a-zA-Z0-9_.' ]+$", obj.ID)
	if !matched {
		return errors.New("name is invalid")
	}

	if !IsValidType(obj.TypeName) {
		return fmt.Errorf("type %s is invalid", obj.TypeName)
	}
	itemsSet := util.Set{}
	itemsSet.Add(obj.Data...)
	obj.Data = itemsSet.ToSlice()

	if err := ValidateListDataForAdmin(obj.TypeName, obj.Data); err != nil {
		return err
	}

	if err := obj.ValidateDataIntersection(); err != nil {
		return err
	}

	return nil
}

func LockGenericNamespacedList() {
	namespacedListUpdateMutex.Lock()
}

func UnlockGenericNamespacedList() {
	namespacedListUpdateMutex.Unlock()
}

func ValidateListData(typeName string, listData []string) error {
	if !IsValidType(typeName) {
		return errors.New("Type is invalid")
	}

	if len(listData) == 0 {
		return errors.New("List must not be empty")
	}

	var invalidAddresses []string
	if typeName == IP_LIST {
		for _, ipAddress := range listData {
			if net.ParseIP(ipAddress) == nil {
				invalidAddresses = append(invalidAddresses, ipAddress)
			}
		}
	} else if typeName == MAC_LIST {
		for _, mac := range listData {
			if !util.IsValidMacAddress(mac) {
				invalidAddresses = append(invalidAddresses, mac)
			}
		}
	}
	if len(invalidAddresses) > 0 {
		return fmt.Errorf("List contains invalid address(es): %v", invalidAddresses)
	}

	return nil
}

func ValidateListDataForAdmin(typeName string, listData []string) error {
	if !IsValidType(typeName) {
		return errors.New("Type is invalid")
	}

	if len(listData) == 0 {
		return errors.New("List must not be empty")
	}

	var invalidAddresses []string
	if typeName == IP_LIST {
		for _, ipAddress := range listData {
			if NewIpAddress(ipAddress) == nil {
				invalidAddresses = append(invalidAddresses, ipAddress)
			}
		}
	} else if typeName == MAC_LIST {
		for _, mac := range listData {
			if !util.IsValidMacAddress(mac) {
				invalidAddresses = append(invalidAddresses, mac)
			}
		}
	}
	if len(invalidAddresses) > 0 {
		return fmt.Errorf("List contains invalid address(es): %v", invalidAddresses)
	}

	return nil
}

// NewGenericNamespacedListInf constructor
func NewGenericNamespacedListInf() interface{} {
	return &GenericNamespacedList{}
}

// TODO Updated is NOT included in the constructor. EVAL if it is ok
func NewGenericNamespacedList(id string, typeName string, data []string) *GenericNamespacedList {
	return &GenericNamespacedList{
		ID:       id,
		TypeName: typeName,
		Data:     data,
	}
}

func NewEmptyGenericNamespacedList() *GenericNamespacedList {
	return &GenericNamespacedList{}
}

func NewMacList() *GenericNamespacedList {
	return &GenericNamespacedList{
		TypeName: MAC_LIST,
	}
}

func NewIpList() *GenericNamespacedList {
	return &GenericNamespacedList{
		TypeName: IP_LIST,
	}
}

func (g *GenericNamespacedList) IsMacList() bool {
	if MAC_LIST == g.TypeName {
		return true
	}
	return false
}

func (g *GenericNamespacedList) IsIpList() bool {
	if IP_LIST == g.TypeName {
		return true
	}
	return false
}

func (obj *GenericNamespacedList) ValidateDataIntersection() error {
	if obj.TypeName == MAC_LIST {
		itemsSet := util.Set{}
		itemsSet.Add(obj.Data...)

		intersectionMap := make(map[string][]string)

		namespacedLists, err := GetGenericNamedListListsByTypeDB(obj.TypeName)
		if err != nil {
			return err
		}

		for _, nsList := range namespacedLists {
			if obj.ID == nsList.ID {
				continue
			}
			intersection := make([]string, 0)
			for _, mac := range nsList.Data {
				if itemsSet.Contains(mac) {
					intersection = append(intersection, mac)
				}
			}
			if len(intersection) > 0 {
				intersectionMap[nsList.ID] = intersection
			}
		}

		if len(intersectionMap) > 0 {
			addSeparator := false
			buffer := bytes.NewBufferString("MAC addresses are already used in other lists: ")
			for key, value := range intersectionMap {
				if addSeparator {
					buffer.WriteString(", ")
				} else {
					addSeparator = true
				}
				buffer.WriteString(fmt.Sprintf("[%s] in %s", strings.Join(value, ", "), key))
			}
			return errors.New(buffer.String())
		}
	}

	return nil
}

func GetGenericNamedListSetByType(typeName string) (*util.Set, error) {
	if !IsValidType(typeName) {
		return nil, fmt.Errorf("Invalid GenericNamespacedList typeName %s", typeName)
	}
	cm := db.GetCacheManager()
	cacheKey := typeName
	cacheInst := cm.ApplicationCacheGet(db.TABLE_GENERIC_NS_LIST, cacheKey)
	if cacheInst != nil {
		return cacheInst.(*util.Set), nil
	}
	result := util.NewSet()
	entry, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_GENERIC_NS_LIST, 0)
	if err != nil {
		return nil, err
	}
	for _, obj := range entry {
		nl := obj.(*GenericNamespacedList)
		if nl.TypeName == typeName {
			result.Add(nl.Data...)
		}
	}
	cm.ApplicationCacheSet(db.TABLE_GENERIC_NS_LIST, cacheKey, &result)
	return &result, nil
}

func GetGenericNamedListOneDB(id string) (*GenericNamespacedList, error) {
	instlst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_GENERIC_NS_LIST, id)
	if err != nil {
		return nil, err
	}

	if instlst == nil {
		return nil, nil
	}

	lstptr := instlst.(*GenericNamespacedList)
	return lstptr, nil
}

func GetGenericNamedListListsDB() ([]*GenericNamespacedList, error) {
	list, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_GENERIC_NS_LIST, 0)
	if err != nil {
		return nil, err
	}

	i := 0
	glist := make([]*GenericNamespacedList, len(list))
	for _, obj := range list {
		nl := obj.(*GenericNamespacedList)
		glist[i] = nl
		i++
	}

	return glist, nil
}

func GetGenericNamedListListsByTypeDB(typeName string) ([]*GenericNamespacedList, error) {
	if !IsValidType(typeName) {
		return nil, fmt.Errorf("Invalid GenericNamespacedList typeName %s", typeName)
	}

	result := []*GenericNamespacedList{}
	list, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_GENERIC_NS_LIST, 0)
	if err != nil {
		return nil, err
	}
	for _, obj := range list {
		nl := obj.(*GenericNamespacedList)
		if nl.TypeName == typeName {
			result = append(result, nl)
		}
	}
	return result, nil
}

func CreateGenericNamedListOneDB(newList *GenericNamespacedList) error {
	newList.Updated = util.GetTimestamp()
	err := db.GetCachedSimpleDao().SetOne(db.TABLE_GENERIC_NS_LIST, newList.ID, newList)
	return err
}

func (g *GenericNamespacedList) String() string {
	return fmt.Sprintf("GenericNamespacedList(%v |%v| %v)", g.ID, g.TypeName, g.Data)
}

func GetGenericNamedListOneByType(id string, typeName string) (*GenericNamespacedList, error) {
	lst, err := GetGenericNamedListOneDB(id)
	if err != nil {
		return nil, err
	}

	if lst.TypeName == typeName {
		return lst, nil
	}

	return nil, nil
}

func GetGenericNamedListOneByTypeNonCached(id string, typeName string) (*GenericNamespacedList, error) {
	instlst, err := db.GetCompressingDataDao().GetOne(db.TABLE_GENERIC_NS_LIST, id)
	if err != nil {
		return nil, err
	}

	if instlst == nil {
		return nil, nil
	}

	lstptr := instlst.(*GenericNamespacedList)
	if typeName != "" && lstptr.TypeName != typeName {
		return nil, nil
	}

	return lstptr, nil
}

func DeleteOneGenericNamedList(id string) error {
	err := db.GetCachedSimpleDao().DeleteOne(db.TABLE_GENERIC_NS_LIST, id)
	if err != nil {
		return err
	}
	return nil
}

func (g *GenericNamespacedList) IsInIpRange(ipAddressStr string) bool {
	ipAddressGroup := NewIpAddressGroupWithAddrStrings("foo", "bar", g.Data)
	return ipAddressGroup.IsInRange(ipAddressStr)
}

func (g *GenericNamespacedList) CreateIpAddressGroupResponse() *IpAddressGroup {
	return &IpAddressGroup{
		Id:             g.ID,
		Name:           g.ID,
		RawIpAddresses: g.Data,
	}
}

func (g *GenericNamespacedList) CreateGenericNamespacedListResponse() *GenericNamespacedList {
	return &GenericNamespacedList{
		ID:   g.ID,
		Data: g.Data,
	}
}

func ConvertToIpAddressGroup(genericIpList *GenericNamespacedList) *IpAddressGroup {
	return NewIpAddressGroupWithAddrStrings(genericIpList.ID, genericIpList.ID, genericIpList.Data)
}

func ConvertFromIpAddressGroup(ipAddressGroup *IpAddressGroup) *GenericNamespacedList {
	ipList := NewIpList()
	ipList.ID = ipAddressGroup.Name
	ipList.Data = ipAddressGroup.RawIpAddresses
	return ipList
}
