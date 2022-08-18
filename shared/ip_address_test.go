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
	"testing"

	"gotest.tools/assert"
)

func TestIpAddress(t *testing.T) {
	// test case 1
	s1 := "192.168.2.1"
	ipaddr1 := NewIpAddress(s1)
	assert.Assert(t, !ipaddr1.IsIpv6())
	assert.Assert(t, !ipaddr1.IsCidrBlock())
	t.Logf("ipaddr1.IP()=%v\n", ipaddr1.IP())
	assert.Equal(t, s1, ipaddr1.GetAddress())

	// test case 2
	s2 := "192.168.2.10/23"
	ipaddr2 := NewIpAddress(s2)
	assert.Assert(t, !ipaddr2.IsIpv6())
	assert.Assert(t, ipaddr2.IsCidrBlock())
	assert.Equal(t, "192.168.2.0/23", ipaddr2.GetAddress())

	s3 := "192.168.2.100"
	assert.Assert(t, ipaddr2.IsInRange(s3))
	x3 := NewIpAddress(s3)
	assert.Assert(t, ipaddr2.IsInRange(*x3))

	s4 := "192.168.3.50"
	assert.Assert(t, ipaddr2.IsInRange(s4))
	x4 := NewIpAddress(s4)
	assert.Assert(t, ipaddr2.IsInRange(*x4))

	s5 := "192.168.4.50"
	assert.Assert(t, !ipaddr2.IsInRange(s5))
	x5 := NewIpAddress(s5)
	assert.Assert(t, !ipaddr2.IsInRange(*x5))

	// test case 3
	s6 := "2001:558:6027:14:c071:6b22:17ae:ea06"
	ipaddr3 := NewIpAddress(s6)
	assert.Assert(t, ipaddr3.IsIpv6())
	assert.Assert(t, !ipaddr3.IsCidrBlock())
	assert.Equal(t, s6, ipaddr3.GetAddress())

	// test case 4
	s7 := "2001:558:6027:180::/57"
	ipaddr4 := NewIpAddress(s7)
	assert.Assert(t, ipaddr4.IsIpv6())
	assert.Assert(t, ipaddr4.IsCidrBlock())
	assert.Equal(t, s7, ipaddr4.GetAddress())

	s8 := "2001:558:6027:180::1234"
	assert.Assert(t, ipaddr4.IsInRange(s8))
	x8 := NewIpAddress(s8)
	assert.Assert(t, ipaddr4.IsInRange(*x8))

	s9 := "2001:558:4321:180::2345"
	assert.Assert(t, !ipaddr4.IsInRange(s9))
	x9 := NewIpAddress(s9)
	assert.Assert(t, !ipaddr4.IsInRange(*x9))
}
