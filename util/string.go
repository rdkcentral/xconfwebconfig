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
package util

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"xconfwebconfig/common"

	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zenazn/pkcs7pad"
	"golang.org/x/crypto/ripemd160"
)

var (
	telemetryFields = [][]string{
		{"version", common.HeaderProfileVersion},
		{"model", common.HeaderModelName},
		{"partnerId", common.HeaderPartnerID},
		{"accountId", common.HeaderAccountID},
		{"firmwareVersion", common.HeaderFirmwareVersion},
	}

	alnumRe    = regexp.MustCompile("[^a-zA-Z0-9]+")
	validMacRe = regexp.MustCompile(`^([0-9a-fA-F]{12}$)|([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})|([0-9A-Fa-f]{4}[.]){2}([0-9A-Fa-f]{4})$`)
)

func ToAlphaNumericString(str string) string {
	return alnumRe.ReplaceAllString(str, "")
}

func ToColonMac(d string) string {
	return fmt.Sprintf("%v:%v:%v:%v:%v:%v", d[:2], d[2:4], d[4:6], d[6:8], d[8:10], d[10:12])
}

func GetAuditId() string {
	u := uuid.New()
	ustr := u.String()
	uustr := strings.ReplaceAll(ustr, "-", "")
	return uustr
}

func GenerateRandomCpeMac() string {
	u := uuid.New().String()
	return strings.ToUpper(u[len(u)-12:])
}

func ValidateMac(mac string) bool {
	if len(mac) != 12 {
		return false
	}
	for _, r := range mac {
		if r < 48 || r > 70 || (r > 57 && r < 65) {
			return false
		}
	}
	return true
}

func GetTelemetryQueryString(header http.Header, mac string) string {
	// build the query parameters in a fixed order
	params := []string{}

	firmwareVersion := header.Get(common.HeaderFirmwareVersion)
	if strings.Contains(firmwareVersion, "PROD") {
		params = append(params, "env=PROD")
	} else if strings.Contains(firmwareVersion, "DEV") {
		params = append(params, "env=DEV")
	}

	for _, pairs := range telemetryFields {
		params = append(params, fmt.Sprintf("%v=%v", pairs[0], header.Get(pairs[1])))
	}

	estbMacAddress := GetEstbMacAddress(mac)
	params = append(params, fmt.Sprintf("estbMacAddress=%v", estbMacAddress))
	params = append(params, fmt.Sprintf("ecmMacAddress=%v", mac))

	return strings.Join(params, "&")
}

func GetEstbMacAddress(mac string) string {
	// if the mac cannot be parsed, then return back the input
	i, err := strconv.ParseInt(mac, 16, 64)
	if err != nil {
		return mac
	}
	return fmt.Sprintf("%012X", i+2)
}

// REMINDER
// a 2-D slices/arrays of strings are chosen, instead of a map, to keep the params ordering
func GetURLQueryParameterString(kvs [][]string) (string, error) {
	params := []string{}
	for _, kv := range kvs {
		if len(kv) != 2 {
			err := fmt.Errorf("len(kv) != 2")
			return "", err
		}
		params = append(params, fmt.Sprintf("%v=%v", kv[0], kv[1]))
	}
	return strings.Join(params, "&"), nil
}

func IsUnknownValue(param string) bool {
	return strings.EqualFold(param, "unknown") || strings.EqualFold(param, "NoAccount")
}

// MACAddressValidator method is to validate MAC address
// ”'
// Validate inputs are:
//
//	11-11-11-11-11-11
//	11 11 11 11 11 11
//	11:11:11:11:11:11
//	11111111111
//
// :param value: A String
// :return: A String, upper case mac address, AABBCCDDEEFF
// ”'
func MACAddressValidator(macAddress string) (bool, error) {

	// Replace all dash, space or colon from MAC address
	macAddress = AlphaNumericMacAddress(macAddress)

	// Check, if updated MAC address has only 12 char
	if len(macAddress) != 12 {
		return false, errors.New("mac address must be 12 char long")
	}

	// Match MAC Address pattern to have only 0-9 & A-F
	match, _ := regexp.MatchString("^([0-9A-F]){12}$", macAddress)
	if !match {
		return match, errors.New("mac address should have only 0-9 and/or A-F chars only")
	}
	return match, nil
}

// AlphaNumericMacAddress is converting MAC address to only alphanumeric
func AlphaNumericMacAddress(macAddress string) string {
	macAddress = strings.Replace(macAddress, "MAC:", "", -1)
	macAddress = ToAlphaNumericString(macAddress)
	macAddress = strings.ToUpper(macAddress)
	return macAddress
}

// MacAddrComplexFormat is to convert mac address from XXXXXXXXXXX to XX:XX:XX:XX:// XXX
func MacAddrComplexFormat(macaddr string) (string, error) {
	// Replace all dash, space or colon from MAC address
	macaddr = AlphaNumericMacAddress(macaddr)

	_, err := MACAddressValidator(macaddr)
	if err != nil {
		return "", err
	}

	runes := []rune(macaddr)
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		string(runes[0:2]),
		string(runes[2:4]),
		string(runes[4:6]),
		string(runes[6:8]),
		string(runes[8:10]),
		string(runes[10:]),
	), nil
}

func IsValidMacAddress(macaddr string) bool {
	_, err := MACAddressValidator(macaddr)
	return err == nil
}

func MACAddressValidatorForAS(macAddress string) (bool, error) {
	if validMacRe.MatchString(macAddress) {
		return true, nil
	}

	return false, errors.New("Invalid MAC address")
}

func IsValidMacAddressForAdminService(macaddr string) bool {
	_, err := MACAddressValidatorForAS(macaddr)
	return err == nil
}

func ValidateAndNormalizeMacAddress(macaddr string) (string, error) {
	// 1st validates the mac address
	_, err := MACAddressValidator(macaddr)
	if err != nil {
		return "", err
	}

	// Replace all dash, colon or period from MAC address
	mac := AlphaNumericMacAddress(macaddr)
	return ToColonMac(mac), nil
}

func NormalizeMacAddress(macAddress string) string {
	l := len(macAddress)
	if l < 1 {
		return ""
	}
	macAddress = AlphaNumericMacAddress(macAddress)
	return ToColonMac(macAddress)
}

func IsBlank(str string) bool {
	return strings.Trim(str, " ") == ""
}

func StringSliceEqual(a, b []string) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func CalculateHash(plainText string) string {
	if plainText == "" {
		return ""
	}
	bytes := []byte(plainText)
	if len(bytes) < 128 {
		bytes = pkcs7pad.Pad(bytes, 128)
	}
	if len(bytes) < 128 {
		log.Error("Exception, must be minimum of 128 bytes for input")
	}
	sha256Hash := sha256.Sum256(bytes)
	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Hash[:])
	result := ripemd160Hasher.Sum(nil)
	version := []byte{0}
	result = append(version, result...)

	// double sha256
	hash := sha256.Sum256(result)
	hash = sha256.Sum256(hash[:])
	checkSum := hash[:4]
	result = append(result, checkSum...)
	return base58.Encode(result)
}
