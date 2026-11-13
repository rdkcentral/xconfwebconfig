## Why

SettingsUtil currently writes the applied rule and device-attribute details into concatenated log messages. This makes the information harder to query consistently than the structured `EstbFirmwareService XCONF_LOG` event; a stable structured event will make settings decisions easier to search and correlate.

## What Changes

- Replace the SettingsUtil log-upload and VOD applied-rule messages with one stable structured completion event per settings request, including unsuccessful requests.
- Preserve the applied-rule identification and relevant device context as structured fields, using the existing filtered logging pattern.
- Keep the events at their current logging level and avoid changing settings selection or response behavior.

## Capabilities

### New Capabilities
- `settings-util-structured-logging`: Structured request-completion events for SettingsUtil, with applied log-upload and VOD rule details when available.

### Modified Capabilities

## Impact

- Affects `dataapi/log_uploader_handler.go`, `shared/logupload/utils.go`, and their logging tests.
- Changes log message format for the two SettingsUtil applied-rule events; log consumers may need to query the new event name and fields.
- No API, persistence, or dependency changes.