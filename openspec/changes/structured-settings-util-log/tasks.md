## 1. Implement Structured Events

- [x] 1.1 Replace both applied-rule message strings in `shared/logupload/utils.go` with the stable `SettingsUtil XCONF_LOG` event and structured settings-type, rule, and essential context fields.
- [x] 1.2 Preserve info-level severity and apply `common.FilterLogFields` to the emitted fields without changing settings selection behavior.

## 2. Verify Logging Behavior

- [x] 2.1 Add focused tests for log-upload and VOD events, asserting the event name, settings type, rule details, and available filtered context fields.
- [x] 2.2 Verify the event message no longer embeds concatenated device attributes or rule conditions, and no applied-rule event is emitted when the corresponding settings are not applied.
- [x] 2.3 Run the focused logupload package tests and relevant caller package tests.

## 3. Emit for Every Settings Request

- [x] 3.1 Move event emission to deferred request completion, including status, selected device context, and early errors.
- [x] 3.2 Collect applied log-upload and VOD rule details on audit fields without duplicate events or response changes.
- [x] 3.3 Verify success, no-settings, and early-error logging, and run relevant package tests. Focused tests and both logupload package suites pass; the full dataapi suite requires the unavailable Cassandra `test_appds` keyspace.