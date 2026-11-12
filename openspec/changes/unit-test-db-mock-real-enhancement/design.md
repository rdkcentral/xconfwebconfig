## Context

The xconfwebconfig project has 113 test files but only 13 test database operations. Tests that do use the database have several critical issues:
- **No abstraction**: Direct calls to `db.GetSimpleDao()`, `db.GetCachedDao()`, etc. prevent dependency injection
- **Inconsistent cleanup**: Mix of `truncateTable()` (deletes ALL data), manual `DeleteOne()`, and missing cleanup
- **No mock support**: Tests require real Cassandra database, cannot run in CI/CD without infrastructure
- **Non-idempotent**: Tests depend on execution order, shared test data via helper functions
- **Coverage gaps**: 14 production files with database operations have ZERO tests

Current test infrastructure in `db/setup_teardown_db.go` provides `SetUp()` and `TearDown()` that truncate entire tables. This is inadequate for modern testing practices.

**Constraints**:
- ZERO production code modifications allowed (only test code changes)
- Must maintain backward compatibility (existing tests must pass)
- Real database tests must continue working for integration scenarios
- Coverage must not regress (baseline or better)

## Goals / Non-Goals

**Goals:**
- Enable all 113 tests to run with either mock or real database via `USE_MOCK_DB` environment variable
- Provide complete mock implementations for all DAO types
- Eliminate all table-level truncation in favor of targeted cleanup
- Achieve 100% idempotent tests (run in any order, no shared state)
- Add comprehensive test coverage for 14 untested production files
- Establish coverage verification protocol to prevent regression
- Create reusable test infrastructure for future development

**Non-Goals:**
- Modifying production database code (DAOs, clients, schema)
- Changing database connection logic or configuration
- Optimizing database query performance (not a testing concern)
- Adding new database tables or modifying schema
- Replacing testify/mock framework (use existing tools)
- Creating integration test harness (focus is unit/component level)

## Decisions

### Decision 1: DAO Interface Abstraction Layer

**Choice**: Create `TestableDAO` interfaces matching existing DAO method signatures

**Rationale**: 
- Existing code uses `db.GetSimpleDao().GetOne()` pattern - we need drop-in replacements
- Interface-based dependency injection is standard Go practice
- Allows tests to use either mock or real implementation without code changes
- Type-safe compile-time checking

**Alternatives Considered**:
- ❌ Modify production DAOs to accept interfaces: Violates zero-production-code constraint
- ❌ Use reflection/monkey patching: Fragile, breaks with Go updates, poor error messages
- ❌ Test-only wrapper functions: Adds indirection, harder to maintain

### Decision 2: USE_MOCK_DB Environment Variable

**Choice**: Single environment variable `USE_MOCK_DB=true|false` controls test execution mode

**Rationale**:
- Simple, explicit control: `make test USE_MOCK_DB=true`
- Follows 12-factor app principles (config via env vars)
- Easy to integrate with CI/CD pipelines
- Default to `real` maintains backward compatibility

**Alternatives Considered**:
- ❌ Build tags (`//go:build mock`): Requires separate binary builds, complex
- ❌ Command-line flags (`-mock`): Harder to pass through make, less standard
- ❌ Config file: Overkill for single boolean decision

### Decision 3: CleanupTracker Pattern

**Choice**: Each test creates a `CleanupTracker` that records insertions and deletes only that data

**Rationale**:
- Surgical cleanup: Only removes what the test inserted
- Explicit cleanup at test end ensures visibility of cleanup operations
- No automatic defer hides test failure state for debugging
- Tracks composite keys (for ListingDAO)
- Simple API: `tracker.Track(table, key)`

**Alternatives Considered**:
- ❌ Transaction rollback: Cassandra doesn't support transactions
- ❌ Test database cloning: Too slow, infrastructure heavy
- ❌ Keep truncateTable(): Violates idempotency, causes test interference

**Implementation**:
```go
type CleanupTracker struct {
    insertedKeys map[string][]string        // table -> []keys
    listingKeys  map[string][]KeyPair       // table -> []rowKey+columnName
    dao          TestableDAO
}

func (ct *CleanupTracker) Track(table, key string) {
    ct.insertedKeys[table] = append(ct.insertedKeys[table], key)
}

func (ct *CleanupTracker) TrackListing(table, rowKey, colName string) {
    ct.listingKeys[table] = append(ct.listingKeys[table], KeyPair{rowKey, colName})
}

func (ct *CleanupTracker) Cleanup() {
    // Delete in reverse order of insertion
    for table, keys := range ct.insertedKeys {
        for _, key := range keys {
            ct.dao.DeleteOne(table, key)
        }
    }
    // ... similar for listing keys
}
```

### Decision 4: Per-Function Coverage Verification

**Choice**: Automated script `verify_coverage.sh` run after each test function refactor

**Rationale**:
- Catches regressions immediately at function level
- Documents baseline → mock → real coverage progression
- Enforces rule: mock coverage must match real coverage
- Creates audit trail of coverage improvements

**Alternatives Considered**:
- ❌ Manual verification: Error-prone, inconsistent
- ❌ Post-phase verification only: Regressions found too late
- ❌ Coverage targets without verification: No enforcement mechanism

**Process**:
1. Record baseline coverage (real DB, before changes)
2. Refactor function to use DAO interface + CleanupTracker
3. Run with USE_MOCK_DB=true, record coverage
4. Run with USE_MOCK_DB=false, record coverage
5. Assert: real >= baseline AND mock ≈ real
6. Document in task completion notes

### Decision 5: Phased Rollout Strategy

**Choice**: 5 phases starting with infrastructure, then core DAOs, then new tests

**Rationale**:
- Phase 0 (infrastructure) creates foundation with low risk
- Phase 1 (core DAO tests) proves pattern works, gets early feedback
- Phase 2-3 (new tests) addresses critical coverage gaps
- Phase 4-5 (integration, tables) builds on proven foundation
- Each phase delivers value independently

**Alternatives Considered**:
- ❌ Big bang: All 113 files at once - too risky, no early feedback
- ❌ File-by-file ad hoc: Loses architectural coherence
- ❌ Module-first: Doesn't address critical gaps early

**Phase Ordering Rationale**:
```
Phase 0: Infrastructure (3-4 days) → Enables everything else
Phase 1: Core DAOs (4-5 days) → Proves pattern, high reuse
Phase 2: DataAPI (8-10 days) → Addresses critical untested code
Phase 3: Shared (12-15 days) → Addresses most gaps (config_change_logs, logupload)
Phase 4: Integration (8-10 days) → Validates cross-module patterns
Phase 5: Tables (6-8 days) → Completes coverage
```

### Decision 6: Mock Implementation Strategy

**Choice**: Use testify/mock for all mock DAOs with manual setup

**Rationale**:
- testify/mock already in project dependencies
- Type-safe, well-documented, widely used in Go
- Supports expectation setting and verification
- Manual setup (not code generation) keeps it simple

**Alternatives Considered**:
- ❌ gomock: Requires code generation, adds build complexity
- ❌ Hand-rolled mocks: Reinventing wheel, more maintenance
- ❌ In-memory Cassandra: Too heavy, not true unit tests

**Mock Pattern**:
```go
type MockSimpleDAO struct {
    mock.Mock
}

func (m *MockSimpleDAO) GetOne(table, key string) (interface{}, error) {
    args := m.Called(table, key)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0), args.Error(1)
}

// In tests:
mockDAO := &MockSimpleDAO{}
mockDAO.On("GetOne", "TABLE_MODEL", "test-id").Return(testModel, nil)
```

## Risks / Trade-offs

### Risk 1: Mock Behavior Divergence from Real DB
**Risk**: Mock implementations may not perfectly match real database behavior (edge cases, error conditions, performance)

**Mitigation**:
- Dual-mode testing: Every test runs with both mock AND real DB
- Coverage comparison enforces behavioral equivalence
- Integration tests continue using real DB for end-to-end validation
- Document any known differences in test comments

**Trade-off**: Maintaining two test modes adds execution time but provides confidence

### Risk 2: Incomplete Cleanup Tracking
**Risk**: Developer forgets to call `tracker.Track()` after insertion, leaving orphaned test data

**Mitigation**:
- Code review checklist includes cleanup verification
- Create helper functions that auto-track (e.g., `createTestModel(dao, tracker)`)
- Document pattern prominently in test infrastructure
- Consider static analysis tool to detect untracked insertions (future)

**Trade-off**: Requires discipline but better than full table truncates

### Risk 3: Coverage Regression During Refactor
**Risk**: Refactoring test may accidentally reduce code coverage

**Mitigation**:
- Per-function `verify_coverage.sh` script enforces baseline
- Automated in task completion criteria
- Git pre-commit hook could run verification (optional)
- Coverage reports committed with each task

**Trade-off**: Slows development slightly but prevents regressions

### Risk 4: Large Scope, Long Timeline
**Risk**: 41-52 day timeline may lose momentum or priorities change

**Mitigation**:
- Each phase delivers independent value
- Phase 0-1 (7-9 days) proves pattern, enables early stop if needed
- Phase 2-3 address critical gaps (config_change_logs, logupload) first
- Can parallelize with 2 developers (20-26 days)
- OpenSpec task tracking provides clear progress visibility

**Trade-off**: Comprehensive coverage takes time but risk is front-loaded

### Risk 5: USE_MOCK_DB Environment Variable Conflicts
**Risk**: Existing tests or environment may use USE_MOCK_DB for different purpose

**Mitigation**:
- Grep codebase for USE_MOCK_DB before implementation (verify no conflicts)
- Use namespaced variable `XCONF_USE_MOCK_DB` if conflict found
- Document in README and test guidelines
- Makefile provides clean interface, hides implementation

**Trade-off**: Namespaced var is less elegant but safer

### Risk 6: Mock Setup Complexity for Complex DAOs
**Risk**: CompressingDataDAO and GroupServiceDAO have complex interfaces, mocks may be hard to set up

**Mitigation**:
- Create test helper functions for common mock setups
- Document mock patterns with examples in test_infrastructure.go
- Start with simpler DAOs (SimpleDAO, ListingDAO) to establish pattern
- Complex DAOs tackled in later phases after pattern proven

**Trade-off**: Initial complexity pays off with reusable patterns

## Migration Plan

Not applicable - this is test infrastructure only, no production deployment.

## Open Questions

None - exploration phase has resolved all architectural questions. Ready for implementation.
