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
	"encoding/json"
	_ "net"
	_ "strings"

	util "xconfwebconfig/util"
)

// IpAddressGroup IpAddressGroupExtended table
type IpAddressGroup struct {
	Id             string      `json:"id" xml:"id"`
	Name           string      `json:"name" xml:"name"`
	IpAddresses    []IpAddress `json:"-" xml:"-"`
	RawIpAddresses []string    `json:"ipAddresses" xml:"ipAddresses"` // custom unmarshal for IpAddresses
}

func (obj *IpAddressGroup) Clone() (*IpAddressGroup, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*IpAddressGroup), nil
}

// NewIpAddressGroupInf constructor
func NewIpAddressGroupInf() interface{} {
	return &IpAddressGroup{}
}

func NewEmptyIpAddressGroup() *IpAddressGroup {
	return &IpAddressGroup{}
}

func NewIpAddressGroupWithAddrStrings(sid string, sname string, addrs []string) *IpAddressGroup {
	res := &IpAddressGroup{
		Id:   sid,
		Name: sname,
	}

	res.SetIpAddresses(addrs)
	return res
}

func NewIpAddressGroup(input *IpAddressGroup) *IpAddressGroup {
	iaddrs := []IpAddress{}
	iaddrs = append(iaddrs, input.IpAddresses...)

	return &IpAddressGroup{
		Id:          input.Id,
		Name:        input.Name,
		IpAddresses: iaddrs,
	}
}

func (g *IpAddressGroup) UnmarshalJSON(bytes []byte) error {
	type ipAddressGroup IpAddressGroup

	err := json.Unmarshal(bytes, (*ipAddressGroup)(g))
	if err != nil {
		return err
	}

	g.SetIpAddresses(g.RawIpAddresses)
	return nil
}

func (g *IpAddressGroup) SetIpAddresses(addrs []string) {
	g.IpAddresses = []IpAddress{}
	g.RawIpAddresses = []string{}
	for _, addr := range addrs {
		g.RawIpAddresses = append(g.RawIpAddresses, addr)
		ipaddr := NewIpAddress(addr)
		if ipaddr != nil {
			g.IpAddresses = append(g.IpAddresses, *ipaddr)
		}
	}
}

func (g *IpAddressGroup) IsInRange(itfs ...interface{}) bool {
	for _, itf := range itfs {
		for _, ipAddress := range g.IpAddresses {
			if ipAddress.IsInRange(itf) {
				return true
			}
		}
	}
	return false
}
