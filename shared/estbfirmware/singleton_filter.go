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
package estbfirmware

import (
	"encoding/json"
	"fmt"
	"strings"

	"xconfwebconfig/util"
)

const (
	PERCENT_FILTER_SINGLETON_ID     = "PERCENT_FILTER_VALUE"
	ROUND_ROBIN_FILTER_SINGLETON_ID = "DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE"
)

type SingletonFilterClass string

const (
	PercentFilterClass    SingletonFilterClass = "com.comcast.xconf.estbfirmware.PercentFilterValue"
	RoundRobinFilterClass SingletonFilterClass = "com.comcast.xconf.estbfirmware.DownloadLocationRoundRobinFilterValue"
)

// SingletonFilterValue table - this struct serves as a container for the two subtypes
type SingletonFilterValue struct {
	ID                                    string                                 `json:"id"`
	PercentFilterValue                    *PercentFilterValue                    `json:"-"`
	DownloadLocationRoundRobinFilterValue *DownloadLocationRoundRobinFilterValue `json:"-"`
}

func (obj *SingletonFilterValue) Clone() (*SingletonFilterValue, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*SingletonFilterValue), nil
}

func NewSingletonFilterValueInf() interface{} {
	return &SingletonFilterValue{}
}

func (sfv *SingletonFilterValue) IsPercentFilterValue() bool {
	return strings.HasSuffix(sfv.ID, PERCENT_FILTER_SINGLETON_ID)
}

func (sfv *SingletonFilterValue) IsDownloadLocationRoundRobinFilterValue() bool {
	return strings.HasSuffix(sfv.ID, ROUND_ROBIN_FILTER_SINGLETON_ID)
}

// UnmarshalJSON custom unmarshal to handle different subclass of SingletonFilterValue
func (sfv *SingletonFilterValue) UnmarshalJSON(bytes []byte) error {
	type singletonFilterValue SingletonFilterValue

	// Unmarshal just the base class to get the ID
	err := json.Unmarshal(bytes, (*singletonFilterValue)(sfv))
	if err != nil {
		return err
	}

	// Unmarshal the subtype based on the ID
	if sfv.IsPercentFilterValue() {
		var obj PercentFilterValue
		err = json.Unmarshal(bytes, &obj)
		if err != nil {
			return err
		}
		sfv.PercentFilterValue = &obj
	} else if sfv.IsDownloadLocationRoundRobinFilterValue() {
		var obj DownloadLocationRoundRobinFilterValue
		err = json.Unmarshal(bytes, &obj)
		if err != nil {
			return err
		}
		sfv.DownloadLocationRoundRobinFilterValue = &obj
	} else {
		return fmt.Errorf("Invalid ID for SingletonFilterValue: %v", string(bytes))
	}

	return nil
}
