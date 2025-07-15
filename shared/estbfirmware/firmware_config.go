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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// ConfigNames ...
type ConfigNames string

const (
	ID                               = "id"
	UPDATED                          = "updated"
	DESCRIPTION                      = "description"
	SUPPORTED_MODEL_IDS              = "supportedModelIds"
	FIRMWARE_DOWNLOAD_PROTOCOL       = "firmwareDownloadProtocol"
	FIRMWARE_FILENAME                = "firmwareFilename"
	FIRMWARE_LOCATION                = "firmwareLocation"
	FIRMWARE_VERSION                 = "firmwareVersion"
	IPV6_FIRMWARE_LOCATION           = "ipv6FirmwareLocation"
	UPGRADE_DELAY                    = "upgradeDelay"
	REBOOT_IMMEDIATELY               = "rebootImmediately"
	APPLICATION_TYPE                 = "applicationType"
	MANDATORY_UPDATE                 = "mandatoryUpdate"
	MAX_ALLOWED_NUMBER_OF_PROPERTIES = 20
)

type Capabilities string

const (
	/**
	 * RCDL indicates that the STB is capable of performing HTTP firmware downloads using DNS resolved URIs.
	 * The download will run in the eSTB. Until this, the eCM performed the download. eCM does not have
	 * DNS and thus requires an IP address.
	 * <p>
	 * Warning!!! Due to a bug in STB code, we have RNG150 boxes that send this parameter but they are NOT
	 * able to do HTTP firmware downloads. To handle this situation we are implementing a hack where
	 * RNG150 boxes will always be told to do TFTP. Once we are confident that all RNG150s have been updated
	 * to versions that actually do support HTTP, we will turn off the hack.
	 */
	RCDL Capabilities = "RCDL"
	/**
	 * lets Xconf know that reboot has been decoupled from firmware download. If not specified in the
	 * rebootImmediately response, the STB will still reboot immediately after firmware download.
	 */
	RebootDecoupled = "rebootDecoupled"
	RebootCoupled   = "rebootDecoupled"
	/**
	 * Lets Xconf know that the STB can accept a full URL for location rather than just an IP address or
	 * domain name.
	 */
	SupportsFullHttpUrl = "supportsFullHttpUrl"
)

type Expression struct {
	TargetedModelIds []string               `json:"targetedModelIds"`
	EnvironmentId    string                 `json:"environmentId,omitempty"`
	ModelId          string                 `json:"modelId,omitempty"`
	IpAddressGroup   *shared.IpAddressGroup `json:"ipAddressGroup,omitempty"`
}

type FirmwareConfigForMacRuleBeanResponse struct {
	ID                       string            `json:"id"`
	Updated                  int64             `json:"updated,omitempty"`
	Description              string            `json:"description"`
	SupportedModelIds        []string          `json:"supportedModelIds"`
	FirmwareFilename         string            `json:"firmwareFilename"`
	FirmwareVersion          string            `json:"firmwareVersion"`
	ApplicationType          string            `json:"applicationType,omitempty"`
	FirmwareDownloadProtocol string            `json:"firmwareDownloadProtocol,omitempty"`
	FirmwareLocation         string            `json:"firmwareLocation,omitempty"`
	Ipv6FirmwareLocation     string            `json:"ipv6FirmwareLocation,omitempty"`
	UpgradeDelay             int64             `json:"upgradeDelay,omitempty"`
	RebootImmediately        bool              `json:"rebootImmediately,omitempty"`
	MandatoryUpdate          bool              `json:"-"`
	Properties               map[string]string `json:"properties,omitempty"`
}

// FirmwareConfig table
type FirmwareConfig struct {
	ID                       string            `json:"id"`
	Updated                  int64             `json:"updated,omitempty"`
	Description              string            `json:"description"`
	SupportedModelIds        []string          `json:"supportedModelIds"`
	FirmwareFilename         string            `json:"firmwareFilename"`
	FirmwareVersion          string            `json:"firmwareVersion"`
	ApplicationType          string            `json:"applicationType,omitempty"`
	FirmwareDownloadProtocol string            `json:"firmwareDownloadProtocol,omitempty"`
	FirmwareLocation         string            `json:"firmwareLocation,omitempty"`
	Ipv6FirmwareLocation     string            `json:"ipv6FirmwareLocation,omitempty"`
	UpgradeDelay             int64             `json:"upgradeDelay,omitempty"`
	RebootImmediately        bool              `json:"rebootImmediately"`
	MandatoryUpdate          bool              `json:"mandatoryUpdate,omitempty"`
	Properties               map[string]string `json:"properties,omitempty"`
}

func (obj *FirmwareConfig) SetApplicationType(appType string) {
	obj.ApplicationType = appType
}

func (obj *FirmwareConfig) GetApplicationType() string {
	return obj.ApplicationType
}

type MacRuleBeanResponse struct {
	Id               string                                `json:"id,omitempty"`
	Name             string                                `json:"name,omitempty"`
	MacAddresses     string                                `json:"macAddresses,omitempty"`
	MacListRef       string                                `json:"macListRef,omitempty"`
	FirmwareConfig   *FirmwareConfigForMacRuleBeanResponse `json:"firmwareConfig"`
	TargetedModelIds *[]string                             `json:"targetedModelIds,omitempty"`
	MacList          *[]string                             `json:"macList,omitempty"`
}

func MacRuleBeanToMacRuleBeanResponse(macRuleBean *MacRuleBean) *MacRuleBeanResponse {
	response := MacRuleBeanResponse{}
	response.Id = macRuleBean.Id
	response.Name = macRuleBean.Name
	response.MacAddresses = macRuleBean.MacAddresses
	response.MacListRef = macRuleBean.MacListRef
	response.TargetedModelIds = macRuleBean.TargetedModelIds
	response.MacList = macRuleBean.MacList
	response.FirmwareConfig = nil
	if macRuleBean.FirmwareConfig != nil {
		response.FirmwareConfig = FirmwareConfigToFirmwareConfigForMacRuleBeanResponse(macRuleBean.FirmwareConfig)
	}
	return &response
}

func FirmwareConfigToFirmwareConfigForMacRuleBeanResponse(firmwareConfig *FirmwareConfig) *FirmwareConfigForMacRuleBeanResponse {
	response := FirmwareConfigForMacRuleBeanResponse{}
	response.ID = firmwareConfig.ID
	response.Updated = firmwareConfig.Updated
	response.Description = firmwareConfig.Description
	response.SupportedModelIds = firmwareConfig.SupportedModelIds
	response.FirmwareFilename = firmwareConfig.FirmwareFilename
	response.FirmwareVersion = firmwareConfig.FirmwareVersion
	response.ApplicationType = firmwareConfig.ApplicationType
	response.FirmwareDownloadProtocol = firmwareConfig.FirmwareDownloadProtocol
	response.FirmwareLocation = firmwareConfig.FirmwareLocation
	response.Ipv6FirmwareLocation = firmwareConfig.Ipv6FirmwareLocation
	response.UpgradeDelay = firmwareConfig.UpgradeDelay
	response.RebootImmediately = firmwareConfig.RebootImmediately
	response.MandatoryUpdate = firmwareConfig.MandatoryUpdate
	response.Properties = firmwareConfig.Properties
	return &response
}

func (obj *FirmwareConfig) Clone() (*FirmwareConfig, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*FirmwareConfig), nil
}

func (obj *FirmwareConfig) Validate() error {
	if obj == nil {
		return errors.New("Firmware config is not present")
	}
	if util.IsBlank(obj.Description) {
		return errors.New("Description is empty")
	}
	if util.IsBlank(obj.FirmwareFilename) {
		return errors.New("File name is empty")
	}
	if util.IsBlank(obj.FirmwareVersion) {
		return errors.New("Version is empty")
	}
	if len(obj.SupportedModelIds) == 0 {
		return errors.New("Supported model list is empty")
	}

	for _, modelId := range obj.SupportedModelIds {
		if !shared.IsExistModel(modelId) {
			return fmt.Errorf("Model: %s does not exist", modelId)
		}
	}

	if !util.IsBlank(obj.FirmwareDownloadProtocol) {
		downloadProtocols := []string{"tftp", "http", "https"}
		if !util.Contains(downloadProtocols, obj.FirmwareDownloadProtocol) {
			return fmt.Errorf("FirmwareDownloadProtocol must be one of %v ", downloadProtocols)
		}
	}

	if len(obj.Properties) > MAX_ALLOWED_NUMBER_OF_PROPERTIES {
		return fmt.Errorf("Max allowed number of properties is %v", MAX_ALLOWED_NUMBER_OF_PROPERTIES)
	}

	for k, v := range obj.Properties {
		if util.IsBlank(k) {
			return errors.New("Key is empty")
		}
		if util.IsBlank(v) {
			return fmt.Errorf("Value is blank for key: %s", k)
		}
	}

	err := shared.ValidateApplicationType(obj.ApplicationType)
	if err != nil {
		return err
	}

	return nil
}

func (obj *FirmwareConfig) ValidateName() error {
	list, err := GetFirmwareConfigAsListDB()
	if err != nil {
		return err
	}

	for _, config := range list {
		if config == nil || obj.ID == config.ID || obj.ApplicationType != config.ApplicationType {
			continue
		}
		if strings.ToUpper(config.Description) == strings.ToUpper(obj.Description) {
			return errors.New("This description is already used in " + config.ID)
		}
	}

	return nil
}

type FirmwareConfigFacadeResponse map[string]interface{}

func CreateFirmwareConfigFacadeResponse(firmwareConfigFacade FirmwareConfigFacade) FirmwareConfigFacadeResponse {
	firmwareConfigResponse := map[string]interface{}{}
	for k, v := range firmwareConfigFacade.Properties {
		if k == "upgradeDelay" {
			int64v, ok := v.(int64)
			if ok && int64v != 0 {
				firmwareConfigResponse[k] = v
			}
		} else if !IsRedundantEntry(k) && v != "" {
			firmwareConfigResponse[k] = v
		}
	}
	// ensure we add CustomProperties after Properties to ensure they are not overwritten
	for k, v := range firmwareConfigFacade.CustomProperties {
		firmwareConfigResponse[k] = v
	}

	// rebootImmediately flag is missing from stb response
	if _, ok := firmwareConfigResponse[common.REBOOT_IMMEDIATELY]; !ok {
		firmwareConfigResponse[common.REBOOT_IMMEDIATELY] = false
	}

	return firmwareConfigResponse
}

func IsRedundantEntry(key string) bool {
	return key == "id" || key == "description" || key == "supportedModelIds" || key == "updated"
}

func NewEmptyFirmwareConfig() *FirmwareConfig {
	return &FirmwareConfig{
		RebootImmediately:        false,
		ApplicationType:          "stb",
		FirmwareDownloadProtocol: "tftp",
	}
}

func NewFirmwareConfigFromMap(dataMap map[string]interface{}) *FirmwareConfig {
	fc := FirmwareConfig{}
	for k, v := range dataMap {
		switch k {
		case common.ID:
			fc.ID = v.(string)
		case common.DESCRIPTION:
			fc.Description = v.(string)
		case common.FIRMWARE_FILENAME:
			fc.FirmwareFilename = v.(string)
		case common.FIRMWARE_VERSION:
			fc.FirmwareVersion = v.(string)
		case common.APPLICATION_TYPE:
			fc.ApplicationType = v.(string)
		case common.FIRMWARE_LOCATION:
			fc.FirmwareLocation = v.(string)
		case common.IPV6_FIRMWARE_LOCATION:
			fc.Ipv6FirmwareLocation = v.(string)
		case common.FIRMWARE_DOWNLOAD_PROTOCOL:
			fc.FirmwareDownloadProtocol = v.(string)
		case common.UPDATED:
			fc.Updated = int64(v.(float64))
		case common.UPGRADE_DELAY:
			fc.UpgradeDelay = v.(int64)
		case common.REBOOT_IMMEDIATELY:
			b, ok := v.(bool)
			if ok {
				fc.RebootImmediately = b
			} else {
				// Boolean is stored as string "True" or "False"
				b, err := strconv.ParseBool(v.(string))
				if err == nil {
					fc.RebootImmediately = b
				} else {
					log.Error(fmt.Sprintf("FirmwareConfigFacade.UnmarshalJSON failed for property %s:%v", k, v))
				}
			}
		case common.MANDATORY_UPDATE:
			b, ok := v.(bool)
			if ok {
				fc.MandatoryUpdate = b
			} else {
				// Boolean is stored as string "True" or "False"
				b, err := strconv.ParseBool(v.(string))
				if err == nil {
					fc.MandatoryUpdate = b
				} else {
					log.Error(fmt.Sprintf("FirmwareConfigFacade.UnmarshalJSON failed for property %s:%v", k, v))
				}
			}
		case common.SUPPORTED_MODEL_IDS:
			aInterface := v.([]interface{})
			fc.SupportedModelIds = make([]string, len(aInterface))
			for i, val := range aInterface {
				fc.SupportedModelIds[i] = val.(string)
			}
		default:
			log.Debug(fmt.Sprintf("FirmwareConfigFacade.UnmarshalJSON ignored property %s:%v", k, v))
		}
	}

	return &fc
}

type FirmwareConfigResponse struct {
	ID                string            `json:"id"`
	Description       string            `json:"description,omitempty"`
	SupportedModelIds []string          `json:"supportedModelIds,omitempty"`
	FirmwareFilename  string            `json:"firmwareFilename,omitempty"`
	FirmwareVersion   string            `json:"firmwareVersion,omitempty"`
	Properties        map[string]string `json:"properties,omitempty"`
}

func (fc *FirmwareConfig) CreateFirmwareConfigResponse() *FirmwareConfigResponse {
	return &FirmwareConfigResponse{
		ID:                fc.ID,
		Description:       fc.Description,
		SupportedModelIds: fc.SupportedModelIds,
		FirmwareFilename:  fc.FirmwareFilename,
		FirmwareVersion:   fc.FirmwareVersion,
		Properties:        fc.Properties,
	}
}

// NewFirmwareConfigInf constructor
func NewFirmwareConfigInf() interface{} {
	return &FirmwareConfig{
		RebootImmediately:        false,
		ApplicationType:          shared.STB,
		SupportedModelIds:        []string{},
		Properties:               map[string]string{},
		FirmwareDownloadProtocol: "tftp",
	}
}

func (fc *FirmwareConfig) ToPropertiesMap() map[string]interface{} {
	dataMap := make(map[string]interface{})

	util.PutIfValuePresent(dataMap, common.ID, fc.ID)
	util.PutIfValuePresent(dataMap, common.UPDATED, fc.Updated)
	util.PutIfValuePresent(dataMap, common.DESCRIPTION, fc.Description)
	util.PutIfValuePresent(dataMap, common.SUPPORTED_MODEL_IDS, fc.SupportedModelIds)
	util.PutIfValuePresent(dataMap, common.FIRMWARE_DOWNLOAD_PROTOCOL, fc.FirmwareDownloadProtocol)
	util.PutIfValuePresent(dataMap, common.FIRMWARE_FILENAME, fc.FirmwareFilename)
	util.PutIfValuePresent(dataMap, common.FIRMWARE_LOCATION, fc.FirmwareLocation)
	util.PutIfValuePresent(dataMap, common.FIRMWARE_VERSION, fc.FirmwareVersion)
	util.PutIfValuePresent(dataMap, common.IPV6_FIRMWARE_LOCATION, fc.Ipv6FirmwareLocation)
	util.PutIfValuePresent(dataMap, common.UPGRADE_DELAY, fc.UpgradeDelay)
	util.PutIfValuePresent(dataMap, common.REBOOT_IMMEDIATELY, fc.RebootImmediately)
	util.PutIfValuePresent(dataMap, common.MANDATORY_UPDATE, fc.MandatoryUpdate)

	return dataMap
}

func (fc *FirmwareConfig) ToFirmwareConfigResponseMap() map[string]interface{} {
	dataMap := make(map[string]interface{})
	util.PutIfValuePresent(dataMap, common.ID, fc.ID)
	util.PutIfValuePresent(dataMap, common.DESCRIPTION, fc.Description)
	util.PutIfValuePresent(dataMap, common.SUPPORTED_MODEL_IDS, fc.SupportedModelIds)
	util.PutIfValuePresent(dataMap, common.FIRMWARE_FILENAME, fc.FirmwareFilename)
	util.PutIfValuePresent(dataMap, common.FIRMWARE_VERSION, fc.FirmwareVersion)
	return dataMap
}

// FirmwareConfigFacade ...
type FirmwareConfigFacade struct {
	Properties       map[string]interface{}
	CustomProperties map[string]string
}

func (fcf *FirmwareConfigFacade) MarshalJSON() ([]byte, error) {
	/**
	 * The order is important for STB team as they read json response via bash commands in upgrade script.
	 * Will exclude from response fields, like id, description, supportedModelIds, updated.
	 * And also empty values and blank strings.
	 */
	fields := []string{
		common.ID,
		common.DESCRIPTION,
		common.FIRMWARE_DOWNLOAD_PROTOCOL,
		common.FIRMWARE_FILENAME,
		common.FIRMWARE_LOCATION,
		common.FIRMWARE_VERSION,
		common.IPV6_FIRMWARE_LOCATION,
		common.UPGRADE_DELAY,
		common.REBOOT_IMMEDIATELY,
		common.SUPPORTED_MODEL_IDS,
		common.MANDATORY_UPDATE,
		common.UPDATED,
	}

	buffer := bytes.NewBufferString("{")
	for _, field := range fields {
		value := fcf.Properties[field]
		if value != nil {
			if field == common.UPGRADE_DELAY {
				int64v, ok := value.(int64)
				if !ok || int64v == 0 {
					continue
				}
			} else if value == "" {
				continue
			}

			jsonValue, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			if buffer.Len() > 1 {
				buffer.WriteString(",")
			}
			buffer.WriteString(fmt.Sprintf("\"%s\":%s", field, string(jsonValue)))
		}
	}
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (fcf *FirmwareConfigFacade) UnmarshalJSON(bytes []byte) error {
	// Unmarshal into a map of generic types and cast them appropriately
	var dataMap map[string]interface{}
	err := json.Unmarshal(bytes, &dataMap)
	if err != nil {
		return err
	}

	fc := NewFirmwareConfigFromMap(dataMap)
	fcf.Properties = fc.ToPropertiesMap()

	return nil
}

// MinimumFirmwareCheckBean ...
type MinimumFirmwareCheckBean struct {
	HasMinimumFirmware bool `json:"hasMinimumFirmware"`
}

// BseConfiguration ...
type BseConfiguration struct {
	Location string `json:"location"`

	Ipv6Location string `json:"ipv6Location"`

	Protocol string `json:"protocol"`

	ModelConfigurations []*ModelFirmwareConfiguration `json:"modelConfigurations"`
}

// ToString ...
func (b *BseConfiguration) ToString() string {
	return "BseConfiguration{" +
		"location='" + b.Location + "\\" +
		", ipv6Location='" + b.Ipv6Location + "\\" +
		", protocol='" + b.Protocol + "\\" +
		//", modelConfigurations=" + b.ModelConfigurations +  todo
		"}"
}

// ModelFirmwareConfiguration ...
type ModelFirmwareConfiguration struct {
	Model            string `json:"model"`
	FirmwareFilename string `json:"firmwareFilename"`
	FirmwareVersion  string `json:"firmwareVersion"`
}

// NewModelFirmwareConfiguration ... new model config struct
func NewModelFirmwareConfiguration(model string, firmwareFilename string, firmwareVersion string) *ModelFirmwareConfiguration {
	return &ModelFirmwareConfiguration{
		Model:            model,
		FirmwareFilename: firmwareFilename,
		FirmwareVersion:  firmwareVersion,
	}
}

func (m *ModelFirmwareConfiguration) ToString() string {
	return "ModelFirmwareConfiguration{" +
		"model='" + m.Model + "\\" +
		", firmwareFilename='" + m.FirmwareFilename + "\\" +
		", firmwareVersion='" + m.FirmwareVersion + "\\" +
		"}"

}

// IpRuleBean ...
type IpRuleBean struct {
	Id             string                 `json:"id"`
	FirmwareConfig *FirmwareConfig        `json:"firmwareConfig"`
	Name           string                 `json:"name"`
	IpAddressGroup *shared.IpAddressGroup `json:"ipAddressGroup"`
	EnvironmentId  string                 `json:"environmentId"`
	ModelId        string                 `json:"modelId"`
	Expression     *Expression            `json:"expression"`
	Noop           bool                   `json:"noop"`
}

// NewFirmwareConfigFacade ...
func NewDefaulttFirmwareConfigFacade() *FirmwareConfigFacade {
	ff := &FirmwareConfigFacade{
		Properties: map[string]interface{}{},
	}

	return ff
}
func NewFirmwareConfigFacade(firmwareConfig *FirmwareConfig) *FirmwareConfigFacade {
	ff := &FirmwareConfigFacade{
		Properties:       firmwareConfig.ToPropertiesMap(),
		CustomProperties: firmwareConfig.Properties,
	}

	return ff
}

// NewEmptyFirmwareConfigFacade ...
func NewFirmwareConfigFacadeEmptyProperties() *FirmwareConfigFacade {
	properties := map[string]interface{}{}
	return &FirmwareConfigFacade{
		Properties: properties,
	}
}

func (ff *FirmwareConfigFacade) PutIfPresent(key string, value interface{}) {
	if value == nil {
		return
	}

	vstr, ok := value.(string)
	if ok && vstr == "" {
		return
	}

	ff.Properties[key] = vstr
}

// GetFirmwareDownloadProtocol ...
func (ff *FirmwareConfigFacade) GetFirmwareDownloadProtocol() string {
	return ff.GetStringValue(common.FIRMWARE_DOWNLOAD_PROTOCOL)
}

// SetFirmwareDownloadProtocol ...
func (ff *FirmwareConfigFacade) SetFirmwareDownloadProtocol(protocol string) {
	ff.SetStringValue(common.FIRMWARE_DOWNLOAD_PROTOCOL, protocol)
}

func (ff *FirmwareConfigFacade) SetRebootImmediately(flag bool) {
	ff.Properties[common.REBOOT_IMMEDIATELY] = flag
}

func (ff *FirmwareConfigFacade) GetRebootImmediately() bool {
	if flag, ok := ff.Properties[common.REBOOT_IMMEDIATELY]; ok {
		if flagval, isBool := flag.(bool); isBool {
			return flagval
		}
	}
	return false
}

func (ff *FirmwareConfigFacade) SetFirmwareLocation(location string) {
	ff.SetStringValue(common.FIRMWARE_LOCATION, location)
}

func (ff *FirmwareConfigFacade) GetFirmwareLocation() string {
	return ff.GetStringValue(common.FIRMWARE_LOCATION)
}

func (ff *FirmwareConfigFacade) GetFirmwareFilename() string {
	return ff.GetStringValue(common.FIRMWARE_FILENAME)
}

func (ff *FirmwareConfigFacade) GetFirmwareVersion() string {
	return ff.GetStringValue(common.FIRMWARE_VERSION)
}
func (ff *FirmwareConfigFacade) GetIpv6FirmwareLocation() string {
	return ff.GetStringValue(common.IPV6_FIRMWARE_LOCATION)
}

func (ff *FirmwareConfigFacade) GetUpgradeDelay() int {
	if val, ok := ff.Properties[common.UPGRADE_DELAY]; ok {
		if val == nil {
			return 0
		}

		intVal, valid := val.(int)
		if !valid {
			return 0
		}
		return intVal
	}
	return 0
}

func (ff *FirmwareConfigFacade) GetStringValue(key string) string {
	val := ff.Properties[key]
	if val == nil {
		return ""
	}
	return val.(string)
}

func (ff *FirmwareConfigFacade) GetValue(key string) interface{} {
	return ff.Properties[key]
}

func (ff *FirmwareConfigFacade) SetStringValue(key string, value string) {
	ff.Properties[key] = value
}

func (ff *FirmwareConfigFacade) PutAll(nmap map[string]interface{}) {
	for k, v := range nmap {
		ff.Properties[k] = v
	}
}

func GetFirmwareConfigOneDB(id string) (*FirmwareConfig, error) {
	if len(id) == 0 {
		return nil, errors.New("id is empty")
	}
	inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_FIRMWARE_CONFIG, id)
	if err != nil {
		return nil, err
	}
	fc, ok := inst.(*FirmwareConfig)
	if !ok {
		return nil, common.NotFirmwareConfig
	}
	if fc.ApplicationType == "" {
		fc.ApplicationType = shared.STB
	}
	return fc, nil
}

func CreateFirmwareConfigOneDB(fc *FirmwareConfig) error {
	// create record in DB
	if util.IsBlank(fc.ID) {
		fc.ID = uuid.New().String()
	}
	fc.Updated = util.GetTimestamp()
	return db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_CONFIG, fc.ID, fc)
}

func DeleteOneFirmwareConfig(id string) error {
	return db.GetCachedSimpleDao().DeleteOne(db.TABLE_FIRMWARE_CONFIG, id)
}

func GetFirmwareConfigAsListDB() ([]*FirmwareConfig, error) {
	rulelst, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_FIRMWARE_CONFIG, 0)
	if err != nil {
		return nil, err
	}

	var lst []*FirmwareConfig

	for _, r := range rulelst {
		cfg, ok := r.(*FirmwareConfig)
		if !ok {
			continue
		}
		lst = append(lst, cfg)
	}

	return lst, nil
}

func GetFirmwareVersion(id string) string {
	fc, err := GetFirmwareConfigOneDB(id)
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareVersion: %v", err))
		return ""
	}

	return fc.FirmwareVersion
}

// MacRuleBean ...
type MacRuleBean struct {
	Id               string          `json:"id,omitempty"`
	Name             string          `json:"name,omitempty"`
	MacAddresses     string          `json:"macAddresses,omitempty"`
	MacListRef       string          `json:"macListRef,omitempty"`
	FirmwareConfig   *FirmwareConfig `json:"firmwareConfig"`
	TargetedModelIds *[]string       `json:"targetedModelIds,omitempty"`
	//in JAva, MacList is in MacRuleBeanWrapper that extends MacRuleBean
	MacList *[]string `json:"macList,omitempty"`
}

type EnvModelBean struct {
	Id             string          `json:"id,omitempty" xml:"id,omitempty"`
	Name           string          `json:"name,omitempty" xml:"name,omitempty"`
	EnvironmentId  string          `json:"environmentId,omitempty" xml:"environmentId,omitempty"`
	ModelId        string          `json:"modelId,omitempty" xml:"modelId,omitempty"`
	FirmwareConfig *FirmwareConfig `json:"firmwareConfig,omitempty" xml:"firmwareConfig,omitempty"`
	Noop           bool            `json:"-"`
}
