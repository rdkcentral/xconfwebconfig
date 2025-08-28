/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
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
package tests

const (
	// arg1: telemetry2rule uuid
	// arg2: telemetry2rule name
	// arg3: referenced telemetry2profile uuid
	MockTelemetryTwoRuleTemplate1 = `{
  "negated": false,
  "compoundParts": [
    {
      "negated": false,
      "condition": {
        "freeArg": {
          "type": "STRING",
          "name": "model"
        },
        "operation": "IS",
        "fixedArg": {
          "bean": {
            "value": {
              "java.lang.String": "TG7777"
            }
          }
        }
      },
      "compoundParts": []
    }
  ],
  "id": "%v",
  "name": "%v",
  "boundTelemetryIds": [
    "%v"
  ],
  "applicationType": "stb"
}`

	// arg1: namelist primary key
	// arg2: telemetry2rule uuid
	// arg3: telemetry2rule name
	// arg4: referenced telemetry2profile uuid
	MockTelemetryTwoRuleTemplate2 = `{
  "negated": false,
  "compoundParts": [
    {
      "negated": false,
      "condition": {
        "freeArg": {
          "type": "STRING",
          "name": "estbMacAddress"
        },
        "operation": "IN_LIST",
        "fixedArg": {
          "bean": {
            "value": {
              "java.lang.String": "%v"
            }
          }
        }
      },
      "compoundParts": []
    }
  ],
  "id": "%v",
  "name": "%v",
  "boundTelemetryIds": [
    "%v"
  ],
  "applicationType": "stb"
}`

	// arg1: namelist primary key
	// arg2: telemetry2rule uuid
	// arg3: telemetry2rule name
	// arg4: referenced telemetry2profile uuid
	MockTelemetryTwoRuleTemplate3 = `{
  "negated": false,
  "compoundParts": [
    {
      "negated": false,
      "condition": {
        "freeArg": {
          "type": "STRING",
          "name": "ipAddress"
        },
        "operation": "IN_LIST",
        "fixedArg": {
          "bean": {
            "value": {
              "java.lang.String": "%v"
            }
          }
        }
      },
      "compoundParts": []
    }
  ],
  "id": "%v",
  "name": "%v",
  "boundTelemetryIds": [
    "%v"
  ],
  "applicationType": "stb"
}`

	// arg1: namelist primary key
	// arg2: telemetry2rule uuid
	// arg3: telemetry2rule name
	// arg4: referenced telemetry2profile uuid
	MockTelemetryTwoRuleTemplate9 = `{
  "negated": false,
  "compoundParts": [
    {
      "negated": false,
      "condition": {
        "freeArg": {
          "type": "STRING",
          "name": "estbMacAddress"
        },
        "operation": "IN_LIST",
        "fixedArg": {
          "bean": {
            "value": {
              "java.lang.String": "%v"
            }
          }
        }
      },
      "compoundParts": []
    },
    {
      "negated": false,
      "relation": "OR",
      "condition": {
        "freeArg": {
          "type": "STRING",
          "name": "env"
        },
        "operation": "IS",
        "fixedArg": {
          "bean": {
            "value": {
              "java.lang.String": "WTEST"
            }
          }
        }
      },
      "compoundParts": []
    },
    {
      "negated": false,
      "relation": "OR",
      "condition": {
        "freeArg": {
          "type": "STRING",
          "name": "model"
        },
        "operation": "IS",
        "fixedArg": {
          "bean": {
            "value": {
              "java.lang.String": "WTEST"
            }
          }
        }
      },
      "compoundParts": []
    },
    {
      "negated": false,
      "relation": "OR",
      "condition": {
        "freeArg": {
          "type": "STRING",
          "name": "partnerId"
        },
        "operation": "IS",
        "fixedArg": {
          "bean": {
            "value": {
              "java.lang.String": "WSMITHPART"
            }
          }
        }
      },
      "compoundParts": []
    },
    {
      "negated": false,
      "relation": "OR",
      "condition": {
        "freeArg": {
          "type": "STRING",
          "name": "accountId"
        },
        "operation": "IS",
        "fixedArg": {
          "bean": {
            "value": {
              "java.lang.String": "112233445566"
            }
          }
        }
      },
      "compoundParts": []
    },
    {
      "negated": false,
      "relation": "OR",
      "condition": {
        "freeArg": {
          "type": "STRING",
          "name": "randomParam"
        },
        "operation": "IS",
        "fixedArg": {
          "bean": {
            "value": {
              "java.lang.String": "1234"
            }
          }
        }
      },
      "compoundParts": []
    }
  ],
  "id": "%v",
  "name": "%v",
  "boundTelemetryIds": [
    "%v"
  ],
  "applicationType": "stb"
}`

	// arg1 telemetry2profile uuid
	// arg2 telemetry2profile name
	MockTelemetryTwoProfileTemplate1 = `{
  "id": "%v",
  "updated": 1617207461643,
  "name": "%v",
  "jsonconfig": "{\n    \"Description\":\"Telemetry 2.0 Test Profile\",\n    \"Version\":\"0.1\",\n    \"Protocol\":\"HTTP\",\n    \"EncodingType\":\"JSON\",\n    \"ReportingInterval\":60,\n    \"TimeReference\":\"0001-01-01T00:00:00Z\",\n    \"ActivationTimeout\":600,\n    \"Parameter\": [\n        {\"type\":\"dataModel\", \"reference\":\"Profile.Name\"} ,\n        {\"type\":\"dataModel\", \"reference\":\"Profile.Description\"} ,\n        {\"type\":\"dataModel\", \"reference\":\"Profile.Version\"} ,\n        {\"type\":\"dataModel\", \"reference\":\"Device.WiFi.Radio.1.MaxBitRate\"} ,\n        {\"type\":\"dataModel\", \"reference\":\"Device.WiFi.Radio.1.OperatingFrequencyBand\"}\n    ],\n    \"HTTP\": {\n        \"URL\":\"https://test.xcal.tv/\",\n        \"Compression\":\"None\",\n        \"Method\":\"POST\",\n        \"RequestURIParameter\": [\n            {\"Name\":\"profileName\", \"Reference\":\"Profile.Name\" },\n            {\"Name\":\"reportVersion\", \"Reference\":\"Profile.Version\" }\n        ]\n    },\n    \"JSONEncoding\": {\n        \"ReportFormat\":\"NameValuePair\",\n        \"ReportTimestamp\": \"None\"\n    }\n}",
  "applicationType": "stb"
}`
	// this profile is not a valid json
	MockTelemetryTwoProfileTemplate2 = `{
  "id": "%v",
  "updated": 1617207461643,
  "name": "%v",
  "jsonconfig": "{\n    \"Description\":\"Telemetry 2.0 Test Profile\",\n    \"Version\":\"0.1\",\n    \"Protocol\":\"HTTP\",\n    \"EncodingType\":\"JSON\",\n    \"ReportingInterval\":60,\n    \"TimeReference\":\"0001-01-01T00:00:00Z\",\n    \"ActivationTimeout\":600,\n    \"Parameter\": [\n        {\"type\":\"dataModel\", \"reference\":\"Profile.Name\"} ,\n        {\"type\":\"dataModel\", \"reference\":\"Profile.Description\"} ,\n        {\"type\":\"dataModel\", \"reference\":\"Profile.Version\"} ,\n        {\"type\":\"dataModel\", \"reference\":\"Device.WiFi.Radio.1.MaxBitRate\"} ,\n        {\"type\":\"dataModel\", \"reference\":\"Device.WiFi.Radio.1.OperatingFrequencyBand\"}\n    ],\n    \"HTTP\": {\n        \"URL\":\"https://test.xcal.tv/\",\n        \"Compression\":\"None\",\n        \"Method\":\"POST\",\n        \"RequestURIParameter\": [\n            {\"Name\":\"profileName\", \"Reference\":\"Profile.Name\" },\n            {\"Name\":\"reportVersion\", \"Reference\":\"Profile.Version\" }\n        ]\n    },\n    \"JSONEncoding\": {\n        \"ReportFormat\":\"NameValuePair\",\n        \"ReportTimestamp\": \"None\"\n    },\n}",
  "applicationType": "stb"
}`
	TestTelemetryTwoProfileJsonConfig = `{
  "Description":"Test for Test",
  "Version":"0.1",
  "Protocol":"HTTP",
  "EncodingType":"JSON",
  "ReportingInterval":300,
  "ActivationTimeOut":900,
  "TimeReference":"0001-01-01T00:00:00Z",
  "Parameter":
      [
          { "type": "dataModel", "reference": "Profile.Name"}, 
          { "type": "dataModel", "reference": "Profile.Version"},
          {"type":"dataModel", "name":"mac", "reference":"Device.DeviceInfo.X_COMCAST-COM_STB_MAC"},
          { "type": "event", "eventName":"RECONNECT_REASON_POWER_KEY",
          "name":"TEST_RECONNECT_REASON_POWER_KEY","component":"receiver","use":"count"}          
      ],
  "HTTP": {
      "URL":"https://test-server.com",
      "Compression":"None",
      "Method":"POST",
      "RequestURIParameter": [
          {"Name":"profileName", "Reference":"Profile.Name" },
          {"Name":"reportVersion", "Reference":"Profile.Version" }
      ]
  },
  "JSONEncoding": {
      "ReportFormat":"NameValuePair",
      "ReportTimestamp": "None"
  }
}`
)
