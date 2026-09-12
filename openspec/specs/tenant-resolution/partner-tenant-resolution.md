# Spec: Partner-to-Tenant Resolution for Device-Facing Requests

## Status
Accepted

## Context
This specification defines tenantId resolution for device-facing requests handled by xconfwebconfig.

xconfadmin tenant behavior remains separate and auth-based; this spec does not apply to xconfadmin SAT/Xerxes admin APIs.

## Definitions
- partnerId: Device-provided partner identifier in request context.
- tenantId: Effective tenant identifier used for data access/rule evaluation.
- default tenant: Configured fallback tenant used when tenant cannot otherwise be determined.
- tenant table: The `tenants` table in Cassandra, containing the set of known canonical tenant IDs.

## Requirements
1. Device-facing tenant resolution MUST use partnerId as the primary input.
2. Device-facing services MUST NOT require tenantId from devices.
3. SAT MUST NOT be part of device-facing tenant resolution for this spec.
4. If partnerId is missing, blank, unknown, or noaccount (case-insensitive), resolver MUST return configured default tenant.
5. If partnerId (uppercased) exactly matches a tenant ID in the tenant table, resolver MUST return that tenant ID.
6. If a tenant ID in the table is a prefix of the uppercased partnerId, resolver MUST return the matching tenant ID. When multiple tenant IDs are prefix matches, the longest match MUST win.
7. If no exact or prefix match is found, resolver MUST return the configured default tenant.
8. Tenant IDs from the table MUST be uppercased before comparison; partnerId input MUST be trimmed and uppercased for matching.
9. Resolver MUST be deterministic for the same input and tenant table state.
10. Resolver MUST never return blank tenantId.
11. The tenant table is the authoritative source for known tenants; no additional config-driven partner-to-tenant mapping is required.
12. The behavior defined by this spec MUST apply consistently across xconfwebconfig device-facing paths.

## Resolution Flow
1. Trim and uppercase partnerId.
2. If blank, unknown, or noaccount → return default tenant.
3. Check tenant table (cached) for exact match → return match.
4. Check tenant table for prefix match → return longest matching prefix.
5. No match → return default tenant.

## Behavior Examples
- partnerId missing → default tenant
- partnerId=unknown → default tenant
- partnerId=noaccount → default tenant
- partnerId=partner1 → PARTNER1 (exact match in tenant table)
- partnerId=partner1-dev → PARTNER1 (prefix match; "PARTNER1" is a prefix of "PARTNER1-DEV")
- partnerId=partner1-dev-foo (with both PARTNER1 and PARTNER1-DEV in table) → PARTNER1-DEV (longest prefix wins)
- partnerId=partner2 (not in tenant table, no prefix match) → default tenant

## Caching
The tenant table lookup result is cached per process using the application cache. The cache is keyed on the default tenant and the `tenants` table name. The cache may be invalidated explicitly when the tenant table changes.

## Compatibility and Non-Goals
- No requirement for devices to send tenantId.
- No SAT/Auth changes.
- No xconfadmin behavior changes.
- This spec resolves tenant context only and does not imply partner authorization.
- Config-driven partner-to-tenant alias mapping is no longer used for device-facing resolution.
