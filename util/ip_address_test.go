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
	"net/http"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/common"

	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

func TestGetIpAddress(t *testing.T) {
	headerIpAddress := "192.0.2.1"
	paramIpAddress := "193.0.2.1"
	remoteIpAddress := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"

	req, _ := http.NewRequest("GET", "www.test.com", nil)

	// test no ipAddress
	ipAddress := GetIpAddress(req, "", log.Fields{})
	assert.Equal(t, ipAddress, "0.0.0.0")

	// test invalid ipAddress
	ipAddress = GetIpAddress(req, "192.0.2", log.Fields{})
	// assert.Equal(t, ipAddress, "0.0.0.0")
	assert.Equal(t, ipAddress, "192.0.2")

	// test remote ipAddress
	req.RemoteAddr = remoteIpAddress
	ipAddress = GetIpAddress(req, "", log.Fields{})
	assert.Equal(t, ipAddress, remoteIpAddress)

	// test ipAddress in header
	req.Header.Set(common.X_FORWARDED_FOR_HEADER, headerIpAddress)
	ipAddress = GetIpAddress(req, "", log.Fields{})
	assert.Equal(t, ipAddress, headerIpAddress)

	req.Header.Del(common.X_FORWARDED_FOR_HEADER)
	req.Header.Set(common.HA_FORWARDED_FOR_HEADER, remoteIpAddress)
	ipAddress = GetIpAddress(req, "", log.Fields{})
	assert.Equal(t, ipAddress, remoteIpAddress)

	// test param ipAddress
	ipAddress = GetIpAddress(req, paramIpAddress, log.Fields{})
	assert.Equal(t, ipAddress, paramIpAddress)
}

func TestTextToNumericFormatV4(t *testing.T) {
	ipAddr := ""
	bytes := TextToNumericFormatV4(ipAddr)
	assert.Assert(t, bytes == nil)

	// 192 x (256)^3 + 168 x (256)^2 + 1 x (256)^1 + 2 (256)^0 = ?
	// 3221225472 + 11010048 + 256 + 2 = 3232235778
	ipAddr = "3232235778" // 192.168.1.2
	bytes = TextToNumericFormatV4(ipAddr)
	assert.Equal(t, 4, len(bytes))
	assert.Equal(t, 192, int(bytes[0]))
	assert.Equal(t, 168, int(bytes[1]))
	assert.Equal(t, 1, int(bytes[2]))
	assert.Equal(t, 2, int(bytes[3]))

	ipAddr = "127.1" // 127.0.0.1
	bytes = TextToNumericFormatV4(ipAddr)
	assert.Equal(t, 4, len(bytes))
	assert.Equal(t, 127, int(bytes[0]))
	assert.Equal(t, 0, int(bytes[1]))
	assert.Equal(t, 0, int(bytes[2]))
	assert.Equal(t, 1, int(bytes[3]))

	ipAddr = "127.65530" // 127.0.255.250
	bytes = TextToNumericFormatV4(ipAddr)
	assert.Equal(t, 4, len(bytes))
	assert.Equal(t, 127, int(bytes[0]))
	assert.Equal(t, 0, int(bytes[1]))
	assert.Equal(t, 255, int(bytes[2]))
	assert.Equal(t, 250, int(bytes[3]))

	ipAddr = "192.168.1" // 192.168.0.1
	bytes = TextToNumericFormatV4(ipAddr)
	assert.Equal(t, 4, len(bytes))
	assert.Equal(t, 192, int(bytes[0]))
	assert.Equal(t, 168, int(bytes[1]))
	assert.Equal(t, 0, int(bytes[2]))
	assert.Equal(t, 1, int(bytes[3]))
}

func TestIsIPv4LiteralAddress(t *testing.T) {
	ipAddr := "4.5.6"
	assert.Assert(t, IsIPv4LiteralAddress(ipAddr))
}
