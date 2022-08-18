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

func TestIpAddressGroup(t *testing.T) {
	ipaddrs := []IpAddress{
		*NewIpAddress("192.168.1.1"),
		*NewIpAddress("192.168.1.2"),
		*NewIpAddress("192.168.1.3"),
		*NewIpAddress("192.168.2.3"),
		*NewIpAddress("192.168.3.3"),
	}

	g := &IpAddressGroup{}

	s1 := "foo"
	g.Id = s1
	s2 := "bar"
	g.Name = s2
	g.IpAddresses = ipaddrs

	assert.Equal(t, s1, g.Id)
	assert.Equal(t, s2, g.Name)
	s3 := "192.168.1.2"
	assert.Assert(t, g.IsInRange(s3))
	ipaddr3 := NewIpAddress(s3)
	assert.Assert(t, g.IsInRange(*ipaddr3))

	assert.Assert(t, g.IsInRange("10.0.0.1", "192.168.3.3"))
	assert.Assert(t, !g.IsInRange("10.0.0.1", "10.0.0.2"))

	assert.Assert(t, !g.IsInRange(
		*NewIpAddress("192.168.0.3"),
		*NewIpAddress("192.168.0.4"),
		*NewIpAddress("192.168.0.5"),
	))
}
