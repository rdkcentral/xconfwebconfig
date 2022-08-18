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
	"net"
	"strings"
)

// not getting the correct IP address. getting it from ip (net.IP) or ipNet (*net.IPNet) which are both nil, as opposed to the Address string value where there's actually data
type IpAddress struct {
	Address     string      `json:"address,omitempty" xml:"address,omitempty"`
	BaseAddress interface{} `json:"baseAddress,omitempty" xml:"baseAddress,omitempty"`
	ip          net.IP      `json:"-"`
	ipNet       *net.IPNet  `json:"-"`
	isIpv6      bool        `json:"-"`
	isCidrBlock bool        `json:"-"`
}

func NewIpAddress(input string) *IpAddress {
	var isIpv6, isCidrBlock bool
	var ip net.IP
	var ipNet *net.IPNet
	var err error
	if strings.Contains(input, "/") {
		ip, ipNet, err = net.ParseCIDR(input)
		if err != nil {
			return nil
		}
		isCidrBlock = true
	} else {
		ip = net.ParseIP(input)
		if ip == nil {
			return nil
		}
	}
	if x := ip.To4(); x == nil {
		isIpv6 = true
	}

	return &IpAddress{
		Address:     input,
		ip:          ip,
		ipNet:       ipNet,
		isIpv6:      isIpv6,
		isCidrBlock: isCidrBlock,
	}
}

// parse() use NewIpAddress

func (a IpAddress) IsIpv6() bool {
	return a.isIpv6
}

func (a IpAddress) IsCidrBlock() bool {
	return a.isCidrBlock
}

func (a IpAddress) GetAddress() string {
	if a.IsCidrBlock() {
		return a.ipNet.String()
	}
	return a.ip.String()
}

func (a IpAddress) IP() net.IP {
	return a.ip
}

func (a IpAddress) IsInRange(itf interface{}) bool {
	switch ty := itf.(type) {
	case string:
		inputIp := net.ParseIP(ty)
		if inputIp == nil {
			return false
		}

		if a.ipNet == nil {
			return a.ip.Equal(inputIp)
		}
		return a.ipNet.Contains(inputIp)
	case IpAddress:
		if a.ipNet == nil {
			return a.ip.Equal(ty.IP())
		}
		return a.ipNet.Contains(ty.IP())
	}
	return false
}

func (a IpAddress) Equals(b IpAddress) bool {
	return a.GetAddress() == b.GetAddress()
}
