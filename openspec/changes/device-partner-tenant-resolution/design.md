# Design: Device Partner-to-Tenant Resolution

## Scope
Applies to device-facing request processing in:
- xconfwebconfig

Out of scope:
- xconfadmin SAT/Xerxes admin APIs
- Authentication and authorization policy

## Design Goals
- Keep device-facing tenant resolution deterministic.
- Preserve legacy behavior for devices missing partnerId.
- Use the tenant table as the single source of truth for known tenants; eliminate config-driven partner alias maps.
- Avoid duplicating tenant-resolution logic across xconfwebconfig device paths.

## Resolution Flow
1. Read partnerId from device request context.
2. Trim and uppercase partnerId.
3. If partnerId is missing/blank, return configured default tenant.
4. If partnerId is unknown or noaccount (case-insensitive), return configured default tenant.
5. Check tenant table (cached) for exact match (case-insensitive via uppercasing); if found, return that tenant ID.
6. Check tenant table for prefix match; if a tenant ID is a prefix of the uppercased partnerId, return the longest matching prefix.
7. If no match, return configured default tenant.
8. Ensure final tenantId is never blank.

## Decision Table
| Input partnerId | Tenant table contains | Output tenantId |
|---|---|---|
| missing/blank | n/a | default tenant |
| unknown | n/a | default tenant |
| noaccount | n/a | default tenant |
| partner1 | PARTNER1 | PARTNER1 |
| partner1-dev | PARTNER1 | PARTNER1 |
| partner1-dev-foo | PARTNER1, PARTNER1-DEV | PARTNER1-DEV |
| partner2 | PARTNER1 | default tenant |

## Tenant Table as Source of Truth
The `tenants` table in Cassandra is the authoritative registry of known canonical tenants.
Tenant IDs stored in this table SHOULD be uppercased. The resolver uppercases all IDs from the table before comparison to guard against inconsistent casing.

The config-driven `partner_tenant_mapping` is no longer used for device-facing resolution and has been removed.

## Caching
Tenant table results are cached using the application cache (same mechanism used for other per-tenant application data). The cache is keyed by `(defaultTenantId, TABLE_TENANTS, "TenantIDList")` and can be explicitly invalidated via `InvalidateTenantCache()`.

## Prefix Matching
Prefix matching allows partner variants (e.g. `partner1-dev`, `partner1-staging`) to resolve to a common tenant without requiring explicit entries for every variant. When multiple tenant IDs are prefix matches, the longest prefix wins to ensure specificity.

## Fallback Behavior
- Missing/blank partnerId uses existing default-tenant behavior.
- partnerId values unknown/noaccount (case-insensitive) use existing default-tenant behavior.
- Unmapped partnerId (no exact or prefix match in the tenant table) resolves to the configured default tenant.
- Resolution must never return blank.

## Compatibility
- Legacy devices not sending partnerId continue using default tenant.
- Devices are not required to send tenantId.
- Existing default-tenant fallback remains intact.
- No SAT/Auth changes.

## Reuse and Centralization
`ResolveTenantIdFromPartner` in `xconfwebconfig/http` is the single shared resolver used by xconfwebconfig device paths.
