/*
*

  - Copyright 2022 Comcast Cable Communications Management, LLC
    *

  - Licensed under the Apache License, Version 2.0 (the "License");

  - you may not use this file except in compliance with the License.

  - You may obtain a copyfunc TestValidateAndNormalizeMacAddress(t *testing.T) {
    // Test valid MAC address - function converts all to uppercase
    result, err := ValidateAndNormalizeMacAddress("aa:bb:cc:dd:ee:ff")
    assert.NilError(t, err)
    assert.Equal(t, result, "AA:BB:CC:DD:EE:FF") // Function converts to uppercase

    result, err = ValidateAndNormalizeMacAddress("aabbccddeeff")
    assert.NilError(t, err)
    assert.Equal(t, result, "AA:BB:CC:DD:EE:FF") // Function converts to uppercase

    result, err = ValidateAndNormalizeMacAddress("AABBCCDDEEFF")
    assert.NilError(t, err)
    assert.Equal(t, result, "AA:BB:CC:DD:EE:FF") // Function preserves uppercasese	result, err = ValidateAndNormalizeMacAddress("AABBCCDDEEFF")
    assert.NilError(t, err)
    assert.Equal(t, result, "AA:BB:CC:DD:EE:FF") // Function preserves uppercase
    *

  - http://www.apache.org/licenses/LICENSE-2.0
    *

  - Unless required by applicable law or agreed to in writing, software

  - distributed under the License is distributed on an "AS IS" BASIS,

  - WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.

  - See the License for the specific language governing permissions and

  - limitations under the License.
    *

  - SPDX-License-Identifier: Apache-2.0
*/
package util

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/common"

	"gotest.tools/assert"
)

func TestString(t *testing.T) {
	s := "112233445566"
	c := ToColonMac(s)
	expected := "11:22:33:44:55:66"
	assert.Equal(t, c, expected)
}

func TestValidateMac(t *testing.T) {
	mac := "001122334455"
	assert.Assert(t, ValidateMac(mac))

	mac = "4444ABCDEF01"
	assert.Assert(t, ValidateMac(mac))

	mac = "00112233445Z"
	assert.Assert(t, !ValidateMac(mac))

	mac = "001122334455Z"
	assert.Assert(t, !ValidateMac(mac))

	mac = "0H1122334455"
	assert.Assert(t, !ValidateMac(mac))

	for i := 0; i < 10; i++ {
		mac := GenerateRandomCpeMac()
		assert.Assert(t, ValidateMac(mac))
	}
}

func TestGetAuditId(t *testing.T) {
	auditId := GetAuditId()
	assert.Equal(t, len(auditId), 32)
}

func TestTelemetryQuery(t *testing.T) {
	header := http.Header{}
	header.Set(common.HeaderProfileVersion, "2.0")
	header.Set(common.HeaderModelName, "TG1682G")
	header.Set(common.HeaderPartnerID, "comcast")
	header.Set(common.HeaderAccountID, "1234567890")
	header.Set(common.HeaderFirmwareVersion, "TG1682_3.14p9s6_PROD_sey")
	mac := "567890ABCDEF"
	qstr := GetTelemetryQueryString(header, mac)

	expected := "env=PROD&version=2.0&model=TG1682G&partnerId=comcast&accountId=1234567890&firmwareVersion=TG1682_3.14p9s6_PROD_sey&estbMacAddress=567890ABCDF1&ecmMacAddress=567890ABCDEF"
	assert.Equal(t, qstr, expected)
}

func TestGetQueryParameters(t *testing.T) {
	// ==== normal ====
	kvs := [][]string{
		{"env", "PROD"},
		{"version", "2.0"},
		{"model", "CGM4140COM"},
		{"partnerId", "comcast"},
		{"accountId", "1234567890"},
		{"firmwareVersion", "CGM4140COM_4.4p1s11_PROD_sey"},
		{"estbMacAddress", "112233445565"},
		{"ecmMacAddress", "112233445567"},
	}
	expected := "env=PROD&version=2.0&model=CGM4140COM&partnerId=comcast&accountId=1234567890&firmwareVersion=CGM4140COM_4.4p1s11_PROD_sey&estbMacAddress=112233445565&ecmMacAddress=112233445567"
	queryParams, err := GetURLQueryParameterString(kvs)
	assert.NilError(t, err)
	assert.Equal(t, expected, queryParams)

	// ==== ill formatted ====
	kvs = [][]string{
		{"env", "PROD"},
		{"version", "2.0"},
		{"model", "CGM4140COM"},
		{"partnerId", "comcast", "cox"},
		{"accountId", "1234567890"},
		{"firmwareVersion", "CGM4140COM_4.4p1s11_PROD_sey"},
		{"estbMacAddress", "112233445565"},
		{"ecmMacAddress", "112233445567"},
	}
	_, err = GetURLQueryParameterString(kvs)
	assert.Assert(t, err != nil)
}

func TestIsUnknownValue(t *testing.T) {
	isUnknown := IsUnknownValue("hello")
	assert.Equal(t, isUnknown, false)

	isUnknown = IsUnknownValue("")
	assert.Equal(t, isUnknown, false)

	isUnknown = IsUnknownValue("UNKNOWN")
	assert.Equal(t, isUnknown, true)

	isUnknown = IsUnknownValue("noaccount")
	assert.Equal(t, isUnknown, true)
}

func TestMACAddressValidator(t *testing.T) {
	fmt.Println("Testing MACAddressValidator...")
	var err error

	// Positive scenarios
	validMac, err := MACAddressValidator("142536ABAC23")
	assert.Equal(t, validMac, true)

	// Positive scenarios
	validMac, err = MACAddressValidator("MAC:142536ABAC23")
	assert.Equal(t, validMac, true)

	validMac, err = MACAddressValidator("14:68:36:AB:DD:23")
	assert.Equal(t, validMac, true)

	validMac, err = MACAddressValidator("11 25 F6 AB AC 23")
	assert.Equal(t, validMac, true)

	validMac, err = MACAddressValidator("14-25-36-AB-AC-23")
	assert.Equal(t, validMac, true)

	validMac, err = MACAddressValidator("bd-c5-9a-7e-fd-23")
	assert.Equal(t, validMac, true)

	validMac, err = MACAddressValidator("ab-bc 9a:7efd23")
	assert.Equal(t, validMac, true)

	// Nagetive scenarios
	validMac, err = MACAddressValidator("14-25-36-LP-AT-23")
	assert.Equal(t, validMac, false)

	validMac, err = MACAddressValidator("14253 6LPAT:23")
	assert.Equal(t, validMac, false)

	validMac, err = MACAddressValidator("14-25-36AC-23")
	assert.Error(t, err, "mac address must be 12 char long")

	validMac, err = MACAddressValidator("14-25-36AC-23:aa 66")
	assert.Error(t, err, "mac address must be 12 char long")

	validMac, err = MACAddressValidator("MAC:142536HBAC23")
	assert.Error(t, err, "mac address should have only 0-9 and/or A-F chars only")

	validMac, err = MACAddressValidator("14536RBAC233")
	assert.Error(t, err, "mac address should have only 0-9 and/or A-F chars only")
}

func TestIsValidMacAddress(t *testing.T) {
	isValidMacAddress := IsValidMacAddress("142536ABAC23")
	assert.Equal(t, isValidMacAddress, true)

	isValidMacAddress = IsValidMacAddress("14:25:36:ab:ac:23")
	assert.Equal(t, isValidMacAddress, true)

	isValidMacAddress = IsValidMacAddress("helloworld")
	assert.Equal(t, isValidMacAddress, false)

	isValidMacAddress = IsValidMacAddress("")
	assert.Equal(t, isValidMacAddress, false)
}

func TestMacAddrComplexFormat(t *testing.T) {
	fmt.Println("Testing MacAddrComplexFormat...")
	var err error

	// Positive scenarios
	validMac, err := MacAddrComplexFormat("142536ABAC23")
	assert.Equal(t, validMac, "14:25:36:AB:AC:23")

	validMac, err = MacAddrComplexFormat("11 25 R6 AB AC 23")
	assert.Error(t, err, "mac address should have only 0-9 and/or A-F chars only")
}

func TestMACAddressValidatorForAS(t *testing.T) {
	// Test valid MAC address
	valid, err := MACAddressValidatorForAS("aa:bb:cc:dd:ee:ff")
	assert.NilError(t, err)
	assert.Assert(t, valid)

	valid, err = MACAddressValidatorForAS("AA:BB:CC:DD:EE:FF")
	assert.NilError(t, err)
	assert.Assert(t, valid)

	valid, err = MACAddressValidatorForAS("aabbccddeeff")
	assert.NilError(t, err)
	assert.Assert(t, valid)

	// Test invalid MAC address
	valid, _ = MACAddressValidatorForAS("invalid")
	assert.Assert(t, !valid)

	valid, _ = MACAddressValidatorForAS("")
	assert.Assert(t, !valid)

	valid, _ = MACAddressValidatorForAS("aa:bb:cc:dd:ee")
	assert.Assert(t, !valid)
}

func TestIsValidMacAddressForAdminService(t *testing.T) {
	// Test valid MAC address
	assert.Assert(t, IsValidMacAddressForAdminService("aa:bb:cc:dd:ee:ff"))
	assert.Assert(t, IsValidMacAddressForAdminService("AA:BB:CC:DD:EE:FF"))
	assert.Assert(t, IsValidMacAddressForAdminService("aabbccddeeff"))
	assert.Assert(t, IsValidMacAddressForAdminService("AABBCCDDEEFF"))

	// Test invalid MAC address
	assert.Assert(t, !IsValidMacAddressForAdminService("invalid"))
	assert.Assert(t, !IsValidMacAddressForAdminService(""))
	assert.Assert(t, !IsValidMacAddressForAdminService("aa:bb:cc:dd:ee"))
}

func TestValidateAndNormalizeMacAddress(t *testing.T) {
	// Test valid MAC address
	result, err := ValidateAndNormalizeMacAddress("aa:bb:cc:dd:ee:ff")
	assert.NilError(t, err)
	assert.Equal(t, result, "AA:BB:CC:DD:EE:FF") // Function converts to uppercase

	result, err = ValidateAndNormalizeMacAddress("aabbccddeeff")
	assert.NilError(t, err)
	assert.Equal(t, result, "AA:BB:CC:DD:EE:FF") // Function converts to uppercase

	result, err = ValidateAndNormalizeMacAddress("AABBCCDDEEFF")
	assert.NilError(t, err)
	assert.Equal(t, result, "AA:BB:CC:DD:EE:FF")

	// Test invalid MAC address
	_, err = ValidateAndNormalizeMacAddress("invalid")
	assert.Assert(t, err != nil, "Expected error for invalid MAC address")

	_, err = ValidateAndNormalizeMacAddress("")
	assert.Assert(t, err != nil, "Expected error for empty MAC address")
}

func TestNormalizeMacAddress(t *testing.T) {
	// Test with colon-separated MAC - function converts all to uppercase
	result := NormalizeMacAddress("aa:bb:cc:dd:ee:ff")
	assert.Equal(t, result, "AA:BB:CC:DD:EE:FF") // Function converts to uppercase

	// Test with non-colon MAC (should add colons and convert to uppercase)
	result = NormalizeMacAddress("aabbccddeeff")
	assert.Equal(t, result, "AA:BB:CC:DD:EE:FF") // Function converts to uppercase

	result = NormalizeMacAddress("AABBCCDDEEFF")
	assert.Equal(t, result, "AA:BB:CC:DD:EE:FF") // Function preserves uppercase

	// Test with invalid length - function will panic if we pass invalid data
	// So let's test with valid length but invalid characters
	result = NormalizeMacAddress("123456789012") // 12 chars but not valid hex
	assert.Equal(t, result, "12:34:56:78:90:12") // Function still formats it

	result = NormalizeMacAddress("")
	assert.Equal(t, result, "") // Should return original if empty
}

func TestStringSliceEqual(t *testing.T) {
	// Test equal slices
	slice1 := []string{"a", "b", "c"}
	slice2 := []string{"a", "b", "c"}
	assert.Assert(t, StringSliceEqual(slice1, slice2))

	// Test different order (should be false as it's exact comparison)
	slice3 := []string{"c", "b", "a"}
	assert.Assert(t, !StringSliceEqual(slice1, slice3))

	// Test different lengths
	slice4 := []string{"a", "b"}
	assert.Assert(t, !StringSliceEqual(slice1, slice4))

	// Test different values
	slice5 := []string{"a", "b", "d"}
	assert.Assert(t, !StringSliceEqual(slice1, slice5))

	// Test empty slices
	assert.Assert(t, StringSliceEqual([]string{}, []string{}))

	// Test one empty
	assert.Assert(t, !StringSliceEqual(slice1, []string{}))
}

func TestCalculateHash(t *testing.T) {
	// Test with string input
	hash := CalculateHash("test string")
	assert.Assert(t, hash != "")
	assert.Assert(t, len(hash) > 0)

	// Test that same input gives same hash
	hash2 := CalculateHash("test string")
	assert.Equal(t, hash, hash2)

	// Test that different input gives different hash
	hash3 := CalculateHash("different string")
	assert.Assert(t, hash != hash3)

	// Test with empty string
	hashEmpty := CalculateHash("")
	assert.Assert(t, hashEmpty == "") // Empty input returns empty string
}
