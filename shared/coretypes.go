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
package shared

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"xconfwebconfig/db"
	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

const (
	STB      = "stb"
	XHOME    = "xhome"
	RDKCLOUD = "rdkcloud"
	SKY      = "sky"
	ALL      = "all"
)

func isValid(at string) bool {
	if at == STB || at == XHOME || at == RDKCLOUD || at == SKY {
		return true
	}
	return false
}

func ValidateApplicationType(applicationType string) error {
	if applicationType != "" && !isValid(applicationType) {
		return fmt.Errorf("ApplicationType %s is not valid", applicationType)
	}
	return nil
}

const (
	TABLE_LOGS_KEY2_FIELD_NAME = "column1"
	LAST_CONFIG_LOG_ID         = "0"
)

const (
	StbContextTime      = "time"
	StbContextModel     = "model"
	MacList             = "MAC_LIST"
	IpList              = "IP_LIST"
	TableGenericNSList  = "GenericXconfNamedList"
	TableFirmwareConfig = "FirmwareConfig"
	TableFirmwareRule   = "FirmwareRule4"
)

const (
	Tftp  = "tftp"
	Http  = "http"
	Https = "https"
)

// XEnvModel is ...
type XEnvModel interface {
	GetId() string
	GetDescription() string
}

// Environment table object
type Environment struct {
	ID          string `json:"id"`
	Updated     int64  `json:"updated"`
	Description string `json:"description"`
}

func (obj *Environment) Clone() (*Environment, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*Environment), nil
}

func (obj *Environment) Validate() error {
	if len(strings.TrimSpace(obj.ID)) > 0 {
		match, _ := regexp.MatchString("^[-a-zA-Z0-9_.' ]+$", obj.ID)
		if match {
			return nil
		}
	}

	return errors.New("Id is invalid")
}

// NewEnvironmentInf constructor
func NewEnvironmentInf() interface{} {
	return &Environment{}
}

// NewEnvironment ...
func NewEnvironment(id string, description string) *Environment {
	if id != "" {
		id = strings.ToUpper(strings.TrimSpace(id))
	}

	return &Environment{
		ID:          id,
		Description: description,
	}
}

type EnvironmentResponse struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
}

func (e *Environment) CreateEnvironmentResponse() *EnvironmentResponse {
	return &EnvironmentResponse{
		ID:          e.ID,
		Description: e.Description,
	}
}

func GetAllEnvironmentList() []*Environment {
	result := []*Environment{}
	list, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_ENVIRONMENT, 0)
	if err != nil {
		log.Warn("no environment found")
		return result
	}
	for _, inst := range list {
		env := inst.(*Environment)
		result = append(result, env)
	}
	return result
}

func GetOneEnvironment(id string) *Environment {
	inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_ENVIRONMENT, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no environment found for " + id))
		return nil
	}
	return inst.(*Environment)
}

func SetOneEnvironment(env *Environment) (*Environment, error) {
	env.Updated = util.GetTimestamp(time.Now())
	err := db.GetCachedSimpleDao().SetOne(db.TABLE_ENVIRONMENT, env.ID, env)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func DeleteOneEnvironment(id string) error {
	err := db.GetCachedSimpleDao().DeleteOne(db.TABLE_ENVIRONMENT, id)
	if err != nil {
		return err
	}
	return nil
}

// Model table object
type Model struct {
	ID          string `json:"id"`
	Updated     int64  `json:"updated"`
	Description string `json:"description"`
}

func (obj *Model) Clone() (*Model, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*Model), nil
}

func (obj *Model) Validate() error {
	if len(strings.TrimSpace(obj.ID)) > 0 {
		match, _ := regexp.MatchString("^[-a-zA-Z0-9_.' ]+$", obj.ID)
		if match {
			return nil
		}
	}

	return errors.New("Id is invalid")
}

// NewModelInf constructor
func NewModelInf() interface{} {
	return &Model{}
}

// NewModel ...
func NewModel(id string, description string) *Model {
	return &Model{
		ID:          strings.ToUpper(id),
		Description: description,
	}
}

func GetAllModelList() []*Model {
	result := []*Model{}
	list, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_MODEL, 0)
	if err != nil {
		log.Warn("no model found")
		return result
	}
	for _, inst := range list {
		model := inst.(*Model)
		result = append(result, model)
	}
	return result
}

func GetOneModel(id string) *Model {
	inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_MODEL, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no model found for " + id))
		return nil
	}
	return inst.(*Model)
}

func SetOneModel(model *Model) (*Model, error) {
	model.Updated = util.GetTimestamp(time.Now())
	err := db.GetCachedSimpleDao().SetOne(db.TABLE_MODEL, model.ID, model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func DeleteOneModel(id string) error {
	err := db.GetCachedSimpleDao().DeleteOne(db.TABLE_MODEL, id)
	if err != nil {
		return err
	}
	return nil
}

func IsExistModel(id string) bool {
	if !util.IsBlank(id) {
		inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_MODEL, id)
		if inst != nil && err == nil {
			return true
		}
	}
	return false
}

type ModelResponse struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
}

func (m *Model) CreateModelResponse() *ModelResponse {
	return &ModelResponse{
		ID:          m.ID,
		Description: m.Description,
	}
}

// StringListWrapper ...
type StringListWrapper struct {
	List []string `json:"list"`
}

func NewStringListWrapper(list []string) *StringListWrapper {
	return &StringListWrapper{List: list}
}
