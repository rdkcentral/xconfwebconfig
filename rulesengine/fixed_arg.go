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
	"fmt"
	"sort"

	copy "github.com/mitchellh/copystructure"
)

// TODO use fixed fields for simplicity for now
type Value struct {
	JLString string  `json:"java.lang.String,omitempty"`
	JLDouble float64 `json:"java.lang.Double,omitempty"`
}

type Bean struct {
	Value Value `json:"value"`
}

func (a *FixedArg) IsValid() bool {
	isCollection := &a.Collection != nil
	isBean := &a.Bean != nil
	if isCollection {
		if isBean {
			return false // cannot be both collection and bean
		}
		return isCollection
	}

	if isBean {
		isString := &a.Bean.Value.JLString != nil
		isDouble := &a.Bean.Value.JLDouble != nil
		if isString && isDouble {
			return false // cannot be both string and double
		}
		return isString || isDouble
	}

	return false
}

func (b *Bean) UnmarshalJSON(bbytes []byte) error {
	dict := make(map[string]interface{})
	err := json.Unmarshal(bbytes, &dict)
	if err != nil {
		return err
	}
	if val, ok := dict["value"]; ok {
		if innerItf, ok := val.(map[string]interface{}); ok {
			if jsvItf, ok := innerItf["java.lang.String"]; ok {
				b.Value.JLString = jsvItf.(string)
			}
			if jdvItf, ok := innerItf["java.lang.Double"]; ok {
				b.Value.JLDouble = jdvItf.(float64)
			}
		}
		if valueStr, ok := val.(string); ok {
			b.Value.JLString = valueStr
		}
	}
	return nil
}

type ValueList struct {
	JLArrayList []string `json:"java.lang.ArrayList,omitempty"`
}

type Collection struct {
	//Value ValueList `json:"value,omitempty"`
	Value []string `json:"value,omitempty"`
}

type FixedArg struct {
	Bean       Bean       `json:"bean"`
	Collection Collection `json:"collection"`
}

func NewFixedArg(itf interface{}) *FixedArg {
	switch ty := itf.(type) {
	case string:
		return &FixedArg{
			Bean: Bean{
				Value: Value{
					JLString: ty,
				},
			},
		}
	case float64:
		return &FixedArg{
			Bean: Bean{
				Value: Value{
					JLDouble: ty,
				},
			},
		}
	case []string:
		return &FixedArg{
			Collection: Collection{
				Value: ty,
			},
		}
	}

	return nil
}

func (a *FixedArg) GetValue() interface{} {
	if a == nil {
		return nil
	}
	if a.Collection.Value != nil && len(a.Collection.Value) > 0 {
		return a.Collection.Value
	}
	if len(a.Bean.Value.JLString) > 0 {
		return a.Bean.Value.JLString
	}
	if a.Bean.Value.JLDouble != float64(0) {
		return a.Bean.Value.JLDouble
	}
	return nil
}

func (a *FixedArg) IsCollectionValue() bool {
	if a.Collection.Value != nil && len(a.Collection.Value) > 0 {
		return true
	}
	return false
}

func (a *FixedArg) IsDoubleValue() bool {
	return &a.Bean != nil && &a.Bean.Value.JLDouble != nil
}

func (a *FixedArg) IsStringValue() bool {
	if len(a.Bean.Value.JLString) > 0 {
		return true
	}
	return false
}

func (a *FixedArg) String() string {
	if len(a.Collection.Value) > 0 {
		return fmt.Sprintf("'%v'", a.Collection.Value)
	}

	if len(a.Bean.Value.JLString) == 0 {
		// return fmt.Sprintf("FixedArg('%v')", a.Bean.Value.JLDouble)
		return fmt.Sprintf("'%v'", a.Bean.Value.JLDouble)
	}
	// return fmt.Sprintf("FixedArg('%v')", a.Bean.Value.JLString)
	return fmt.Sprintf("'%v'", a.Bean.Value.JLString)
}

func (a *FixedArg) Copy() *FixedArg {
	cloneObj, _ := copy.Copy(a)
	return cloneObj.(*FixedArg)
}

func (a *FixedArg) Equals(x *FixedArg) bool {
	if a.Collection.Value != nil && len(a.Collection.Value) > 0 && x.Collection.Value != nil && len(x.Collection.Value) > 0 {
		// Two Collections can be equal when their contents are same. Order does not matter.
		// Equality testing should not alter the objects being compared. So sort a copy of the objects
		atmp := a.Collection.Value
		xtmp := x.Collection.Value

		sort.Strings(atmp)
		sort.Strings(xtmp)

		atmpLen := removeDuplicates(atmp)
		xtmpLen := removeDuplicates(xtmp)
		if atmpLen != xtmpLen {
			return false
		}
		for i := 0; i < atmpLen; i++ {
			if atmp[i] != xtmp[i] {
				return false
			}
		}
		return true
	}
	return a.GetValue() == x.GetValue()
}

func removeDuplicates(s []string) int {
	if len(s) == 0 {
		return 0
	}
	j := 1
	for i := 1; i < len(s); i++ {
		if s[i] != s[i-1] {
			s[j] = s[i]
			j++
		}
	}
	return j
}
