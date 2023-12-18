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

type Set map[string]struct{}

func NewSet(items ...string) Set {
	s := Set{}
	s.Add(items...)
	return s
}

func (s Set) Contains(obj string) bool {
	if _, ok := s[obj]; ok {
		return true
	} else {
		return false
	}
}

func (s Set) Add(items ...string) {
	for _, x := range items {
		s[x] = struct{}{}
	}
}

func (s Set) Remove(x string) {
	delete(s, x)
}

func (s Set) ToSlice() []string {
	var list []string
	for k := range s {
		list = append(list, k)
	}
	return list
}
