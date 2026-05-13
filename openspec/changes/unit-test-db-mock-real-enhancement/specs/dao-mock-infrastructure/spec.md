## ADDED Requirements

### Requirement: TestableDAO Interface for SimpleDAO
The system SHALL provide a `TestableSimpleDAO` interface that mirrors all methods of the production `SimpleDAO` to enable dependency injection in tests.

#### Scenario: Interface matches production DAO
- **WHEN** test code uses `TestableSimpleDAO` interface
- **THEN** both mock and real implementations SHALL be assignable to the interface

#### Scenario: GetOne operation
- **WHEN** test calls `dao.GetOne(table, key)`
- **THEN** interface SHALL return `(interface{}, error)` matching production signature

#### Scenario: SetOne operation
- **WHEN** test calls `dao.SetOne(table, key, data)`
- **THEN** interface SHALL return `error` matching production signature

#### Scenario: DeleteOne operation
- **WHEN** test calls `dao.DeleteOne(table, key)`
- **THEN** interface SHALL return `error` matching production signature

### Requirement: TestableDAO Interface for CachedDAO
The system SHALL provide a `TestableCachedDAO` interface that mirrors all methods of the production `CachedSimpleDAO` including cache invalidation methods.

#### Scenario: Cache operations included
- **WHEN** test uses `TestableCachedDAO`
- **THEN** interface SHALL include `InvalidateCache(table, key)` method

#### Scenario: Wrapper for real CachedDAO
- **WHEN** test runs with real DB mode
- **THEN** wrapper SHALL delegate all calls to `db.GetCachedSimpleDao()`

### Requirement: TestableDAO Interface for ListingDAO
The system SHALL provide a `TestableListingDAO` interface that supports composite key operations (rowKey + columnName pattern).

#### Scenario: Composite key GetOne
- **WHEN** test calls `dao.GetOne(table, rowKey, columnName)`
- **THEN** interface SHALL return `([]byte, error)` for listing data

#### Scenario: Composite key SetOne
- **WHEN** test calls `dao.SetOne(table, rowKey, columnName, data)`
- **THEN** interface SHALL store data with composite key

#### Scenario: GetAll for rowKey
- **WHEN** test calls `dao.GetAll(table, rowKey)`
- **THEN** interface SHALL return all column data for that rowKey

#### Scenario: GetRange operation
- **WHEN** test calls `dao.GetRange(table, rowKey, rangeInfo)`
- **THEN** interface SHALL return time-ordered range of entries

### Requirement: TestableDAO Interface for CompressingDAO
The system SHALL provide a `TestableCompressingDAO` interface that handles compressed and split data operations.

#### Scenario: Compression flag support
- **WHEN** test uses CompressingDAO for compressed tables
- **THEN** interface SHALL handle transparent compression/decompression

#### Scenario: Split data support
- **WHEN** data exceeds size threshold
- **THEN** interface SHALL support chunk-based operations

### Requirement: TestableDAO Interface for GroupServiceDAO
The system SHALL provide a `TestableGroupServiceDAO` interface for group service cache operations.

#### Scenario: GetGroupServiceFeatureTags
- **WHEN** test calls `dao.GetGroupServiceFeatureTags(cacheKey)`
- **THEN** interface SHALL return `map[string]string` of feature tags

#### Scenario: SetGroupServiceFeatureTags
- **WHEN** test calls `dao.SetGroupServiceFeatureTags(cacheKey, tags)`
- **THEN** interface SHALL persist feature tags

### Requirement: MockSimpleDAO Implementation
The system SHALL provide a `MockSimpleDAO` struct using testify/mock that implements `TestableSimpleDAO`.

#### Scenario: Mock expectations setup
- **WHEN** test sets up mock with `mockDAO.On("GetOne", table, key).Return(data, nil)`
- **THEN** subsequent call to `GetOne` SHALL return the configured response

#### Scenario: Mock call verification
- **WHEN** test completes
- **THEN** `mockDAO.AssertExpectations(t)` SHALL verify all expected calls occurred

#### Scenario: Mock default behavior
- **WHEN** mock method called without expectation
- **THEN** mock SHALL return zero values and record unexpected call

### Requirement: MockCachedDAO Implementation
The system SHALL provide a `MockCachedDAO` struct that implements `TestableCachedDAO` with cache invalidation tracking.

#### Scenario: Cache invalidation tracked
- **WHEN** test calls `mockDAO.InvalidateCache(table, key)`
- **THEN** mock SHALL record the invalidation call for verification

#### Scenario: Cache hit simulation
- **WHEN** mock configured with cached data
- **THEN** subsequent `GetOne` SHALL return cached value without DB call

### Requirement: MockListingDAO Implementation
The system SHALL provide a `MockListingDAO` struct that implements `TestableListingDAO` for composite key operations.

#### Scenario: Composite key mocking
- **WHEN** test sets `mockDAO.On("GetOne", table, rowKey, colName).Return(data, nil)`
- **THEN** mock SHALL handle three-parameter GetOne calls

#### Scenario: Range query mocking
- **WHEN** mock configured with range data
- **THEN** `GetRange` SHALL return time-ordered entries

### Requirement: MockCompressingDAO Implementation
The system SHALL provide a `MockCompressingDAO` struct that implements `TestableCompressingDAO`.

#### Scenario: Compression simulation
- **WHEN** mock stores compressed data
- **THEN** retrieval SHALL return decompressed data

### Requirement: MockGroupServiceDAO Implementation
The system SHALL provide a `MockGroupServiceDAO` struct that implements `TestableGroupServiceDAO`.

#### Scenario: Feature tags mocking
- **WHEN** mock configured with feature tags
- **THEN** `GetGroupServiceFeatureTags` SHALL return configured map

### Requirement: getTestDAO Factory Function
The system SHALL provide a `getTestDAO(t *testing.T, daoType string)` function that returns appropriate DAO implementation based on TEST_MODE.

#### Scenario: Mock mode returns mock
- **WHEN** `TEST_MODE=mock` and test calls `getTestDAO(t, "simple")`
- **THEN** function SHALL return `MockSimpleDAO` instance

#### Scenario: Real mode returns real DAO
- **WHEN** `TEST_MODE=real` and test calls `getTestDAO(t, "simple")`
- **THEN** function SHALL return wrapper around `db.GetSimpleDao()`

#### Scenario: Real mode without DB skips test
- **WHEN** `TEST_MODE=real` but `!db.IsCassandraClient()`
- **THEN** function SHALL call `t.Skip("Real DB not available")`

#### Scenario: Default mode
- **WHEN** `TEST_MODE` not set
- **THEN** function SHALL default to `real` mode for backward compatibility

### Requirement: RealDAOWrapper Implementations
The system SHALL provide wrapper structs (`RealSimpleDAOWrapper`, `RealCachedDAOWrapper`, etc.) that implement Testable interfaces by delegating to production DAOs.

#### Scenario: Simple wrapper delegation
- **WHEN** wrapper `GetOne` called
- **THEN** wrapper SHALL call `db.GetSimpleDao().GetOne()` and return result

#### Scenario: Error propagation
- **WHEN** production DAO returns error
- **THEN** wrapper SHALL propagate error unchanged

### Requirement: Test Infrastructure Documentation
The system SHALL provide comprehensive documentation in `db/test_infrastructure.go` with examples of mock usage patterns.

#### Scenario: Usage example included
- **WHEN** developer reads test_infrastructure.go
- **THEN** file SHALL contain commented example of setting up mocks and using getTestDAO

#### Scenario: Interface documentation
- **WHEN** developer views interface definition
- **THEN** each method SHALL have godoc comment explaining parameters and return values
