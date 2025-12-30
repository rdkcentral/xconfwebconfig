package tests

import (
	"fmt"

	"github.com/rdkcentral/xconfwebconfig/shared"
)

const (
	FirmwareConfigId1                                       = "de529a04-3bab-41e3-ad79-f1e583723b47"
	FirmwareConfigId2                                       = "393e2152-9d50-4f30-aab9-c74977471632"
	FirmwareConfigId3                                       = "e4b10a02-094b-4941-8aee-6b10a996829d"
	FirmwareConfigId4                                       = "e4b10a02-094b-4941-8aee-6b10a996829e"
	firmwareRuleId1                                         = "e05a5b92-8605-4309-bfe5-25646e888137"
	firmwareRuleId2                                         = "aa534186-ef60-4516-8c47-c254f9066c22"
	firmwareRuleId3                                         = "64a19e12-21d0-4a72-9f0e-346fa53c3c67"
	firmwareRuleId4                                         = "64a19e12-21d0-4a72-9f0e-346fa53c3c68"
	mac1                                                    = "11:11:22:22:33:33"
	mac2                                                    = "22:22:33:33:44:44"
	mac2a                                                   = "22:22:33:33:44:AA"
	mac3                                                    = "33:33:44:44:55:55"
	namespaceListKey                                        = "scarletoverkill"
	NamespaceIPListKey                                      = "myipaddresstests"
	IpAddress1                                              = "10.0.0.101"
	IPAddress2                                              = "10.0.0.1"
	IpAddress3                                              = "10.0.0.12"
	IpAddress4                                              = "10.0.0.11"
	IPAddressV61                                            = "2600:1f18:227b:c01:b111:3d17:7a86:ab36"
	IPAddressV62                                            = "2600:1f18:227b:c01:b111:3d17:7a86:ab37"
	DownloadLocationRoundRobinFilterHTTPFULLURLLOCATION     = "http://comcast.com"
	DownloadLocationRoundRobinFilterHTTPLOCATION            = "comcast.com"
	DownloadLocationRoundRobinFilterIPADDRESS               = "192.168.1.1"
	RDKCLOUD_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE_KEY = "RDKCLOUD_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE"
	SKY_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE_KEY      = "SKY_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE"
	FIREBOLT_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE_KEY = "FIREBOLT_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE"
)

var (
	NamespaceIPList = []string{"10.0.0.11", "10.0.0.12", "10.0.0.101"}

	firmwareConfigJsonTemplate1 = `{
  "id": "%v",
  "updated": 1591807259972,
  "description": "1-3939",
  "supportedModelIds": [
    "DPC9999",
    "DPC9999T"
  ],
  "firmwareFilename": "DPC9999_4.2p1s8_DEV_sey-signed.bin",
  "firmwareVersion": "DPC9999_4.2p1s8_DEV_sey-signed",
  "applicationType": "stb"
}`
	firmwareConfigJsonTemplate2 = `{
  "id": "%v",
  "updated": 1591807259972,
  "description": "1-3939",
  "supportedModelIds": [
    "DPC8888",
    "DPC8888T"
  ],
  "firmwareFilename": "DPC8888_4.2p1s8_DEV_sey-signed.bin",
  "firmwareVersion": "DPC8888_4.2p1s8_DEV_sey-signed",
  "applicationType": "stb"
}`
	firmwareConfigJsonTemplate3 = `{
  "id": "%v",
  "updated": 1591807259972,
  "description": "1-3939",
  "supportedModelIds": [
    "DPC7777",
    "DPC7777T"
  ],
  "firmwareFilename": "DPC7777_4.2p1s8_DEV_sey-signed.bin",
  "firmwareVersion": "DPC7777_4.2p1s8_DEV_sey-signed",
  "applicationType": "stb"
}`

	fwRuleJsonTemplate1 = `{
  "id": "%v",
  "name": "1-3939",
  "rule": {
    "negated": false,
    "condition": {
      "freeArg": {
        "type": "STRING",
        "name": "eStbMac"
      },
      "operation": "IS",
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
  "applicableAction": {
    "type": ".RuleAction",
    "actionType": "RULE",
    "configId": "%v",
    "configEntries": [],
    "active": true,
    "useAccountPercentage": false,
    "firmwareCheckRequired": false,
    "rebootImmediately": false
  },
  "type": "IV_RULE",
  "active": true,
  "applicationType": "stb"
}`

	fwRuleJsonTemplate2 = `{
  "id": "%v",
  "name": "1717_LED_AXG1v1",
  "rule": {
    "negated": false,
    "compoundParts": [
      {
        "negated": false,
        "condition": {
          "freeArg": {
            "type": "STRING",
            "name": "eStbMac"
          },
          "operation": "IN_LIST",
          "fixedArg": {
            "bean": {
              "value": {
                "java.lang.String": "1717_LED_AXG1v3"
              }
            }
          }
        }
      },
      {
        "negated": false,
        "relation": "OR",
        "condition": {
          "freeArg": {
            "type": "STRING",
            "name": "eStbMac"
          },
          "operation": "IS",
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
            "name": "eStbMac"
          },
          "operation": "IN",
          "fixedArg": {
            "collection": {
              "value": [
                "%v"
              ]
            }
          }
        },
        "compoundParts": []
      }
    ]
  },
  "applicableAction": {
    "type": ".RuleAction",
    "actionType": "RULE",
    "configId": "%v",
    "configEntries": [],
    "active": true,
    "useAccountPercentage": false,
    "firmwareCheckRequired": true,
    "rebootImmediately": true
  },
  "type": "MAC_RULE",
  "active": true,
  "applicationType": "stb"
}`

	fwRuleJsonTemplate3 = `{
  "id": "%v",
  "name": "000ipPerformanceTestRule",
  "rule": {
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
        }
      },
      {
        "negated": false,
        "relation": "AND",
        "condition": {
          "freeArg": {
            "type": "STRING",
            "name": "env"
          },
          "operation": "IS",
          "fixedArg": {
            "bean": {
              "value": {
                "java.lang.String": "TEST"
              }
            }
          }
        }
      },
      {
        "negated": false,
        "relation": "AND",
        "condition": {
          "freeArg": {
            "type": "STRING",
            "name": "model"
          },
          "operation": "IS",
          "fixedArg": {
            "bean": {
              "value": {
                "java.lang.String": "XCONFTESTMODEL"
              }
            }
          }
        }
      }
    ]
  },
  "applicableAction": {
    "type": ".RuleAction",
    "actionType": "RULE",
    "configId": "%v",
    "configEntries": [],
    "active": true,
    "useAccountPercentage": false,
    "firmwareCheckRequired": false,
    "rebootImmediately": false
  },
  "type": "IP_RULE",
  "active": true,
  "applicationType": "stb"
}`
	fwRuleJsonTemplate4 = `{
		"id": "%v",
		"name": "000ipPerformanceTestRule",
		"rule": {
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
              }
            }
          ]
		},
		"applicableAction": {
		  "type": ".RuleAction",
		  "actionType": "RULE",
		  "configId": "%v",
		  "configEntries": [],
		  "active": true,
		  "useAccountPercentage": false,
		  "firmwareCheckRequired": false,
		  "rebootImmediately": false,
		  "properties": {
			  "firmwareLocation": "http://127.0.1.1/app/download",
			  "ipv6FirmwareLocation": "http://127.0.1.1/app/downloadv6",
			  "irmwareDownloadProtocol": "https"
		  }
		},
		"type": "IP_RULE",
		"active": true,
		"applicationType": "stb"
	  }`

	RDKCLOUD_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE_TEMPLATE = `{
     "type":"com.comcast.xconf.estbfirmware.DownloadLocationRoundRobinFilterValue",
     "id":"RDKCLOUD_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE",
     "updated":1581038891097,
     "applicationType":"%v",
     "locations":[{"locationIp":"%v","percentage":100.0},{"locationIp":"%v","percentage":0.0}],
     "ipv6locations":[{"locationIp":"%v","percentage":50.0},{"locationIp":"%v","percentage":50.0}],
     "httpLocation":"dac15cdlserver.ae.ccp.xcal.tv",
     "httpFullUrlLocation":"https://dac15cdlserver.ae.ccp.xcal.tv/Images"
    }`
	SKY_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE_TEMPLATE = `{
    "type":"com.comcast.xconf.estbfirmware.DownloadLocationRoundRobinFilterValue",
    "id":"SKY_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE",
    "updated":1615812930296,"applicationType":"sky",
    "locations":[{"locationIp":"96.114.220.246","percentage":100.0},{"locationIp":"69.252.106.162","percentage":0.0}],
    "ipv6locations":[{"locationIp":"2600:1f18:227b:c01:b111:3d17:7a86:ab36","percentage":50.0},{"locationIp":"2001:558:1020:1:250:56ff:fe94:646f","percentage":50.0}],
    "httpLocation":"dac15cdlserver.ae.ccp.xcal.tv",
    "httpFullUrlLocation":"https://dac15cdlserver.ae.ccp.xcal.tv/Images"
  }`
	FIREBOLT_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE_TEMPLATE = `{
    "type":"com.comcast.xconf.estbfirmware.DownloadLocationRoundRobinFilterValue",
    "id":"FIREBOLT_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE",
    "updated":1538081702207,
    "applicationType":"firebolt",
    "locations":[{"locationIp":"11.11.11.11","percentage":100.0}],
    "ipv6locations":[],
    "rogueModels":[{"id":"NEWW1","updated":1490803583884,"ttlMap":{},"description":"newww"}],
    "httpLocation":"test.com",
    "httpFullUrlLocation":"http://test.com:8080/Images",
    "neverUseHttp":false,
    "firmwareVersions":"SERICAM2_3.1s1_VBNsd\nABC\nTG3482SHW_DEV_2.8_p14axb6_20171222031047sdy\nDPC3941_2.9p1s5_DEV_sey\nSERXW3_2.6s3_VBNsd\nSERXW3_VBN_master_043018152018sd_NOCHK_2054\nSERICAM2_3.1s2_VBNsd\nTG3482_2.8p19s1_DEV_sey\nSERICAM2_VBN_master_042007592018sd_NOCHK_GRT\nTG1682COX_DEV_master_20180103230428sdy_N\nSERXW3_3.0p3s1_PRODsd\nCGA4131COM_2.9s6_DEV_sey\nCGM4140COMCOX_DEV_master_20171227230711sdy_NG\nPX5001_VBN_master_20171221160245sdy\nTG1682COX_DEV_master_20180101230410sdy_N\nTG3482SHW_2.8p22s1_DEV_sey\nSERXW3_VBN_master_042703462018sd_NOCHK\nCGM4140COMCOX_DEV_master_20180103230730sdy_NG\nSERICAM2_VBN_master_042703462018sd_NOCHK\nSERICAM2_VBN_1808_sprint_080700412018sd_NOCHK_NG\nSERXW3_VBN_master_071809322017sd\nSERICAM2_VBN_master_052722042018sd_NOCHK_NG"
  }`
	firmwareConfig1Bytes []byte
	firmwareConfig2Bytes []byte
	firmwareConfig3Bytes []byte
	firmwareRule1Bytes   []byte
	firmwareRule2Bytes   []byte
	firmwareRule3Bytes   []byte

	firmwareRuleTemplateTemplateOne = `{
    "id":"IP_RULE",
    "rule":{
        "negated":false,
        "compoundParts":[
          {
             "negated":false,
            "condition":
              {"freeArg": {"type":"ANY","name":"Tag31"},
              "operation":"EXISTS",
              "fixedArg":{"bean":{"value":{"java.lang.String":""}}}         
            },
            "compoundParts":[]
          },
          {
            "negated":false,
            "relation":"AND",
            "condition":
              {"freeArg":{"type":"ANY","name":"Tag32"},
              "operation":"EXISTS",
              "fixedArg":{"bean":{"value":{"java.lang.String":""}}}
              },
              "compoundParts":[]
          },
          {
              "negated":false,
              "relation":"AND",
              "condition":{"freeArg":
                            {"type":"ANY",
                            "name":"Tag33"
                          },
                          "operation":"EXISTS",
                          "fixedArg":{"bean":{"value":{"java.lang.String":""}}}
                          },
              "compoundParts":[]
          },
          {
            "negated":false,
             "relation":"AND","condition":{"freeArg":{"type":"ANY","name":"Tag34"},
             "operation":"EXISTS","fixedArg":{"bean":{"value":{"java.lang.String":""}}}
             },
             "compoundParts":[]
          },
          {
            "negated":false,"relation":"AND","condition":{"freeArg":{"type":"ANY","name":"Tag35"},
                "operation":"EXISTS","fixedArg":{"bean":{"value":{"java.lang.String":""}}}
            },
            "compoundParts":[]
          },
          {
            "negated":false,"relation":"AND","condition":{"freeArg":{"type":"ANY","name":"Tag36"},"operation":"EXISTS","fixedArg":{"bean":{"value":{"java.lang.String":""}}}
            },
            "compoundParts":[]
          },
          {
            "negated":false,"relation":"AND","condition":{"freeArg":{"type":"ANY","name":"Tag37"},"operation":"EXISTS","fixedArg":{"bean":{"value":{"java.lang.String":""}}}
            },
            "compoundParts":[]
          },
          {
            "negated":false,"relation":"AND","condition":{"freeArg":{"type":"ANY","name":"Tag38"},"operation":"EXISTS","fixedArg":{"bean":{"value":{"java.lang.String":""}}}},
            "compoundParts":[]
          },
          {
            "negated":false,"relation":"AND","condition":{"freeArg":{"type":"ANY","name":"Tag39"},"operation":"EXISTS","fixedArg":{"bean":{"value":{"java.lang.String":""}}}},
            "compoundParts":[]
          },
          {
            "negated":false,"relation":"AND","condition":{"freeArg":{"type":"ANY","name":"Tag40"},"operation":"EXISTS","fixedArg":{"bean":{"value":{"java.lang.String":""}}}},
            "compoundParts":[]
          },
          {
            "negated":false,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":""}}}},
            "compoundParts":[]
          },
          {
            "negated":false,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"env"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":""}}}},
            "compoundParts":[]
          }
        ]
      },
      "applicableAction":
      {
        "type":".RuleAction",
        "actionType": "RULE_TEMPLATE",
        "active":true,
        "useAccountPercentage":false,
        "firmwareCheckRequired":false,
        "rebootImmediately":false
      },
      "priority":55,
      "requiredFields":[],
      "byPassFilters":[],
      "editable":true
  }`
	firmwareRuleTemplateTemplateTwo = `{ 
    "id":"MAC_RULE",
    "rule":
    {
        "negated":false,
        "compoundParts":
        [
          {
            "negated":false,
            "condition":
            {
              "freeArg":
              {
                "type":"STRING","name":"eStbMac"
              },
              "operation":"IN_LIST",
              "fixedArg":{"bean":{"value":{"java.lang.String":"AKHIL-MAC-LIST2"}}}
          },
          "compoundParts":[]
          },
          {
            "negated":false,
            "relation":"AND",
            "condition":
            {
              "freeArg":
              {
                "type":"ANY",
                "name":"additionalFwVerInfo"
              },
              "operation":"EXISTS",
              "fixedArg":
              {
                "bean":{"value":{"java.lang.String":""}}}
              },
              "compoundParts":[]
            },
          {
            "negated":false,
            "relation":"AND",
            "condition":
            {
              "freeArg":
              {
                "type":"STRING",
                "name":"model"
              },
              "operation":"IS",
              "fixedArg":
              {"bean":
                {"value":
                {"java.lang.String":"SKXI11ANS"}}}
              },
              "compoundParts":[]},
            {
              "negated":false,
              "relation":"AND",
              "condition":
                {
                  "freeArg":
                  {
                    "type":"STRING",
                    "name":"model"
                  },
                  "operation":"IS"
                  ,
                  "fixedArg":
                  {
                    "bean":
                    {"value":{"java.lang.String":"SKXI11AIS"}}}
                  },
                  "compoundParts":[]
                },
                {
                  "negated":false,
                  "relation":"AND",
                  "condition":
                    {"freeArg":
                      {"type":"STRING","name":"model"
                      },
                      "operation":"IS",
                      "fixedArg":
                      {"bean":
                        {"value":{"java.lang.String":"SKXI11ADS"}}}
                        },
                        "compoundParts":[]
                }
      ]
    },
    "applicableAction":
    {
        "type":".RuleAction",
        "actionType":"RULE_TEMPLATE",
        "active":true,
        "useAccountPercentage":false,
        "firmwareCheckRequired":false,
        "rebootImmediately":false
    },
    "priority":37,
    "requiredFields":[],
    "byPassFilters":[],
    "editable":true
  }`
	firmwareRuleTemplateTemplateThree = `{ 
      "id":"GLOBAL_PERCENT",
      "rule":
      {
        "negated":false,
        "compoundParts":
        [
          {"negated":true,"condition":{"freeArg":{"type":"STRING","name":"matchedRuleType"},"operation":"IN","fixedArg":{"collection":{"value":["ENV_MODEL_RULE","MIN_CHECK_RULE","IV_RULE"]}}}},
          {"negated":false,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"eStbMac"},"operation":"PERCENT","fixedArg":{"bean":{"value":{"java.lang.Double":0.0}}}}},
          {"negated":true,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"ipAddress"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":""}}}}}
        ]
      },
      "applicableAction":
      {
        "type":".ApplicableAction","ttlMap":{},"actionType":"BLOCKING_FILTER_TEMPLATE"
      },
      "priority":13,
      "requiredFields":[],
      "byPassFilters":[],
      "editable":true
  }`
	firmwareRuleTemplateTemplateFour = `{
    "id":"REMCTR_XR15-20_ENV_MODEL_RULE",
    "rule":{
      "negated":false,
      "compoundParts":[
        {
          "negated":false,
          "condition":{
            "freeArg":{
              "type":"STRING",
              "name":"model"
            },
            "operation":"IS",
            "fixedArg":{
              "bean":{
                "value":{
                  "java.lang.String":""
                }
              }
            }
          },
          "compoundParts":[]
        },
        {
          "negated":false,
          "relation":"AND",
          "condition":{
            "freeArg":{
              "type":"ANY",
              "name":"remCtrlXR15-20"
            },
            "operation":"EXISTS",
            "fixedArg":{
              "bean":{
                "value":{
                  "java.lang.String":""
                }
              }
            }
          },
          "compoundParts":[]
        },
        {
          "negated":false,
          "relation":"AND",
          "condition":{
            "freeArg":{
              "type":"STRING",
              "name":"eStbMac"
            },
            "operation":"IN_LIST",
            "fixedArg":{
              "bean":{
                "value":{
                  "java.lang.String":""
                }
              }
            }
          },
          "compoundParts":[]
        }
      ]
    },
    "applicableAction":{
      "type":".DefinePropertiesTemplateAction",
      "actionType":"DEFINE_PROPERTIES_TEMPLATE",
      "properties":{
        "remCtrlXR15-20":{
          "value":"",
          "optional":false,
          "validationTypes":[
            "STRING"
          ]
        },
        "remCtrlXR15-20Audio":{
          "value":"",
          "optional":true,
          "validationTypes":[
            "STRING"
          ]
        }
      },
      "byPassFilters":[],
      "firmwareVersionRegExs":[]
    },
    "priority":80,
    "requiredFields":[],
    "byPassFilters":[],
    "editable":true
  }`
)

func init() {
	firmwareConfig1Bytes = []byte(fmt.Sprintf(firmwareConfigJsonTemplate1, FirmwareConfigId1))
	firmwareConfig2Bytes = []byte(fmt.Sprintf(firmwareConfigJsonTemplate2, FirmwareConfigId2))
	firmwareConfig3Bytes = []byte(fmt.Sprintf(firmwareConfigJsonTemplate3, FirmwareConfigId3))
	firmwareRule1Bytes = []byte(fmt.Sprintf(fwRuleJsonTemplate1, firmwareRuleId1, mac1, FirmwareConfigId1))
	firmwareRule2Bytes = []byte(fmt.Sprintf(fwRuleJsonTemplate2, firmwareRuleId2, mac2a, mac2, FirmwareConfigId2))
	firmwareRule3Bytes = []byte(fmt.Sprintf(fwRuleJsonTemplate3, firmwareRuleId3, namespaceListKey, FirmwareConfigId3))
}

func GetFirmwareConfigStr1() string {
	return fmt.Sprintf(firmwareConfigJsonTemplate1, FirmwareConfigId1)
}

func GetFirmwareConfigStr2() string {
	return fmt.Sprintf(firmwareConfigJsonTemplate2, FirmwareConfigId2)
}

func GetFirmwareRuleStr1() string {
	return fmt.Sprintf(fwRuleJsonTemplate1, firmwareRuleId1, mac1, FirmwareConfigId1)
}

func GetFirmwareRuleStr2() string {
	return fmt.Sprintf(fwRuleJsonTemplate2, firmwareRuleId2, mac2a, mac2, FirmwareConfigId2)
}

func GetFirmwareRuleStr3() string {
	return fmt.Sprintf(fwRuleJsonTemplate3, firmwareRuleId3, namespaceListKey, FirmwareConfigId3)
}

func GetFirmwareRuleStr4() string {
	return fmt.Sprintf(fwRuleJsonTemplate4, firmwareRuleId4, NamespaceIPListKey, FirmwareConfigId1)
}

func GetRDKCDownloadLocationROUNDROBINFILTERVALUE() string {
	return fmt.Sprintf(RDKCLOUD_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE_TEMPLATE, shared.STB, IpAddress3, IpAddress4, IPAddressV61, IPAddressV62)
}

func GetFirmwareTemplateStr1() string {
	return firmwareRuleTemplateTemplateOne
}

func GetFirmwareTemplateStr2() string {
	return firmwareRuleTemplateTemplateTwo
}

func GetFirmwareTemplateStr3() string {
	return firmwareRuleTemplateTemplateThree
}

func GetFirmwareTemplateStr4() string {
	return firmwareRuleTemplateTemplateFour
}

var (
	formulaData01 = []byte(`{
  "formula": {
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
            "java.lang.String": "CORE_NW_MAC_LIST"
          }
        }
      }
    },
    "compoundParts": [],
    "id": "e7f7ff14-a39d-4627-9e9d-a1690babf4c1",
    "name": "CORE_NW_DCM_FORMULA",
    "description": "CORE_NW_DCM_FORMULA",
    "priority": 1,
    "percentage": 100,
    "percentageL1": 0,
    "percentageL2": 0,
    "percentageL3": 0,
    "applicationType": "stb"
  },
  "deviceSettings": {
    "id": "e7f7ff14-a39d-4627-9e9d-a1690babf4c2",
    "name": "CORE_NW_DCM_FORMULA",
    "checkOnReboot": true,
    "settingsAreActive": true,
    "schedule": {
      "type": "ActNow",
      "expression": "15 1 * * *",
      "timeZone": "UTC",
      "timeWindowMinutes": 0
    },
    "applicationType": "stb"
  },
  "logUploadSettings": {
    "id": "e7f7ff14-a39d-4627-9e9d-a1690babf4c3",
    "name": "CORE_NW_DCM_FORMULA",
    "uploadOnReboot": false,
    "numberOfDays": 0,
    "areSettingsActive": true,
    "schedule": {
      "type": "ActNow",
      "expression": "1 0 * * *",
      "timeZone": "UTC",
      "expressionL1": "",
      "expressionL2": "",
      "expressionL3": "",
      "timeWindowMinutes": 0
    },
    "uploadRepositoryId": "982a7ac4-0049-489c-8b63-4539f525aa39",
    "applicationType": "stb"
  },
  "vodSettings": {
    "id": "e7f7ff14-a39d-4627-9e9d-a1690babf4c4",
    "name": "CORE_NW_DCM_FORMULA",
    "locationsURL": "https://test.net",
    "ipNames": [],
    "ipList": [],
    "srmIPList": {},
    "applicationType": "stb"
  }
}`)

	ruleData01 = []byte(`{
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
            "java.lang.String": "CORE_NW_MAC_LIST"
          }
        }
      }
    },
    "compoundParts": [],
    "id": "e7f7ff14-a39d-4627-9e9d-a1690babf4c5",
    "name": "CORE_NW_DCM_FORMULA",
    "description": "CORE_NW_DCM_FORMULA",
    "priority": 1,
    "percentage": 100,
    "percentageL1": 0,
    "percentageL2": 0,
    "percentageL3": 0,
    "applicationType": "stb"
  }`)

	rawdata01 = []byte(`
{
    "id": "6bfb5b5d-e800-4e3e-9da3-34eb16a070bd",
    "updated": 0,
    "ttlmap": null,
    "name": "123",
    "Rule": {
        "compoundParts": [
            {
                "compoundParts": null,
                "condition": {
                    "freeArg": {
                        "type": "STRING",
                        "name": "timeZone"
                    },
                    "operation": "IS",
                    "fixedArg": {
                        "bean": {
                            "value": {
                                "java.lang.String": "UTC",
                                "java.lang.Double": 0
                            }
                        }
                    }
                },
                "negated": false,
                "relation": ""
            },
            {
                "compoundParts": null,
                "condition": {
                    "freeArg": {
                        "type": "STRING",
                        "name": "time"
                    },
                    "operation": "GTE",
                    "fixedArg": {
                        "bean": {
                            "value": {
                                "java.lang.String": "00:09:00",
                                "java.lang.Double": 0
                            }
                        }
                    }
                },
                "negated": false,
                "relation": "AND"
            },
            {
                "compoundParts": null,
                "condition": {
                    "freeArg": {
                        "type": "STRING",
                        "name": "time"
                    },
                    "operation": "LTE",
                    "fixedArg": {
                        "bean": {
                            "value": {
                                "java.lang.String": "00:10:00",
                                "java.lang.Double": 0
                            }
                        }
                    }
                },
                "negated": false,
                "relation": "AND"
            },
            {
                "compoundParts": null,
                "condition": {
                    "freeArg": {
                        "type": "ANY",
                        "name": "rebootDecoupled"
                    },
                    "operation": "EXISTS",
                    "fixedArg": null
                },
                "negated": true,
                "relation": "AND"
            },
            {
                "compoundParts": null,
                "condition": {
                    "freeArg": {
                        "type": "STRING",
                        "name": "firmware_download_protocol"
                    },
                    "operation": "IS",
                    "fixedArg": {
                        "bean": {
                            "value": {
                                "java.lang.String": "http",
                                "java.lang.Double": 0
                            }
                        }
                    }
                },
                "negated": true,
                "relation": "AND"
            },
            {
                "compoundParts": null,
                "condition": {
                    "freeArg": {
                        "type": "STRING",
                        "name": "ipAddress"
                    },
                    "operation": "IN_LIST",
                    "fixedArg": {
                        "bean": {
                            "value": {
                                "java.lang.String": "_-",
                                "java.lang.Double": 0
                            }
                        }
                    }
                },
                "negated": true,
                "relation": "AND"
            },
            {
                "compoundParts": null,
                "condition": {
                    "freeArg": {
                        "type": "STRING",
                        "name": "env"
                    },
                    "operation": "IS",
                    "fixedArg": {
                        "bean": {
                            "value": {
                                "java.lang.String": "AA",
                                "java.lang.Double": 0
                            }
                        }
                    }
                },
                "negated": true,
                "relation": "AND"
            },
            {
                "compoundParts": null,
                "condition": {
                    "freeArg": {
                        "type": "STRING",
                        "name": "model"
                    },
                    "operation": "IS",
                    "fixedArg": {
                        "bean": {
                            "value": {
                                "java.lang.String": "12",
                                "java.lang.Double": 0
                            }
                        }
                    }
                },
                "negated": true,
                "relation": "OR"
            }
        ],
        "condition": null,
        "negated": false,
        "relation": ""
    },
    "ApplicableAction": {
        "actionType": "BLOCKING_FILTER",
        "type": ".BlockingFilterAction"
    },
    "type": "TIME_FILTER",
    "active": true,
    "applicationtype": "stb"
}`)
)
