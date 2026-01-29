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
* [Project Folder Structure](#project-folder-structure)
    * [Folder Details](#folder-details)
* [Core Idea and Project Overview](#core-idea-and-project-overview)
    * [Key Concepts](#key-concepts)
    * [Typical Use Cases](#typical-use-cases)
* [Community & Contribution](#community--contribution)
* [🧪 Testing](#-testing)
    * [Run All Tests](#run-all-tests)
* [Development Workflow](#development-workflow)
* [📄 License](#-license)
* [🆘 Support](#-support)
* [🔗 Related Projects](#-related-projects)
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

---

## Project Folder Structure

Below is the typical folder structure for `xconfwebconfig`:

```
xconfwebconfig/
├── bin/                # Compiled binaries
├── common/             # Common utilities and constants
├── config/             # Configuration files (e.g., sample_xconfwebconfig.conf)
├── dataapi/            # Data service API handlers and routers
│   ├── dcm/            # Device Configuration Manager API
│   ├── estbfirmware/   # STB firmware API
│   └── featurecontrol/ # Feature Control API
├── db/                 # Database schema and migration scripts (e.g., db_init.cql)
├── http/               # HTTP server, middleware, and utilities
├── protobuf/           # Protobuf definitions and generated code
├── rulesengine/        # Rule engine logic and components
├── security/           # Security-related code (auth, JWT, etc.)
├── shared/             # Shared logic (firmware, rfc, logupload, estbfirmware, etc.)
├── tagging/            # Tagging service and related logic
├── tests/              # Unit and integration tests
├── tracing/            # Distributed tracing utilities
├── util/               # Utility functions and helpers
├── Makefile            # Build automation
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
└── README.md           # Project documentation
```

### Folder Details

- **bin/**: Contains the compiled application binaries after running `make`.
- **common/**: Common utilities, helpers, and constants used across the project.
- **config/**: Configuration files. `sample_xconfwebconfig.conf` is a template for environment-specific settings.
- **dataapi/**: Data service API handlers and routers.
  - **dcm/**: Device Configuration Manager API logic.
  - **estbfirmware/**: STB firmware API logic.
  - **featurecontrol/**: Feature Control API logic.
- **db/**: Cassandra database schema files and migration scripts. `db_init.cql` initializes the required tables and types.
- **http/**: HTTP server setup, middleware, and related utilities.
- **protobuf/**: Protobuf definitions and generated code for gRPC or other protocol-based communication.
- **rulesengine/**: Rule engine logic and components for evaluating device rules.
- **security/**: Security-related code, including authentication, authorization, and JWT handling.
- **shared/**: Shared logic and modules (e.g., firmware, rfc, logupload, estbfirmware) used by multiple parts of the application.
- **tagging/**: Tagging service and related logic for device or configuration tagging.
- **tests/**: Unit and integration tests for the project.
- **tracing/**: Distributed tracing utilities and instrumentation.
- **util/**: Utility functions and helpers not specific to other modules.
- **Makefile**: Defines build, test, and run commands for automation.
- **go.mod, go.sum**: Go module dependency management files.
- **README.md**: This documentation file.

---

## Core Idea and Project Overview

**xconfwebconfig** is a microservice designed to manage and deliver configuration and firmware information to remote devices (primarily set-top boxes, or STBs) in the field, warehouses, and test environments.

### Key Concepts

- **Centralized Configuration Management**:  
xconfwebconfig acts as the authoritative source for device firmware versions, feature flags, and operational settings. It does not push firmware, but tells devices what version to use and where to get it.

- **Rule-Based Decision Engine**:  
The service uses a flexible rule engine to determine which configuration or firmware a device should receive, based on device attributes (MAC, model, environment, etc.).

- **RESTful API**:  
Devices interact with xconfwebconfig via HTTP endpoints, sending identifying information and receiving JSON responses with configuration details.

- **Extensible for Multiple Application Types**:  
While primarily used for STBs, the architecture supports other device types (e.g., xhome, rdkcloud) via the `{applicationType}` parameter.

- **Separation of Concerns**:  
The codebase is organized to separate API handling, business logic, and data persistence, making it maintainable and extensible.

### Typical Use Cases

- **Firmware Management**:  
Devices query which firmware version to use, and xconfwebconfig responds based on current rules and device attributes.

- **Settings Distribution**:  
Devices fetch operational settings (e.g., log upload schedules, telemetry profiles) tailored to their environment and model.

- **Feature Control**:  
Feature flags can be enabled/disabled remotely for specific devices or groups, allowing for controlled rollouts and testing.

---

## Community & Contribution

- **Extensible Design**:  
The modular structure allows contributors to add new endpoints, rule types, or support for additional device types with minimal friction.

- **Open for Collaboration**:  
Contributions are welcome! Please follow the existing code organization and submit pull requests with clear descriptions.

---

## 🧪 Testing

### Run All Tests

```bash
make test
```

## Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:
- Create an issue in the GitHub repository
- Check the documentation (if available)
- Review existing issues and discussions

## 🔗 Related Projects

- [xconfadmin](https://github.com/rdkcentral/xconfadmin) - admin service
- [RDK Central](https://github.com/rdkcentral) - RDK Central organization

---

**Note**: This is a configuration management service for RDK devices. Ensure proper security measures are in place when deploying in production environments.