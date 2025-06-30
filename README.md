# Table of Contents
<!--ts-->
* [XConf DataService Overview](#xconf-dataservice-overview)
* [Architecture](#architecture)
* [Run the application](#run-the-application)
* [Endpoints](#endpoints)
    * [XConf Primary API](#xconf-primary-api)
    * [Device Configuration Manager](#device-configuration-manager-dcm)
    * [RDK Feature Control](#rdk-feature-control-rfc)
* [Examples](#examples)
    * [Get STB firmware version](#get-stb-firmware-version)
    * [Get STB settings](#get-stb-settings)
    * [Get Feature Settings](#get-feature-settings)
    * [Rule Structure](#rule-structure)
<!--te-->

## XConf DataService Overview

Xconf is slated to be the single entity for managing firmware on set-top boxes both in the field
and in various warehouses and test environments.

Xconf's primary purpose is to tell set-top boxes (STBs) what version of firmware they should be running. 
Xconf does not push firmware to the STB, nor is not involved in any way in the actual download / upgrade process. 
It simply tells the STB which version to use. Xconf also tells STBs when, where (host), and how (protocol) to get the firmware.

Xconf DataService is the app that the STBs talk to. Xconf Admin allows humans to enter all the information necessary for Xconf to provide the correct information to STBs.

The interface between STBs and Xconf is simple. STBs make HTTP requests to Xconf sending information like MAC address, environment, and model. Xconf then applies various rules to determine which firmware information to return. The information is returned in JSON format.

## Architecture

* Golang
* gorilla/mux
* Cassandra DB

## Run the application
1. Run cassandra DB and create a corresponding schema using `db_init.cql` file in `db` folder.
2. Use config/sample_xconfwebconfig.conf to add/override specific environments properties.
4. Build the binary:
```shell
cd .../xconfwebconfig
make
```
5. Run the application:
```shell
export SAT_CLIENT_ID='xxxxxx'
export SAT_CLIENT_SECRET='yyyyyy'
export SECURITY_TOKEN_KEY='zzzzzz'
mkdir -p /app/logs/xconfwebconfig
cd .../xconfwebconfig
bin/xconfwebconfig-linux-amd64 -f config/sample_xconfwebconfig.conf
```

## Endpoints

### XConf Primary API

| PATH | METHOD | MAIN QUERY PARAMETERS | DESCRIPTION |
|------|--------|-----------------------|-------------|
| `/xconf/swu/{applicationType}` | `GET` |`eStbMac`,<br>`ipAddress`,<br>`env`,<br>`model`,<br>`firmwareVersion`,<br>`partnerId`,<br>`accountId`,<br>`{tag name}` - any tag from Tagging Service,<br>`controllerId`,<br>`channelMapId`,<br>`vodId`| Returns firmwareVersion to STB box <br> `{applicationType}` - for now supported `stb`, `xhome`, `rdkcloud` |
| `/xconf/swu/bse` | `GET` | `ipAddress` - required | Returns BSE configuration |
| `/xconf/{applicationType}/runningFirmwareVersion/info` | `GET` | `mac` - required | Return if device has Activation Minimum Firmware and Minimum Firmware version |
| `/estbfirmware/checkMinimumFirmware` | `GET` | `mac` - required | Return if device has Minimum Firmware version |

#### Headers 
For `/xconf/swu/{applictionType}` API: <br>
`HA-Haproxy-xconf-http` to indicate if connection is secure

### Device Configuration Manager (DCM)

Remote devices like set top boxes and DVRs have settings to control certain activities. For instance, STBs need to know when to upload log files, or when to check for a new firmware update. In order to remotely manage a large population of devices, we need a solution that lets support staff define instructions and get the instructions to the devices.

| PATH | METHOD | MAIN QUERY PARAMETERS | DESCRIPTION |
|------|--------|-----------------------|-------------|
| `/loguploader/getSettings/{applicationType}` | `GET` | `estbMacAddress`,<br>`ipAddress`,<br>`env`,<br>`model`,<br>`firmwareVersion`,<br>`partnerId`,<br>`accountId`,<br>`{tag name}` - any tag from Tagging Service,<br>`checkNow` - boolean,<br>`version`,<br>`settingsType`,<br>`ecmMacAddress`,<br>`controllerId`,<br>`channelMapId`,<br>`vodId` | Returns settings to STB box <br> `{applicationType}` - for now supported `stb`, `xhome`, `rdkcloud`, <br>field is optional and `stb` application is used by default |
| `/loguploader/getT2Settings/{applicationType}` | `GET` | The same as a previous | Returns telemetry configuration in the new format. If the component name has been defined for an entry, <br>the response will be in the new format. The second and third columns for that entry will not be used in the response. <br>The content field comes from the fifth column (component name). The type field will be a constant string `<event>` |
| `/loguploader/getTelemetryProfiles/{applicationType}` | `GET` | The same as a previous | Returns Telemetry 2.0 profiles based on Telemetry 2.0 rules |

### RDK Feature Control (RFC)

| PATH | METHOD | MAIN QUERY PARAMETERS | DESCRIPTION |
|------|--------|-----------------------|-------------|
| `/featureControl/getSettings/{applicationType}` | `GET` | `estbMacAddress`,<br>`ipAddress`,<br>`env`,<br>`model`,<br>`firmwareVersion`,<br>`partnerId`,<br>`accountId`,<br>`{tag name}` - any tag from Tagging Service,<br>`ecmMacAddress`,<br>`controllerId`,<br>`channelMapId`,<br>`vodId`| Returns enabled/disable features <br> `{applicationType}` - for now supported `stb`, `xhome`, `rdkcloud`, field is optional and `stb` application is used by default |

#### Headers
`HA-Haproxy-xconf-http` - indicate if connection is secure <br>
`configsethash` - hash of previous response to return `304 Not Modified` http status

## Examples

### Get STB firmware version
#### Request
```shell script
curl --location --request GET 'https://${xconf-path}/xconf/swu/stb?eStbMac=AA:AA:AA:AA:AA:AA&env=DEV&model=TEST_MODEL&ipAddress=10.10.10.10'
```
#### Positive response
```json
{
    "firmwareDownloadProtocol": "http",
    "firmwareFilename": "FIRMWARE-NAME.bin",
    "firmwareLocation": "http://ssr-url.com/cgi-bin/x1-sign-redirect.pl?K=10&F=stb_cdl",
    "firmwareVersion": "FIRMWARE_VERSION",
    "rebootImmediately": true
}
```

### Get STB settings
#### Request
```shell script
curl --location --request GET 'https://${xconf-path}/loguploader/getSettings/stb?estbMacAddress=AA:AA:AA:AA:AA:AA&env=DEV&model=TEST_MODEL&ipAddress=10.10.10.10'
```
#### Positive response
```json
{
  "urn:settings:GroupName": "TEST_GROUP_NAME_1",
  "urn:settings:CheckOnReboot": false,
  "urn:settings:CheckSchedule:cron": "19 7 * * *",
  "urn:settings:CheckSchedule:DurationMinutes": 180,
  "urn:settings:LogUploadSettings:Message": null,
  "urn:settings:LogUploadSettings:Name": "LUS-NAME",
  "urn:settings:LogUploadSettings:NumberOfDays": 0,
  "urn:settings:LogUploadSettings:UploadRepositoryName": "TEST-NAME",
  "urn:settings:LogUploadSettings:UploadRepository:URL": "https://upload-repository-url.com",
  "urn:settings:LogUploadSettings:UploadRepository:uploadProtocol": "HTTP",
  "urn:settings:LogUploadSettings:UploadOnReboot": false,
  "urn:settings:LogUploadSettings:UploadImmediately": false,
  "urn:settings:LogUploadSettings:upload": true,
  "urn:settings:LogUploadSettings:UploadSchedule:cron": "8 20 * * *",
  "urn:settings:LogUploadSettings:UploadSchedule:levelone:cron": null,
  "urn:settings:LogUploadSettings:UploadSchedule:leveltwo:cron": null,
  "urn:settings:LogUploadSettings:UploadSchedule:levelthree:cron": null,
  "urn:settings:LogUploadSettings:UploadSchedule:DurationMinutes": 420,
  "urn:settings:VODSettings:Name": null,
  "urn:settings:VODSettings:LocationsURL": null,
  "urn:settings:VODSettings:SRMIPList": null,
  "urn:settings:TelemetryProfile": {
    "id": "c34518e8-0af5-4524-b96d-c2efb1904458",
    "telemetryProfile": [
      {
        "header": "MEDIA_ERROR_NETWORK_ERROR",
        "content": "NETWORK ERROR(10)",
        "type": "receiver.log",
        "pollingFrequency": "0"
      }
    ],
    "schedule": "*/15 * * * *",
    "expires": 0,
    "telemetryProfile:name": "RDKV_DEVprofile",
    "uploadRepository:URL": "https://upload-repository-host.tv",
    "uploadRepository:uploadProtocol": "HTTP"
  }
}
```

### Get Feature Settings
#### Request
```shell
curl --location --request GET 'http://${xconf-path}/featureControl/getSettings?estbMacAddress=AA:AA:AA:AA:AA:AA' \
--header 'Accept: application/json'
```

#### Positive Response
```json
{
    "featureControl": {
        "features": [
            {
                "name": "TEST_INSPECTOR",
                "effectiveImmediate": false,
                "enable": true,
                "configData": {},
                "featureInstance": "TEST_INSPECTOR"
            }
        ]
    }
}
```

### Rule Structure
There are 6 different rule types: FirmwareRule, DCM Rule (Formula), TelemetryRule, TelemetryTwoRule, SettingRule and FeatureRule (RFC Rule).

Extended from `Rule` object: `DCM Rule`, `TelemetryRule`, `TelemetryTwoRule`. It means that rule object itself has rule structure, corresponding rule fields like `negated`, `condition`, `compoundParts` and `relation` are located in root json object itself.

Otherwise there is rule field.

`TelemetryRule` json extended by `Rule` object:

```json
{
    "negated": false,
    "compoundParts": [
    {
        "negated": false,
        "condition":
        {
            "freeArg":
            {
                "type": "STRING",
                "name": "model"
            },
            "operation": "IS",
            "fixedArg":
            {
                "bean":
                {
                    "value":
                    {
                        "java.lang.String": "TEST_MODEL1"
                    }
                }
            }
        },
        "compoundParts": []
    },
    {
        "negated": false,
        "relation": "OR",
        "condition":
        {
            "freeArg":
            {
                "type": "STRING",
                "name": "model"
            },
            "operation": "IS",
            "fixedArg":
            {
                "bean":
                {
                    "value":
                    {
                        "java.lang.String": "TEST_MODEL2"
                    }
                }
            }
        },
        "compoundParts": []
    }],
    "boundTelemetryId": "ad10dd05-d2ff-4d00-8f52-b0ca6956cde6",
    "id": "fb4210ac-8187-4cba-9301-eb8f27fcdaa8",
    "name": "Arris_SVG",
    "applicationType": "stb"
}
```


`FeatureRule` json, which contains `Rule` object
```json
{
    "id": "018bfb79-4aaf-426e-9e45-17e26d52ad49",
    "name": "Test",
    "rule":
    {
        "negated": false,
        "condition":
        {
            "freeArg":
            {
                "type": "STRING",
                "name": "estbMacAddress"
            },
            "operation": "IS",
            "fixedArg":
            {
                "bean":
                {
                    "value":
                    {
                        "java.lang.String": "AA:AA:AA:AA:AA:AA"
                    }
                }
            }
        },
        "compoundParts": []
    },
    "priority": 13,
    "featureIds": ["68add112-cdc9-47be-ae3b-86e753c8d23e"],
    "applicationType": "stb"
}
```

#### Rule Object
There are following fields there:<br>
`negated` - means condition is with `not` operand.<br>
`condition` - key and value statement.<br>
`compoundParts` - list with multiple conditions.<br>
`relation` - operation between multiple conditions, possible values `OR`, `AND`.

#### Condition structure
Each condition has `freeArg` and `fixedArg` field.
freeArg typed key.
fixedArg value meaning.

If rule has only one condition there are no `compoundParts`, `relation` field is empty.
If there are more than one condition - they are located in `compoundParts` object. First condition does not have any relation, next one has a relation.