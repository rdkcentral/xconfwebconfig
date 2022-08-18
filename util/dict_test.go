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
	"testing"
)

func TestUtilPrettyPrint(t *testing.T) {
	line := `{"foo":"bar", "enabled": true, "age": 30}`
	t.Logf(PrettyJson(line))

	a := Dict{
		"broadcast_ssid":  true,
		"radio_index":     10100,
		"ssid_enabled":    true,
		"ssid_index":      10101,
		"ssid_name":       "hello",
		"wifi_security":   4,
		"wifi_passphrase": "password1",
	}
	t.Logf(PrettyJson(a))

	b := []Dict{
		Dict{
			"broadcast_ssid":  true,
			"radio_index":     10000,
			"ssid_enabled":    true,
			"ssid_index":      10001,
			"ssid_name":       "ssid_2g",
			"wifi_security":   4,
			"wifi_passphrase": "password2",
		},
		Dict{
			"broadcast_ssid":  true,
			"radio_index":     10100,
			"ssid_enabled":    true,
			"ssid_index":      10101,
			"ssid_name":       "ssid_5g",
			"wifi_security":   4,
			"wifi_passphrase": "password5",
		},
	}
	t.Logf(PrettyJson(b))
}
