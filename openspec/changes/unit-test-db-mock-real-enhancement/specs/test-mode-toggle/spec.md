## ADDED Requirements

### Requirement: TEST_MODE Environment Variable
The system SHALL support a `TEST_MODE` environment variable that controls whether tests use mock or real database implementations.

#### Scenario: Mock mode activation
- **WHEN** `TEST_MODE=mock` is set
- **THEN** all tests SHALL use mock DAO implementations

#### Scenario: Real mode activation
- **WHEN** `TEST_MODE=real` is set
- **THEN** all tests SHALL use real database DAO implementations

#### Scenario: Default behavior
- **WHEN** `TEST_MODE` is not set
- **THEN** tests SHALL default to `real` mode for backward compatibility

#### Scenario: Invalid mode value
- **WHEN** `TEST_MODE` set to invalid value (not "mock" or "real")
- **THEN** tests SHALL default to `real` mode and log warning

### Requirement: Makefile Test Targets
The system SHALL provide Makefile targets that simplify running tests in different modes.

#### Scenario: test-mock target
- **WHEN** developer runs `make test-mock`
- **THEN** Makefile SHALL execute `TEST_MODE=mock go test ./... -cover -count=1`

#### Scenario: test-real target
- **WHEN** developer runs `make test-real`
- **THEN** Makefile SHALL execute `TEST_MODE=real go test ./... -cover -count=1`

#### Scenario: Default test target backward compatibility
- **WHEN** developer runs `make test`
- **THEN** Makefile SHALL run tests with TEST_MODE=real (existing behavior preserved)

#### Scenario: cover-mock target
- **WHEN** developer runs `make cover-mock`
- **THEN** Makefile SHALL generate `coverage-mock.out` with mock DAO coverage

#### Scenario: cover-real target
- **WHEN** developer runs `make cover-real`
- **THEN** Makefile SHALL generate `coverage-real.out` with real DAO coverage

### Requirement: Coverage Comparison Target
The system SHALL provide `make compare-coverage` target that displays side-by-side coverage comparison.

#### Scenario: Coverage comparison output
- **WHEN** developer runs `make compare-coverage`
- **THEN** output SHALL display total coverage percentage for both mock and real modes

#### Scenario: Coverage files required
- **WHEN** `make compare-coverage` run without coverage files
- **THEN** Makefile SHALL display helpful error message directing to run cover-mock and cover-real first

### Requirement: CI/CD Integration
The system SHALL support TEST_MODE in CI/CD pipelines without Makefile dependency.

#### Scenario: Direct go test invocation
- **WHEN** CI runs `TEST_MODE=mock go test ./...`
- **THEN** tests SHALL execute in mock mode

#### Scenario: Parallel test execution
- **WHEN** CI runs mock and real tests in parallel jobs
- **THEN** both test suites SHALL complete successfully without interference

### Requirement: Mode Detection in Tests
The system SHALL provide helper function `getTestMode()` that returns current test mode.

#### Scenario: Mode detection from environment
- **WHEN** test calls `getTestMode()`
- **THEN** function SHALL return "mock" or "real" based on TEST_MODE environment variable

#### Scenario: Test mode logging
- **WHEN** test suite starts
- **THEN** framework SHALL log which TEST_MODE is active for debugging

### Requirement: Mode-Specific Test Skipping
The system SHALL allow tests to skip execution based on test mode when necessary.

#### Scenario: Mock-only test
- **WHEN** test requires mock-specific behavior
- **THEN** test MAY call `skipIfRealMode(t)` to skip in real mode

#### Scenario: Real-only test
- **WHEN** test requires real DB behavior (e.g., transaction testing)
- **THEN** test MAY call `skipIfMockMode(t)` to skip in mock mode
