## Why

Currently, 113 test files exist across the xconfwebconfig codebase, but only 13 (11.5%) actually test database operations. Of the 18 production files that use database DAOs, 14 (78%) have NO database test coverage. This creates critical risk: production code interacting with 33 database tables has minimal test coverage, and tests that do exist use inconsistent patterns including full table truncation, lack of cleanup tracking, and no mock support. This prevents reliable testing, causes test interdependencies, and makes it impossible to run tests without a real Cassandra database.

## What Changes

- **Add DAO Interface Abstraction**: Create `TestableDAO` interfaces for all DAO types (SimpleDAO, CachedDAO, ListingDAO, CompressingDAO, GroupServiceDAO) to enable dependency injection in tests
- **Mock Infrastructure**: Implement complete mock implementations for all DAO types to enable testing without a real database
- **USE_MOCK_DB Environment Variable**: Add command-line control to run tests with either mock or real database via `make test USE_MOCK_DB=true` or `make test USE_MOCK_DB=false`
- **Cleanup Tracking System**: Implement `CleanupTracker` to track all data insertions per test and delete ONLY test-inserted data (no more full table truncates)
- **Idempotent Test Pattern**: Refactor all database tests to be self-contained with unique test data and explicit cleanup at test completion
- **Coverage Verification Loop**: Establish per-function verification protocol to ensure coverage with mock matches real DB and no regression occurs
- **New Test Files**: Create 14 new test files for production code that currently has zero database test coverage
- **Makefile Enhancements**: Add `test-mock`, `test-real`, `cover-mock`, `cover-real`, and `compare-coverage` targets

## Capabilities

### New Capabilities
- `dao-mock-infrastructure`: Complete mock DAO implementations and test infrastructure for all database access patterns
- `test-mode-toggle`: Command-line control to run tests with mock or real database
- `cleanup-tracking`: Per-test data insertion tracking and targeted cleanup system
- `idempotent-tests`: Self-contained test pattern ensuring tests can run in any order
- `coverage-verification`: Automated verification that mock and real DB tests achieve equivalent coverage
- `missing-test-coverage`: Comprehensive test coverage for 14 currently untested production files and 15 untested database tables

### Modified Capabilities
- `existing-dao-tests`: All 13 existing test files using database operations will be refactored to use new infrastructure (requirements: support both mock and real DB, eliminate table truncates, track cleanup)
- `test-execution`: Makefile test targets enhanced with USE_MOCK_DB support (requirement: backward compatible, default to real DB)

## Impact

**Affected Code**:
- **Test Files**: All 113 test files will support USE_MOCK_DB toggle (13 actively refactored, 100 enhanced)
- **Build System**: Makefile updated with new test targets
- **New Files Created**: 
  - `db/test_infrastructure.go` (DAO interfaces, mocks, cleanup tracker)
  - 14 new test files for untested production code
- **Production Code**: ZERO modifications (critical constraint)

**Affected Tables**: All 33 database tables gain test coverage including:
- Critical gaps: LOGS, FIRMWARE_CONFIG, XCONF_FEATURE, DCM_RULE, LOG_UPLOAD_SETTINGS
- Untested tables: APPLICATION_TYPES, APP_SETTINGS, TAG, all CHANGE tables

**Dependencies**:
- Go testing framework (existing)
- testify/mock (existing)
- Cassandra test database (existing, for real DB mode)

**Risk Mitigation**:
- Per-function coverage verification prevents regression
- Dual-mode testing ensures mock behavior matches real DB
- Zero production code changes eliminates deployment risk
- Incremental rollout allows early feedback
