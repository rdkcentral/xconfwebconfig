/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
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
package tests

import (
	"fmt"
	"reflect"
	"testing"

	ds "github.com/rdkcentral/xconfwebconfig/db"

	"gotest.tools/assert"
)

func TestCloneFunctionExists(t *testing.T) {
	tables := ds.GetAllTableInfo()
	assert.Assert(t, tables != nil)

	// Make sure Clone() function is defined for cached DAO object
	for _, table := range tables {
		if table.CacheData {
			obj := table.ConstructorFunc()
			assert.Assert(t, obj != nil)
			value := reflect.ValueOf(obj)
			method := value.MethodByName("Clone")
			valid := method.IsValid()
			if !valid {
				reflect.TypeOf(obj).Elem().Name()
				fmt.Println("Missing Clone function for DAO object:", reflect.TypeOf(obj).Elem().Name())
			}
			assert.Assert(t, valid)
		}
	}
}
