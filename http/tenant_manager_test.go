package http

import (
	"testing"

	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/stretchr/testify/assert"
)

func TestResolveTenantIdFromPartner_BlankPartnerUsesDefault(t *testing.T) {
	resolved := resolveTenantIdFromPartnerWithIds("", []string{})
	assert.Equal(t, db.GetDefaultTenantId(), resolved)
}

func TestResolveTenantIdFromPartner_UnknownOrNoaccountUsesDefault(t *testing.T) {
	testCases := []string{"unknown", "Unknown", "noaccount", "NoAccount"}
	for _, partnerId := range testCases {
		resolved := resolveTenantIdFromPartnerWithIds(partnerId, []string{})
		assert.Equal(t, db.GetDefaultTenantId(), resolved)
	}
}

func TestResolveTenantIdFromPartner_ExactMatchCaseInsensitive(t *testing.T) {
	testCases := []struct {
		partnerId        string
		expectedTenantId string
	}{
		{"partner1", "PARTNER1"},
		{"PARTNER1", "PARTNER1"},
		{"test1", "TEST1"},
		{"Test2", "TEST2"},
	}
	tenantIds := []string{"PARTNER1", "TEST1", "TEST2"}
	for _, tc := range testCases {
		resolved := resolveTenantIdFromPartnerWithIds(tc.partnerId, tenantIds)
		assert.Equal(t, tc.expectedTenantId, resolved)
	}
}

func TestResolveTenantIdFromPartner_PrefixMatch(t *testing.T) {
	resolved := resolveTenantIdFromPartnerWithIds("partner-dev", []string{"PARTNER"})
	assert.Equal(t, "PARTNER", resolved)
}

func TestResolveTenantIdFromPartner_LongestPrefixWins(t *testing.T) {
	resolved := resolveTenantIdFromPartnerWithIds("partner-dev-foo", []string{"PARTNER", "PARTNER-DEV"})
	assert.Equal(t, "PARTNER-DEV", resolved)
}

func TestResolveTenantIdFromPartner_UnmappedReturnsDefault(t *testing.T) {
	resolved := resolveTenantIdFromPartnerWithIds("other", []string{"PARTNER", "TEST"})
	assert.Equal(t, db.GetDefaultTenantId(), resolved)
}

func TestResolveTenantIdFromPartner_NeverReturnsEmpty(t *testing.T) {
	resolved := resolveTenantIdFromPartnerWithIds("   ", []string{})
	assert.NotEmpty(t, resolved)
}

func TestResolveTenantIdFromPartner_WhitespaceIsTrimmed(t *testing.T) {
	resolved := resolveTenantIdFromPartnerWithIds("  partner1  ", []string{"PARTNER1"})
	assert.Equal(t, "PARTNER1", resolved)
}
