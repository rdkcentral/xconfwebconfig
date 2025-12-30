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

// Test IsValidIpAddress
func TestIsValidIpAddress_ValidIPv4(t *testing.T) {
	assert.Assert(t, IsValidIpAddress("192.168.1.1"))
	assert.Assert(t, IsValidIpAddress("10.0.0.1"))
	assert.Assert(t, IsValidIpAddress("172.16.0.1"))
	assert.Assert(t, IsValidIpAddress("255.255.255.255"))
	assert.Assert(t, IsValidIpAddress("0.0.0.0"))
}

func TestIsValidIpAddress_ValidIPv4CIDR(t *testing.T) {
	assert.Assert(t, IsValidIpAddress("192.168.1.0/24"))
	assert.Assert(t, IsValidIpAddress("10.0.0.0/8"))
	assert.Assert(t, IsValidIpAddress("172.16.0.0/16"))
}

func TestIsValidIpAddress_ValidIPv6(t *testing.T) {
	assert.Assert(t, IsValidIpAddress("2001:558:6027:14:c071:6b22:17ae:ea06"))
	assert.Assert(t, IsValidIpAddress("::1"))
	assert.Assert(t, IsValidIpAddress("fe80::1"))
	assert.Assert(t, IsValidIpAddress("2001:db8::1"))
}

func TestIsValidIpAddress_ValidIPv6CIDR(t *testing.T) {
	assert.Assert(t, IsValidIpAddress("2001:558:6027:180::/57"))
	assert.Assert(t, IsValidIpAddress("2001:db8::/32"))
	assert.Assert(t, IsValidIpAddress("fe80::/10"))
}

func TestIsValidIpAddress_Invalid(t *testing.T) {
	assert.Assert(t, !IsValidIpAddress(""))
	assert.Assert(t, !IsValidIpAddress("invalid"))
	assert.Assert(t, !IsValidIpAddress("999.999.999.999"))
	assert.Assert(t, !IsValidIpAddress("192.168.1"))
	assert.Assert(t, !IsValidIpAddress("192.168.1.1.1"))
	assert.Assert(t, !IsValidIpAddress("not-an-ip"))
	assert.Assert(t, !IsValidIpAddress("192.168.1.1/"))
	assert.Assert(t, !IsValidIpAddress("192.168.1.1/33"))
}

// Test Equals
func TestIpAddress_Equals_SameIPv4(t *testing.T) {
	ip1 := NewIpAddress("192.168.1.1")
	ip2 := NewIpAddress("192.168.1.1")

	assert.Assert(t, ip1.Equals(*ip2))
}

func TestIpAddress_Equals_DifferentIPv4(t *testing.T) {
	ip1 := NewIpAddress("192.168.1.1")
	ip2 := NewIpAddress("192.168.1.2")

	assert.Assert(t, !ip1.Equals(*ip2))
}

func TestIpAddress_Equals_SameIPv6(t *testing.T) {
	ip1 := NewIpAddress("2001:db8::1")
	ip2 := NewIpAddress("2001:db8::1")

	assert.Assert(t, ip1.Equals(*ip2))
}

func TestIpAddress_Equals_DifferentIPv6(t *testing.T) {
	ip1 := NewIpAddress("2001:db8::1")
	ip2 := NewIpAddress("2001:db8::2")

	assert.Assert(t, !ip1.Equals(*ip2))
}

func TestIpAddress_Equals_SameCIDR(t *testing.T) {
	ip1 := NewIpAddress("192.168.1.0/24")
	ip2 := NewIpAddress("192.168.1.0/24")

	assert.Assert(t, ip1.Equals(*ip2))
}

func TestIpAddress_Equals_DifferentCIDR(t *testing.T) {
	ip1 := NewIpAddress("192.168.1.0/24")
	ip2 := NewIpAddress("192.168.1.0/25")

	assert.Assert(t, !ip1.Equals(*ip2))
}

func TestIpAddress_Equals_IPv4VsIPv6(t *testing.T) {
	ip1 := NewIpAddress("192.168.1.1")
	ip2 := NewIpAddress("2001:db8::1")

	assert.Assert(t, !ip1.Equals(*ip2))
}

// Test edge cases for NewIpAddress
func TestNewIpAddress_InvalidCIDR(t *testing.T) {
	ip := NewIpAddress("192.168.1.1/invalid")
	assert.Assert(t, ip == nil)
}

func TestNewIpAddress_InvalidCIDRRange(t *testing.T) {
	ip := NewIpAddress("192.168.1.1/33")
	assert.Assert(t, ip == nil)
}

func TestNewIpAddress_EmptyString(t *testing.T) {
	ip := NewIpAddress("")
	assert.Assert(t, ip == nil)
}

func TestNewIpAddress_InvalidFormat(t *testing.T) {
	ip := NewIpAddress("not-an-ip-address")
	assert.Assert(t, ip == nil)
}

// Test IsInRange with invalid input
func TestIpAddress_IsInRange_InvalidString(t *testing.T) {
	ip := NewIpAddress("192.168.1.0/24")
	assert.Assert(t, !ip.IsInRange("invalid-ip"))
	assert.Assert(t, !ip.IsInRange("999.999.999.999"))
}

func TestIpAddress_IsInRange_UnsupportedType(t *testing.T) {
	ip := NewIpAddress("192.168.1.1")
	// Passing unsupported type (int) should return false
	assert.Assert(t, !ip.IsInRange(12345))
	assert.Assert(t, !ip.IsInRange([]string{"192.168.1.1"}))
}

// Test GetAddress
func TestIpAddress_GetAddress_RegularIPv4(t *testing.T) {
	ip := NewIpAddress("192.168.1.1")
	assert.Equal(t, "192.168.1.1", ip.GetAddress())
}

func TestIpAddress_GetAddress_RegularIPv6(t *testing.T) {
	ip := NewIpAddress("2001:db8::1")
	assert.Equal(t, "2001:db8::1", ip.GetAddress())
}

func TestIpAddress_GetAddress_CIDR(t *testing.T) {
	ip := NewIpAddress("192.168.1.5/24")
	// CIDR should return normalized network address
	assert.Equal(t, "192.168.1.0/24", ip.GetAddress())
}

// Test IP() method
func TestIpAddress_IP_Method(t *testing.T) {
	ip := NewIpAddress("192.168.1.1")
	netIP := ip.IP()

	assert.Assert(t, netIP != nil)
	assert.Equal(t, "192.168.1.1", netIP.String())
}

// Test IsIpv6 and IsCidrBlock edge cases
func TestIpAddress_IPv6Detection(t *testing.T) {
	// IPv4 should not be detected as IPv6
	ip4 := NewIpAddress("192.168.1.1")
	assert.Assert(t, !ip4.IsIpv6())

	// IPv6 should be detected
	ip6 := NewIpAddress("2001:db8::1")
	assert.Assert(t, ip6.IsIpv6())

	// IPv6 loopback
	ip6Loopback := NewIpAddress("::1")
	assert.Assert(t, ip6Loopback.IsIpv6())
}

func TestIpAddress_CIDRDetection(t *testing.T) {
	// Regular IP should not be CIDR
	ip := NewIpAddress("192.168.1.1")
	assert.Assert(t, !ip.IsCidrBlock())

	// CIDR notation should be detected
	cidr := NewIpAddress("192.168.1.0/24")
	assert.Assert(t, cidr.IsCidrBlock())

	// IPv6 CIDR
	cidr6 := NewIpAddress("2001:db8::/32")
	assert.Assert(t, cidr6.IsCidrBlock())
}
