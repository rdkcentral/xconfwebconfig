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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

var TelemetryTwoProfileSchema *gojsonschema.Schema

const (
	TelemetryTwoProfileJSONSchema = `{
  "$schema": "http://json-schema.org/draft-06/schema#",
  "$id": "https://github.com/cfry002/telemetry2/schemas/t2_reportProfileSchema.schema.json",
  "title":"Telemetry 2.0 Report Profile Description",
  "version": "2.0.10",
  "type": "object",
  "description": "<b>Telemetry 2.0 Report Profile Description</b><p>A Telemetry 2.0 Report Profile is a configuration, authored in JSON, that can be sent to any RDK device which supports Telemetry 2.0.  A Report Profile contains properties that are interpreted by the CPE in order to generate and upload a telemetry report. These properties define the details of a generated report, including: <br/>&bull;&nbsp;&nbsp;Scheduling (how often the report should be generated) <br/>&bull;&nbsp;&nbsp;Parameters (what key/value pairs should be in the report) <br/>&bull;&nbsp;&nbsp;Encoding (the format of the generated report) <br/>&bull;&nbsp;&nbsp;Protocol (protocol to use to send generated report)</p>",

  "definitions": {
    "parmUse": {
        "type":  "string",
        "enum": [ "count", "absolute", "csv", "accumulate" ],
        "default": "absolute",
        "example":"\"use\":\"count\""
    },
    "dataModelMethod": {
        "type":  "string",
        "enum": [ "get", "subscribe" ],
        "default": "get",
        "example":"\"method\":\"subscribe\""
    },
    "parmReportTimestamp":{
        "type":"string",
        "enum": ["Unix-Epoch","None"],
        "default": "None",
        "example":"\"reportTimestamp\":\"Unix-Epoch\""
        
    },
    "parmDefinitions": {
        "title":"Definitions for the different Parameter types",
        "type":"object",
        "properties":{
            "grep": {
                "title":"\"grep\" Parameter",
                "type": "object",
                "properties": {
                    "type":    { "type": "string", "const": "grep", "description": "Defines a grep parameter"  },
                    "marker":  { "type": "string", "description": "The key name to be used for this data in the generated report." },
                    "search":  { "type": "string", "description": "The string for which to search within the log file." },
                    "logFile": { "type": "string", "description": "The name of the log file to be searched."},
                    "use":     { "$ref":    "#/definitions/parmUse", "description": "This property indicates how the data for this parameter should be gathered and reported.<br/>&bull;&nbsp;&nbsp;\"count\": Indicates that the value to report for this parameter is the number of times it has occurred during the reporting interval..<br/>&bull;&nbsp;&nbsp;\"absolute\": Indicates that the value to report for this parameter is the last actual value received, in the case of events, or found in the log file, in the case of greps.<br/>&bull;&nbsp;&nbsp;\"csv\": Indicates that the value to report for this parameter is a comma separated list of all the actual values received, in the case of events, or found in the log file, in the case of greps. <b>NOTE:</b> \"csv\" is not currently supported in Telemetry 2.0."},
                    "reportEmpty": { "type": "boolean", "default":"false", "description": "Should this marker name be included in the generated report even if the search string was not found in the log file?"},
                    "firstSeekFromEOF": { "type": "integer", "description": "An offset, in bytes, backward from logFile EOF to where the grep for search must begin for the first report this profile generates.  See documentation."}
                },
                "additionalProperties": false,
                "required": ["type", "marker", "search", "logFile"],
                "description": "A grep parameter defines report data that comes from searching a log file for a particular string.",
                "example":"{ \"type\": \"grep\", \"marker\": \"T2_btime_dsLock_split\", \"search\": \"Downstream Lock Success=\", \"logFile\": \"BootTime.log\", \"use\": \"absolute\"}"
            },
            "event": {
                "title":"\"event\" Parameter",
                "type": "object",
                "properties": {
                    "type":      { "type": "string", "const": "event", "description": "Defines an event parameter.  This data comes from a component that has been instrumented to send its telemetry data via Telemetry 2 APIs."  },
                    "name":      { "type": "string", "description": "Optional: The key name to be used for this data in the generated report." },
                    "eventName":  { "type": "string", "description": "The event name by which the component will report this data to Telemetry 2" },
                    "component":  { "type": "string", "description": "The name of the component from which this data should be expected.  Telemetry 2 will use this name to register its interest with the component." },
                    "use":        { "$ref": "#/definitions/parmUse", "description": "This property indicates how the data for this parameter should be gathered and reported.<br/>&bull;&nbsp;&nbsp;\"count\": Indicates that the value to report for this parameter is the number of times it has occurred during the reporting interval..<br/>&bull;&nbsp;&nbsp;\"absolute\": Indicates that the value to report for this parameter is the last actual value received, in the case of events, or found in the log file, in the case of greps.<br/>&bull;&nbsp;&nbsp;\"csv\": Indicates that the value to report for this parameter is a comma separated list of all the actual values received, in the case of events, or found in the log file, in the case of greps. <b>NOTE:</b> \"csv\" is not currently supported in Telemetry 2.0."},
                    "reportEmpty": { "type": "boolean", "default":"false", "description": "Should this marker name be included in the generated report even if the search string was not found in the log file?"},
                    "reportTimestamp": { "$ref": "#/definitions/parmReportTimestamp", "description": "This property indicates whether or not a timestamp should be encoded to indicate the time at which this parameter data was received."}
                },
                "additionalProperties": false,
                "required": ["type", "eventName", "component"],
                "description": "An event parameter defines data that will come from a component event.",
                "example": "{ \"type\": \"event\", \"name\": \"XH_RSSI_1_split\", \"eventName\":\"xh_rssi_3_split\",\"component\": \"ccsp-wifi-agent\", \"use\":\"absolute\" }, "
            },
            "dataModel": {
                "title":"\"dataModel\" Parameter",
                "type": "object",
                "properties": {
                    "type":      { "type": "string", "const": "dataModel", "description": "Defines a data model parameter, e.g., TR-181 data."  },
                    "name":      { "type": "string", "description": "Optional: The key name to be used for this data in the generated report." },
                    "reference":  { "type": "string", "description": "The data model object or property name whose value is to be in the generated report, e.g., \"Device.DeviceInfo.HardwareVersion\"" },
                    "reportEmpty": { "type": "boolean", "default":"false", "description": "Should this marker name be included in the generated report even if the search string was not found in the log file?"},
                    "method":     { "$ref": "#/definitions/dataModelMethod", "description": "How should the value for this parameter be retrieved?  \"get\" will do a parameter GET; \"subscribe\" will subscribe for changes on the event indicated by \"reference\"." },
                    "use":        { "$ref": "#/definitions/parmUse", "description": "This property indicates how the data for this parameter should be gathered and reported.<br/>&bull;&nbsp;&nbsp;\"count\": Indicates that the value to report for this parameter is the number of times it has occurred during the reporting interval..<br/>&bull;&nbsp;&nbsp;\"absolute\": Indicates that the value to report for this parameter is the last actual value received.<br/>&bull;&nbsp;&nbsp;\"csv\": Indicates that the value to report for this parameter is a comma separated list of all the actual values received. <b>NOTE:</b> \"csv\" is not currently supported in Telemetry 2.0."},
                    "reportTimestamp": { "$ref": "#/definitions/parmReportTimestamp", "description": "This property indicates whether or not a timestamp should be encoded to indicate the time at which this parameter data was received."}
                },
                "additionalProperties": false,
                "description":"A dataModel parameter defines data that will come from the CPE data model, e.g., TR-181",
                "example":"",
                "required": ["type", "reference"],
                "anyOf": [
                    {
                        "properties": {
                           "reportTimestamp": { "const": "Unix-Epoch" },
                           "method":{ "const":"subscribe" }
                        },
                        "required": ["method"]
                    },
                    {
                        "properties": {
                           "reportTimestamp": { "const": "None" }
                        }
                    }                  
                ]
            }   
        }     
    },
    "protocolDefinitions": {
        "title":"Definitions for the supported Protocols",
        "type":"object",
        "properties": {
            "HTTP": { 
                "title":"HTTP Definition",
                "type": "object",
                "properties": {
                    "URL": {"type":"string", "description": "The URL to which the generated report should be uploaded."},
                    "Compression": {"type":"string", "enum": ["None"], "description": "Compression scheme to be used in the generated report. <b>NOTE:</b> Only \"None\" is currently supported in Telemetry 2.0."},
                    "Method": {"type":"string", "enum":["POST", "PUT"], "description": "HTTP method to be used to upload the generated report. <b>NOTE:</b> Only \"POST\" is currently supported in Telemetry 2.0."},
                    "RequestURIParameter": {
                        "title":"RequestURIParameter",
                        "type":"array",
                        "items": {
                            "type":"object",
                            "properties": {
                                "Name": {"type":"string", "description": "Value to be used as the Name in the query parameter name/value pair."} ,
                                "Reference": {"type":"string", "description": "Value to be used as the Value in the query parameter name/value pair.  Must be a data model reference."}
                            },
                            "required": ["Name", "Reference"]
                        }, 
                        "description": "Optional: Query parameters to be included in the report's upload HTTP URL."
                    }
                },
                "additionalProperties": false,
                "required": ["URL", "Compression", "Method"],
                "description": "HTTP Protocol details that will be used when Protocol=\"HTTP\"."
            },
            "RBUS_METHOD": {
                "title": "RBUS_METHOD Definition",
                "type": "object",
                "properties": {
                    "Method": {"type":"string", "description": "The name of the method to invoke via rbus. This is the method name that the provider component registers with rbus."},
                    "Parameters": {
                        "title":"Parameters to send to the provider method",
                        "type":"array",
                        "items": {
                            "type":"object",
                            "properties": {
                                "name": {"type":"string", "description": "Value to be used as the name in the method input parameter name/value pair."} ,
                                "value": {"type":"string", "description": "Value to be used as the value in the method input parameter name/value pair."}
                            },
                            "required": ["name", "value"]

                        }
                    }
                },
                "additionalProperties": false,
                "required": ["Method", "Parameters"],
                "description": "RBUS_METHOD Protocol details that will be used when Protocol=\"RBUS_METHOD\".  These method and parameters to pass to RBUS for transport of generated report."
            }
        }
    },
    "encodingDefinitions": {
        "title":"Definitions for the supported Encoding types",
        "type":"object",
        "properties": {
            "JSONEncodingDefinition": { 
                "title":"JSONEncoding Definition",
                "type": "object",
                "properties": {
                    "ReportFormat": {"type":"string", "enum": ["NameValuePair"], "description":"JSON Format to be used for JSON encoding in the generated report. <b>NOTE:</b> Only \"NameValuePair\" is currently supported in Telemetry 2.0."},
                    "ReportTimestamp": {"type":"string","enum": ["None"], "description":"Timestamp format to be used in generated report. <b>NOTE:</b> Only \"None\" is currently supported in Telemetry 2.0."}
                },
                "required": ["ReportFormat", "ReportTimestamp"],
                "description": "JSON Encoding details that will be used when EncodingType=\"JSON\"."
            }
        }
    },
    "conditionDefinitions": {
      "type":"object",
      "properties":{
          "dataModel": {
              "type": "object",
              "properties": {
                  "type":      { "type": "string", "enum": [ "dataModel" ]  },
                  "operator":  { "type": "string", "enum": [ "any", "lt", "gt", "eq" ]  },
                  "threshold": {"type":"integer"},
                  "minThresholdDuration": {"type":"integer"},
                  "reference":  { "type": "string" },
                  "report":     { "type":"boolean", "default":"true", "description": "Should this dataModel event be included in the generated report if this condition caused report generation?"}
              },
              "additionalProperties": false,
              "required": ["type", "operator", "reference"]
              }   
          }     
    },
    "ReportingAdjustments": {
        "type":"object",
        "properties":{
            "ReportOnUpdate": {"title":"ReportOnUpdate","type":"boolean", "default": false, "description": "Indicates if a report should be generated and sent before a new version of this profile is activated."},
            "FirstReportingInterval": {"title":"FirstReportingInterval","type":"integer", "default": 0, "description": "The number of seconds to wait after profile activation before generating a report once."},
            "MaxUploadLatency": {"title":"ReportOnUpdate","type":"integer", "default": 0, "description": "If present, this value is used to randomize the upload time of a generated report. Only valid when TimeReference is not equal to the default value."}
        },
        "additionalProperties": false,
        "minProperties": 1
    }     
    },




    "properties": {
        "Description": {
            "type":"string", 
            "title":"Description",
            "description":"Text describing the purpose of this Report Profile."
        },
        "Version":  { 
            "title":"Version",
            "type":"string", 
            "description":"Version of the profile. This value is opaque to the Telemetry 2 component, but can be used by server processing to indicate specifics about data available in the generated report."
        },
        "Protocol": { 
            "title":"Protocol",
            "type":"string", 
            "enum":["HTTP", "RBUS_METHOD"], 
            "description":"The protocol to be used for the upload of report generated by this profile." 
        },
        "EncodingType": { "title":"EncodingType","type":"string", "enum":["JSON"], "description": "The encoding type to be used in the report generated by this profile." },
        "ReportingInterval": {"title":"ReportingInterval","type":"integer", "description": "The interval, in seconds, at which this profile shall cause a report to be generated."},
        "ActivationTimeOut": {"title":"ActivationTimeOut","type":"integer", "description": "The amount of time, in seconds, that this profile shall remain active on the device.  This is the amount of time from which the profile is received until the CPE will consider the profile to be disabled. After this time, no further reports will be generated for this report."},
        "DeleteOnTimeOut": {"title":"DeleteOnTimeout","type":"boolean", "default":false, "description": "Indicates whether this profile should be removed from memory when the ActivationTimeOut is reached."},
        "TimeReference": {"title":"TimeReference", "type":"string", "default":"0001-01-01T00:00:00Z", "description": "An absolute time reference in UTC that indicates when a report shall be sent."},
        "GenerateNow": {"title":"GenerateNow","type":"boolean", "default": false, "description": "When true, indicates that the report for this Report Profile should be generated immediately upon receipt of the profile."},
        "Parameter": { 
            "title":"Parameter",
            "type":"array",
            "maxItems":800, 
            "items": {
                "type":"object",
                "title":"Parameter items",
                "oneOf": [
                    { "$ref": "#/definitions/parmDefinitions/properties/grep", "title": "grep" },
                    { "$ref": "#/definitions/parmDefinitions/properties/event", "title": "event" },
                    { "$ref": "#/definitions/parmDefinitions/properties/dataModel", "title": "dataModel" }
                    ]
            },
            "description": "An array of objects which defines the data to be included in the generated report. Each object defines the type of data, the source of the data and an optional name to be used as the name (marker) for this data in the generated report. "
        },
        "TriggerCondition": {
            "type":"array",
            "maxItems":50,
            "items": 
                    { "$ref": "#/definitions/conditionDefinitions/properties/dataModel"}
        },
        "HTTP": { "$ref": "#/definitions/protocolDefinitions/properties/HTTP"},
        "RBUS_METHOD": { "$ref": "#/definitions/protocolDefinitions/properties/RBUS_METHOD"},
        "JSONEncoding": {"$ref": "#/definitions/encodingDefinitions/properties/JSONEncodingDefinition"},
        "RootName": {  "title":"RootName","type":"string","default":"Report","description": "The name to be used for the root of the JSON report, e.g., \"searchResult\"." },
        "ReportingAdjustments": { "$ref":"#/definitions/ReportingAdjustments", "title": "ReportingAdjustments"}
    },
    "required": ["Protocol", "EncodingType","Parameter"],
    "additionalProperties": false,
    "anyOf": [
        {
            "properties": {
            "Protocol": { "const": "HTTP" }
            },
            "required": ["HTTP"]
        },
        {
            "properties": {
            "Protocol": { "const": "RBUS_METHOD" }
            },
            "required": ["RBUS_METHOD"]
        }
    ],
    "dependencies": {
        "DeleteOnTimeOut": { "required": ["ActivationTimeOut"] }
      }
}`
)

const PermanentTelemetryProfileConst = "PermanentTelemetryProfile"

func init() {
	var err error

	schemaLoader := gojsonschema.NewSchemaLoader()
	schemaLoader.Draft = gojsonschema.Draft6
	schemaLoader.Validate = true

	// TODO - figure out how to get the schema from the file system that will work with unit tests
	// refLoader := gojsonschema.NewReferenceLoader("file://xconf/schema/telemetrytwoprofile-schema.json")
	refLoader := gojsonschema.NewStringLoader(TelemetryTwoProfileJSONSchema)

	TelemetryTwoProfileSchema, err = schemaLoader.Compile(refLoader)
	if err != nil {
		panic(fmt.Errorf("fatal error loads and compiles JSON schema: %+v", err))
	}
}

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
	return &TelemetryProfile{ApplicationType: shared.STB}
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

func (s *PermanentTelemetryProfile) Equals(t *PermanentTelemetryProfile) bool {
	if t == nil {
		return false
	}

	if s.ID != t.ID || s.Schedule != t.Schedule || s.Expires != t.Expires || s.Name != t.Name || s.UploadRepository != t.UploadRepository ||
		s.UploadProtocol != t.UploadProtocol || s.ApplicationType != t.ApplicationType || checkEqualTelemetryElements(s.TelemetryProfile, t.TelemetryProfile) {
		return false
	}

	return true
}

func (s *PermanentTelemetryProfile) SetApplicationType(appType string) {
	s.ApplicationType = appType
}

func (s *PermanentTelemetryProfile) GetApplicationType() string {
	return s.ApplicationType
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

	return count == len(s)
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
		return common.NewRemoteErrorAS(http.StatusBadRequest, "Type is required")
	}
	if util.IsBlank(obj.Name) {
		return common.NewRemoteErrorAS(http.StatusBadRequest, "Name is empty")
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
		return common.NewRemoteErrorAS(http.StatusBadRequest, "URL is invalid")
	}

	if elements := obj.TelemetryProfile; len(elements) < 1 {
		return common.NewRemoteErrorAS(http.StatusBadRequest, "Should contain at least one profile entry")
	} else {
		for i, element := range elements {
			_, err := strconv.Atoi(element.PollingFrequency)
			if err != nil {
				return common.NewRemoteErrorAS(http.StatusBadRequest, "Polling frequency is not a number")
			}
			for j := i + 1; j < len(elements); j++ {
				if element.EqualTelemetryData(&elements[j]) {
					return common.NewRemoteErrorAS(http.StatusBadRequest, fmt.Sprintf("Profile has duplicated telemetry entry: %v", elements[j]))
				}
			}
		}
	}
	return nil
}

// NewPermanentTelemetryProfileInf constructor
func NewPermanentTelemetryProfileInf() interface{} {
	return &PermanentTelemetryProfile{
		ApplicationType: shared.STB,
	}
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

func (t *TelemetryElement) EqualTelemetryData(o *TelemetryElement) bool {
	if t == o {
		return true
	}
	if o == nil {
		return false
	}
	return t.Header == o.Header && t.Content == o.Content && t.Type == o.Type && t.PollingFrequency == o.PollingFrequency && t.Component == o.Component
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
	return &TelemetryRule{
		ApplicationType: shared.STB,
	}
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
	NoOp              bool     `json:"noOp"`
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
	return &TelemetryTwoRule{
		ApplicationType: shared.STB,
	}
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

func (obj *TelemetryTwoProfile) SetApplicationType(appType string) {
	obj.ApplicationType = appType
}

func (obj *TelemetryTwoProfile) GetApplicationType() string {
	return obj.ApplicationType
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
		return common.NewRemoteErrorAS(http.StatusBadRequest, "Type is required")
	}
	if util.IsBlank(entity.Name) {
		return common.NewRemoteErrorAS(http.StatusBadRequest, "Name is not present")
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
			return common.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("TelemetryTwo Profile with such name exists: %s", entity.Name))
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
	return &TelemetryTwoProfile{
		Type:            "TelemetryTwoProfile",
		ApplicationType: shared.STB,
	}
}

//var cachedSimpleDao ds.CachedSimpleDao

var GetCachedSimpleDaoFunc = db.GetCachedSimpleDao

func DeleteExpiredTelemetryProfile(cacheUpdateWindowSize int64) {
	telemetryProfileMapInst, err := GetCachedSimpleDaoFunc().GetAllAsMap(db.TABLE_TELEMETRY)
	if err != nil {
		log.Warn("no telemetryProfileList found for ExpireTemporaryTelemetryRules()")
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
		log.Warn(fmt.Sprintf("no telemetryProfile found for:%s ", rowKey))
		return nil
	}
	telemetry := telemetryInst.(TelemetryProfile)
	return &telemetry
}

func GetTimestampedRules() []TimestampedRule {
	timestampedRuleSet, err := GetCachedSimpleDaoFunc().GetKeys(db.TABLE_TELEMETRY)
	if err != nil {
		log.Warn("no TimestampedRule found")
		return nil
	}
	rules := make([]TimestampedRule, 0, len(timestampedRuleSet))
	for idx := range timestampedRuleSet {
		timestampedRule := timestampedRuleSet[idx].(TimestampedRule)
		rules = append(rules, timestampedRule)
	}
	return rules
}

func GetRulesFromTimestampedRules() []re.Rule {
	timestampedRuleSet, err := GetCachedSimpleDaoFunc().GetKeys(db.TABLE_TELEMETRY)
	if err != nil {
		log.Warn("no TimestampedRule found")
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
		log.Warn("no telemetryProfileMap found")
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
		log.Warn("no TelemetryProfile found")
		return nil
	}
	for idx := range tRuleList {
		tProfile := tRuleList[idx].(TelemetryProfile)
		all = append(all, &tProfile)
	}
	return all
}

func GetTelemetryRuleListForAs() []*TelemetryRule {
	all := []*TelemetryRule{}
	tRuleList, err := GetCachedSimpleDaoFunc().GetAllAsList(db.TABLE_TELEMETRY_RULES, 0)
	if err != nil {
		log.Warn("no TelemetryRule found")
		return nil
	}
	for idx := range tRuleList {
		tRule := tRuleList[idx].(*TelemetryRule)
		all = append(all, tRule)
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

	if len(tRuleList) == 0 {
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
		log.Warn(fmt.Sprintf("no telemetryProfile found for:%s ", rowKey))
		return nil
	}
	telemetry := telemetryInst.(*PermanentTelemetryProfile)
	return telemetry
}

func GetPermanentTelemetryProfileList() []*PermanentTelemetryProfile {
	all := []*PermanentTelemetryProfile{}
	tRuleList, err := GetCachedSimpleDaoFunc().GetAllAsList(db.TABLE_PERMANENT_TELEMETRY, 0)
	if err != nil {
		log.Warn("no TelemetryProfile found")
		return nil
	}
	for idx := range tRuleList {
		tProfile := tRuleList[idx].(*PermanentTelemetryProfile)
		all = append(all, tProfile)
	}
	return all
}

func GetTelemetryTwoRuleList() []*TelemetryTwoRule {
	cm := db.GetCacheManager()
	cacheKey := "TelemetryTwoRuleList"
	cacheInst := cm.ApplicationCacheGet(db.TABLE_TELEMETRY_TWO_RULES, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*TelemetryTwoRule)
	}
	tRuleList, err := GetCachedSimpleDaoFunc().GetAllAsList(db.TABLE_TELEMETRY_TWO_RULES, 0)
	if err != nil {
		log.Warn("no TelemetryTwoRule found")
		return nil
	}

	if len(tRuleList) == 0 {
		return nil
	}

	all := make([]*TelemetryTwoRule, 0, len(tRuleList))

	for _, itf := range tRuleList {
		if telemetryTwoRule, ok := itf.(*TelemetryTwoRule); ok {
			all = append(all, telemetryTwoRule)
		}
	}
	cm.ApplicationCacheSet(db.TABLE_TELEMETRY_TWO_RULES, cacheKey, all)

	return all
}

func GetTelemetryTwoRuleListForAS() []*TelemetryTwoRule {
	all := []*TelemetryTwoRule{}
	tRuleList, err := GetCachedSimpleDaoFunc().GetAllAsList(db.TABLE_TELEMETRY_TWO_RULES, 0)
	if err != nil {
		log.Warn("no TelemetryTwoRule found")
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
		log.Warn(fmt.Sprintf("no TelemetryTwoProfile found for: %s ", rowKey))
		return nil
	}
	telemetry := telemetryInst.(*TelemetryTwoProfile)
	return telemetry
}

// ValidateTelemetryTwoProfileJson validates JSON against the schema
func ValidateTelemetryTwoProfileJson(json string) error {
	jsonLoader := gojsonschema.NewStringLoader(json)
	result, err := TelemetryTwoProfileSchema.Validate(jsonLoader)
	if err != nil {
		return err
	}
	if result.Valid() {
		return nil
	} else {
		var errList []string
		for _, err := range result.Errors() {
			errList = append(errList, err.String())
		}
		log.Errorf("Invalid Telemetry 2.0 Profile JSON config data: %s", errors.New(strings.Join(errList, ". ")))

		return common.NewRemoteErrorAS(http.StatusBadRequest, "Please provide the valid Telemetry 2.0 Profile JSON config data.")
	}
}
