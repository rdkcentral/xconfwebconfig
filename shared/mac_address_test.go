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
package shared

import (
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestMacAddress(t *testing.T) {
	expected := "F8:A0:97:1E:D6:74"

	// parse "F8A0971ED674" and lower
	s1 := "F8A0971ED674"
	m1, err := NewMacAddress(s1)
	assert.NilError(t, err)
	assert.Equal(t, m1.String(), expected)

	m1, err = NewMacAddress("f8A0971ed674")
	assert.NilError(t, err)
	assert.Equal(t, m1.String(), expected)

	m2, err := NewMacAddress(strings.ToLower(s1))
	assert.NilError(t, err)
	assert.Equal(t, m2.String(), expected)

	// parse "F8:A0:97:1E:D6:74" lower
	m3, err := NewMacAddress(expected)
	assert.NilError(t, err)
	assert.Equal(t, m3.String(), expected)

	m4, err := NewMacAddress(strings.ToLower(expected))
	assert.NilError(t, err)
	assert.Equal(t, m4.String(), expected)

	// parse "F8-A0-97-1E-D6-74"
	s5 := "F8-A0-97-1E-D6-74"
	m5, err := NewMacAddress(s5)
	assert.NilError(t, err)
	assert.Equal(t, m5.String(), expected)

	m6, err := NewMacAddress(strings.ToLower(s5))
	assert.NilError(t, err)
	assert.Equal(t, m6.String(), expected)

	// error
	_, err = NewMacAddress("F8A0971ED674A")
	assert.Assert(t, err != nil)

	_, err = NewMacAddress("X8A0971ED674")
	assert.Assert(t, err != nil)

	_, err = NewMacAddress("x8:a0971ED674")
	assert.Assert(t, err != nil)
}
