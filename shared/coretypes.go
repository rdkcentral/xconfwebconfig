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
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

const (
	STB      = "stb"
	XHOME    = "xhome"
	RDKCLOUD = "rdkcloud"
	SKY      = "sky"
	ALL      = "all"
)

// AppSettings table object
type AppSetting struct {
	ID      string      `json:"id"`
	Updated int64       `json:"updated"`
	Value   interface{} `json:"value"`
}

// ApplicationType table object
type ApplicationType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedBy   string `json:"createdBy"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt,omitempty"`
}

func (obj *ApplicationType) Clone() (*ApplicationType, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*ApplicationType), nil
}

func ValidateApplicationType(applicationType string) error {
	if applicationType == "" {
		return common.NewRemoteError(http.StatusBadRequest, "ApplicationType is empty")
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
func NewApplicationTypeInf() interface{} {
	return &ApplicationType{}
}

// NewEnvironmentInf constructor
func NewEnvironmentInf() interface{} {
	return &Environment{}
}

// InitializeStaticApplicationTypes creates predefined application types on first run only
func InitializeStaticApplicationTypes() error {
	dao := db.GetCachedSimpleDao()
	allTypes, err := dao.GetAllAsList(db.TABLE_APPLICATION_TYPES, 0)
	if err == nil && len(allTypes) > 0 {
		log.Info("Application types already initialized, skipping...")
		return nil
	}

	log.Info("Creating static application types...")
	staticTypes := map[string]string{
		STB:   "Set-Top Box application type",
		XHOME: "Home security and automation application type",
		SKY:   "Sky platform application type",
	}

	timestamp := time.Now().Unix()
	for name, description := range staticTypes {
		uid := uuid.New().String()
		appType := &ApplicationType{
			ID:          uid,
			Name:        name,
			Description: description,
			CreatedBy:   "system",
			CreatedAt:   timestamp,
		}
		if err := dao.SetOne(db.TABLE_APPLICATION_TYPES, uid, appType); err != nil {
			return fmt.Errorf("failed to create static application type '%s': %w", name, err)
		}
		log.Infof("Created static application type: %s ", name)
	}
	return nil
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
		log.Warn(fmt.Sprintf("no environment found for:%s ", id))
		return nil
	}
	return inst.(*Environment)
}

func SetOneEnvironment(env *Environment) (*Environment, error) {
	env.Updated = util.GetTimestamp()
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
	if len(obj.ID) > 0 {
		match, _ := regexp.MatchString("^[-a-zA-Z0-9_.' ]+$", obj.ID)
		if match {
			return nil
		}
	}

	return errors.New("Id is invalid. Valid Characters: alphanumeric _ . -")
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
		log.Warn(fmt.Sprintf("no model found for:%s ", id))
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

func (obj *AppSetting) Clone() (*AppSetting, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*AppSetting), nil
}

// NewAppSettingInf constructor
func NewAppSettingInf() interface{} {
	return &AppSetting{}
}

func GetBooleanAppSetting(key string, vargs ...bool) bool {
	defaultVal := false
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn(fmt.Sprintf("no AppSetting found for %s", key))
		return defaultVal
	}

	setting := inst.(*AppSetting)
	return setting.Value.(bool)
}

func GetIntAppSetting(key string, vargs ...int) int {
	defaultVal := -1
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn(fmt.Sprintf("no AppSetting found for %s", key))
		return defaultVal
	}

	setting := inst.(*AppSetting)

	// Note: json.Unmarshal numbers into float64 when target type is of type interface{}
	if val, ok := setting.Value.(float64); ok {
		return int(val)
	} else {
		return setting.Value.(int)
	}
}

func GetFloat64AppSetting(key string, vargs ...float64) float64 {
	defaultVal := -1.0
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn(fmt.Sprintf("no AppSetting found for %s", key))
		return defaultVal
	}

	setting := inst.(*AppSetting)
	return setting.Value.(float64)
}

func GetTimeAppSetting(key string, vargs ...time.Time) time.Time {
	var defaultVal time.Time
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn(fmt.Sprintf("no AppSetting found for %s", key))
		return defaultVal
	}

	setting := inst.(*AppSetting)
	timeStr := setting.Value.(string)
	time, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		log.Error(fmt.Sprintf("error getting AppSetting for %s: %s ", key, err.Error()))
	}

	return time
}

func GetStringAppSetting(key string, vargs ...string) string {
	defaultVal := ""
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn(fmt.Sprintf("no AppSetting found for:%s ", key))
		return defaultVal
	}

	setting := inst.(*AppSetting)
	return setting.Value.(string)
}

func GetAppSettings() (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	list, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_APP_SETTINGS, 0)
	if err != nil {
		return settings, err
	}
	for _, v := range list {
		p := *v.(*AppSetting)
		settings[p.ID] = p.Value
	}
	return settings, nil
}

func SetAppSetting(key string, value interface{}) (*AppSetting, error) {
	setting := AppSetting{
		ID:      key,
		Updated: util.GetTimestamp(time.Now()),
		Value:   value,
	}

	err := db.GetCachedSimpleDao().SetOne(db.TABLE_APP_SETTINGS, setting.ID, &setting)
	if err != nil {
		return nil, err
	}
	return &setting, nil
}
