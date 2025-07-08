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
