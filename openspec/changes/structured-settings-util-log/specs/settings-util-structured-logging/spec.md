## ADDED Requirements

### Requirement: SettingsUtil request-completion events are structured
For every request handled by `GetLogUploaderSettings`, including T2 settings and `checkNow` requests, the system SHALL emit exactly one info-level `SettingsUtil XCONF_LOG` event on completion, whether successful or unsuccessful. The event SHALL include the HTTP `status` and retain available essential device context and caller-provided audit fields, applying the existing log-field filtering. Applied settings SHALL be identified by structured `logUploadRule` and `vodRule` objects containing rule ID, rule name, and base rule condition. The message SHALL NOT concatenate rule conditions or device attributes into free-form text.

#### Scenario: Log-upload settings are applied
- **WHEN** `CopySettings` applies log-upload settings for a matching rule
- **THEN** the request-completion event includes `logUploadRule` with the rule ID, rule name, and base condition, along with HTTP status and filtered available device and audit fields

#### Scenario: VOD settings are applied
- **WHEN** `CopySettings` applies VOD settings for a matching rule
- **THEN** the same request-completion event includes `vodRule` with the rule ID, rule name, and base condition, without an additional event

#### Scenario: Settings are not applied
- **WHEN** no matching settings are found
- **THEN** the request emits `SettingsUtil XCONF_LOG` with status 404 and without applied-rule objects

#### Scenario: Request fails early
- **WHEN** the settings handler returns an error before rule evaluation
- **THEN** it still emits exactly one `SettingsUtil XCONF_LOG` with the error status and available device context