package http

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/util"
)

const (
	tenantCacheKey = "TenantIDList"
	tenantCacheTTL = 5 * time.Minute // Lightweight in-process cache TTL
)

// tenantIdsCacheEntry holds cached tenant IDs with a timestamp for TTL validation
type tenantIdsCacheEntry struct {
	ids       []string
	timestamp int64 // Unix nanoseconds
}

// tenantIdsCache is a lightweight in-process cache that works even when optional application cache is disabled.
// This prevents per-request DB queries on device-facing request paths.
var tenantIdsCache atomic.Value // stores *tenantIdsCacheEntry or nil

// getCachedTenantIds returns all tenant IDs from cache or DB.
// Uses a lightweight in-process cache to avoid per-request DB queries when application cache is disabled.
// Declared as a variable so tests can substitute a lightweight stub.
var getCachedTenantIds = func() []string {
	dbClient := db.GetDatabaseClient()
	if dbClient == nil {
		return []string{}
	}

	// Check lightweight in-process cache first (works even if optional application cache is disabled)
	if entry, ok := tenantIdsCache.Load().(*tenantIdsCacheEntry); ok {
		if time.Now().UnixNano()-entry.timestamp < int64(tenantCacheTTL) {
			return entry.ids
		}
	}

	// Cache miss or expired; query database
	tenants := dbClient.GetAllTenants()
	tenantIds := make([]string, 0, len(tenants))
	for _, t := range tenants {
		if t == nil || util.IsBlank(t.ID) {
			continue
		}
		tenantIds = append(tenantIds, strings.ToUpper(t.ID))
	}

	// Store in lightweight cache
	tenantIdsCache.Store(&tenantIdsCacheEntry{
		ids:       tenantIds,
		timestamp: time.Now().UnixNano(),
	})

	// Also store in optional application cache if enabled
	cm := db.GetCacheManager()
	defaultTenantId := db.GetDefaultTenantId()
	cm.ApplicationCacheSet(defaultTenantId, db.TABLE_TENANTS, tenantCacheKey, tenantIds)

	return tenantIds
}

// ResolveTenantIdFromPartner resolves a partnerId to a tenantId using the tenant table.
//
// Resolution order:
//  1. Blank, unknown, or noaccount partnerId → default tenant.
//  2. Exact match (case-insensitive) against a tenant ID in the tenant table → that tenant.
//  3. A tenant ID in the table is a prefix of the partnerId → longest matching prefix wins.
//  4. No match → default tenant.
func ResolveTenantIdFromPartner(partnerId string) string {
	tenantIds := getCachedTenantIds()
	return resolveTenantIdFromPartnerWithIds(partnerId, tenantIds)
}

// resolveTenantIdFromPartnerWithIds resolves a partnerId to a tenantId using the provided tenant IDs.
// This is a pure helper function used for both production code and unit testing.
// It accepts tenantIds as a parameter to avoid global state mutations in tests.
func resolveTenantIdFromPartnerWithIds(partnerId string, tenantIds []string) string {
	defaultTenantId := db.GetDefaultTenantId()
	trimmedPartnerId := strings.TrimSpace(partnerId)
	if util.IsUnknownValue(trimmedPartnerId) || trimmedPartnerId == "" {
		return strings.ToUpper(defaultTenantId)
	}

	normalizedPartnerId := strings.ToUpper(trimmedPartnerId)

	// Exact match
	for _, tid := range tenantIds {
		if tid == normalizedPartnerId {
			return tid
		}
	}

	// Longest-prefix match
	bestMatch := ""
	for _, tid := range tenantIds {
		if strings.HasPrefix(normalizedPartnerId, tid) && len(tid) > len(bestMatch) {
			bestMatch = tid
		}
	}
	if bestMatch != "" {
		return bestMatch
	}

	return strings.ToUpper(defaultTenantId)
}

// InvalidateTenantCache clears both the lightweight in-process cache and the optional application cache,
// forcing a fresh DB lookup on the next call to ResolveTenantIdFromPartner.
// Intended for use in integration tests after inserting or deleting tenants.
func InvalidateTenantCache() {
	// Clear lightweight in-process cache
	tenantIdsCache.Store((*tenantIdsCacheEntry)(nil))

	// Also clear optional application cache if enabled
	cm := db.GetCacheManager()
	defaultTenantId := db.GetDefaultTenantId()
	cm.ApplicationCacheDelete(defaultTenantId, db.TABLE_TENANTS, tenantCacheKey)
}
