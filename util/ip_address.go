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
	"net"
	"net/http"
	"strconv"
	"strings"

	"xconfwebconfig/common"
)

const (
	INADDR4SZ = 4
)

func GetIpAddress(r *http.Request, ipAddress string) string {
	if IsIPv4LiteralAddress(ipAddress) {
		return ipAddress
	}
	if net.ParseIP(ipAddress) != nil {
		return ipAddress
	}

	// First we check 'HA-Forwarded-For' header, then 'X-Forwarded-For' if it exists and contains valid ip address.
	// Usually format of header is 'HA-Forwarded-For: client, proxy1, proxy2' so we split string by "[,]" and take first part.
	xff_headers := []string{common.HA_FORWARDED_FOR_HEADER, common.X_FORWARDED_FOR_HEADER}
	for _, header := range xff_headers {
		ipFromHeader := r.Header.Get(header)
		if !IsBlank(ipFromHeader) {
			parts := strings.Split(ipFromHeader, ",")
			if len(parts) > 0 {
				ipFromHeader = strings.Trim(parts[0], " ")
				if net.ParseIP(ipFromHeader) != nil {
					return ipFromHeader
				}
			}
		}
	}

	remoteIp := r.RemoteAddr
	if net.ParseIP(remoteIp) != nil {
		return remoteIp
	}
	if len(ipAddress) > 0 {
		return ipAddress
	}
	return "0.0.0.0"
}

/*
 * Converts IPv4 address in its textual presentation form
 * into its numeric binary form.
 *
 * @param src a String representing an IPv4 address in standard format
 * @return a byte array representing the IPv4 numeric address
 */
func TextToNumericFormatV4(src string) []byte {
	if len(src) == 0 {
		return nil
	}

	var res = make([]byte, INADDR4SZ)
	s := strings.Split(src, ".")
	switch len(s) {
	case 1:
		/*
		 * When only one part is given, the value is stored directly in
		 * the network address without any byte rearrangement.
		 */
		val, err := strconv.ParseInt(s[0], 10, 64)
		if err != nil {
			return nil
		}
		if val < 0 || val > 0xffffffff {
			return nil
		}
		res[0] = (byte)((val >> 24) & 0xff)
		res[1] = (byte)(((val & 0xffffff) >> 16) & 0xff)
		res[2] = (byte)(((val & 0xffff) >> 8) & 0xff)
		res[3] = (byte)(val & 0xff)
	case 2:
		/*
		 * When a two part address is supplied, the last part is
		 * interpreted as a 24-bit quantity and placed in the right
		 * most three bytes of the network address. This makes the
		 * two part address format convenient for specifying Class A
		 * network addresses as net.host.
		 */
		i64, err := strconv.ParseInt(s[0], 10, 32)
		if err != nil {
			return nil
		}
		val := int32(i64)
		if val < 0 || val > 0xff {
			return nil
		}
		res[0] = (byte)(val & 0xff)
		i64, err = strconv.ParseInt(s[1], 10, 32)
		if err != nil {
			return nil
		}
		val = int32(i64)
		if val < 0 || val > 0xffffff {
			return nil
		}
		res[1] = (byte)((val >> 16) & 0xff)
		res[2] = (byte)(((val & 0xffff) >> 8) & 0xff)
		res[3] = (byte)(val & 0xff)
	case 3:
		/*
		 * When a three part address is specified, the last part is
		 * interpreted as a 16-bit quantity and placed in the right
		 * most two bytes of the network address. This makes the
		 * three part address format convenient for specifying
		 * Class B net- work addresses as 128.net.host.
		 */
		for i := 0; i < 2; i++ {
			i64, err := strconv.ParseInt(s[i], 10, 32)
			if err != nil {
				return nil
			}
			val := int32(i64)
			if val < 0 || val > 0xff {
				return nil
			}
			res[i] = (byte)(val & 0xff)
		}
		i64, err := strconv.ParseInt(s[2], 10, 32)
		if err != nil {
			return nil
		}
		val := int32(i64)
		if val < 0 || val > 0xffff {
			return nil
		}
		res[2] = (byte)((val >> 8) & 0xff)
		res[3] = (byte)(val & 0xff)
	case 4:
		/*
		 * When four parts are specified, each is interpreted as a
		 * byte of data and assigned, from left to right, to the
		 * four bytes of an IPv4 address.
		 */
		for i := 0; i < 4; i++ {
			i64, err := strconv.ParseInt(s[i], 10, 32)
			if err != nil {
				return nil
			}
			val := int32(i64)
			if val < 0 || val > 0xff {
				return nil
			}
			res[i] = (byte)(val & 0xff)
		}
	default:
		return nil
	}
	return res
}

/**
 * @param src a String representing an IPv4 address in textual format
 * @return a boolean indicating whether src is an IPv4 literal address
 */
func IsIPv4LiteralAddress(src string) bool {
	return TextToNumericFormatV4(src) != nil
}
