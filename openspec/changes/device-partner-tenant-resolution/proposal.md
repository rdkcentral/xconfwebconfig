# Proposal: Device Partner-to-Tenant Resolution

## Summary
Introduce partnerId-based tenant resolution for device-facing requests in xconfwebconfig.

Today, device-facing paths generally resolve tenantId to the configured default tenant when tenantId is not provided. Devices do not use SAT, and many devices (especially legacy) do not send tenantId. Devices commonly send partnerId. This change defines deterministic tenant resolution using partnerId with configurable partner-to-tenant mapping.

## Why This Is Needed
- Device-facing services need tenant resolution for correct data lookup and rule evaluation.
- Devices are not expected to participate in admin-side auth tenant mechanisms (SAT/Xerxes).
- Requiring tenantId from devices is not backward compatible.
- partnerId is broadly available in device requests and can drive tenant resolution.

## Proposed Behavior
- Primary input for device-facing tenant resolution is partnerId.
- tenantId header is optional and not required from devices.
- SAT is not involved in this resolution path.
- Tenant is resolved using configurable partner-to-tenant alias mapping with deterministic fallback.

High-level resolution:
1. Missing/blank partnerId -> default tenant.
2. partnerId present -> lookup in mapping.
3. If mapped alias -> mapped tenant.
4. If not mapped -> tenantId = partnerId.
5. Never return blank tenant.
6. Case-insensitive matching is preferred.

## Impact
- Device-facing request routing/evaluation in xconfwebconfig becomes multi-tenant aware by partnerId.
- Legacy device behavior is preserved via default-tenant fallback.
- No change to xconfadmin behavior.
- No SAT/Auth behavior changes.

## Non-Goals
- No changes to xconfadmin SAT/Xerxes admin APIs.
- No authentication/authorization design changes.
- No requirement for devices to send tenantId.
- No hardcoded partner lists in code; mapping remains externally configurable.

## Risks and Mitigations
- Risk: inconsistent behavior across xconfwebconfig device-facing paths.
  - Mitigation: centralize resolver helper and reuse in device paths where practical.
- Risk: ambiguous partner casing.
  - Mitigation: prefer case-insensitive matching and cover with tests.
- Risk: missing mapping entries.
  - Mitigation: deterministic fallback to partnerId itself and default tenant for missing partnerId.
