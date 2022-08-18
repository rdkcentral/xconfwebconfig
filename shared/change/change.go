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
package change

import (
	"strings"

	"xconfwebconfig/shared/logupload"
)

// EntityType enum
type EntityType string

const (
	TelemetryProfile EntityType = "TELEMETRY_PROFILE"
)

// ChangeOperation enum
type ChangeOperation string

const (
	Create ChangeOperation = "CREATE"
	Update                 = "UPDATE"
	Delete                 = "DELETE"
)

type EntityInterface interface {
	getName() string
}

// Change XconfChange table
type Change struct {
	ID              string                              `json:"id"`
	Updated         int64                               `json:"updated"`
	EntityID        string                              `json:"entityId"`
	EntityType      EntityType                          `json:"entityType"`
	ApplicationType string                              `json:"applicationType"`
	NewEntity       logupload.PermanentTelemetryProfile `json:"newEntity"`
	OldEntity       logupload.PermanentTelemetryProfile `json:"oldEntity"`
	Operation       ChangeOperation                     `json:"operation"`
	Author          string                              `json:"author"`
	ApprovedUser    string                              `json:"approvedUser"`
}

// NewChangeInf constructor
func NewChangeInf() interface{} {
	return &Change{}
}

func (c *Change) byAuthor(author string) bool {
	return strings.EqualFold(c.Author, author)
}

func (c *Change) byTelemetryProfileName(name string) bool {
	return strings.EqualFold(c.NewEntity.Name, name) || strings.EqualFold(c.OldEntity.Name, name)
}

// ApprovedChange XconfApprovedChange table
type ApprovedChange Change

// NewApprovedChangeInf constructor
func NewApprovedChangeInf() interface{} {
	return &ApprovedChange{}
}
