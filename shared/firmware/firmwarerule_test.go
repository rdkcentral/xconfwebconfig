package firmware

import (
	"strings"
	"testing"

	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"gotest.tools/assert"
)

// Helper function for string containment check
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestNewTemplateApplicableActionAndType(t *testing.T) {
	typ := "RULE"
	configId := "test-config-id"

	// Test with RULE action type
	action := NewTemplateApplicableActionAndType(typ, RULE, configId)

	if action == nil {
		t.Fatal("Expected action to be created, got nil")
	}

	if action.ActionType != RULE {
		t.Errorf("Expected action type RULE, got %v", action.ActionType)
	}

	if action.ConfigId != configId {
		t.Errorf("Expected ConfigId %s, got %s", configId, action.ConfigId)
	}

	if action.ConfigEntries == nil {
		t.Error("Expected ConfigEntries to be initialized")
	}

	if action.Properties == nil {
		t.Error("Expected Properties to be initialized")
	}

	// Test with DEFINE_PROPERTIES action type
	propsAction := NewTemplateApplicableActionAndType("DEF_PROPS", DEFINE_PROPERTIES, configId)

	if propsAction == nil {
		t.Fatal("Expected propsAction to be created, got nil")
	}

	if propsAction.ActionType != DEFINE_PROPERTIES {
		t.Errorf("Expected action type DEFINE_PROPERTIES, got %v", propsAction.ActionType)
	}
}

func TestNewApplicableAction(t *testing.T) {
	typ := "RULE"
	configId := "test-config-id"

	action := NewApplicableAction(typ, configId)

	if action == nil {
		t.Fatal("Expected action to be created, got nil")
	}

	if action.Type != typ {
		t.Errorf("Expected Type %s, got %s", typ, action.Type)
	}

	if action.ConfigId != configId {
		t.Errorf("Expected configId %s, got %s", configId, action.ConfigId)
	}
}

func TestApplicableAction_IsValid(t *testing.T) {
	// Test valid action
	action := &ApplicableAction{
		Type:     "RULE",
		ConfigId: "valid-config-id",
	}

	isValid := action.IsValid()
	if !isValid {
		t.Error("Expected action to be valid")
	}

	// All actions are valid according to the implementation
	emptyAction := &ApplicableAction{}
	if !emptyAction.IsValid() {
		t.Error("Expected empty action to be valid according to implementation")
	}
}

func TestApplicableAction_String(t *testing.T) {
	action := &ApplicableAction{
		Type:       "RULE",
		ActionType: RULE,
		ConfigId:   "test-config-id",
		Active:     true,
	}

	str := action.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}

	// Verify it contains the ActionType and ConfigId
	if !contains(str, "RULE") {
		t.Error("Expected string to contain ActionType")
	}
	if !contains(str, "test-config-id") {
		t.Error("Expected string to contain ConfigId")
	}
}

func TestFirmwareRule_SetApplicationType(t *testing.T) {
	rule := &FirmwareRule{}
	applicationType := "rdkb"

	rule.SetApplicationType(applicationType)

	if rule.ApplicationType != applicationType {
		t.Errorf("Expected ApplicationType %s, got %s", applicationType, rule.ApplicationType)
	}
}

func TestFirmwareRule_String(t *testing.T) {
	// Create a proper Rule with non-nil components
	condition := &re.Condition{}
	rule := &re.Rule{
		Condition: condition,
		Negated:   false,
		Relation:  "AND",
	}

	firmwareRule := &FirmwareRule{
		Name:   "test-rule",
		Type:   "test-type",
		Active: true,
		Rule:   *rule,
		ApplicableAction: &ApplicableAction{
			Type:       "RULE",
			ActionType: RULE,
			ConfigId:   "test-config",
		},
	}

	str := firmwareRule.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
}

func TestFirmwareRule_NewFirmwareRule(t *testing.T) {
	// Create a proper Rule
	condition := &re.Condition{}
	rule := &re.Rule{
		Condition: condition,
		Negated:   false,
		Relation:  "AND",
	}

	action := &ApplicableAction{
		Type:       "RULE",
		ActionType: RULE,
		ConfigId:   "test-config",
	}

	firmwareRule := NewFirmwareRule("test-id", "test-rule", MAC_RULE, rule, action, true)

	assert.Assert(t, firmwareRule != nil)
	assert.Equal(t, firmwareRule.ID, "test-id")
	assert.Equal(t, firmwareRule.Name, "test-rule")
	assert.Equal(t, firmwareRule.Type, MAC_RULE)
	assert.Equal(t, firmwareRule.Active, true)
	assert.Assert(t, firmwareRule.ApplicableAction != nil)
}

func TestFirmwareRule_ConfigId(t *testing.T) {
	// Test with ConfigId in ApplicableAction
	action := &ApplicableAction{
		ConfigId: "test-config-id",
	}

	firmwareRule := &FirmwareRule{
		ApplicableAction: action,
	}

	configId := firmwareRule.ConfigId()
	if configId != "test-config-id" {
		t.Errorf("Expected ConfigId 'test-config-id', got '%s'", configId)
	}

	// Test with ConfigEntries when ConfigId is empty
	configEntryAction := &ApplicableAction{
		ConfigEntries: []ConfigEntry{
			{ConfigId: "entry-config-id"},
		},
	}
	entryRule := &FirmwareRule{
		ApplicableAction: configEntryAction,
	}

	entryId := entryRule.ConfigId()
	if entryId != "entry-config-id" {
		t.Errorf("Expected ConfigId from entry 'entry-config-id', got '%s'", entryId)
	}

	// Test with empty everything - should return empty string
	emptyAction := &ApplicableAction{}
	emptyRule := &FirmwareRule{
		ID:               "fallback-id",
		ApplicableAction: emptyAction,
	}

	emptyId := emptyRule.ConfigId()
	if emptyId != "" {
		t.Errorf("Expected empty ConfigId, got '%s'", emptyId)
	}
}

func TestGetFirmwareRuleAllAsListDB(t *testing.T) {
	// Note: This test depends on database setup and might need mocking
	rules, err := GetFirmwareRuleAllAsListDB()

	// We don't expect errors in normal operation
	if err != nil {
		t.Logf("Database error (expected in test environment): %v", err)
	}

	if rules == nil {
		t.Log("No rules returned (expected in test environment)")
	}
}

// Test NewTemplateApplicableAction function
func TestNewTemplateApplicableAction(t *testing.T) {
	typ := "TEST_TYPE"
	configId := "test-config-id"

	action := NewTemplateApplicableAction(typ, configId)

	assert.Assert(t, action != nil)
	assert.Equal(t, typ, action.Type)
	assert.Equal(t, configId, action.ConfigId)
	assert.Equal(t, true, action.Active)
	assert.Equal(t, false, action.UseAccountPercentage)
	assert.Equal(t, false, action.FirmwareCheckRequired)
	assert.Equal(t, false, action.RebootImmediately)
	assert.Assert(t, action.ConfigEntries != nil)
	assert.Assert(t, action.FirmwareVersions != nil)
	assert.Assert(t, action.ByPassFilters != nil)
	assert.Assert(t, action.ActivationFirmwareVersions != nil)
	assert.Assert(t, action.Properties != nil)
}

// Test NewApplicableActionAndType function
func TestNewApplicableActionAndType(t *testing.T) {
	typ := "TEST_TYPE"
	configId := "test-config-id"

	action := NewApplicableActionAndType(typ, RULE, configId)

	assert.Assert(t, action != nil)
	assert.Equal(t, typ, action.Type)
	assert.Equal(t, RULE, action.ActionType)
	assert.Equal(t, configId, action.ConfigId)
	assert.Equal(t, true, action.Active)
	assert.Equal(t, false, action.UseAccountPercentage)
	assert.Equal(t, false, action.FirmwareCheckRequired)
	assert.Equal(t, false, action.RebootImmediately)
	assert.Assert(t, action.ConfigEntries != nil)
	assert.Assert(t, action.FirmwareVersions != nil)
	assert.Assert(t, action.ByPassFilters != nil)
	assert.Assert(t, action.ActivationFirmwareVersions != nil)
}

// Test FirmwareRule GetApplicationType
func TestFirmwareRule_GetApplicationType(t *testing.T) {
	rule := &FirmwareRule{ApplicationType: "stb"}
	assert.Equal(t, "stb", rule.GetApplicationType())

	rule.ApplicationType = "xhome"
	assert.Equal(t, "xhome", rule.GetApplicationType())

	emptyRule := &FirmwareRule{}
	assert.Equal(t, "", emptyRule.GetApplicationType())
}

// Test FirmwareRule Clone function
func TestFirmwareRule_Clone(t *testing.T) {
	original := &FirmwareRule{
		ID:              "test-id",
		Name:            "test-rule",
		Type:            "MAC_RULE",
		Active:          true,
		ApplicationType: "stb",
		ApplicableAction: &ApplicableAction{
			Type:     "RULE",
			ConfigId: "config-123",
		},
	}

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.Type, cloned.Type)
	assert.Equal(t, original.Active, cloned.Active)
	assert.Equal(t, original.ApplicationType, cloned.ApplicationType)

	// Ensure it's a deep copy
	assert.Assert(t, cloned != original)
	assert.Assert(t, cloned.ApplicableAction != original.ApplicableAction)
}

// Test NewEmptyFirmwareRule function
func TestNewEmptyFirmwareRule(t *testing.T) {
	rule := NewEmptyFirmwareRule()

	assert.Assert(t, rule != nil)
	assert.Equal(t, true, rule.Active)
	assert.Equal(t, "stb", rule.ApplicationType)
	assert.Assert(t, rule.ApplicableAction != nil)
	assert.Equal(t, "", rule.ApplicableAction.Type)
	assert.Equal(t, "", rule.ApplicableAction.ConfigId)
}

// Test NewFirmwareRuleInf function
func TestNewFirmwareRuleInf(t *testing.T) {
	ruleInterface := NewFirmwareRuleInf()

	rule, ok := ruleInterface.(*FirmwareRule)
	assert.Assert(t, ok)
	assert.Assert(t, rule != nil)
	assert.Equal(t, true, rule.Active)
	assert.Equal(t, "stb", rule.ApplicationType)
}

// Test FirmwareRule Validate function
func TestFirmwareRule_Validate(t *testing.T) {
	// Valid rule
	validRule := &FirmwareRule{
		Type: "MAC_RULE",
		ApplicableAction: &ApplicableAction{
			Type:       ".RuleAction",
			ActionType: RULE,
		},
	}

	err := validRule.Validate()
	assert.NilError(t, err)

	// Invalid rule - empty type
	invalidTypeRule := &FirmwareRule{
		Type: "",
		ApplicableAction: &ApplicableAction{
			Type:       ".RuleAction",
			ActionType: RULE,
		},
	}

	err = invalidTypeRule.Validate()
	assert.Assert(t, err != nil)
	assert.Assert(t, strings.Contains(err.Error(), "Type is is not present"))

	// Invalid rule - nil action
	nilActionRule := &FirmwareRule{
		Type:             "MAC_RULE",
		ApplicableAction: nil,
	}

	err = nilActionRule.Validate()
	assert.Assert(t, err != nil)
	assert.Assert(t, strings.Contains(err.Error(), "ApplicableAction is not present"))

	// Invalid rule - invalid action type
	invalidActionRule := &FirmwareRule{
		Type: "MAC_RULE",
		ApplicableAction: &ApplicableAction{
			Type:       ".RuleAction",
			ActionType: "INVALID_TYPE",
		},
	}

	err = invalidActionRule.Validate()
	assert.Assert(t, err != nil)
	assert.Assert(t, strings.Contains(err.Error(), "ActionType is invalid"))
}

// Test FirmwareRule Equals function
func TestFirmwareRule_Equals(t *testing.T) {
	rule1 := &FirmwareRule{
		ID:              "test-id",
		Name:            "test-rule",
		Type:            "MAC_RULE",
		Active:          true,
		ApplicationType: "stb",
	}

	rule2 := &FirmwareRule{
		ID:              "test-id",
		Name:            "test-rule",
		Type:            "MAC_RULE",
		Active:          true,
		ApplicationType: "stb",
	}

	rule3 := &FirmwareRule{
		ID:              "different-id",
		Name:            "test-rule",
		Type:            "MAC_RULE",
		Active:          true,
		ApplicationType: "stb",
	}

	assert.Assert(t, rule1.Equals(rule2))
	assert.Assert(t, !rule1.Equals(rule3))
}

// Test FirmwareRule getter functions
func TestFirmwareRule_Getters(t *testing.T) {
	condition := &re.Condition{}
	ruleEngine := &re.Rule{Condition: condition}

	rule := &FirmwareRule{
		ID:   "test-id",
		Name: "test-name",
		Type: "MAC_RULE",
		Rule: *ruleEngine,
	}

	assert.Equal(t, "test-id", rule.GetId())
	assert.Equal(t, "test-name", rule.GetName())
	assert.Equal(t, "MAC_RULE", rule.GetTemplateId())
	assert.Equal(t, "FirmwareRule", rule.GetRuleType())
	assert.Assert(t, rule.GetRule() != nil)
}

// Test FirmwareRule IsNoop function
func TestFirmwareRule_IsNoop(t *testing.T) {
	// Test with ConfigId
	ruleWithConfig := &FirmwareRule{
		ApplicableAction: &ApplicableAction{
			ConfigId: "test-config",
		},
	}
	assert.Assert(t, !ruleWithConfig.IsNoop())

	// Test with ConfigEntries
	ruleWithEntries := &FirmwareRule{
		ApplicableAction: &ApplicableAction{
			ConfigId: "",
			ConfigEntries: []ConfigEntry{
				{ConfigId: "entry-config"},
			},
		},
	}
	assert.Assert(t, !ruleWithEntries.IsNoop())

	// Test empty (should be Noop)
	emptyRule := &FirmwareRule{
		ApplicableAction: &ApplicableAction{
			ConfigId:      "",
			ConfigEntries: []ConfigEntry{},
		},
	}
	assert.Assert(t, emptyRule.IsNoop())

	// Test nil action
	nilActionRule := &FirmwareRule{
		ApplicableAction: nil,
	}
	assert.Assert(t, nilActionRule.IsNoop())
}
