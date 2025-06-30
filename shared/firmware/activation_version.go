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
package firmware

const (
	SINGLETON_ID = "DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE"
)

type ActivationVersion struct {
	ID                 string   `json:"id"`
	ApplicationType    string   `json:"applicationType,omitempty"`
	Description        string   `json:"description,omitempty"`
	Model              string   `json:"model,omitempty"`
	PartnerId          string   `json:"partnerId,omitempty"`
	RegularExpressions []string `json:"regularExpressions"`
	FirmwareVersions   []string `json:"firmwareVersions"`
}

// setApplicationType implements queries.T.
func (obj *ActivationVersion) SetApplicationType(appType string) {
	obj.ApplicationType = appType
}

// getApplicationType implements queries.T.
func (obj *ActivationVersion) GetApplicationType() string {
	return obj.ApplicationType
}

// NewActivationVersion constructor
func NewActivationVersion() *ActivationVersion {
	return &ActivationVersion{
		RegularExpressions: []string{},
		FirmwareVersions:   []string{}}
}
