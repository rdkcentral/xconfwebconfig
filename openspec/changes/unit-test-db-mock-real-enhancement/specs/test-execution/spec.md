## MODIFIED Requirements

### Requirement: Test execution supports mode selection
The test execution system SHALL support running tests with either mock or real database via USE_MOCK_DB environment variable.

#### Scenario: Mock mode execution via environment
- **WHEN** developer runs `USE_MOCK_DB=true go test ./...`
- **THEN** all tests SHALL execute using mock DAOs

#### Scenario: Real mode execution via environment
- **WHEN** developer runs `USE_MOCK_DB=false go test ./...`
- **THEN** all tests SHALL execute using real database DAOs

#### Scenario: Makefile target for mock mode
- **WHEN** developer runs `make test-mock`
- **THEN** tests SHALL execute in mock mode with coverage

#### Scenario: Makefile target for real mode
- **WHEN** developer runs `make test-real`
- **THEN** tests SHALL execute in real mode with coverage

### Requirement: Coverage generation supports both modes
The test execution system SHALL generate separate coverage reports for mock and real modes.

#### Scenario: Mock mode coverage generation
- **WHEN** developer runs `make cover-mock`
- **THEN** system SHALL generate `coverage-mock.out` file

#### Scenario: Real mode coverage generation
- **WHEN** developer runs `make cover-real`
- **THEN** system SHALL generate `coverage-real.out` file

#### Scenario: Coverage comparison
- **WHEN** developer runs `make compare-coverage`
- **THEN** system SHALL display side-by-side coverage percentages from both modes

#### Scenario: HTML coverage report
- **WHEN** developer runs `make coverage-report`
- **THEN** system SHALL generate HTML visualization of coverage data

### Requirement: Test execution includes verification loop
The test execution workflow SHALL include coverage verification after each task completion.

#### Scenario: Task completion verification
- **WHEN** developer completes implementation task
- **THEN** developer SHALL run both `make cover-mock` and `make cover-real`

#### Scenario: Coverage recording in tasks
- **WHEN** task completes
- **THEN** developer SHALL document coverage metrics in tasks.md

#### Scenario: Coverage regression detection
- **WHEN** new tests added
- **THEN** developer SHALL verify coverage improved or remained stable

### Requirement: Test execution supports parallel execution
The test execution system SHALL safely support parallel test execution in both modes.

#### Scenario: Parallel mock tests
- **WHEN** running `USE_MOCK_DB=true go test -parallel=4 ./...`
- **THEN** tests SHALL execute concurrently without interference

#### Scenario: Parallel real tests
- **WHEN** running `USE_MOCK_DB=false go test -parallel=4 ./...`
- **THEN** tests SHALL execute concurrently with isolated cleanup

### Requirement: Test execution supports selective targeting
The test execution system SHALL support running specific tests or packages.

#### Scenario: Single test execution
- **WHEN** developer runs `USE_MOCK_DB=true go test -run TestSpecificFunction`
- **THEN** only specified test SHALL execute

#### Scenario: Package-level execution
- **WHEN** developer runs `USE_MOCK_DB=false go test ./db/...`
- **THEN** only tests in db package SHALL execute

#### Scenario: Coverage for specific package
- **WHEN** developer runs `USE_MOCK_DB=true go test -cover ./shared/estbfirmware/...`
- **THEN** coverage SHALL be reported for estbfirmware package only
