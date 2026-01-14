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
package logupload

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/rdkcentral/xconfwebconfig/db"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	util "github.com/rdkcentral/xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

// UploadProtocol enum
type UploadProtocol string

const (
	TFTP  UploadProtocol = "TFTP"
	SFTP  UploadProtocol = "SFTP"
	SCP   UploadProtocol = "SCP"
	HTTP  UploadProtocol = "HTTP"
	HTTPS UploadProtocol = "HTTPS"
	S3    UploadProtocol = "S3"
)

var urlRe = regexp.MustCompile(`^[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$`)

const (
	EstbIp            string = "estbIP"
	EstbMacAddress    string = "estbMacAddress"
	EcmMac            string = "ecmMacAddress"
	Env               string = "env"
	Model             string = "model"
	AccountMgmt       string = "accountMgmt"
	SerialNum         string = "serialNum"
	PartnerId         string = "partnerId"
	FirmwareVersion   string = "firmwareVersion"
	ControllerId      string = "controllerId"
	ChannelMapId      string = "channelMapId"
	VodId             string = "vodId"
	UploadImmediately string = "uploadImmediately"
	Timezone          string = "timezone"
	Application       string = "applicationType"
	AccountHash       string = "accountHash"
	AccountId         string = "accountId"
	ConfigSetHash     string = "configSetHash"
)

/*
	LogUpload tables
*/

// UploadRepository table
type UploadRepository struct {
	ID              string `json:"id"`
	Updated         int64  `json:"updated"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	URL             string `json:"url"`
	ApplicationType string `json:"applicationType"`
	Protocol        string `json:"protocol"`
}

func (obj *UploadRepository) Clone() (*UploadRepository, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*UploadRepository), nil
}

// NewUploadRepositoryInf constructor
func NewUploadRepositoryInf() interface{} {
	return &UploadRepository{
		ApplicationType: shared.STB,
	}
}

// LogFile table
type LogFile struct {
	ID             string `json:"id"`
	Updated        int64  `json:"updated"`
	Name           string `json:"name"`
	DeleteOnUpload bool   `json:"deleteOnUpload"`
}

func (obj *LogFile) Clone() (*LogFile, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*LogFile), nil
}

// NewLogFileInf constructor
func NewLogFileInf() interface{} {
	return &LogFile{}
}

// LogFilesGroups table
type LogFilesGroups struct {
	ID         string   `json:"id"`
	Updated    int64    `json:"updated"`
	GroupName  string   `json:"groupName"`
	LogFileIDs []string `json:"logFileIds"`
}

func (obj *LogFilesGroups) Clone() (*LogFilesGroups, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*LogFilesGroups), nil
}

// NewLogFilesGroupsInf constructor
func NewLogFilesGroupsInf() interface{} {
	return &LogFilesGroups{}
}

// LogFileList LogFileList table
type LogFileList struct {
	Updated int64      `json:"updated"`
	Data    []*LogFile `json:"data"`
}

func (obj *LogFileList) Clone() (*LogFileList, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*LogFileList), nil
}

// NewLogFileListInf constructor
func NewLogFileListInf() interface{} {
	return &LogFileList{}
}

type Schedule struct {
	Type              string      `json:"type,omitempty"`
	Expression        string      `json:"expression,omitempty"`
	TimeZone          string      `json:"timeZone,omitempty"`
	ExpressionL1      string      `json:"expressionL1,omitempty"`
	ExpressionL2      string      `json:"expressionL2,omitempty"`
	ExpressionL3      string      `json:"expressionL3,omitempty"`
	StartDate         string      `json:"startDate,omitempty"`
	EndDate           string      `json:"endDate,omitempty"`
	TimeWindowMinutes json.Number `json:"timeWindowMinutes,omitempty"`
}
type ConfigurationServiceURL struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

// DcmRule DcmRule table
type DCMGenericRule struct {
	re.Rule
	ID              string      `json:"id"`
	Updated         int64       `json:"updated"`
	Name            string      `json:"name,omitempty"`
	Description     string      `json:"description,omitempty"`
	Priority        int         `json:"priority,omitempty"`
	RuleExpression  string      `json:"ruleExpression,omitempty"`
	Percentage      int         `json:"percentage,omitempty"`
	PercentageL1    json.Number `json:"percentageL1,omitempty"`
	PercentageL2    json.Number `json:"percentageL2,omitempty"`
	PercentageL3    json.Number `json:"percentageL3,omitempty"`
	ApplicationType string      `json:"applicationType"`
}

func (obj *DCMGenericRule) GetPriority() int {
	return obj.Priority
}

func (obj *DCMGenericRule) SetPriority(priority int) {
	obj.Priority = priority
}

func (obj *DCMGenericRule) GetID() string {
	return obj.ID
}

func (obj *DCMGenericRule) Clone() (*DCMGenericRule, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*DCMGenericRule), nil
}

func NewDCMGenericRuleInf() interface{} {
	return &DCMGenericRule{
		Percentage:      100,
		ApplicationType: shared.STB,
	}
}

type DCMFormula struct {
	Formula DCMGenericRule `json:"formula"`
}

// GetId XRule interface
func (r *DCMGenericRule) GetId() string {
	return r.ID
}

// GetRule XRule interface
func (r *DCMGenericRule) GetRule() *re.Rule {
	return &r.Rule
}

// GetName XRule interface
func (r *DCMGenericRule) GetName() string {
	return r.Name
}

// GetTemplateId XRule interface
func (r *DCMGenericRule) GetTemplateId() string {
	return ""
}

// GetRuleType XRule interface
func (r *DCMGenericRule) GetRuleType() string {
	return "DCMGenericRule"
}

func (dcm *DCMGenericRule) ToStringOnlyBaseProperties() string {
	if dcm.Rule.IsCompound() {
		var sb strings.Builder
		for _, compoundPart := range dcm.Rule.CompoundParts {
			sb.WriteString(compoundPart.String())
		}
		return sb.String()
	}
	return dcm.Rule.Condition.String()
}

func GetDCMGenericRuleList() []*DCMGenericRule {
	cm := db.GetCacheManager()
	cacheKey := "DCMGenericRuleList"
	cacheInst := cm.ApplicationCacheGet(db.TABLE_DCM_RULE, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*DCMGenericRule)
	}

	dmcRuleList, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_DCM_RULE, 0)
	if err != nil {
		log.Warn("no dmcRule found")
		return []*DCMGenericRule{}
	}

	all := make([]*DCMGenericRule, 0, len(dmcRuleList))

	for idx := range dmcRuleList {
		if dmcRuleList[idx] != nil {
			dmcRule := dmcRuleList[idx].(*DCMGenericRule)
			all = append(all, dmcRule)
		}
	}

	if len(all) > 0 {
		cm.ApplicationCacheSet(db.TABLE_DCM_RULE, cacheKey, all)
	}

	return all
}

func GetDCMGenericRuleListForAS() []*DCMGenericRule {
	all := []*DCMGenericRule{}
	dmcRuleList, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_DCM_RULE, 0)
	if err != nil {
		log.Warn("no dmcRule found")
		return all
	}
	for idx := range dmcRuleList {
		if dmcRuleList[idx] != nil {
			dmcRule := dmcRuleList[idx].(*DCMGenericRule)
			all = append(all, dmcRule)
		}
	}
	return all
}

func GetOneDCMGenericRule(id string) *DCMGenericRule {
	dmcRuleInst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_DCM_RULE, id)
	if err != nil {
		log.Warn("no dmcRule found for: " + id)
		return nil
	}
	dmcRule := dmcRuleInst.(*DCMGenericRule)
	return dmcRule
}
