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
	"reflect"
	"sort"
	"strings"
)

func Contains(collection interface{}, element interface{}) bool {
	switch ty := element.(type) {
	case string:
		if elements, ok := collection.([]string); ok {
			for _, e := range elements {
				if e == ty {
					return true
				}
			}
		}
	case int:
		if elements, ok := collection.([]int); ok {
			for _, e := range elements {
				if e == ty {
					return true
				}
			}
		}
	case float64:
		if elements, ok := collection.([]float64); ok {
			for _, e := range elements {
				if e == ty {
					return true
				}
			}
		}
	}
	return false
}

func ContainsAny(collection1 []string, collection2 []string) bool {
	if len(collection1) < len(collection2) {
		for _, e := range collection1 {
			if Contains(collection2, e) {
				return true
			}
		}
	} else {
		for _, e := range collection2 {
			if Contains(collection1, e) {
				return true
			}
		}
	}
	return false
}

// TODO keep it for backward compatibility in "webconfig" for now
//      plan to remove it later
func ContainsInt(data []int, x int) bool {
	for _, d := range data {
		if d == x {
			return true
		}
	}
	return false
}

func CaseInsensitiveContains(data []string, x string) bool {
	for _, d := range data {
		if strings.ToLower(x) == strings.ToLower(d) {
			return true
		}
	}
	return false
}

func StringElementsMatch(a, b []string) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func StringAppendIfMissing(collection []string, s string) []string {
	for _, e := range collection {
		if e == s {
			return collection
		}
	}
	return append(collection, s)
}

func PutIfValuePresent(m map[string]interface{}, k string, v interface{}) {
	if v == nil {
		return
	}

	vstr, ok := v.(string)
	if ok && vstr == "" {
		return
	}

	switch reflect.TypeOf(v).Kind() {
	case reflect.Map, reflect.Array, reflect.Slice:
		value := reflect.ValueOf(v)
		if value.IsNil() || reflect.ValueOf(v).Len() == 0 {
			return
		}
	}

	m[k] = v
}

func NewStringSet(collection []string) map[string]struct{} {
	if collection == nil {
		return nil
	}

	set := make(map[string]struct{})

	for _, e := range collection {
		set[e] = struct{}{}
	}

	return set
}
