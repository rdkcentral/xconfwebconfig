package rulesengine

import (
	"testing"

	"gotest.tools/assert"
)

func TestNewParser(t *testing.T) {
	// Test creating a new parser
	t.Run("Create_parser_with_string", func(t *testing.T) {
		parser := NewParser("test string")
		assert.Assert(t, parser != nil, "Parser should not be nil")
		assert.Equal(t, "test string", parser.str)
		assert.Assert(t, parser.rule != nil, "Parser rule should not be nil")
	})

	// Test with empty string
	t.Run("Create_parser_with_empty_string", func(t *testing.T) {
		parser := NewParser("")
		assert.Assert(t, parser != nil, "Parser should not be nil")
		assert.Equal(t, "", parser.str)
		assert.Assert(t, parser.rule != nil, "Parser rule should not be nil")
	})

	// Test with complex string
	t.Run("Create_parser_with_complex_string", func(t *testing.T) {
		complexStr := "model IS TG1682G AND env IN (QA,PROD)"
		parser := NewParser(complexStr)
		assert.Assert(t, parser != nil, "Parser should not be nil")
		assert.Equal(t, complexStr, parser.str)
		assert.Assert(t, parser.rule != nil, "Parser rule should not be nil")
	})
}

func TestParse(t *testing.T) {
	// Test parse function returns empty rule
	t.Run("Parse_returns_empty_rule", func(t *testing.T) {
		rule := parse()
		assert.Assert(t, rule != nil, "Parse should return a rule")
		
		// Verify it's an empty rule
		assert.Assert(t, rule.Condition == nil, "Condition should be nil")
		assert.Assert(t, len(rule.CompoundParts) == 0, "CompoundParts should be empty")
		assert.Equal(t, "", rule.Xxid)
		assert.Equal(t, false, rule.Negated)
	})
}
