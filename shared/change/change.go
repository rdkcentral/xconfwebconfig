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
	"errors"
	"reflect"
	"strings"
	"xconfwebconfig/shared"
	"xconfwebconfig/shared/logupload"
	"xconfwebconfig/util"
)

// EntityType enum
type EntityType string

const (
	TelemetryProfile EntityType = "TELEMETRY_PROFILE"
)

// ChangeOperation enum
type ChangeOperation string

// TelemetryTwoChange XconfApprovedTelemetryTwoChange table
type ApprovedTelemetryTwoChange TelemetryTwoChange

// TelemetryTwoChange XconfTelemetryTwoChange table
type TelemetryTwoChange struct {
	ID              string                         `json:"id"`
	Updated         int64                          `json:"updated"`
	EntityID        string                         `json:"entityId"`
	EntityType      string                         `json:"entityType"`
	ApplicationType string                         `json:"applicationType"`
	NewEntity       *logupload.TelemetryTwoProfile `json:"newEntity,omitempty"`
	OldEntity       *logupload.TelemetryTwoProfile `json:"oldEntity,omitempty"`
	Operation       ChangeOperation                `json:"operation"`
	Author          string                         `json:"author"`
	ApprovedUser    string                         `json:"approvedUser,omitempty"`
}

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

func (c Change) GetID() string {
	return c.ID
}

func (c Change) GetEntityID() string {
	return c.EntityID
}

func (c Change) GetEntityType() EntityType {
	return c.EntityType
}

func (c Change) GetApplicationType() string {
	return c.ApplicationType
}

func (c Change) GetNewEntity() *logupload.PermanentTelemetryProfile {
	return &c.NewEntity
}

func (c Change) GetOldEntity() *logupload.PermanentTelemetryProfile {
	return &c.OldEntity
}

func (c Change) GetOperation() ChangeOperation {
	return c.Operation
}

func (c Change) GetAuthor() string {
	return c.Author
}

func (c Change) GetApprovedUser() string {
	return c.ApprovedUser
}

func (c ApprovedChange) GetID() string {
	return c.ID
}

func (c ApprovedChange) GetEntityID() string {
	return c.EntityID
}

func (c ApprovedChange) GetEntityType() EntityType {
	return c.EntityType
}

func (c ApprovedChange) GetApplicationType() string {
	return c.ApplicationType
}

func (c ApprovedChange) GetNewEntity() *logupload.PermanentTelemetryProfile {
	return &c.NewEntity
}

func (c ApprovedChange) GetOldEntity() *logupload.PermanentTelemetryProfile {
	return &c.OldEntity
}

func (c ApprovedChange) GetOperation() ChangeOperation {
	return c.Operation
}

func (c ApprovedChange) GetAuthor() string {
	return c.Author
}

func (c ApprovedChange) GetApprovedUser() string {
	return c.ApprovedUser
}

func (obj *TelemetryTwoChange) Validate() error {
	if util.IsBlank(obj.EntityID) {
		return errors.New("Entity id is empty")
	}
	if util.IsBlank(obj.Author) {
		return errors.New("Author is empty")
	}
	if util.IsBlank(string(obj.Operation)) {
		return errors.New("Operation is empty")
	}
	if (obj.Operation == Create || obj.Operation == Update) && obj.NewEntity == nil {
		return errors.New("New entity is empty")
	}
	if (obj.Operation == Update || obj.Operation == Delete) && obj.OldEntity == nil {
		return errors.New("Old entity is empty")
	}
	return nil
}

func (obj *ApprovedTelemetryTwoChange) Validate() error {
	change := TelemetryTwoChange(*obj)
	if err := change.Validate(); err != nil {
		return err
	}
	if util.IsBlank(obj.ApprovedUser) {
		return errors.New("Approved user is empty")
	}
	return nil
}

type PendingChange interface {
	GetID() string
	GetEntityID() string
	GetEntityType() EntityType
	GetApplicationType() string
	GetNewEntity() *logupload.PermanentTelemetryProfile
	GetOldEntity() *logupload.PermanentTelemetryProfile
	GetOperation() ChangeOperation
	GetAuthor() string
	GetApprovedUser() string
}

func (c *Change) EqualChangeData(c2 *Change) bool {
	if c == c2 {
		return true
	}
	return c.EntityType == c2.EntityType &&
		c.ApplicationType == c2.ApplicationType &&
		reflect.DeepEqual(c.NewEntity, c2.NewEntity) &&
		reflect.DeepEqual(c.OldEntity, c2.OldEntity) &&
		c.Operation == c2.Operation
}

// NewApprovedTelemetryTwoChangeInf constructor
func NewApprovedTelemetryTwoChangeInf() interface{} {
	return &ApprovedTelemetryTwoChange{
		ApplicationType: shared.STB,
	}
}

// NewTelemetryTwoChangeInf constructor
func NewTelemetryTwoChangeInf() interface{} {
	return &TelemetryTwoChange{
		ApplicationType: shared.STB,
	}
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
