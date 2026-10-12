# Tasks: Device Partner-to-Tenant Resolution

## Documentation
- [X] Add tenant-resolution OpenSpec docs for partnerId-based device resolution.
- [X] Update design and spec to reflect tenant-table-based resolution (replaces config-driven alias map).

## Discovery
- [X] Identify current default-tenant resolution points in xconfwebconfig device-facing request paths.

## Implementation
- [X] Implement a centralized partnerId -> tenantId resolver reusable by xconfwebconfig where practical.
- [X] Replace config-driven partner-to-tenant mapping with tenant table lookup (exact + longest-prefix match).
- [ ] Wire resolver into device request paths in xconfwebconfig.
- [ ] Complete remaining xconfwebconfig API migration to partner-based tenant mapping (/info/refreshAll).
- [ ] Complete remaining xconfwebconfig API migration to partner-based tenant mapping (/info/refresh/{tableName}).
- [ ] Complete remaining xconfwebconfig API migration to partner-based tenant mapping (/info/statistics).
- [ ] Complete remaining xconfwebconfig API migration to partner-based tenant mapping (/estbfirmware/changelogs).
- [ ] Complete remaining xconfwebconfig API migration to partner-based tenant mapping (/estbfirmware/lastlog).

## Required Behavior Validation
- [X] Ensure missing/blank partnerId resolves to configured default tenant.
- [X] Ensure exact-match partner resolves to matching tenant ID from tenant table.
- [X] Ensure prefix-match partner resolves to matching tenant ID (longest prefix wins).
- [X] Ensure unmapped partnerId resolves to configured default tenant.
- [X] Ensure tenant resolution is deterministic and never returns blank.
- [X] Ensure matching behavior is case-insensitive (uppercase normalization applied to both partnerId and tenant IDs).

## Testing
- [X] Add unit tests for blank, unknown/noaccount, exact match, prefix match, longest-prefix-wins, unmapped fallback.
- [ ] Add integration/handler tests in xconfwebconfig where feasible.
- [X] Verify legacy device compatibility (no partnerId -> default tenant fallback preserved).

## Out of Scope Guardrails
- [X] Confirm no xconfadmin SAT/Xerxes behavior changes are introduced.
- [X] Confirm no authN/authZ behavior is introduced by this change.
