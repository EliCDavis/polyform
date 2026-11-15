package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func variableNamed(t *testing.T, session sessionT, path string) polyformmcp.VariableSummary {
	t.Helper()
	var list polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &list)
	for _, v := range list.Variables {
		if v.Path == path {
			return v
		}
	}
	require.Fail(t, "no variable", path)
	return polyformmcp.VariableSummary{}
}

func TestUpdateVariableTypeKeepsReferencesProfilesAndVariantSets(t *testing.T) {
	session, placed := nestedVariableReader(t)
	before := variableNamed(t, session, "Scale")
	callTool(t, session, "update_variable", map[string]any{"path": "Scale", "description": "How big"}, &polyformmcp.UpdateVariableOutput{})
	callTool(t, session, "save_profile", map[string]any{"name": "Big", "variables": map[string]any{"Scale": "9"}}, &polyformmcp.SaveProfileOutput{})
	callTool(t, session, "create_variant_set", map[string]any{
		"name":       "Scales",
		"dimensions": []map[string]any{{"path": "Scale", "type": "discrete", "data": `{"values":[1,4]}`}},
	}, &polyformmcp.CreateVariantSetOutput{})

	var updated polyformmcp.UpdateVariableOutput
	callTool(t, session, "update_variable", map[string]any{"path": "Scale", "type": "int"}, &updated)
	assert.True(t, updated.Updated)

	after := variableNamed(t, session, "Scale")
	assert.Equal(t, "int", after.Type)
	assert.Equal(t, "How big", after.Description)
	assert.Equal(t, before.References, after.References)
	assert.InDelta(t, 6.0, evaluateFloat(t, session, placed, "Out"), 1e-9)

	callTool(t, session, "update_variable", map[string]any{"path": "Scale", "value": "5"}, &polyformmcp.UpdateVariableOutput{})
	assert.InDelta(t, 15.0, evaluateFloat(t, session, placed, "Out"), 1e-9)

	var profile polyformmcp.GetProfileOutput
	callTool(t, session, "get_profile", map[string]any{"name": "Big"}, &profile)
	assert.Equal(t, map[string]string{"Scale": "9"}, profile.Variables)
	assert.Empty(t, profile.Stale)
	callTool(t, session, "apply_profile", map[string]any{"name": "Big"}, &polyformmcp.ApplyProfileOutput{})
	assert.InDelta(t, 27.0, evaluateFloat(t, session, placed, "Out"), 1e-9)

	var sets polyformmcp.ListVariantSetsOutput
	callTool(t, session, "list_variant_sets", map[string]any{}, &sets)
	require.Len(t, sets.VariantSets, 1)

	path := t.TempDir() + "/graph.json"
	callTool(t, session, "save_graph", map[string]any{"path": path}, &polyformmcp.SaveGraphOutput{})
	callTool(t, session, "load_graph", map[string]any{"path": path}, &polyformmcp.LoadGraphOutput{})
	assert.Equal(t, "int", variableNamed(t, session, "Scale").Type)
}

func TestUpdateVariableTypeRefusesATypeAReaderCannotTake(t *testing.T) {
	session, placed := nestedVariableReader(t)

	message := callToolExpectingError(t, session, "update_variable", map[string]any{"path": "Scale", "type": "string", "value": `"wide"`})

	assert.Contains(t, message, "inner")
	assert.Contains(t, message, "Values")
	assert.Equal(t, "float64", variableNamed(t, session, "Scale").Type)
	assert.InDelta(t, 6.0, evaluateFloat(t, session, placed, "Out"), 1e-9)
}

func TestUpdateVariablesLeavesARefusedRetypeWired(t *testing.T) {
	session, placed := nestedVariableReader(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Count", "type": "int", "value": "2"}},
	}, &polyformmcp.CreateVariablesOutput{})

	var out polyformmcp.UpdateVariablesOutput
	callTool(t, session, "update_variables", map[string]any{"variables": []map[string]any{
		{"path": "Scale", "type": "string", "value": `"wide"`},
		{"path": "Count", "value": "7"},
	}}, &out)

	assert.Equal(t, 1, out.Failed)
	assert.Equal(t, "float64", variableNamed(t, session, "Scale").Type)
	assert.InDelta(t, 6.0, evaluateFloat(t, session, placed, "Out"), 1e-9)
	callTool(t, session, "update_variable", map[string]any{"path": "Scale", "value": "5"}, &polyformmcp.UpdateVariableOutput{})
	assert.InDelta(t, 15.0, evaluateFloat(t, session, placed, "Out"), 1e-9)
}

func TestUpdateVariableTypeRefusesWhenTheValueDoesNotFit(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Width", "type": "float64", "value": "2.5"}},
	}, &polyformmcp.CreateVariablesOutput{})

	message := callToolExpectingError(t, session, "update_variable", map[string]any{"path": "Width", "type": "int"})

	assert.Contains(t, message, "2.5")
	assert.Equal(t, "float64", variableNamed(t, session, "Width").Type)

	callTool(t, session, "update_variable", map[string]any{"path": "Width", "type": "int", "value": "3"}, &polyformmcp.UpdateVariableOutput{})
	width := variableNamed(t, session, "Width")
	assert.Equal(t, "int", width.Type)
	assert.EqualValues(t, 3, width.Value)
}

func TestUpdateVariableTypeRefusesWhenAProfileValueDoesNotFit(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Width", "type": "float64", "value": "2"}},
	}, &polyformmcp.CreateVariablesOutput{})
	callTool(t, session, "save_profile", map[string]any{"name": "Odd", "variables": map[string]any{"Width": "0.5"}}, &polyformmcp.SaveProfileOutput{})

	message := callToolExpectingError(t, session, "update_variable", map[string]any{"path": "Width", "type": "int"})

	assert.Contains(t, message, "Odd")
	assert.Equal(t, "float64", variableNamed(t, session, "Width").Type)
}

func TestUpdateVariableTypeRefusesWhenAVariantSetValueDoesNotFit(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Width", "type": "float64", "value": "2"}},
	}, &polyformmcp.CreateVariablesOutput{})
	callTool(t, session, "create_variant_set", map[string]any{
		"name":       "Widths",
		"dimensions": []map[string]any{{"path": "Width", "type": "numericRange", "data": `{"min":0,"max":1,"samples":3}`}},
	}, &polyformmcp.CreateVariantSetOutput{})

	message := callToolExpectingError(t, session, "update_variable", map[string]any{"path": "Width", "type": "int"})

	assert.Contains(t, message, "Widths")
	assert.Equal(t, "float64", variableNamed(t, session, "Width").Type)
}

func TestUpdateVariableTypeRefusesAValueOutsideANormRange(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Wear", "type": "float64", "value": "2.5"}},
	}, &polyformmcp.CreateVariablesOutput{})

	message := callToolExpectingError(t, session, "update_variable", map[string]any{"path": "Wear", "type": "norm.UFloat64"})

	assert.Contains(t, message, "2.5")
	assert.Equal(t, "float64", variableNamed(t, session, "Wear").Type)
}

func TestUpdateVariableTypeRefusesAProfileValueOutsideANormRange(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Lean", "type": "float64", "value": "0.25"}},
	}, &polyformmcp.CreateVariablesOutput{})
	callTool(t, session, "save_profile", map[string]any{"name": "Toppled", "variables": map[string]any{"Lean": "-3"}}, &polyformmcp.SaveProfileOutput{})

	message := callToolExpectingError(t, session, "update_variable", map[string]any{"path": "Lean", "type": "norm.SFloat64"})

	assert.Contains(t, message, "Toppled")
	assert.Equal(t, "float64", variableNamed(t, session, "Lean").Type)
}

func TestUpdateVariableTypeTakesAValueInsideANormRange(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Wear", "type": "float64", "value": "0.35"}},
	}, &polyformmcp.CreateVariablesOutput{})

	callTool(t, session, "update_variable", map[string]any{"path": "Wear", "type": "norm.UFloat64"}, &polyformmcp.UpdateVariableOutput{})

	assert.Equal(t, "norm.UFloat64", variableNamed(t, session, "Wear").Type)
}
