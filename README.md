# XConf WebConfig – Comprehensive Documentation

## Folder Structure Diagram

```
xconfwebconfig/
├── main.go
├── config/
│   └── sample_xconfwebconfig.conf
├── db/
│   └── db_init.cql
├── shared/
│   └── ...
├── dataapi/
│   └── ...
├── rulesengine/
│   └── ...
├── http/
│   └── ...
├── protobuf/
│   └── ...
├── security/
│   └── ...
├── Makefile
└── ...
```

## Technology Stack

- **Language:** Go (Golang)
- **Web Framework:** [gorilla/mux](https://github.com/gorilla/mux) (HTTP routing)
- **Database:** Cassandra (NoSQL persistent storage)
- **Configuration:** Environment variables & `.conf` files
- **Build Tool:** Makefile
- **Tracing:** OpenTelemetry
- **Security:** Token-based authentication, SSL/TLS support
- **Logging:** Configurable log levels and formats

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

## Overview

XConf WebConfig is a backend service for remote device management, primarily targeting set-top boxes (STBs) and similar devices. It provides APIs for firmware versioning, device configuration, telemetry, and feature toggling. Devices query XConf for instructions; XConf responds based on rules and device attributes.

---

## Purpose & Core Functionalities

- **Firmware Management:** Centralizes firmware versioning for STBs in field, warehouse, and test environments. Determines which firmware a device should use, and provides download details (protocol, location, filename, reboot instructions).
- **Device Configuration:** Supplies device-specific settings such as log upload schedules, telemetry profiles, VOD parameters, and more. All settings are dynamically selected based on device context and rules.
- **Feature Control:** Enables or disables features for devices using rule-based logic. Returns feature toggles and configuration data.
- **Telemetry Profiles:** Provides telemetry configuration and profiles, supporting both legacy and new formats.
- **Security & Auditing:** Validates tokens and headers, logs all requests and rule evaluations for audit and troubleshooting.

---

## Architecture & Codebase Structure

- **Language:** Go (Golang)
- **Web Framework:** gorilla/mux (HTTP routing)
- **Database:** Cassandra (persistent storage)
- **Configuration:** Environment variables and config files
- **Build Tool:** Makefile

### Directory Structure & Key Files

- `main.go` – Service entry point; initializes configuration, database, HTTP server, and routes.
- `config/` – Configuration files and templates (e.g., `sample_xconfwebconfig.conf`).
- `db/` – Cassandra schema (`db_init.cql`) and initialization scripts.
- `shared/` – Common Go packages (e.g., logupload types, utilities).
- `dataapi/` – API logic for firmware, features, telemetry, etc. Contains handlers for each endpoint.
- `rulesengine/` – Rule parsing, evaluation, and matching logic. Implements the core decision engine.
- `http/` – HTTP server, routing, connectors, security, middleware.
- `protobuf/` – Protocol buffer definitions for structured data exchange.
- `security/` – Encryption and security helpers.

---

## Data Flow & Operation

1. **Device Request:** Devices send HTTP requests with identifying parameters (MAC, model, env, etc.) to XConf endpoints.
2. **Routing:** gorilla/mux routes requests to the appropriate handler in `dataapi/`.
3. **Rule Evaluation:** The handler invokes the rule engine (`rulesengine/`) to match device attributes against rules stored in Cassandra.
4. **Response Construction:** Based on rule matches, the handler assembles a JSON response with firmware, settings, or feature data.
5. **Logging & Security:** All requests and responses are logged; security tokens and headers are validated.

---

## Configuration

- **Config File:** `config/sample_xconfwebconfig.conf` defines environment-specific properties (DB connection, logging, etc.).
- **Environment Variables:** Used for secrets and runtime overrides (e.g., `SAT_CLIENT_ID`, `SAT_CLIENT_SECRET`, `SECURITY_TOKEN_KEY`).
- **Schema:** Cassandra tables are defined in `db/db_init.cql` and must be initialized before running the service.

---

## Endpoints

### XConf Primary API

| PATH | METHOD | MAIN QUERY PARAMETERS | DESCRIPTION |
|------|--------|-----------------------|-------------|
| `/xconf/swu/{applicationType}` | `GET` |`eStbMac`,<br>`ipAddress`,<br>`env`,<br>`model`,<br>`firmwareVersion`,<br>`partnerId`,<br>`accountId`,<br>`{tag name}` - any tag from Tagging Service,<br>`controllerId`,<br>`channelMapId`,<br>`vodId`| Returns firmwareVersion to STB box <br> `{applicationType}` - for now supported `stb`, `xhome`, `rdkcloud` |
| `/xconf/swu/bse` | `GET` | `ipAddress` - required | Returns BSE configuration |
| `/xconf/{applicationType}/runningFirmwareVersion/info` | `GET` | `mac` - required | Return if device has Activation Minimum Firmware and Minimum Firmware version |
| `/estbfirmware/checkMinimumFirmware` | `GET` | `mac` - required | Return if device has Minimum Firmware version |

#### Sample Requests & Responses

**GET `/xconf/swu/{applicationType}`**
Request:
```shell
curl 'https://<host>/xconf/swu/stb?eStbMac=AA:AA:AA:AA:AA:AA&env=DEV&model=TEST_MODEL&ipAddress=10.10.10.10'
```
Response:
```json
{
    "firmwareDownloadProtocol": "http",
    "firmwareFilename": "FIRMWARE-NAME.bin",
    "firmwareLocation": "http://ssr-url.com/cgi-bin/x1-sign-redirect.pl?K=10&F=stb_cdl",
    "firmwareVersion": "FIRMWARE_VERSION",
    "rebootImmediately": true
}
```

**GET `/xconf/swu/bse`**
Request:
```shell
curl 'https://<host>/xconf/swu/bse?ipAddress=10.10.10.10'
```
Response:
```json
{
  "bseConfig": {
    "param1": "value1",
    "param2": "value2"
  }
}
```

**GET `/xconf/{applicationType}/runningFirmwareVersion/info`**
Request:
```shell
curl 'https://<host>/xconf/stb/runningFirmwareVersion/info?mac=AA:AA:AA:AA:AA:AA'
```
Response:
```json
{
  "minimumFirmwareVersion": "V1.2.3",
  "activationStatus": true
}
```

**GET `/estbfirmware/checkMinimumFirmware`**
Request:
```shell
curl 'https://<host>/estbfirmware/checkMinimumFirmware?mac=AA:AA:AA:AA:AA:AA'
```
Response:
```json
{
  "compliance": true
}
```

---

### Device Configuration Manager (DCM)

| PATH | METHOD | MAIN QUERY PARAMETERS | DESCRIPTION |
|------|--------|-----------------------|-------------|
| `/loguploader/getSettings/{applicationType}` | `GET` | `estbMacAddress`,<br>`ipAddress`,<br>`env`,<br>`model`,<br>`firmwareVersion`,<br>`partnerId`,<br>`accountId`,<br>`{tag name}` - any tag from Tagging Service,<br>`checkNow` - boolean,<br>`version`,<br>`settingsType`,<br>`ecmMacAddress`,<br>`controllerId`,<br>`channelMapId`,<br>`vodId` | Returns settings to STB box <br> `{applicationType}` - for now supported `stb`, `xhome`, `rdkcloud`, <br>field is optional and `stb` application is used by default |
| `/loguploader/getT2Settings/{applicationType}` | `GET` | The same as a previous | Returns telemetry configuration in the new format. If the component name has been defined for an entry, <br>the response will be in the new format. The second and third columns for that entry will not be used in the response. <br>The content field comes from the fifth column (component name). The type field will be a constant string `<event>` |
| `/loguploader/getTelemetryProfiles/{applicationType}` | `GET` | The same as a previous | Returns Telemetry 2.0 profiles based on Telemetry 2.0 rules |

#### Sample Requests & Responses

**GET `/loguploader/getSettings/{applicationType}`**
Request:
```shell
curl 'https://<host>/loguploader/getSettings/stb?estbMacAddress=AA:AA:AA:AA:AA:AA&env=DEV&model=TEST_MODEL&ipAddress=10.10.10.10'
```
Response:
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

**GET `/loguploader/getT2Settings/{applicationType}`**
Request:
```shell
curl 'https://<host>/loguploader/getT2Settings/stb?estbMacAddress=AA:AA:AA:AA:AA:AA&env=DEV&model=TEST_MODEL&ipAddress=10.10.10.10'
```
Response:
```json
{
  "telemetryT2Settings": {
    "component": "COMPONENT_NAME",
    "type": "event",
    "content": "..."
  }
}
```

**GET `/loguploader/getTelemetryProfiles/{applicationType}`**
Request:
```shell
curl 'https://<host>/loguploader/getTelemetryProfiles/stb?estbMacAddress=AA:AA:AA:AA:AA:AA&env=DEV&model=TEST_MODEL&ipAddress=10.10.10.10'
```
Response:
```json
{
  "telemetryProfiles": [
    {
      "id": "profile-id",
      "name": "ProfileName",
      "settings": { }
    }
  ]
}
```

---

### RDK Feature Control (RFC)

| PATH | METHOD | MAIN QUERY PARAMETERS | DESCRIPTION |
|------|--------|-----------------------|-------------|
| `/featureControl/getSettings/{applicationType}` | `GET` | `estbMacAddress`,<br>`ipAddress`,<br>`env`,<br>`model`,<br>`firmwareVersion`,<br>`partnerId`,<br>`accountId`,<br>`{tag name}` - any tag from Tagging Service,<br>`ecmMacAddress`,<br>`controllerId`,<br>`channelMapId`,<br>`vodId`| Returns enabled/disable features <br> `{applicationType}` - for now supported `stb`, `xhome`, `rdkcloud`, field is optional and `stb` application is used by default |

#### Sample Requests & Responses

**GET `/featureControl/getSettings/{applicationType}`**
Request:
```shell
curl 'http://<host>/featureControl/getSettings?estbMacAddress=AA:AA:AA:AA:AA:AA'
```
Response:
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

---

## Authentication & Security

- All API endpoints require secure tokens and validated headers.
- Tokens are checked using environment variables (`SAT_CLIENT_ID`, `SAT_CLIENT_SECRET`, `SECURITY_TOKEN_KEY`).
- Requests without valid tokens or headers are rejected with `401 Unauthorized` or `403 Forbidden`.
- All requests and rule evaluations are logged for audit and troubleshooting.

---

## Data Models & Storage

- **Firmware Rules, Device Settings, Feature Toggles, and Telemetry Profiles** are stored in Cassandra tables defined in `db/db_init.cql`.
- Each rule or configuration object includes metadata (ID, name, applicationType, etc.) and a rule structure for matching device attributes.
- Data is retrieved and updated via Go data access layers in the `db/` and `shared/` directories.

---

## Rule Engine – Internals & Usage

Rules are the backbone of XConf's decision logic. They are stored in Cassandra and evaluated at runtime.

### Rule Types
- FirmwareRule: Governs firmware assignment logic
- DCM Rule (Formula): Governs device configuration logic
- TelemetryRule: Governs telemetry profile assignment
- TelemetryTwoRule: Advanced telemetry logic
- SettingRule: Governs device settings
- FeatureRule (RFC Rule): Governs feature toggling

### Rule Structure
- `negated`: Boolean to invert the condition
- `condition`: Key-value pair specifying the match criteria
- `compoundParts`: Array of additional conditions
- `relation`: Logical operator (`AND`, `OR`) between conditions

#### Example Rule
```json
{
  "negated": false,
  "compoundParts": [
    {
      "negated": false,
      "condition": {
        "freeArg": { "type": "STRING", "name": "model" },
        "operation": "IS",
        "fixedArg": { "bean": { "value": { "java.lang.String": "TEST_MODEL1" } } }
      },
      "compoundParts": []
    },
    {
      "negated": false,
      "relation": "OR",
      "condition": {
        "freeArg": { "type": "STRING", "name": "model" },
        "operation": "IS",
        "fixedArg": { "bean": { "value": { "java.lang.String": "TEST_MODEL2" } } }
      },
      "compoundParts": []
    }
  ],
  "boundTelemetryId": "ad10dd05-d2ff-4d00-8f52-b0ca6956cde6",
  "id": "fb4210ac-8187-4cba-9301-eb8f27fcdaa8",
  "name": "Arris_SVG",
  "applicationType": "stb"
}
```

#### Rule Evaluation
- Each rule is evaluated against incoming device parameters.
- Compound rules allow complex logic (AND/OR, negation).
- Matching rules determine which firmware, settings, or features are returned.
- Rules are extensible: new types or conditions can be added by extending the `rulesengine/` module.

---

## Application Logic Flow

1. Device sends HTTP request to XConf endpoint with identifying parameters.
2. gorilla/mux routes the request to the correct handler in `dataapi/`.
3. Handler validates authentication tokens and headers.
4. Handler invokes the rule engine to evaluate device attributes against rules in Cassandra.
5. Matching rule is selected; response is constructed (firmware, settings, features, etc.).
6. Response is returned as JSON; request and response are logged for auditing.
7. If no rule matches, appropriate error response is sent.

---

## Setup & Running

1. **Prepare Cassandra:**
   - Initialize the database using `db/db_init.cql`.
2. **Configure:**
   - Edit `config/sample_xconfwebconfig.conf` for your environment.
3. **Build:**
   ```shell
   make
   ```
4. **Run:**
   ```shell
   export SAT_CLIENT_ID='xxxxxx'
   export SAT_CLIENT_SECRET='yyyyyy'
   export SECURITY_TOKEN_KEY='zzzzzz'
   mkdir -p /app/logs/xconfwebconfig
   bin/xconfwebconfig-linux-amd64 -f config/sample_xconfwebconfig.conf
   ```

---

## Contribution & Support

- See `CONTRIBUTING.md` for guidelines.
- For schema changes, update `db/db_init.cql`.
- For new APIs, add handlers in `dataapi/` and update routing in `http/router.go`.
- For extending rule logic, update `rulesengine/` and related Cassandra schema.

---

## License

See `LICENSE` and `COPYING` for details.

---

## Operational Notes & Best Practices

- **Logging:** All requests and rule evaluations are logged for audit and troubleshooting.
- **Security:** Use secure tokens and validate headers for all API requests.
- **Extensibility:** Add new rule types or API endpoints by extending `rulesengine/` and `dataapi/` modules.
- **Testing:** Use sample requests and configuration files to validate new rules and API changes.
- **Monitoring:** Monitor logs and Cassandra health for production deployments.
- **Scalability:** The system is designed to handle large device populations and high request rates. Cassandra and Go provide horizontal scalability.
- **Error Handling:** All API endpoints return clear error messages and status codes for troubleshooting.
- **Developer Extension Points:** To add new features, create new handlers in `dataapi/`, update routing in `http/`, and extend rule logic in `rulesengine/`.

---
