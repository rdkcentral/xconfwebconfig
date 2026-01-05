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

func TestContains(t *testing.T) {
	days := []string{"mon", "tue", "wed", "thu"}
	c1 := Contains(days, "wed")
	assert.Assert(t, c1)
	c2 := Contains(days, "fri")
	assert.Assert(t, !c2)

	assert.Assert(t, Contains([]int{1, 2, 3, 4}, 3))
	assert.Assert(t, !Contains([]int{1, 2, 3, 4}, 9))
	assert.Assert(t, Contains([]string{"red", "orange", "yellow", "green", "blue"}, "orange"))
	assert.Assert(t, !Contains([]string{"red", "orange", "yellow", "green", "blue"}, "violet"))
	assert.Assert(t, Contains([]float64{1.1, 2.2, 3.3, 4.4}, 3.3))
	assert.Assert(t, !Contains([]float64{1.1, 2.2, 3.3, 4.4}, 9.2))
}

func TestContainsInt(t *testing.T) {
	values := []int{1, 2, 3, 4}
	c1 := ContainsInt(values, 3)
	assert.Assert(t, c1)
	c2 := ContainsInt(values, 5)
	assert.Assert(t, !c2)
}

func TestCaseInsensitiveContains(t *testing.T) {
	days := []string{"lon", "tue", "Wed", "thu"}
	c1 := CaseInsensitiveContains(days, "weD")
	assert.Assert(t, c1)
	c2 := CaseInsensitiveContains(days, "fri")
	assert.Assert(t, !c2)
}

func TestContainsAny(t *testing.T) {
	c1 := []string{"dog", "cat", "hamster", "fish"}
	c2 := []string{"dog", "cat"}
	found := ContainsAny(c1, c2)
	assert.Assert(t, found)

	c3 := []string{"bird", "squirrel"}
	found = ContainsAny(c1, c3)
	assert.Assert(t, !found)
}

func TestStringElementsMatch(t *testing.T) {
	// Test with matching elements (same order)
	list1 := []string{"a", "b", "c"}
	list2 := []string{"a", "b", "c"}

	result := StringElementsMatch(list1, list2)
	assert.Assert(t, result)

	// Test with matching elements (different order)
	list3 := []string{"c", "a", "b"}
	result = StringElementsMatch(list1, list3)
	assert.Assert(t, result)

	// Test with different elements
	list4 := []string{"a", "b", "d"}
	result = StringElementsMatch(list1, list4)
	assert.Assert(t, !result)

	// Test with different lengths
	list5 := []string{"a", "b"}
	result = StringElementsMatch(list1, list5)
	assert.Assert(t, !result)

	// Test with empty lists
	result = StringElementsMatch([]string{}, []string{})
	assert.Assert(t, result)

	// Test with duplicates
	list6 := []string{"a", "b", "a"}
	list7 := []string{"a", "a", "b"}
	result = StringElementsMatch(list6, list7)
	assert.Assert(t, result)
}

func TestStringAppendIfMissing(t *testing.T) {
	// Test appending new element
	list := []string{"a", "b", "c"}
	result := StringAppendIfMissing(list, "d")

	assert.Assert(t, len(result) == 4)
	assert.Assert(t, Contains(result, "d"))

	// Test appending existing element (should not add)
	result = StringAppendIfMissing(list, "b")
	assert.Assert(t, len(result) == 3)

	// Test with empty list
	emptyList := []string{}
	result = StringAppendIfMissing(emptyList, "first")
	assert.Assert(t, len(result) == 1)
	assert.Assert(t, result[0] == "first")

	// Test case sensitivity
	result = StringAppendIfMissing(list, "B")
	assert.Assert(t, len(result) == 4)
	assert.Assert(t, Contains(result, "B"))
}

func TestPutIfValuePresent(t *testing.T) {
	// Test with non-empty value
	result := make(map[string]interface{})
	PutIfValuePresent(result, "key1", "value1")

	val, exists := result["key1"]
	assert.Assert(t, exists)
	assert.Assert(t, val == "value1")

	// Test with empty value (should not add)
	PutIfValuePresent(result, "key2", "")
	_, exists = result["key2"]
	assert.Assert(t, !exists)

	// Test with nil value (should not add)
	PutIfValuePresent(result, "key3", nil)
	_, exists = result["key3"]
	assert.Assert(t, !exists)

	// Test overwriting existing key with non-empty value
	PutIfValuePresent(result, "key1", "newvalue")
	val, exists = result["key1"]
	assert.Assert(t, exists)
	assert.Assert(t, val == "newvalue")

	// Test attempting to overwrite with empty value (should not change)
	PutIfValuePresent(result, "key1", "")
	val, exists = result["key1"]
	assert.Assert(t, exists)
	assert.Assert(t, val == "newvalue")
}

func TestNewStringSet(t *testing.T) {
	// Test with string slice
	input := []string{"a", "b", "c", "a"} // includes duplicate
	result := NewStringSet(input)

	assert.Assert(t, len(result) == 3) // should remove duplicate

	// Check if keys exist in the map
	_, aExists := result["a"]
	_, bExists := result["b"]
	_, cExists := result["c"]
	assert.Assert(t, aExists)
	assert.Assert(t, bExists)
	assert.Assert(t, cExists)

	// Test with empty slice
	emptyResult := NewStringSet([]string{})
	assert.Assert(t, len(emptyResult) == 0)

	// Test with single element
	singleResult := NewStringSet([]string{"single"})
	assert.Assert(t, len(singleResult) == 1)
	_, singleExists := singleResult["single"]
	assert.Assert(t, singleExists)
}
