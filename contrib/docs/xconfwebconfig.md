# XConf WebConfig Documentation

## Overview

The XConf WebConfig API provides high-performance configuration delivery to RDK devices through a RESTful interface. It manages firmware configurations, device settings, telemetry profiles, feature rules, and various configuration management operations.

**Base URL**: `/xconfWebconfig`

### Key Features

- **High-Performance Configuration Delivery**: Optimized request processing with rule evaluation for minimal latency
- **Firmware Management**: Determines appropriate firmware versions and download locations based on device characteristics
- **Feature Control Service (RFC)**: Delivers feature flag configurations with percentage-based rollouts and conditional targeting
- **Device Control Manager (DCM) Integration**: Provides DCM settings including log upload policies and device configurations
- **Telemetry Profile Management**: Delivers telemetry collection profiles and data upload configurations
- **Rule-Based Evaluation Engine**: Real-time evaluation of complex conditional logic for targeted configuration delivery
- **Multi-Protocol Support**: Supports HTTP, HTTPS, and TFTP protocols for firmware distribution
- **Security and Validation**: Comprehensive request validation, authentication, and device verification

---

## API Overview

### Configuration Management
1. [Firmware Configuration](#firmware-configuration)
2. [Feature Control Service](#feature-control-service)
3. [Device Control Manager (DCM)](#device-control-manager-dcm)
4. [Telemetry Management](#telemetry-management)
5. [Firmware Version Check](#firmware-version-check)
6. [System Statistics](#system-statistics)

---

## Architecture Overview

The XConf WebConfig service follows a stateless, high-performance architecture optimized for handling thousands of concurrent device requests. The design emphasizes speed and reliability through efficient rule evaluation engines and database connection pooling. The service implements a request-response pattern with sophisticated context enrichment from external services.

### Prerequisites and Dependencies

**MUST Requirements:**
- XConf WebConfig database backend (supporting device configuration tables and rules)
- Go 1.19+ runtime environment
- Cassandra database cluster for persistent configuration storage
- HTTP server infrastructure with Gorilla Mux routing and middleware support
- Rules engine supporting complex conditional logic evaluation

**SHOULD Requirements:**
- External service integrations (Group Service, Account Service, Device Service) for context enrichment
- Metrics collection infrastructure (Prometheus) for operational observability
- Distributed tracing system (OpenTelemetry) for request flow analysis
- Load balancer infrastructure for high-availability deployments

### Internal Modules

| Module | Description | Purpose |
|--------|-------------|---------|
| **Firmware Handler** | Processes firmware configuration requests | Firmware version and location determination |
| **Feature Control Handler** | Manages feature flag requests | Feature flag delivery with rollouts |
| **DCM Handler** | Delivers device control management settings | Log upload and device-specific configurations |
| **Telemetry Handler** | Provides telemetry profile configurations | Data collection and reporting |
| **Rules Engine** | High-performance rule evaluation | Conditional logic processing |
| **Context Enrichment** | External service integration | Device context and feature tag resolution |

---

## Firmware Configuration

### Retrieve Firmware Configuration

**GET** `http://<host>:<port>/xconf/swu/{applicationType}`

**Headers:**
- Accept = application/json

**Parameters:**
- `{applicationType}`: Device application type (stb, rdkv)
- Query parameters: Device-specific attributes (MAC address, model, environment, version)

**Response:** 200 OK; 404 Not Found; 400 Bad Request; 500 Internal Server Error

**Request Example:**
```
http://localhost:9091/xconf/swu/stb?eStbMac=AA:BB:CC:DD:EE:FF&model=MODEL_X&env=PROD&version=1.0.0
```

**JSON Response:**
```json
{
  "firmwareVersion": "2.1.0-PROD",
  "firmwareFilename": "firmware-2.1.0-PROD.bin",
  "firmwareDownloadURL": "https://cdn.example.com/firmware/firmware-2.1.0-PROD.bin",
  "firmwareLocation": "cdn.example.com",
  "ipv6FirmwareLocation": "2001:db8::1",
  "firmwareDownloadProtocol": "http",
  "rebootImmediately": false,
  "forceHttp": false,
  "upgradeDelay": 0,
  "mandatoryUpdate": false
}
```

---

## Feature Control Service

### Retrieve Feature Control Settings

**GET** `http://<host>:<port>/featureControl/getSettings/{applicationType}`

**Headers:**
- Accept = application/json

**Parameters:**
- `{applicationType}`: Device application type (optional)
- Query parameters: Device context for feature evaluation

**Response:** 200 OK; 404 Not Found; 400 Bad Request; 500 Internal Server Error

**Request Example:**
```
http://localhost:9091/featureControl/getSettings/stb?eStbMac=AA:BB:CC:DD:EE:FF&model=MODEL_X&partnerId=PARTNER_1
```

**JSON Response:**
```json
{
  "featureControl": {
    "features": [
      {
        "name": "ENABLE_NEW_UI",
        "featureInstance": "ENABLE_NEW_UI_INSTANCE",
        "enable": true,
        "configData": {
          "ui_theme": "dark",
          "animation_enabled": true
        },
        "applicationType": "stb"
      },
      {
        "name": "VIDEO_OPTIMIZATION",
        "featureInstance": "VIDEO_OPT_INSTANCE",
        "enable": false,
        "configData": {
          "optimization_level": "standard"
        },
        "applicationType": "stb"
      }
    ]
  },
  "effectiveImmediate": false
}
```

---

## Device Control Manager (DCM)

### Retrieve DCM Settings

**GET** `http://<host>:<port>/loguploader/getSettings/{applicationType}`

**Headers:**
- Accept = application/json

**Parameters:**
- `{applicationType}`: Device application type (optional)
- `settingType`: Specific setting types to retrieve (optional)

**Response:** 200 OK; 404 Not Found; 400 Bad Request; 500 Internal Server Error

**Request Example:**
```
http://localhost:9091/loguploader/getSettings/stb?eStbMac=AA:BB:CC:DD:EE:FF&settingType=LogUpload&settingType=DeviceSettings
```

**JSON Response:**
```json
{
  "urn:settings:LogUpload": {
    "Name": "LogUploadSettings_STB",
    "NumberOfDays": 3,
    "AreSettingsActive": true,
    "LogUploadSettings": {
      "MocaLogPeriod": 1,
      "WifiLogPeriod": 1,
      "UploadRepositoryName": "LOG_UPLOAD_REPO",
      "UploadOnReboot": true,
      "UploadRepositoryURL": "https://logs.example.com/upload",
      "NumberOfDays": 3,
      "UploadRepositoryUploadProtocol": "HTTP",
      "LogsUploadFrequency": 1440
    }
  },
  "urn:settings:DeviceSettings": {
    "Name": "DeviceSettings_STB",
    "CheckOnReboot": true,
    "SettingsAreActive": true,
    "Schedule": {
      "TimeZone": "UTC",
      "DurationMinutes": 30,
      "StartDate": "2025-01-01",
      "EndDate": "2025-12-31"
    },
    "ConfigurationServiceURL": "https://config.example.com/api"
  }
}
```

---

## Telemetry Management

### Retrieve Telemetry Profiles

**GET** `http://<host>:<port>/loguploader/getTelemetryProfiles/{applicationType}`

**Headers:**
- Accept = application/json

**Parameters:**
- `{applicationType}`: Device application type (optional)
- Query parameters: Device context for profile selection

**Response:** 200 OK; 404 Not Found; 400 Bad Request; 500 Internal Server Error

**Request Example:**
```
http://localhost:9091/loguploader/getTelemetryProfiles/stb?eStbMac=AA:BB:CC:DD:EE:FF&model=MODEL_X
```

**JSON Response:**
```json
{
  "profiles": [
    {
      "name": "BASIC_TELEMETRY",
      "description": "Basic telemetry collection profile",
      "schedule": {
        "type": "ActNow",
        "expression": "*/15 * * * *"
      },
      "uploadRepository": {
        "name": "TELEMETRY_REPO",
        "url": "https://telemetry.example.com/upload",
        "protocol": "HTTP"
      },
      "telemetryProfile": [
        {
          "header": "SYSTEM_METRICS",
          "content": "CPU_Usage,Memory_Usage,Disk_Usage",
          "type": "2",
          "pollingFrequency": "300"
        },
        {
          "header": "NETWORK_METRICS",
          "content": "Bandwidth_Usage,Packet_Loss",
          "type": "2",
          "pollingFrequency": "60"
        }
      ]
    }
  ]
}
```

---

## Firmware Version Check

### Check Minimum Firmware Version

**GET** `http://<host>:<port>/estbfirmware/checkMinimumFirmware`

**Headers:**
- Accept = application/json

**Parameters:**
- `eStbMac`: Device MAC address
- `firmwareVersion`: Current firmware version
- `model`: Device model

**Response:** 200 OK; 400 Bad Request; 500 Internal Server Error

**Request Example:**
```
http://localhost:9091/estbfirmware/checkMinimumFirmware?eStbMac=AA:BB:CC:DD:EE:FF&firmwareVersion=1.5.0&model=MODEL_X
```

**JSON Response:**
```json
{
  "requiredVersion": "2.0.0",
  "explanation": "Your firmware version 1.5.0 is below the minimum required version 2.0.0",
  "firmwareVersions": ["2.0.0", "2.1.0", "2.2.0"]
}
```

---

## System Statistics

### Get System Statistics

**GET** `http://<host>:<port>/info/statistics`

**Headers:**
- Accept = application/json

**Response:** 200 OK; 500 Internal Server Error

**Request Example:**
```
http://localhost:9091/info/statistics
```

**JSON Response:**
```json
{
  "requestStats": {
    "totalRequests": 1000000,
    "averageResponseTime": "45ms",
    "errorRate": 0.001
  },
  "databaseStats": {
    "connectionPoolSize": 20,
    "activeConnections": 12,
    "avgQueryTime": "15ms"
  }
}
```

---

## Configuration Rule Evaluation

The WebConfig service implements a sophisticated rules engine that evaluates configuration policies in real-time based on device characteristics and deployment contexts.

### Rule Types and Evaluation Order

1. **Priority-Based Rules**: Higher priority rules override lower priority rules
2. **Conditional Logic**: Complex boolean expressions with AND/OR operators
3. **Percentage Rollouts**: Statistical distribution for gradual deployments
4. **Time-Based Rules**: Temporal activation and expiration conditions
5. **Geographic Rules**: Location-based configuration targeting
6. **Device Cohort Rules**: Model, firmware version, and capability-based targeting

### Rule Evaluation Context

The rules engine maintains rich context information for accurate evaluation:

- Device MAC address and identifier
- Device model and capabilities
- Environment and deployment context
- Firmware version and current state
- Partner and account information
- Geographic location and timezone
- Feature tags and group memberships

### Performance Optimizations

- **Rule Compilation**: Rules are pre-compiled into optimized evaluation trees
- **Context Caching**: Device context is cached to reduce external service calls
- **Parallel Evaluation**: Independent rule sets are evaluated concurrently
- **Short-Circuit Logic**: Evaluation stops early when definitive results are reached

---

## Security Model

### Request Validation
- **Parameter Sanitization**: All input parameters are validated and sanitized
- **MAC Address Validation**: Device MAC addresses are verified for format and authenticity
- **Protocol Validation**: Request protocols and headers are validated against expected patterns
- **Rate Limiting**: Requests are rate-limited per device and IP address to prevent abuse

### Device Authentication
- **Certificate-Based**: X.509 certificate validation for device identity
- **MAC Address Verification**: Cross-reference with authorized device databases
- **Capability Validation**: Device-reported capabilities are validated against known models
- **Anti-Spoofing**: Multiple validation layers prevent device identity spoofing

### Response Security
- **Content Sanitization**: All response content is sanitized to prevent injection attacks
- **HTTPS Enforcement**: Secure transport layer encryption for sensitive configuration data
- **Response Signing**: Critical responses can be cryptographically signed for integrity
- **Audit Logging**: Complete audit trail for all configuration deliveries

---

## Performance Characteristics

### Scalability Metrics
- **Request Throughput**: Designed to handle 10,000+ requests per second per instance
- **Response Latency**: Sub-100ms response times for configurations
- **Concurrent Connections**: Support for 1,000+ concurrent device connections
- **Memory Efficiency**: Optimized memory usage with configurable limits

### Database Optimization
- **Connection Pooling**: Efficient database connection management
- **Query Optimization**: Optimized queries with proper indexing strategies
- **Read Replicas**: Read operations distributed across database replicas
- **Batch Operations**: Bulk operations for improved throughput

---

## Monitoring and Observability

### Metrics Collection
The WebConfig service exposes comprehensive metrics for operational monitoring:

- **Request Metrics**: Request count, response times, error rates per endpoint
- **Database Metrics**: Query times, connection pool usage, error rates
- **External Service Metrics**: Response times and error rates for external dependencies
- **Business Metrics**: Configuration delivery success rates, feature activation rates

### Health Checks
- **Liveness Probe**: Confirms service is running and accepting requests
- **Readiness Probe**: Validates all dependencies are available and healthy
- **Deep Health Check**: Comprehensive validation of all system components
- **Dependency Health**: Monitors health of external service dependencies

### Distributed Tracing
- **Request Tracing**: Complete request flow tracing across all components
- **Context Propagation**: Trace context propagated to external service calls
- **Performance Analysis**: Detailed timing analysis for performance optimization
- **Error Correlation**: Error tracking and correlation across service boundaries

---

## Deployment Considerations

### High Availability Setup
- **Load Balancer Configuration**: Multiple service instances behind load balancers
- **Health Check Integration**: Load balancer health check configuration
- **Graceful Shutdown**: Proper connection draining during deployments
- **Circuit Breaker Pattern**: Fail-fast behavior when dependencies are unavailable

### Scaling Strategies
- **Horizontal Scaling**: Add more service instances to handle increased load
- **Database Scaling**: Scale database reads through replica distribution
- **Geographic Distribution**: Deploy instances closer to device populations
- **Request Optimization**: Optimize database queries and request patterns

### Configuration Management
- **Environment Separation**: Separate configurations for dev, staging, and production
- **Feature Flags**: Runtime feature toggling for safe deployments
- **Configuration Validation**: Automated validation of configuration changes
- **Rollback Procedures**: Quick rollback capabilities for problematic deployments

---

This comprehensive documentation provides complete coverage of the XConf WebConfig API and operational characteristics. The service serves as the high-performance component for delivering configuration data to RDK devices at scale, with robust security, monitoring, and deployment capabilities.
