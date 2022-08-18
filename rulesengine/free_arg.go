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
	"fmt"
)

type FreeArg struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// NOTE: public FreeArg() {} not used, hence skipped

func NewFreeArg(ttype string, name string) *FreeArg {
	return &FreeArg{
		Type: ttype,
		Name: name,
	}
}

func (a *FreeArg) GetType() string {
	return a.Type
}

func (a *FreeArg) SetType(ttype string) {
	a.Type = ttype
}

func (a *FreeArg) GetName() string {
	return a.Name
}

func (a *FreeArg) SetName(name string) {
	a.Name = name
}

func (a *FreeArg) String() string {
	//return fmt.Sprintf("FreeArg(Type='%v', Name='%v')", a.Type, a.Name)

	if a.Type == "STRING" {
		// return fmt.Sprintf("FreeArg('%v')", a.Name)
		return fmt.Sprintf("'%v'", a.Name)
	} else {
		// return fmt.Sprintf("FreeArg('%v(%v)')", a.Name, a.Type)
		return fmt.Sprintf("'%v(%v)'", a.Name, a.Type)
	}
}

func (a *FreeArg) Copy() *FreeArg {
	return NewFreeArg(a.GetType(), a.GetName())
}

func (a *FreeArg) Equals(x *FreeArg) bool {
	return a.GetName() == x.GetName() && a.GetType() == x.GetType()
}
