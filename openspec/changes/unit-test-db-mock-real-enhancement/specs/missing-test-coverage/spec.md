## ADDED Requirements

### Requirement: Test Coverage for shared/estbfirmware Package
The system SHALL provide comprehensive tests for shared/estbfirmware package which currently has ZERO test coverage despite handling critical firmware configuration logging.

#### Scenario: config_change_logs.go GetLastConfigLog tested
- **WHEN** test calls `GetLastConfigLog(mac, env)`
- **THEN** test SHALL verify correct retrieval from LOGS table via ListingDAO

#### Scenario: config_change_logs.go SetConfigChangeLog tested
- **WHEN** test calls `SetConfigChangeLog(changeLog)`
- **THEN** test SHALL verify log entry persisted with correct rowKey and timestamp

#### Scenario: Both mock and real modes
- **WHEN** tests run in mock mode
- **THEN** tests SHALL use MockListingDAO
- **WHEN** tests run in real mode
- **THEN** tests SHALL use real database with CleanupTracker

### Requirement: Test Coverage for shared/logupload Package
The system SHALL provide comprehensive tests for shared/logupload package which handles DCM log upload settings without any current tests.

#### Scenario: logupload.go DAO operations tested
- **WHEN** test exercises log upload setting persistence
- **THEN** test SHALL verify operations on LOG_UPLOAD_SETTINGS table

#### Scenario: DCM rule retrieval tested
- **WHEN** test retrieves DCM rules
- **THEN** test SHALL verify correct data from DCM_RULE table

#### Scenario: Upload repository operations tested
- **WHEN** test manages upload repositories
- **THEN** test SHALL verify UPLOAD_REPOSITORY table operations

### Requirement: Test Coverage for shared/firmware Package
The system SHALL provide comprehensive tests for shared/firmware package DAO operations.

#### Scenario: Firmware config CRUD tested
- **WHEN** test exercises firmware config lifecycle
- **THEN** test SHALL verify FIRMWARE_CONFIG table operations

#### Scenario: Firmware rule operations tested
- **WHEN** test manages firmware rules
- **THEN** test SHALL verify FIRMWARE_RULE and FIRMWARE_RULE_TEMPLATE table operations

### Requirement: Test Coverage for shared/dcm Package
The system SHALL provide comprehensive tests for shared/dcm package database operations.

#### Scenario: Device settings operations tested
- **WHEN** test manages device settings
- **THEN** test SHALL verify DEVICE_SETTINGS table operations

#### Scenario: Log upload settings tested
- **WHEN** test exercises log upload configuration
- **THEN** test SHALL verify LOG_UPLOAD_SETTINGS table operations

#### Scenario: VOD settings operations tested
- **WHEN** test manages VOD settings
- **THEN** test SHALL verify VOD_SETTINGS table operations

### Requirement: Test Coverage for shared/rfc Package
The system SHALL provide comprehensive tests for shared/rfc package feature control operations.

#### Scenario: Feature control retrieval tested
- **WHEN** test retrieves feature controls
- **THEN** test SHALL verify XCONF_FEATURE table operations

#### Scenario: Feature rule operations tested
- **WHEN** test manages feature rules
- **THEN** test SHALL verify FEATURE_CONTROL_RULE table operations

### Requirement: Test Coverage for db/cache_dao.go
The system SHALL provide tests for db/cache_dao.go which implements CachedSimpleDAO without current test coverage.

#### Scenario: Cache hit path tested
- **WHEN** data exists in cache
- **THEN** test SHALL verify GetOne returns cached data without DB query

#### Scenario: Cache miss path tested
- **WHEN** data not in cache
- **THEN** test SHALL verify GetOne queries DB and populates cache

#### Scenario: Cache invalidation tested
- **WHEN** test calls InvalidateCache
- **THEN** subsequent GetOne SHALL query DB again

### Requirement: Test Coverage for db/listing_dao.go
The system SHALL provide tests for db/listing_dao.go composite key operations.

#### Scenario: Composite key GetOne tested
- **WHEN** test calls GetOne with rowKey and columnName
- **THEN** test SHALL verify correct data retrieval

#### Scenario: Composite key SetOne tested
- **WHEN** test calls SetOne with rowKey, columnName, and data
- **THEN** test SHALL verify data persisted correctly

#### Scenario: GetAll for rowKey tested
- **WHEN** test calls GetAll with rowKey
- **THEN** test SHALL verify all columns for that row returned

#### Scenario: GetRange tested
- **WHEN** test calls GetRange with time bounds
- **THEN** test SHALL verify time-ordered entries returned

### Requirement: Test Coverage for db/compressing_data_dao.go
The system SHALL provide tests for db/compressing_data_dao.go compression and splitting logic.

#### Scenario: Small data without compression tested
- **WHEN** data below compression threshold
- **THEN** test SHALL verify data stored uncompressed

#### Scenario: Large data with compression tested
- **WHEN** data above compression threshold
- **THEN** test SHALL verify data compressed before storage

#### Scenario: Split data operations tested
- **WHEN** data exceeds maximum size
- **THEN** test SHALL verify data split across multiple chunks

### Requirement: Test Coverage for db/group_service_dao.go
The system SHALL provide tests for db/group_service_dao.go group service cache operations.

#### Scenario: GetGroupServiceFeatureTags tested
- **WHEN** test retrieves feature tags
- **THEN** test SHALL verify correct map[string]string returned

#### Scenario: SetGroupServiceFeatureTags tested
- **WHEN** test persists feature tags
- **THEN** test SHALL verify tags stored correctly

### Requirement: Test Coverage for util/common.go
The system SHALL provide tests for util/common.go utility functions that interact with database.

#### Scenario: Database utility functions tested
- **WHEN** util functions query or modify DB
- **THEN** tests SHALL verify correct behavior with both mock and real DAOs

### Requirement: Test Coverage for util/firmware_util.go
The system SHALL provide tests for util/firmware_util.go firmware-related utilities.

#### Scenario: Firmware utility DAO operations tested
- **WHEN** utility functions access firmware data
- **THEN** tests SHALL verify correct table operations

### Requirement: Test Coverage for util/upload_util.go
The system SHALL provide tests for util/upload_util.go upload-related utilities.

#### Scenario: Upload utility DAO operations tested
- **WHEN** utility functions manage upload settings
- **THEN** tests SHALL verify correct persistence

### Requirement: Test Coverage for rulesengine/legacy_converter.go
The system SHALL provide tests for rulesengine/legacy_converter.go database operations.

#### Scenario: Legacy conversion DAO operations tested
- **WHEN** converter migrates legacy data
- **THEN** tests SHALL verify correct table operations

### Requirement: Test Coverage for rulesengine/rule_processor.go
The system SHALL provide tests for rulesengine/rule_processor.go rule processing with database.

#### Scenario: Rule processor DAO operations tested
- **WHEN** processor retrieves or stores rules
- **THEN** tests SHALL verify correct database interactions

### Requirement: Test Coverage for dataapi/dcm Package
The system SHALL provide tests for dataapi/dcm package handlers that currently lack DB tests.

#### Scenario: DCM handlers DAO operations tested
- **WHEN** handlers process DCM requests
- **THEN** tests SHALL verify correct database persistence

### Requirement: Untested Table Coverage - Authentication Tables
The system SHALL provide tests exercising XCONF_WHITELIST_UPDATES and MAC_LIST tables.

#### Scenario: XCONF_WHITELIST_UPDATES operations tested
- **WHEN** whitelist updates occur
- **THEN** tests SHALL verify correct table operations

#### Scenario: MAC_LIST operations tested
- **WHEN** MAC list managed
- **THEN** tests SHALL verify correct persistence

### Requirement: Untested Table Coverage - Metadata Tables
The system SHALL provide tests exercising XCONF_METADATA, ENV_MODEL_BEAN, and ESTB_FIRMWARE_VERSION_SUPPORT tables.

#### Scenario: XCONF_METADATA operations tested
- **WHEN** metadata stored or retrieved
- **THEN** tests SHALL verify correct table operations

#### Scenario: ENV_MODEL_BEAN operations tested
- **WHEN** environment model beans managed
- **THEN** tests SHALL verify correct persistence

#### Scenario: ESTB_FIRMWARE_VERSION_SUPPORT operations tested
- **WHEN** firmware version support checked
- **THEN** tests SHALL verify correct table queries

### Requirement: Untested Table Coverage - NS List Tables
The system SHALL provide tests exercising NS_LIST, IP_ADDRESS_GROUP, and GENERIC_NS_LIST tables.

#### Scenario: NS_LIST operations tested
- **WHEN** namespace lists managed
- **THEN** tests SHALL verify correct table operations

#### Scenario: IP_ADDRESS_GROUP operations tested
- **WHEN** IP address groups managed
- **THEN** tests SHALL verify correct persistence

#### Scenario: GENERIC_NS_LIST operations tested
- **WHEN** generic namespace lists used
- **THEN** tests SHALL verify correct table operations

### Requirement: Untested Table Coverage - Firmware Tables
The system SHALL provide tests exercising FIRMWARE_RULE, FIRMWARE_RULE_TEMPLATE, and FIRMWARE_CONFIG_LOGS tables.

#### Scenario: FIRMWARE_RULE operations tested
- **WHEN** firmware rules managed
- **THEN** tests SHALL verify correct table operations

#### Scenario: FIRMWARE_RULE_TEMPLATE operations tested
- **WHEN** firmware rule templates used
- **THEN** tests SHALL verify correct persistence

#### Scenario: FIRMWARE_CONFIG_LOGS operations tested
- **WHEN** firmware configuration logged
- **THEN** tests SHALL verify correct log table operations

### Requirement: Untested Table Coverage - DCM Tables
The system SHALL provide tests exercising DEVICE_SETTINGS, VOD_SETTINGS, LOG_FILE_LIST, and UPLOAD_REPOSITORY tables.

#### Scenario: DEVICE_SETTINGS operations tested
- **WHEN** device settings managed
- **THEN** tests SHALL verify correct table operations

#### Scenario: VOD_SETTINGS operations tested
- **WHEN** VOD settings configured
- **THEN** tests SHALL verify correct persistence

#### Scenario: LOG_FILE_LIST operations tested
- **WHEN** log file lists managed
- **THEN** tests SHALL verify correct table operations

#### Scenario: UPLOAD_REPOSITORY operations tested
- **WHEN** upload repositories configured
- **THEN** tests SHALL verify correct persistence

### Requirement: Untested Table Coverage - Feature Control Tables
The system SHALL provide tests exercising FEATURE_CONTROL_RULE and XCONF_FEATURE tables.

#### Scenario: FEATURE_CONTROL_RULE operations tested
- **WHEN** feature control rules managed
- **THEN** tests SHALL verify correct table operations

#### Scenario: XCONF_FEATURE operations tested (with existing tests)
- **WHEN** features managed
- **THEN** tests SHALL augment existing partial coverage to be comprehensive

### Requirement: Test File Organization
The system SHALL organize new test files following project conventions.

#### Scenario: Package-level test files
- **WHEN** creating tests for shared/estbfirmware
- **THEN** tests SHALL be in shared/estbfirmware/config_change_logs_test.go

#### Scenario: DAO test files in db package
- **WHEN** creating tests for DAOs
- **THEN** tests SHALL be in db/ directory with _test.go suffix

#### Scenario: Util test files
- **WHEN** creating tests for utilities
- **THEN** tests SHALL be in util/ directory with _test.go suffix
