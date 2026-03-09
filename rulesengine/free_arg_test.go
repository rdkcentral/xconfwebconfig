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
package rulesengine

import (
	"testing"

	"gotest.tools/assert"
)

func TestFreeArgs(t *testing.T) {
	freeArgTypes := [...]string{
		StandardFreeArgTypeString,
		StandardFreeArgTypeLong,
		StandardFreeArgTypeVoid,
		StandardFreeArgTypeAny,
		AuxFreeArgTypeTime,
		AuxFreeArgTypeIpAddress,
		AuxFreeArgTypeMacAddress,
	}

	for _, type1 := range freeArgTypes {
		for _, type2 := range freeArgTypes {

			a1 := NewFreeArg(type1, "foo")
			a2 := NewFreeArg(type2, "foo")

			if type1 == type2 {
				assert.Assert(t, a1.Equals(a2))
			} else {
				assert.Assert(t, !a1.Equals(a2))
			}

			a3 := NewFreeArg(type1, "bar")
			assert.Assert(t, !a1.Equals(a3))
		}
	}
}

func TestFreeArg_SetType(t *testing.T) {
	freeArg := NewFreeArg(StandardFreeArgTypeString, "testArg")
	
	freeArg.SetType(StandardFreeArgTypeLong)
	
	assert.Equal(t, StandardFreeArgTypeLong, freeArg.GetType(), "SetType should update the type")
}

func TestFreeArg_SetName(t *testing.T) {
	freeArg := NewFreeArg(StandardFreeArgTypeString, "originalName")
	
	freeArg.SetName("newName")
	
	assert.Equal(t, "newName", freeArg.GetName(), "SetName should update the name")
}

func TestFreeArg_String(t *testing.T) {
	testCases := []struct {
		name     string
		freeArg  *FreeArg
		expected string
	}{
		{
			name:     "STRING type",
			freeArg:  NewFreeArg(StandardFreeArgTypeString, "model"),
			expected: "'model'",
		},
		{
			name:     "LONG type",
			freeArg:  NewFreeArg(StandardFreeArgTypeLong, "version"),
			expected: "'version(LONG)'",
		},
		{
			name:     "VOID type",
			freeArg:  NewFreeArg(StandardFreeArgTypeVoid, "void"),
			expected: "'void(VOID)'",
		},
		{
			name:     "TIME type",
			freeArg:  NewFreeArg(AuxFreeArgTypeTime, "timeValue"),
			expected: "'timeValue(TIME)'",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.freeArg.String()
			assert.Equal(t, tc.expected, result, "String representation should match expected")
		})
	}
}
