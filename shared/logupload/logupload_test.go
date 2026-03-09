package logupload

import (
	"testing"

	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/stretchr/testify/assert"
)

// TestDCMGenericRule_Getters tests all getter methods
func TestDCMGenericRule_Getters(t *testing.T) {
	testStr := "test-value"
	rule := &DCMGenericRule{
		ID:          "test-id-123",
		Name:        "test-rule-name",
		Description: "test description",
		Rule: re.Rule{
			Condition: &re.Condition{
				FreeArg: &re.FreeArg{Name: "testParam"},
				FixedArg: &re.FixedArg{
					Bean: &re.Bean{Value: re.Value{JLString: &testStr}},
				},
				Operation: re.StandardOperationIs,
			},
		},
		Priority: 5,
	}

	// Test GetId
	assert.Equal(t, "test-id-123", rule.GetId())

	// Test GetName
	assert.Equal(t, "test-rule-name", rule.GetName())

	// Test GetRule
	gotRule := rule.GetRule()
	assert.NotNil(t, gotRule)
	assert.Equal(t, &rule.Rule, gotRule)

	// Test GetTemplateId (always returns empty string)
	assert.Equal(t, "", rule.GetTemplateId())

	// Test GetRuleType
	assert.Equal(t, "DCMGenericRule", rule.GetRuleType())
}

// TestDCMGenericRule_ToStringOnlyBaseProperties tests string conversion
func TestDCMGenericRule_ToStringOnlyBaseProperties(t *testing.T) {
	testValue := "XG1v4"
	// Test with simple condition
	rule := &DCMGenericRule{
		ID:   "test-id",
		Name: "test-rule",
		Rule: re.Rule{
			Condition: &re.Condition{
				FreeArg: &re.FreeArg{Name: "model"},
				FixedArg: &re.FixedArg{
					Bean: &re.Bean{Value: re.Value{JLString: &testValue}},
				},
				Operation: re.StandardOperationIs,
			},
		},
	}

	result := rule.ToStringOnlyBaseProperties()
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "model")
}

// TestDCMGenericRule_ToStringOnlyBaseProperties_Compound tests compound rules
func TestDCMGenericRule_ToStringOnlyBaseProperties_Compound(t *testing.T) {
	value1 := "value1"
	value2 := "value2"
	// Test with compound condition
	rule := &DCMGenericRule{
		ID:   "compound-test",
		Name: "compound-rule",
		Rule: re.Rule{
			CompoundParts: []re.Rule{
				{
					Condition: &re.Condition{
						FreeArg: &re.FreeArg{Name: "param1"},
						FixedArg: &re.FixedArg{
							Bean: &re.Bean{Value: re.Value{JLString: &value1}},
						},
						Operation: re.StandardOperationIs,
					},
				},
				{
					Condition: &re.Condition{
						FreeArg: &re.FreeArg{Name: "param2"},
						FixedArg: &re.FixedArg{
							Bean: &re.Bean{Value: re.Value{JLString: &value2}},
						},
						Operation: re.StandardOperationIs,
					},
				},
			},
		},
	}

	result := rule.ToStringOnlyBaseProperties()
	assert.NotEmpty(t, result)
}
