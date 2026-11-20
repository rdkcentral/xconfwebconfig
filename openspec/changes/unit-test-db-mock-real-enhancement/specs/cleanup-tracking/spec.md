## ADDED Requirements

### Requirement: CleanupTracker Structure
The system SHALL provide a `CleanupTracker` struct that records all data inserted during tests for precise cleanup.

#### Scenario: Tracker initialization
- **WHEN** test initializes tracker with `tracker := NewCleanupTracker(dao)`
- **THEN** tracker SHALL create empty tracking maps for each table

#### Scenario: Insertion tracking
- **WHEN** test calls `tracker.Insert(table, key, data)`
- **THEN** tracker SHALL record the table/key combination for later cleanup

#### Scenario: Composite key tracking
- **WHEN** test uses ListingDAO with `tracker.InsertListing(table, rowKey, colName, data)`
- **THEN** tracker SHALL record composite key for cleanup

### Requirement: Explicit Cleanup Pattern
The system SHALL support explicit cleanup execution at test completion without using defer.

#### Scenario: Cleanup at test end
- **WHEN** test completes execution
- **THEN** test SHALL explicitly call `tracker.Cleanup(t)` before returning

#### Scenario: Cleanup after test assertions
- **WHEN** all test assertions complete
- **THEN** test SHALL execute cleanup as final statement

#### Scenario: Cleanup on test failure
- **WHEN** test fails before cleanup
- **THEN** cleanup SHALL NOT execute automatically (test framework handles partial state)

### Requirement: Surgical Data Deletion
The system SHALL delete ONLY data inserted during the test, preserving all other table data.

#### Scenario: Delete tracked keys only
- **WHEN** tracker cleanup executes
- **THEN** tracker SHALL call `dao.DeleteOne()` only for tracked table/key pairs

#### Scenario: Skip untracked data
- **WHEN** cleanup runs on table with pre-existing data
- **THEN** tracker SHALL NOT delete data not inserted by current test

#### Scenario: Multiple test isolation
- **WHEN** two tests run sequentially
- **THEN** first test's cleanup SHALL NOT affect second test's data

### Requirement: Error Handling During Cleanup
The system SHALL handle cleanup errors gracefully without masking test failures.

#### Scenario: Cleanup error logged
- **WHEN** cleanup encounters deletion error
- **THEN** tracker SHALL log error with `t.Logf()` and continue cleanup

#### Scenario: Partial cleanup success
- **WHEN** cleanup fails for one key but succeeds for others
- **THEN** tracker SHALL delete all successfully deletable keys

#### Scenario: Cleanup error does not fail passing test
- **WHEN** test passed but cleanup encounters error
- **THEN** tracker SHALL log error but NOT call `t.Error()` to avoid false failure

### Requirement: Nested Cleanup Support
The system SHALL support multiple cleanup trackers within a single test function.

#### Scenario: Multiple trackers
- **WHEN** test uses separate trackers for different tables
- **THEN** each tracker SHALL independently track and clean its data

#### Scenario: Cleanup order control
- **WHEN** multiple trackers used in single test
- **THEN** test SHALL explicitly call cleanup in desired order (first-created first-cleaned)

### Requirement: Real vs Mock Mode Cleanup
The system SHALL execute cleanup operations appropriate to the test mode.

#### Scenario: Mock mode cleanup
- **WHEN** `USE_MOCK_DB=true`
- **THEN** cleanup SHALL call mock DAO delete methods (verified via mock assertions)

#### Scenario: Real mode cleanup
- **WHEN** `USE_MOCK_DB=false`
- **THEN** cleanup SHALL execute actual database DELETE operations

### Requirement: Cleanup Verification
The system SHALL provide optional cleanup verification for test development.

#### Scenario: Verify cleanup success
- **WHEN** test enables verification with `tracker.SetVerifyCleanup(true)`
- **THEN** tracker SHALL call `dao.GetOne()` after delete and assert not found

#### Scenario: Verification disabled by default
- **WHEN** tracker created without verification flag
- **THEN** cleanup SHALL not perform verification queries (performance optimization)

### Requirement: Deprecation of truncateTable Pattern
The system SHALL eliminate all usage of `truncateTable()` and table-wide DELETE operations in tests.

#### Scenario: No truncateTable calls
- **WHEN** reviewing test code
- **THEN** no test SHALL call `truncateTable()` or `db.TRUNCATE` operations

#### Scenario: No double cleanup pattern
- **WHEN** reviewing test code
- **THEN** no test SHALL have both cleanup-at-start AND cleanup-at-end patterns

### Requirement: CleanupTracker Helper Methods
The system SHALL provide convenience methods for common cleanup patterns.

#### Scenario: InsertAndTrack method
- **WHEN** test calls `tracker.InsertAndTrack(table, key, data)`
- **THEN** tracker SHALL both insert data via DAO AND record for cleanup

#### Scenario: TrackExisting method
- **WHEN** test needs to track data created by code under test
- **THEN** test MAY call `tracker.Track(table, key)` to register for cleanup without insertion

### Requirement: CleanupTracker Documentation
The system SHALL provide clear documentation with examples in `db/test_helpers_test.go`.

#### Scenario: Usage example
- **WHEN** developer reads test_helpers_test.go
- **THEN** file SHALL contain commented example of tracker initialization and explicit cleanup pattern

#### Scenario: Anti-pattern documentation
- **WHEN** developer reads test_helpers_test.go
- **THEN** file SHALL document why truncateTable is unsafe, why defer should be avoided, and how tracker solves it
