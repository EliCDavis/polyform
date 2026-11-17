package mcp_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/generator/variable"
	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func variableValue(t *testing.T, session sessionT, path string) any {
	t.Helper()
	var list polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &list)
	for _, v := range list.Variables {
		if v.Path == path {
			return v.Value
		}
	}
	t.Fatalf("no variable %q", path)
	return nil
}

func TestGetProfileReadsValuesWithoutApplyingThem(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Width", "type": "float64", "value": "0.9"}},
	}, &polyformmcp.CreateVariablesOutput{})
	callTool(t, session, "save_profile", map[string]any{"name": "Wide", "variables": map[string]any{"Width": "1.8"}}, &polyformmcp.SaveProfileOutput{})
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Finish", "type": "int", "value": "0"}},
	}, &polyformmcp.CreateVariablesOutput{})

	var got polyformmcp.GetProfileOutput
	callTool(t, session, "get_profile", map[string]any{"name": "Wide"}, &got)
	assert.Equal(t, map[string]string{"Width": "1.8"}, got.Variables)
	assert.Equal(t, []string{"Finish"}, got.Missing)
	assert.EqualValues(t, 0.9, variableValue(t, session, "Width"), "reading leaves the live value alone")

	msg := callToolExpectingError(t, session, "get_profile", map[string]any{"name": "Narrow"})
	assert.Contains(t, msg, "Wide")
}

func TestListProfilesReportsMissingAndStaleKeys(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{
			{"path": "Width", "type": "float64", "value": "0.9"},
			{"path": "Finish", "type": "int", "value": "0"},
		},
	}, &polyformmcp.CreateVariablesOutput{})
	callTool(t, session, "save_profile", map[string]any{"name": "Complete"}, &polyformmcp.SaveProfileOutput{})
	inst.Profiles().Set("Old", variable.Profile{"Width": json.RawMessage("2"), "Gone": json.RawMessage("1")})

	var list polyformmcp.ListProfilesOutput
	callTool(t, session, "list_profiles", map[string]any{}, &list)
	require.Len(t, list.Profiles, 2)
	assert.Equal(t, polyformmcp.ProfileSummary{Name: "Complete", Pinned: 2}, list.Profiles[0])
	assert.Equal(t, polyformmcp.ProfileSummary{Name: "Old", Pinned: 1, Missing: []string{"Finish"}, Stale: []string{"Gone"}}, list.Profiles[1])
	assert.Contains(t, list.Note, "1 profile(s)")
}

func TestApplyProfileSkipsKeysForVariablesThatNoLongerExist(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Width", "type": "float64", "value": "0.9"}},
	}, &polyformmcp.CreateVariablesOutput{})
	inst.Profiles().Set("Old", variable.Profile{"Width": json.RawMessage("2"), "Gone": json.RawMessage("1")})

	var applied polyformmcp.ApplyProfileOutput
	callTool(t, session, "apply_profile", map[string]any{"name": "Old"}, &applied)
	assert.Equal(t, []string{"Gone"}, applied.Skipped)
	assert.Empty(t, applied.Unset)
	assert.Contains(t, applied.Note, "no longer exist")
	assert.EqualValues(t, 2, variableValue(t, session, "Width"))
}

func TestUpdateProfileMergesValuesWithoutTouchingTheLiveGraph(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Width", "type": "float64", "value": "0.9"}},
	}, &polyformmcp.CreateVariablesOutput{})
	callTool(t, session, "save_profile", map[string]any{"name": "Wide", "variables": map[string]any{"Width": "1.8"}}, &polyformmcp.SaveProfileOutput{})
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Finish", "type": "int", "value": "0"}},
	}, &polyformmcp.CreateVariablesOutput{})

	var updated polyformmcp.UpdateProfileOutput
	callTool(t, session, "update_profile", map[string]any{"name": "Wide", "variables": map[string]any{"Finish": `"2"`}}, &updated)
	assert.Equal(t, map[string]string{"Width": "1.8", "Finish": "2"}, updated.Variables, "a quoted number is stored as the variable encodes it")
	assert.Empty(t, updated.Missing)
	assert.EqualValues(t, 0, variableValue(t, session, "Finish"))
	assert.EqualValues(t, 0.9, variableValue(t, session, "Width"))

	callTool(t, session, "apply_profile", map[string]any{"name": "Wide"}, &polyformmcp.ApplyProfileOutput{})
	assert.EqualValues(t, 2, variableValue(t, session, "Finish"))

	msg := callToolExpectingError(t, session, "update_profile", map[string]any{"name": "Wide", "variables": map[string]any{"Nope": "1"}})
	assert.Contains(t, msg, "Nope")
	msg = callToolExpectingError(t, session, "update_profile", map[string]any{"name": "Wide", "variables": map[string]any{"Width": `"red"`}})
	assert.Contains(t, msg, "Width")
}
