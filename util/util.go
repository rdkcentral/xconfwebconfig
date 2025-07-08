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
	"bytes"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net"
	"strconv"
	"strings"
	"time"
	"xconfwebconfig/common"

	"github.com/go-akka/configuration"
	"github.com/gocql/gocql"
	copy "github.com/mitchellh/copystructure"
)

var (
	TZ, _ = time.LoadLocation("UTC")
)

// UtcCurrentTimestamp - return current timestamp in UTC timezone
func UtcCurrentTimestamp() time.Time {
	return time.Now().In(TZ)
}

// UtcOffsetTimestamp currect timestamp
func UtcOffsetTimestamp(sec int) time.Time {
	return UtcCurrentTimestamp().Add(time.Duration(sec) * time.Second)
}

// UtcTimeInNano - return current time in nano in UTC timezone
func UtcTimeInNano() int64 {
	return UtcCurrentTimestamp().UnixNano()
}

func Copy(obj interface{}) (interface{}, error) {
	// Create a deep copy of the object
	cloneObj, err := copy.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj, nil
}

// GetTimestamp - return current timestamp in Millisecond in UTC timezone or convert specified time to Millisecond
func GetTimestamp(args ...time.Time) int64 {
	var unixNano int64
	if args == nil {
		unixNano = UtcTimeInNano()
	} else {
		unixNano = args[0].UnixNano()
	}
	return unixNano / int64(time.Millisecond)
}

func IsValidAppSetting(key string) bool {
	return Contains(common.AllAppSettings, key)
}

func RemoveNonAlphabeticSymbols(macAddress string) string {
	macAddress = strings.Replace(macAddress, ":", "", -1)
	macAddress = strings.Replace(macAddress, "-", "", -1)
	macAddress = strings.Replace(macAddress, "\\.", "", -1)
	macAddress = strings.ToUpper(macAddress)
	return strings.TrimSpace(macAddress)
}

// UUIDFromTime gocql method implementation
func UUIDFromTime(timestamp int64, node int64, clockSeq uint32) (gocql.UUID, error) {
	microseconds := int64(time.Duration(timestamp) * time.Microsecond)
	intervals := (microseconds * 10) + 0x01b21dd213814000

	timeLow := intervals & 0xffffffff
	timeMid := (intervals >> 32) & 0xffff
	timeHiVersion := (intervals>>48)&0x0fff + 0x1000

	clockSeqLow := clockSeq & 0xff
	clockSeqHiVariant := 0x80 | ((clockSeq >> 8) & 0x3f)

	/*
		Ref: https://tools.ietf.org/html/rfc4122
		     UUID                   = time-low "-" time-mid "-"
		                             time-high-and-version "-"
		                             clock-seq-and-reserved
		                             clock-seq-low "-" node
		    time-low               = 4hexOctet
		    time-mid               = 2hexOctet
		    time-high-and-version  = 2hexOctet
		    clock-seq-and-reserved = hexOctet
		    clock-seq-low          = hexOctet
		    node                   = 6hexOctet
		  hexOctet               = hexDigit hexDigit
	*/
	uuid := fmt.Sprintf("%08x", int64(timeLow)) + "-" +
		fmt.Sprintf("%04x", int64(timeMid)) + "-" +
		fmt.Sprintf("%04x", int64(timeHiVersion)) + "-" +
		fmt.Sprintf("%02x", int64(clockSeqHiVariant)) +
		fmt.Sprintf("%02x", int64(clockSeqLow)) + "-" +
		fmt.Sprintf("%012x", int64(node))
	return gocql.ParseUUID(uuid)
}

/*
JSONMarshal is used to marshal struct to string Without escaping special character like &, <, >

Note: JSONMarshal will terminate each value with a newline
*/
func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func XConfJSONMarshal(v interface{}, safeEncoding bool) ([]byte, error) {
	b, err := json.Marshal(v)

	if safeEncoding {
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}

func ApiVersionGreaterOrEqual(version string, value float32) bool {
	if len(version) > 0 {
		fValue, err := strconv.ParseFloat(version, 32)
		if err != nil {
			return false
		}
		float := float32(fValue)
		if float >= value {
			return true
		}
	}
	return false
}

func GetCRC32HashValue(text string) string {
	table := crc32.MakeTable(crc32.IEEE)
	hashValue := crc32.Update(0, table, []byte(text))
	return fmt.Sprintf("%08x", hashValue)
}

func IsVersionGreaterOrEqual(version string, value float64) bool {
	if version != "" {
		floatVersion, err := strconv.ParseFloat(version, 64)
		if err == nil {
			return floatVersion >= value
		}
	}
	return false
}

func CreateKeyValuePairsFromMap(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}

func GetEcmMacAddress(mac string) string {
	// if the mac cannot be parsed, then return back the input
	i, err := strconv.ParseInt(mac, 16, 64)
	if err != nil {
		return mac
	}
	return fmt.Sprintf("%012X", i-2)
}

// check ip address is Ipv4 or Ipv6
func IsIPv4(ip net.IP) bool {
	if ip == nil {
		return false
	}
	return ip.To4() != nil
}

func IsInSameNetwork(ip1Str, ip2Str string, mask net.IPMask) (bool, error) {
	ip1 := net.ParseIP(ip1Str)
	if ip1 == nil {
		return false, fmt.Errorf("invalid IP address: %s", ip1Str)
	}

	ip2 := net.ParseIP(ip2Str)
	if ip2 == nil {
		return false, fmt.Errorf("invalid IP address: %s", ip1Str)
	}

	//mask := net.IPMask(maskStr)

	if mask == nil {
		return false, fmt.Errorf("invalid netmask")
	}

	ip1Network := ip1.Mask(mask)
	ip2Network := ip2.Mask(mask)

	return ip1Network.Equal(ip2Network), nil
}

// generate Ipv6 network mask string
func Ipv6NetworkMask(prefixLength int) net.IPMask {
	mask := net.CIDRMask(prefixLength, 128)
	return mask
}

func Ipv4NetworkMask(prefixLength int) net.IPMask {
	mask := net.CIDRMask(prefixLength, 32)
	return mask
}

func CreateConfigMapStringBool(conf *configuration.Config, s string) map[string]bool {
	m := map[string]bool{}
	config := conf.GetNode(s)
	if config != nil {
		for key, val := range config.GetObject().Items() {
			m[key] = val.GetBoolean()
		}
	}
	return m
}

func CreateConfigMapStringString(conf *configuration.Config, s string) map[string]string {
	m := map[string]string{}
	config := conf.GetNode(s)
	if config != nil {
		for key, val := range config.GetObject().Items() {
			m[key] = val.GetString()
		}
	}
	return m
}
