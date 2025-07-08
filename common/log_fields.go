/**
 * Copyright 2021 Comcast Cable Communications Management, LLC
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
package common

import (
	log "github.com/sirupsen/logrus"
)

var (
	unloggedFields = map[string]bool{
		"moneytrace":      true,
		"out_traceparent": true,
		"out_tracestate":  true,
		"xpc_trace":       true,
	}
)

func FilterLogFields(src log.Fields, excludes ...string) log.Fields {
	fields := make(log.Fields, len(src))
	for k, v := range src {
		if !unloggedFields[k] { // Check exclusion *before* copying
			fields[k] = v
		}
	}

	for _, x := range excludes {
		delete(fields, x)
	}
	return fields
}

func CopyLogFields(src log.Fields) log.Fields {
	fields := log.Fields{}
	for k, v := range src {
		fields[k] = v
	}
	return fields
}

func UpdateLogFields(fields, newfields log.Fields) {
	for k, v := range newfields {
		fields[k] = v
	}
}

func FieldsGetString(fields log.Fields, keyName string, args ...string) string {
	var ret string
	if len(args) == 1 {
		ret = args[0]
	}

	if itf, ok := fields[keyName]; ok {
		ret = itf.(string)
	}
	return ret
}
