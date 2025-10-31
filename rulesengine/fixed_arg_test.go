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
	"encoding/json"
	"testing"

	"gotest.tools/assert"
)

func TestFixedArgsEquality(t *testing.T) {
	s := [...]string{
		`{
			"bean": {
				"value": {
					"java.lang.String": "AA:AA:AA:AA:AA:AA"
				}
			}
		}`,
		`{
			"collection": {
				"value": [
				  "TG1234",
				  "TG1234"
				]
			  }		
		}`,
		`{
			"bean": {
				"value": {
					"java.lang.Double": 40.0
				}
			}
		}`,
	}

	var fixedArgs []FixedArg
	for _, raw := range s {
		var f FixedArg
		err := json.Unmarshal([]byte(raw), &f)
		assert.NilError(t, err)
		fixedArgs = append(fixedArgs, f)
	}

	for i, f1 := range fixedArgs {
		for j, f2 := range fixedArgs {
			if i == j {
				assert.Assert(t, f1.Equals(&f2))
			} else {
				assert.Assert(t, !f1.Equals(&f2))
			}
		}
	}
}

func TestFixedArgCollection(t *testing.T) {
	baseF := FixedArg{
		Collection: &Collection{
			Value: []string{
				"one",
				"two",
			},
		},
	}

	type TestInput struct {
		input    FixedArg
		expected bool
	}

	var testData = []TestInput{
		TestInput{
			input: FixedArg{
				Collection: &Collection{
					Value: []string{
						"one",
						"two",
					},
				},
			},
			expected: true,
		},
		TestInput{
			input: FixedArg{
				Collection: &Collection{
					Value: []string{"two", "one"},
				},
			},
			expected: true,
		},
		TestInput{
			input: FixedArg{
				Collection: &Collection{
					Value: []string{"one", "two", "three"},
				},
			},
			expected: false,
		},
		TestInput{
			input: FixedArg{
				Collection: &Collection{
					Value: []string{"one"},
				},
			},
			expected: false,
		},
		TestInput{
			input: FixedArg{
				Collection: &Collection{
					Value: []string{"different"},
				},
			},
			expected: false,
		},
		TestInput{
			input: FixedArg{
				Collection: &Collection{
					Value: []string{"one", "two", "one"},
				},
			},
			expected: true,
		},
	}
	for _, datum := range testData {
		if datum.expected {
			assert.Assert(t, baseF.Equals(&datum.input))
		} else {
			assert.Assert(t, !baseF.Equals(&datum.input))
		}
	}
}

func TestFixedArg_IsValid(t *testing.T) {
	testCases := []struct {
		name     string
		fixedArg *FixedArg
		expected bool
	}{
		{
			name: "Valid string bean",
			fixedArg: &FixedArg{
				Bean: &Bean{
					Value: Value{
						JLString: stringPtr("test"),
					},
				},
			},
			expected: true,
		},
		{
			name: "Valid double bean",
			fixedArg: &FixedArg{
				Bean: &Bean{
					Value: Value{
						JLDouble: float64Ptr(42.0),
					},
				},
			},
			expected: true,
		},
		{
			name: "Valid collection",
			fixedArg: &FixedArg{
				Collection: &Collection{
					Value: []string{"item1", "item2"},
				},
			},
			expected: true,
		},
		{
			name: "Invalid - both collection and bean",
			fixedArg: &FixedArg{
				Bean: &Bean{
					Value: Value{
						JLString: stringPtr("test"),
					},
				},
				Collection: &Collection{
					Value: []string{"item1"},
				},
			},
			expected: false,
		},
		{
			name: "Invalid - both string and double in bean",
			fixedArg: &FixedArg{
				Bean: &Bean{
					Value: Value{
						JLString: stringPtr("test"),
						JLDouble: float64Ptr(42.0),
					},
				},
			},
			expected: false,
		},
		{
			name: "Invalid - empty bean",
			fixedArg: &FixedArg{
				Bean: &Bean{
					Value: Value{},
				},
			},
			expected: false,
		},
		{
			name: "Invalid - nil everything",
			fixedArg: &FixedArg{},
			expected: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.fixedArg.IsValid()
			assert.Equal(t, tc.expected, result, "IsValid result should match expected")
		})
	}
}

func TestFixedArg_IsDoubleValue(t *testing.T) {
	testCases := []struct {
		name     string
		fixedArg *FixedArg
		expected bool
	}{
		{
			name:     "Double value",
			fixedArg: NewFixedArg(float64(42.5)),
			expected: true,
		},
		{
			name:     "String value",
			fixedArg: NewFixedArg("test"),
			expected: false,
		},
		{
			name: "Collection value",
			fixedArg: NewFixedArg([]string{"item1", "item2"}),
			expected: false,
		},
		{
			name: "Nil bean",
			fixedArg: &FixedArg{},
			expected: false,
		},
		{
			name: "Empty bean value",
			fixedArg: &FixedArg{
				Bean: &Bean{
					Value: Value{},
				},
			},
			expected: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.fixedArg.IsDoubleValue()
			assert.Equal(t, tc.expected, result, "IsDoubleValue result should match expected")
		})
	}
}

func TestFixedArg_String(t *testing.T) {
	testCases := []struct {
		name     string
		fixedArg *FixedArg
		expected string
	}{
		{
			name:     "String value",
			fixedArg: NewFixedArg("hello"),
			expected: "'hello'",
		},
		{
			name:     "Double value",
			fixedArg: NewFixedArg(float64(42.5)),
			expected: "'42.5'",
		},
		{
			name:     "Collection value",
			fixedArg: NewFixedArg([]string{"item1", "item2"}),
			expected: "'[item1 item2]'",
		},
		{
			name: "Nil value",
			fixedArg: &FixedArg{},
			expected: "''",
		},
		{
			name: "Empty bean",
			fixedArg: &FixedArg{
				Bean: &Bean{
					Value: Value{},
				},
			},
			expected: "''",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.fixedArg.String()
			assert.Equal(t, tc.expected, result, "String result should match expected")
		})
	}
}

func TestCollection_MarshalJSON(t *testing.T) {
	testCases := []struct {
		name       string
		collection *Collection
		expected   string
	}{
		{
			name: "Normal collection",
			collection: &Collection{
				Value: []string{"item1", "item2"},
			},
			expected: `{"value":["item1","item2"]}`,
		},
		{
			name: "Empty collection",
			collection: &Collection{
				Value: []string{},
			},
			expected: `{"value":[]}`,
		},
		{
			name: "Nil value collection",
			collection: &Collection{
				Value: nil,
			},
			expected: `{"value":[]}`,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.collection.MarshalJSON()
			assert.NilError(t, err, "MarshalJSON should not return error")
			assert.Equal(t, tc.expected, string(result), "JSON should match expected")
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func TestCollection_MarshalJSON_Extended(t *testing.T) {
	// Test with nil value
	t.Run("Nil_value", func(t *testing.T) {
		collection := &Collection{Value: nil}
		
		result, err := collection.MarshalJSON()
		assert.NilError(t, err)
		assert.Equal(t, `{"value":[]}`, string(result)) // Should output empty array for nil
	})

	// Test with string slice
	t.Run("String_slice", func(t *testing.T) {
		collection := &Collection{Value: []string{"item1", "item2", "item3"}}
		
		result, err := collection.MarshalJSON()
		assert.NilError(t, err)
		expected := `{"value":["item1","item2","item3"]}`
		assert.Equal(t, expected, string(result))
	})

	// Test with empty slice
	t.Run("Empty_slice", func(t *testing.T) {
		collection := &Collection{Value: []string{}}
		
		result, err := collection.MarshalJSON()
		assert.NilError(t, err)
		assert.Equal(t, `{"value":[]}`, string(result))
	})

	// Test with single item slice
	t.Run("Single_item", func(t *testing.T) {
		collection := &Collection{Value: []string{"single"}}
		
		result, err := collection.MarshalJSON()
		assert.NilError(t, err)
		assert.Equal(t, `{"value":["single"]}`, string(result))
	})
}
