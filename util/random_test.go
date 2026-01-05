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

func TestRandomDouble(t *testing.T) {
	value := RandomDouble()

	// Should return a value between 0 and 1
	assert.Assert(t, value >= 0.0)
	assert.Assert(t, value <= 1.0)

	// Test multiple calls should work
	value2 := RandomDouble()
	assert.Assert(t, value2 >= 0.0)
	assert.Assert(t, value2 <= 1.0)
}

func TestRandomPercentage(t *testing.T) {
	value := RandomPercentage()

	// Should return a value between 0 and 99
	assert.Assert(t, value >= 0)
	assert.Assert(t, value <= 99)

	// Test multiple calls should work and be valid
	for i := 0; i < 10; i++ {
		val := RandomPercentage()
		assert.Assert(t, val >= 0)
		assert.Assert(t, val <= 99)
	}
}
