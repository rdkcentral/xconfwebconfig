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

	"gotest.tools/assert"
)

func TestSet(t *testing.T) {
	s1 := NewSet("red", "orange", "yellow", "yellow", "red", "red", "red", "green")
	assert.Equal(t, len(s1), 4)
	assert.Assert(t, s1.Contains("red"))

	s2 := Set{}
	s2.Add("green")
	s2.Add("green")
	s2.Add("green")
	s2.Add("orange")
	s2.Add("yellow")
	s2.Add("yellow")
	s2.Add("orange")
	s2.Add("red")
	s2.Add("orange")
	s2.Add("red")
	s2.Add("red")
	s2.Add("green")
	assert.Equal(t, len(s2), 4)
	assert.DeepEqual(t, s1, s2)

	s1.Remove("orange")
	s1.Remove("green")
	assert.Equal(t, len(s1), 2)

	s1.Remove("green")
	assert.Equal(t, len(s1), 2)

	s3 := Set{}
	s3.Add("green", "red", "orange", "red", "yellow", "yellow", "green")
	assert.Equal(t, len(s3), 4)
	assert.DeepEqual(t, s2, s3)

	slice := s3.ToSlice()
	assert.Equal(t, len(slice), 4)
	assert.Assert(t, ContainsAny(slice, []string{"green", "red", "orange", "yellow"}))

	s4 := Set{}
	slice = s4.ToSlice()
	assert.Equal(t, len(slice), 0)
}
