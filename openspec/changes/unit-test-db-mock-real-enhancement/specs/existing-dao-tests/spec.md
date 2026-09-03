## MODIFIED Requirements

### Requirement: Existing DAO tests use CleanupTracker
Tests in tests/dao_test.go and tests/cached_simple_dao_test.go SHALL be refactored to use CleanupTracker instead of truncateTable pattern.

#### Scenario: Remove truncateTable calls
- **WHEN** refactoring existing DAO tests
- **THEN** all calls to `truncateTable()` SHALL be removed and replaced with CleanupTracker

#### Scenario: Add CleanupTracker to TestCRUD
- **WHEN** TestCRUD executes
- **THEN** test SHALL initialize tracker with `tracker := NewCleanupTracker(dao)` and defer cleanup

#### Scenario: Add CleanupTracker to TestGetAllByKeys
- **WHEN** TestGetAllByKeys executes
- **THEN** test SHALL use tracker to record inserted test data

#### Scenario: Add CleanupTracker to TestGetAllAsList
- **WHEN** TestGetAllAsList executes
- **THEN** test SHALL use tracker for surgical cleanup

#### Scenario: Add CleanupTracker to TestGetAllAsMap
- **WHEN** TestGetAllAsMap executes
- **THEN** test SHALL use tracker for cleanup

#### Scenario: Add CleanupTracker to TestGetKeys
- **WHEN** TestGetKeys executes
- **THEN** test SHALL use tracker for cleanup

#### Scenario: Add CleanupTracker to cached DAO tests
- **WHEN** cached_simple_dao_test.go tests execute
- **THEN** all 7 test functions SHALL use CleanupTracker

#### Scenario: Remove double cleanup pattern
- **WHEN** refactoring cached DAO tests
- **THEN** remove both `truncate at start` AND `defer truncate at end` patterns, replace with single tracker defer

### Requirement: Existing DAO tests support USE_MOCK_DB
Tests in tests/dao_test.go and tests/cached_simple_dao_test.go SHALL support both mock and real database execution.

#### Scenario: Replace direct DAO access with getTestDAO
- **WHEN** existing tests access DAO
- **THEN** tests SHALL obtain DAO via `dao := getTestDAO(t, "simple")` instead of `db.GetSimpleDao()`

#### Scenario: Mock expectations in existing tests
- **WHEN** USE_MOCK_DB=true
- **THEN** existing tests SHALL set up mock expectations for all DAO calls

#### Scenario: Real DB execution preserved
- **WHEN** USE_MOCK_DB=false
- **THEN** existing tests SHALL execute against real database as they currently do

### Requirement: Existing DAO tests are idempotent
Tests in tests/dao_test.go and tests/cached_simple_dao_test.go SHALL be fully idempotent.

#### Scenario: No cross-test dependencies
- **WHEN** tests run in random order
- **THEN** all existing tests SHALL pass

#### Scenario: Unique key generation
- **WHEN** generateTestModels helper called
- **THEN** helper SHALL generate unique keys (e.g., UUID-based) instead of predictable keys

#### Scenario: No shared state
- **WHEN** tests execute
- **THEN** tests SHALL not rely on global mutable variables

### Requirement: Existing DAO test coverage verified
Tests in tests/dao_test.go and tests/cached_simple_dao_test.go SHALL have coverage verified in both modes.

#### Scenario: Coverage baseline established
- **WHEN** refactoring begins
- **THEN** current coverage SHALL be recorded as baseline

#### Scenario: Coverage maintained or improved
- **WHEN** refactoring completes
- **THEN** coverage SHALL not decrease and ideally improve

#### Scenario: Both modes achieve similar coverage
- **WHEN** coverage verification runs
- **THEN** mock and real mode coverage SHALL be within 5% of each other
