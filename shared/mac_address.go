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

	"github.com/rdkcentral/xconfwebconfig/util"
)

type MacAddress struct {
	hwaddr net.HardwareAddr
}

func NewMacAddress(s string) (*MacAddress, error) {
	mstr := s
	if len(s) == 12 && !strings.Contains(s, ":") {
		mstr = util.ToColonMac(s)
	}

	hwaddr, err := net.ParseMAC(mstr)
	if err != nil {
		return nil, err
	}

	return &MacAddress{
		hwaddr: hwaddr,
	}, nil
}

func (a *MacAddress) String() string {
	return strings.ToUpper(a.hwaddr.String())
}
