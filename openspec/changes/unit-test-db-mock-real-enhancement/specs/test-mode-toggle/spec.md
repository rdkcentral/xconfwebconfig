## ADDED Requirements

### Requirement: USE_MOCK_DB Environment Variable
The system SHALL support a `USE_MOCK_DB` environment variable that controls whether tests use mock or real database implementations.

#### Scenario: Mock mode activation
- **WHEN** `USE_MOCK_DB=true` is set
- **THEN** all tests SHALL use mock DAO implementations

#### Scenario: Real mode activation
- **WHEN** `USE_MOCK_DB=false` is set
- **THEN** all tests SHALL use real database DAO implementations

#### Scenario: Default behavior
- **WHEN** `USE_MOCK_DB` is not set
- **THEN** tests SHALL default to `false` (real DB) for backward compatibility

#### Scenario: Invalid mode value
- **WHEN** `USE_MOCK_DB` set to invalid value (not "true" or "false")
- **THEN** tests SHALL default to `false` (real DB) and log warning

### Requirement: Makefile Test Targets
The system SHALL provide Makefile targets that simplify running tests in different modes.

#### Scenario: test-mock target
- **WHEN** developer runs `make test-mock`
- **THEN** Makefile SHALL execute `USE_MOCK_DB=true go test ./... -cover -count=1`

#### Scenario: test-real target
- **WHEN** developer runs `make test-real`
- **THEN** Makefile SHALL execute `USE_MOCK_DB=false go test ./... -cover -count=1`

#### Scenario: Default test target backward compatibility
- **WHEN** developer runs `make test`
- **THEN** Makefile SHALL run tests with USE_MOCK_DB=false (existing behavior preserved)

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
The system SHALL support USE_MOCK_DB in CI/CD pipelines without Makefile dependency.

#### Scenario: Direct go test invocation
- **WHEN** CI runs `USE_MOCK_DB=true go test ./...`
- **THEN** tests SHALL execute in mock mode

#### Scenario: Parallel test execution
- **WHEN** CI runs mock and real tests in parallel jobs
- **THEN** both test suites SHALL complete successfully without interference

### Requirement: Mode Detection in Tests
The system SHALL provide helper function `getTestMode()` that returns current test mode.

#### Scenario: Mode detection from environment
- **WHEN** test calls `getTestMode()`
- **THEN** function SHALL return "true" or "false" based on USE_MOCK_DB environment variable

#### Scenario: Test mode logging
- **WHEN** test suite starts
- **THEN** framework SHALL log which USE_MOCK_DB value is active for debugging

### Requirement: Mode-Specific Test Skipping
The system SHALL allow tests to skip execution based on test mode when necessary.

#### Scenario: Mock-only test
- **WHEN** test requires mock-specific behavior
- **THEN** test MAY call `skipIfRealMode(t)` to skip when USE_MOCK_DB=false

#### Scenario: Real-only test
- **WHEN** test requires real DB behavior (e.g., transaction testing)
- **THEN** test MAY call `skipIfMockMode(t)` to skip when USE_MOCK_DB=true
