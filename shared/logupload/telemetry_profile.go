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
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"xconfwebconfig/common"
	"xconfwebconfig/db"
	re "xconfwebconfig/rulesengine"
	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

// TelemetryElement a telemetry element
type TelemetryElement struct {
	ID               string `json:"id,omitempty"`
	Header           string `json:"header"`
	Content          string `json:"content"`
	Type             string `json:"type"`
	PollingFrequency string `json:"pollingFrequency"`
	Component        string `json:"component,omitempty"`
}

// TelemetryProfile Telemetry table
type TelemetryProfile struct {
	ID               string             `json:"id"`
	TelemetryProfile []TelemetryElement `json:"telemetryProfile"`
	Schedule         string             `json:"schedule"`
	Expires          int64              `json:"expires"`
	Name             string             `json:"telemetryProfile:name"`
	UploadRepository string             `json:"uploadRepository:URL"`
	UploadProtocol   UploadProtocol     `json:"uploadRepository:uploadProtocol"`
	ApplicationType  string             `json:"applicationType"`
}

func (obj *TelemetryProfile) Clone() (*TelemetryProfile, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*TelemetryProfile), nil
}

// NewTelemetryProfileInf constructor
func NewTelemetryProfileInf() interface{} {
	return &TelemetryProfile{}
}

type TelemetryProfileDescriptor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewTelemetryProfileDescriptor() *TelemetryProfileDescriptor {
	return &TelemetryProfileDescriptor{}
}

// PermanentTelemetryProfile PermanentTelemetry table
type PermanentTelemetryProfile struct {
	Type             string             `json:"@type,omitempty"`
	ID               string             `json:"id"`
	TelemetryProfile []TelemetryElement `json:"telemetryProfile"`
	Schedule         string             `json:"schedule"`
	Expires          int64              `json:"expires"`
	Name             string             `json:"telemetryProfile:name"`
	UploadRepository string             `json:"uploadRepository:URL"`
	UploadProtocol   UploadProtocol     `json:"uploadRepository:uploadProtocol"`
	ApplicationType  string             `json:"applicationType,omitempty"`
}

func (obj *PermanentTelemetryProfile) Clone() (*PermanentTelemetryProfile, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*PermanentTelemetryProfile), nil
}

func (obj *PermanentTelemetryProfile) IsEmpty() bool {
	if obj.Type == "" && obj.ID == "" && obj.TelemetryProfile == nil && obj.Schedule == "" && obj.Name == "" && obj.UploadRepository == "" && obj.UploadProtocol == "" && obj.ApplicationType == "" {
		return true
	}
	return false
}

func (s *PermanentTelemetryProfile) EqualChangeData(t *PermanentTelemetryProfile) bool {
	if t == nil {
		return false
	}

	return s.Schedule == t.Schedule && s.Expires == t.Expires && s.Name == t.Name && s.UploadRepository == t.UploadRepository &&
		s.UploadProtocol == t.UploadProtocol && s.ApplicationType == t.ApplicationType && checkEqualTelemetryElements(s.TelemetryProfile, t.TelemetryProfile)
}

// TODO rework it
func checkEqualTelemetryElements(s, t []TelemetryElement) bool {
	if len(s) != len(t) {
		return false
	}

	count := 0

	for i := 0; i < len(s); i++ {
		for j := 0; j < len(t); j++ {
			if s[i].Header == t[j].Header && s[i].Content == t[j].Content && s[i].Type == t[j].Type && s[i].PollingFrequency == t[j].PollingFrequency && s[i].Component == t[j].Component {
				count = count + 1
			}
		}
	}

	if count != len(s) {
		return false
	}

	return true
}

func IsValidUploadProtocol(p string) bool {
	str := strings.ToUpper(p)
	if str == string(TFTP) || str == string(SFTP) || str == string(SCP) || str == string(HTTP) || str == string(HTTPS) || str == string(S3) {
		return true
	}
	return false
}

func IsValidUrl(str string) bool {
	u, err := url.ParseRequestURI(str)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	if !IsValidUploadProtocol(u.Scheme) {
		return false
	}
	return urlRe.MatchString(u.Host)
}

func (obj *PermanentTelemetryProfile) Validate() error {
	if util.IsBlank(obj.Type) {
		return common.NewRemoteError(http.StatusBadRequest, "Type is required")
	}
	if util.IsBlank(obj.Name) {
		return common.NewRemoteError(http.StatusBadRequest, "Name is empty")
	}

	protocol := obj.UploadProtocol
	host := obj.UploadRepository
	var url string
	if strings.Contains(host, "://") || protocol == "" {
		url = host
	} else {
		url = strings.ToLower(string(protocol)) + "://" + host
	}

	if !IsValidUrl(url) {
		return common.NewRemoteError(http.StatusBadRequest, "URL is invalid")
	}

	if elements := obj.TelemetryProfile; len(elements) < 1 {
		return common.NewRemoteError(http.StatusBadRequest, "Should contain at least one profile entry")
	} else {
		for i, element := range elements {
			_, err := strconv.Atoi(element.PollingFrequency)
			if err != nil {
				return common.NewRemoteError(http.StatusBadRequest, "Polling frequency is not a number")
			}
			for j := i + 1; j < len(elements); j++ {
				if element.Equals(&elements[j]) {
					return common.NewRemoteError(http.StatusBadRequest, "Profile entity has duplicate entries")
				}
			}
		}
	}
	return nil
}

// NewPermanentTelemetryProfileInf constructor
func NewPermanentTelemetryProfileInf() interface{} {
	return &PermanentTelemetryProfile{}
}

func NullifyUnwantedFieldsPermanentTelemetryProfile(profile *PermanentTelemetryProfile) *PermanentTelemetryProfile {
	if len(profile.TelemetryProfile) > 0 {
		for index := range profile.TelemetryProfile {
			profile.TelemetryProfile[index].ID = ""
			profile.TelemetryProfile[index].Component = ""
		}
	}

	profile.ApplicationType = ""
	return profile
}

// TelemetryRule TelemetryRules table
type TelemetryRule struct {
	re.Rule
	ID               string `json:"id"`
	Updated          int64  `json:"updated"`
	BoundTelemetryID string `json:"boundTelemetryId"`
	Name             string `json:"name"`
	ApplicationType  string `json:"applicationType"`
}

func (obj *TelemetryRule) Clone() (*TelemetryRule, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*TelemetryRule), nil
}

func (t *TelemetryElement) Equals(o *TelemetryElement) bool {
	if t == o {
		return true
	}
	if o == nil {
		return false
	}
	if t.ID == o.ID && t.Header == o.Header && t.Content == o.Content && t.Type == o.Type && t.PollingFrequency == o.PollingFrequency && t.Component == o.Component {
		return true
	}
	return false
}

func (r *TelemetryRule) GetApplicationType() string {
	if len(r.ApplicationType) > 0 {
		return r.ApplicationType
	}
	return "stb"
}

// GetId XRule interface
func (r *TelemetryRule) GetId() string {
	return r.ID
}

// GetRule XRule interface
func (r *TelemetryRule) GetRule() *re.Rule {
	return &r.Rule
}

// GetName XRule interface
func (r *TelemetryRule) GetName() string {
	return r.Name
}

// GetTemplateId XRule interface
func (r *TelemetryRule) GetTemplateId() string {
	return ""
}

// GetRuleType XRule interface
func (r *TelemetryRule) GetRuleType() string {
	return "TelemetryRule"
}

// NewTelemetryRuleInf constructor
func NewTelemetryRuleInf() interface{} {
	return &TelemetryRule{}
}

type PermanentTelemetryRuleDescriptor struct {
	RuleId   string `json:"ruleId"`
	RuleName string `json:"ruleName"`
}

func NewPermanentTelemetryRuleDescriptor() *PermanentTelemetryRuleDescriptor {
	return &PermanentTelemetryRuleDescriptor{}
}

type TimestampedRule struct {
	re.Rule
	Timestamp int64
}

func NewTimestampedRule() *TimestampedRule {
	return &TimestampedRule{}
}

func (t *TimestampedRule) ToString() string {
	timestampRuleString := strconv.FormatInt(t.Timestamp, 10) + t.Rule.String()
	return timestampRuleString
}

func (t *TimestampedRule) Equals(x *TimestampedRule) bool {
	if t.Timestamp == x.Timestamp && t.Equals(x) {
		return true
	} else {
		return false
	}
}

// TelemetryTwoRule TelemetryTwoRules table
type TelemetryTwoRule struct {
	re.Rule
	ID                string   `json:"id"`
	Updated           int64    `json:"updated"`
	Name              string   `json:"name"`
	ApplicationType   string   `json:"applicationType"`
	BoundTelemetryIDs []string `json:"boundTelemetryIds"`
}

func (obj *TelemetryTwoRule) Clone() (*TelemetryTwoRule, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*TelemetryTwoRule), nil
}

func (t *TelemetryTwoRule) String() string {
	return fmt.Sprintf("TelemetryTwoRule(ID=%v, Name='%v', ApplicationType='%v'\n  BoundTelemetryIDs='%v'\n  %v\n)",
		t.ID, t.Name, t.ApplicationType, t.BoundTelemetryIDs, t.Rule.String())
}

// GetId XRule interface
func (r *TelemetryTwoRule) GetId() string {
	return r.ID
}

// GetRule XRule interface
func (r *TelemetryTwoRule) GetRule() *re.Rule {
	return &r.Rule
}

// GetName XRule interface
func (r *TelemetryTwoRule) GetName() string {
	return r.Name
}

// GetTemplateId XRule interface
func (r *TelemetryTwoRule) GetTemplateId() string {
	return ""
}

// GetRuleType XRule interface
func (r *TelemetryTwoRule) GetRuleType() string {
	return "TelemetryTwoRule"
}

// NewTelemetryTwoRuleInf constructor
func NewTelemetryTwoRuleInf() interface{} {
	return &TelemetryTwoRule{}
}

func (t *TelemetryTwoRule) Equals(o *TelemetryTwoRule) bool {
	if t == o {
		return true
	}
	if o == nil {
		return false
	}
	if !t.Rule.Equals(&o.Rule) {
		return false
	}
	if t.ID != o.ID {
		return false
	}
	if t.Name != o.Name {
		return false
	}
	if t.ApplicationType != o.ApplicationType {
		return false
	}
	if !util.StringSliceEqual(t.BoundTelemetryIDs, o.BoundTelemetryIDs) {
		return false
	}
	return true
}

// TelemetryTwoProfile TelemetryTwoProfiles table
type TelemetryTwoProfile struct {
	Type            string `json:"@type,omitempty"`
	ID              string `json:"id"`
	Updated         int64  `json:"updated"`
	Name            string `json:"name"`
	Jsonconfig      string `json:"jsonconfig"`
	ApplicationType string `json:"applicationType"`
}

func (obj *TelemetryTwoProfile) Clone() (*TelemetryTwoProfile, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*TelemetryTwoProfile), nil
}

func (entity *TelemetryTwoProfile) Validate() error {
	if util.IsBlank(entity.Type) {
		return common.NewRemoteError(http.StatusBadRequest, "Type is required")
	}
	if util.IsBlank(entity.Name) {
		return common.NewRemoteError(http.StatusBadRequest, "Name is not present")
	}
	if err := ValidateTelemetryTwoProfileJson(entity.Jsonconfig); err != nil {
		return err
	}

	return nil
}

func (s *TelemetryTwoProfile) EqualChangeData(t *TelemetryTwoProfile) bool {
	if t == nil {
		return false
	}

	return s.Name == t.Name && s.Jsonconfig == t.Jsonconfig && s.ApplicationType == t.ApplicationType
}

func (entity *TelemetryTwoProfile) ValidateAll(existingEntities []*TelemetryTwoProfile) error {
	for _, profile := range existingEntities {
		if !(profile.ID == entity.ID) && profile.Name == entity.Name {
			return common.NewRemoteError(http.StatusConflict, fmt.Sprintf("TelemetryTwo Profile with such name exists: %s", entity.Name))
		}
	}

	return nil
}

func (s *TelemetryTwoProfile) Equals(t *TelemetryTwoProfile) bool {
	if t == nil {
		return false
	}
	if s.ID != t.ID || s.Name != t.Name || s.Jsonconfig != t.Jsonconfig || s.ApplicationType != t.ApplicationType {
		return false
	}

	return true
}

// NewTelemetryTwoProfileInf constructor
func NewTelemetryTwoProfileInf() interface{} {
	return &TelemetryTwoProfile{}
}

//var cachedSimpleDao ds.CachedSimpleDao

var GetCachedSimpleDaoFunc = db.GetCachedSimpleDao

func DeleteExpiredTelemetryProfile(cacheUpdateWindowSize int64) {
	telemetryProfileMapInst, err := GetCachedSimpleDaoFunc().GetAllAsMap(db.TABLE_TELEMETRY)
	if err != nil {
		log.Warn(fmt.Sprintf("no telemetryProfileList found for ExpireTemporaryTelemetryRules()"))
	} else {
		for k, v := range telemetryProfileMapInst {
			timestampedRule := k.(string)
			telemetryProfile := v.(TelemetryProfile)
			if (telemetryProfile.Expires + cacheUpdateWindowSize) <= time.Now().UTC().Unix()*1000 {
				log.Debugf("{%s} is expired, removing", timestampedRule)
				GetCachedSimpleDaoFunc().DeleteOne(db.TABLE_TELEMETRY, timestampedRule)
			}
		}
	}
}

func DeleteTelemetryProfile(rowKey string) {
	GetCachedSimpleDaoFunc().DeleteOne(db.TABLE_TELEMETRY, rowKey)
}

func SetTelemetryProfile(rowKey string, telemetry TelemetryProfile) {
	GetCachedSimpleDaoFunc().SetOne(db.TABLE_TELEMETRY, rowKey, telemetry)
}

func GetOneTelemetryProfile(rowKey string) *TelemetryProfile {
	telemetryInst, err := GetCachedSimpleDaoFunc().GetOne(db.TABLE_TELEMETRY, rowKey)
	if err != nil {
		log.Warn(fmt.Sprintf("no telemetryProfile found for " + rowKey))
		return nil
	}
	telemetry := telemetryInst.(TelemetryProfile)
	return &telemetry
}

func GetTimestampedRules() []TimestampedRule {
	timestampedRuleSet, err := GetCachedSimpleDaoFunc().GetKeys(db.TABLE_TELEMETRY)
	if err != nil {
		log.Warn(fmt.Sprintf("no TimestampedRule found"))
		return nil
	}
	rules := []TimestampedRule{}
	for idx := range timestampedRuleSet {
		timestampedRule := timestampedRuleSet[idx].(TimestampedRule)
		rules = append(rules, timestampedRule)
	}
	return rules
}

func GetRulesFromTimestampedRules() []re.Rule {
	timestampedRuleSet, err := GetCachedSimpleDaoFunc().GetKeys(db.TABLE_TELEMETRY)
	if err != nil {
		log.Warn(fmt.Sprintf("no TimestampedRule found"))
		return nil
	}
	rules := []re.Rule{}
	for idx := range timestampedRuleSet {
		timestampedRule := timestampedRuleSet[idx].(TimestampedRule)
		rules = append(rules, timestampedRule.Rule)
	}
	return rules
}

func GetTelemetryProfileMap() *map[string]TelemetryProfile {
	telemetryProfileMap, err := GetCachedSimpleDaoFunc().GetAllAsMap(db.TABLE_TELEMETRY)
	if err != nil {
		log.Warn(fmt.Sprintf("no telemetryProfileMap found"))
		return nil
	}
	finalMap := make(map[string]TelemetryProfile)
	for k, v := range telemetryProfileMap {
		mapK := k.(string)
		mapV := v.(TelemetryProfile)
		finalMap[mapK] = mapV
	}
	return &finalMap
}

func GetTelemetryProfileList() []*TelemetryProfile {
	all := []*TelemetryProfile{}
	tRuleList, err := GetCachedSimpleDaoFunc().GetAllAsList(db.TABLE_TELEMETRY, 0)
	if err != nil {
		log.Warn(fmt.Sprintf("no TelemetryProfile found"))
		return nil
	}
	for idx := range tRuleList {
		tProfile := tRuleList[idx].(TelemetryProfile)
		all = append(all, &tProfile)
	}
	return all
}

func GetTelemetryRuleList() []*TelemetryRule {
	cm := db.GetCacheManager()
	cacheKey := "TelemetryRuleList"
	cacheInst := cm.ApplicationCacheGet(db.TABLE_TELEMETRY_RULES, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*TelemetryRule)
	}

	tRuleList, err := GetCachedSimpleDaoFunc().GetAllAsList(db.TABLE_TELEMETRY_RULES, 0)
	if err != nil {
		log.Warn("no TelemetryRule found")
		return []*TelemetryRule{}
	}

	all := make([]*TelemetryRule, 0, len(tRuleList))

	for _, v := range tRuleList {
		tRule := v.(*TelemetryRule)
		all = append(all, tRule)
	}

	if len(all) > 0 {
		cm.ApplicationCacheSet(db.TABLE_TELEMETRY_RULES, cacheKey, all)
	}

	return all
}

func GetOnePermanentTelemetryProfile(rowKey string) *PermanentTelemetryProfile {
	telemetryInst, err := GetCachedSimpleDaoFunc().GetOne(db.TABLE_PERMANENT_TELEMETRY, rowKey)
	if err != nil {
		log.Warn(fmt.Sprintf("no telemetryProfile found for " + rowKey))
		return nil
	}
	telemetry := telemetryInst.(*PermanentTelemetryProfile)
	return telemetry
}

func GetPermanentTelemetryProfileList() []*PermanentTelemetryProfile {
	all := []*PermanentTelemetryProfile{}
	tRuleList, err := GetCachedSimpleDaoFunc().GetAllAsList(db.TABLE_PERMANENT_TELEMETRY, 0)
	if err != nil {
		log.Warn(fmt.Sprintf("no TelemetryProfile found"))
		return nil
	}
	for idx := range tRuleList {
		tProfile := tRuleList[idx].(*PermanentTelemetryProfile)
		all = append(all, tProfile)
	}
	return all
}

func GetTelemetryTwoRuleList() []*TelemetryTwoRule {
	all := []*TelemetryTwoRule{}
	tRuleList, err := GetCachedSimpleDaoFunc().GetAllAsList(db.TABLE_TELEMETRY_TWO_RULES, 0)
	if err != nil {
		log.Warn(fmt.Sprintf("no TelemetryTwoRule found"))
		return nil
	}
	for _, itf := range tRuleList {
		if telemetryTwoRule, ok := itf.(*TelemetryTwoRule); ok {
			all = append(all, telemetryTwoRule)
		}
	}
	return all
}

func GetOneTelemetryTwoProfile(rowKey string) *TelemetryTwoProfile {
	telemetryInst, err := GetCachedSimpleDaoFunc().GetOne(db.TABLE_TELEMETRY_TWO_PROFILES, rowKey)
	if err != nil {
		log.Warn(fmt.Sprintf("no TelemetryTwoProfile found for " + rowKey))
		return nil
	}
	telemetry := telemetryInst.(*TelemetryTwoProfile)
	return telemetry
}
