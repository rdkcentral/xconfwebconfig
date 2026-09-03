## ADDED Requirements

### Requirement: Test Function Independence
The system SHALL ensure each test function can execute in isolation without depending on other test functions.

#### Scenario: No cross-function calls
- **WHEN** reviewing test code
- **THEN** no test function SHALL invoke another test function directly

#### Scenario: Shared setup via helpers
- **WHEN** multiple tests need common setup
- **THEN** tests SHALL use shared helper functions (not test functions) for setup

#### Scenario: Independent data creation
- **WHEN** test requires test data
- **THEN** test SHALL create its own data or use fixture helpers, not rely on other tests

### Requirement: Test Execution Order Independence
The system SHALL ensure tests can run in any order without failures.

#### Scenario: Random order execution
- **WHEN** tests run with `go test -shuffle=on`
- **THEN** all tests SHALL pass regardless of execution order

#### Scenario: Parallel execution safety
- **WHEN** tests run with `go test -parallel`
- **THEN** tests SHALL not interfere with each other

#### Scenario: Single test execution
- **WHEN** developer runs single test with `go test -run TestSpecificFunction`
- **THEN** test SHALL pass without requiring other tests to run first

### Requirement: No Shared Mutable State
The system SHALL eliminate all shared mutable global variables used by tests.

#### Scenario: Test-local variables
- **WHEN** test needs state tracking
- **THEN** test SHALL use local variables or test-scoped structs

#### Scenario: Mock instances per test
- **WHEN** test uses mocks
- **THEN** each test SHALL create fresh mock instances, not share global mocks

#### Scenario: No global DAO instances
- **WHEN** test needs DAO
- **THEN** test SHALL obtain DAO via `getTestDAO()`, not use global DAO variable

### Requirement: Fixture Data Generation
The system SHALL provide fixture generation helpers that create fresh data for each test invocation.

#### Scenario: generateTestModels helper
- **WHEN** test calls `generateTestModels(n)`
- **THEN** helper SHALL return n unique model instances with distinct keys

#### Scenario: Unique key generation
- **WHEN** fixture helper called multiple times
- **THEN** each invocation SHALL generate unique keys (e.g., UUID, timestamp-based, or counter)

#### Scenario: Fixture determinism
- **WHEN** test requires reproducible data
- **THEN** fixture helper MAY accept seed parameter for deterministic generation

### Requirement: Test State Cleanup
The system SHALL ensure all test state is cleaned up after test execution.

#### Scenario: CleanupTracker cleanup
- **WHEN** test uses CleanupTracker
- **THEN** defer cleanup SHALL remove all test-created data

#### Scenario: Mock state reset
- **WHEN** using testify mocks
- **THEN** each test SHALL create new mock instances (automatic reset)

#### Scenario: No leaked goroutines
- **WHEN** test spawns goroutines
- **THEN** test SHALL wait for all goroutines to complete before returning

### Requirement: Helper Function Conventions
The system SHALL distinguish between test helpers and test functions using naming conventions.

#### Scenario: Helper function naming
- **WHEN** creating test helper function
- **THEN** function name SHALL NOT start with "Test" prefix

#### Scenario: Helper function signature
- **WHEN** creating test helper
- **THEN** function MAY accept `*testing.T` as parameter but SHALL NOT be registered as test

#### Scenario: Setup helper pattern
- **WHEN** creating setup helper
- **THEN** helper SHALL return cleanup function for defer usage: `cleanup := setupHelper(t); defer cleanup()`

### Requirement: Assertion Independence
The system SHALL ensure tests make independent assertions without relying on previous test assertions.

#### Scenario: Self-contained verification
- **WHEN** test verifies behavior
- **THEN** test SHALL assert all preconditions and outcomes independently

#### Scenario: No assumption of DB state
- **WHEN** test begins execution
- **THEN** test SHALL NOT assume specific data exists from previous tests

### Requirement: Idempotency Verification Script
The system SHALL provide a script to verify test idempotency.

#### Scenario: verify_idempotency.sh execution
- **WHEN** developer runs `./scripts/verify_idempotency.sh`
- **THEN** script SHALL run test suite 3 times in different orders and report any failures

#### Scenario: Shuffle verification
- **WHEN** verify_idempotency.sh runs
- **THEN** one run SHALL use `-shuffle=on` flag to randomize order

#### Scenario: Parallel verification
- **WHEN** verify_idempotency.sh runs
- **THEN** one run SHALL use `-parallel=4` flag to test concurrent safety

### Requirement: Documentation of Idempotency Patterns
The system SHALL provide documentation of idempotent test patterns in test files.

#### Scenario: Idempotent test example
- **WHEN** developer reads test documentation
- **THEN** examples SHALL show proper use of fixtures, cleanup, and isolation

#### Scenario: Anti-pattern documentation
- **WHEN** developer reads test documentation
- **THEN** documentation SHALL explicitly show forbidden patterns (test calling test, shared globals, etc.)
