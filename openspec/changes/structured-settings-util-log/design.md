## Context

`logupload.CopySettings` emits separate info logs when log-upload settings and VOD settings are applied. Both messages concatenate device-attribute and rule-condition text, while the `EstbFirmwareService XCONF_LOG` convention uses a stable event name and structured fields. `CopySettings` already receives the device context and audit fields needed to produce searchable events.

## Goals / Non-Goals

**Goals:**
- Emit exactly one stable `SettingsUtil XCONF_LOG` event per settings request, including unsuccessful requests and the `checkNow` and T2 settings paths.
- Preserve the applied rule identity and condition, plus essential available device context, as structured fields.
- Continue filtering fields through `common.FilterLogFields` and keep existing settings behavior unchanged.

**Non-Goals:**
- Change rule evaluation, settings selection, or device responses.
- Redesign logging for other services or change the shared log-field filtering policy.
- Retain the old free-form message format as a compatibility contract.

## Decisions

- Defer the event in `GetLogUploaderSettings` so normal completion and early errors emit once, using the response writer's final HTTP `status`. A missing audit writer or unwritten response is recorded as status 500; panics are not recovered or suppressed.
- Record rule ID, rule name, and the base rule condition in `logUploadRule` and `vodRule` objects on the existing audit fields when settings are applied. Do not emit separate applied-rule events.
- Populate the completion event from a copy of the audit fields and selected final device context (MAC, IP, model, firmware, environment, application, tenant, partner, and `checkNow`); pass the result through `common.FilterLogFields`. Avoid logging arbitrary context values.
- Keep info-level severity, rule evaluation, rollout, and response behavior unchanged. No-match requests still emit the event with status 404 and without applied-rule objects.

## Risks / Trade-offs

- [Log consumers may depend on the old text] -> Use the stable event name and document the new fields; update queries that parse the old message.
- [Selected context fields may be incomplete for a consumer] -> Preserve caller-provided audit fields and cover the required rule and device identifiers in tests.