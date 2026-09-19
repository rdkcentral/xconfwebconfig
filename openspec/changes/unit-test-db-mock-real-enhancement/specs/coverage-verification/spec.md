## ADDED Requirements

### Requirement: Per-Function Coverage Verification
The system SHALL provide a script that measures and reports test coverage for individual functions.

#### Scenario: Function-level coverage
- **WHEN** developer runs `./scripts/verify_coverage.sh functionName`
- **THEN** script SHALL report coverage percentage for that specific function

#### Scenario: Test file targeting
- **WHEN** developer runs `./scripts/verify_coverage.sh -file path/to/file_test.go`
- **THEN** script SHALL report coverage for all functions tested in that file

#### Scenario: Coverage threshold checking
- **WHEN** developer runs `./scripts/verify_coverage.sh -threshold 80`
- **THEN** script SHALL exit with error if coverage below threshold

### Requirement: Dual-Mode Coverage Comparison
The system SHALL execute coverage verification in both mock and real DB modes and compare results.

#### Scenario: verify_coverage.sh runs both modes
- **WHEN** developer runs `./scripts/verify_coverage.sh`
- **THEN** script SHALL run tests with USE_MOCK_DB=true and USE_MOCK_DB=false and display both coverage percentages

#### Scenario: Coverage delta reporting
- **WHEN** script completes dual-mode coverage
- **THEN** output SHALL show difference between mock and real coverage percentages

#### Scenario: Mode-specific coverage files
- **WHEN** script runs coverage
- **THEN** script SHALL generate separate `coverage-mock.out` and `coverage-real.out` files

### Requirement: Task-Level Coverage Tracking
The system SHALL require coverage verification after completing each implementation task.

#### Scenario: Task completion verification
- **WHEN** developer completes task (e.g., "Add mock for SimpleDAO")
- **THEN** developer SHALL run `make cover-mock` and `make cover-real` and record coverage in task notes

#### Scenario: Coverage regression detection
- **WHEN** new test added
- **THEN** verification SHALL confirm coverage increased or remained stable

#### Scenario: Task sign-off criteria
- **WHEN** marking task complete
- **THEN** task SHALL include coverage metrics for both mock and real modes

### Requirement: CI/CD Coverage Enforcement
The system SHALL enforce minimum coverage thresholds in CI/CD pipeline.

#### Scenario: CI coverage check
- **WHEN** CI pipeline runs tests
- **THEN** pipeline SHALL fail if coverage drops below configured threshold (e.g., 80%)

#### Scenario: Coverage trend reporting
- **WHEN** CI completes
- **THEN** pipeline SHALL publish coverage trend chart showing mock vs real coverage over time

### Requirement: Coverage Report Generation
The system SHALL generate human-readable coverage reports with function-level detail.

#### Scenario: HTML coverage report
- **WHEN** developer runs `make coverage-report`
- **THEN** system SHALL generate HTML report showing line-by-line coverage

#### Scenario: Uncovered code highlighting
- **WHEN** viewing coverage report
- **THEN** uncovered lines SHALL be highlighted in report

#### Scenario: Mode comparison report
- **WHEN** developer runs `make coverage-compare-report`
- **THEN** system SHALL generate side-by-side HTML comparing mock and real coverage

### Requirement: Coverage Metrics in Tasks.md
The system SHALL document expected and actual coverage in tasks.md after each task completion.

#### Scenario: Pre-task baseline
- **WHEN** task begins
- **THEN** developer SHALL record current coverage baseline

#### Scenario: Post-task verification
- **WHEN** task completes
- **THEN** developer SHALL update task with actual achieved coverage (mock and real)

#### Scenario: Coverage delta tracking
- **WHEN** task completes
- **THEN** task notes SHALL include coverage improvement (e.g., "+12% mock, +10% real")

### Requirement: Zero-Coverage Detection
The system SHALL identify functions with zero test coverage.

#### Scenario: Untested function report
- **WHEN** developer runs `./scripts/verify_coverage.sh -report-untested`
- **THEN** script SHALL list all functions with 0% coverage

#### Scenario: New function detection
- **WHEN** new production code added
- **THEN** coverage verification SHALL detect and report functions without tests

### Requirement: Mock vs Real Coverage Parity Checking
The system SHALL detect and report significant differences between mock and real mode coverage.

#### Scenario: Coverage divergence warning
- **WHEN** mock coverage is >10% different from real coverage
- **THEN** verification script SHALL emit warning about potential test gaps

#### Scenario: Mock-only coverage paths
- **WHEN** certain code paths only covered in mock mode
- **THEN** report SHALL highlight these paths for real-mode test addition

#### Scenario: Real-only coverage paths
- **WHEN** certain code paths only covered in real mode
- **THEN** report SHALL highlight these paths for mock-mode test addition
