package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfilesSnapshotApplyAndDelete(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{
			{"path": "Width", "type": "float64", "value": "0.9"},
			{"path": "Leaves", "type": "int", "value": "1"},
		},
	}, &polyformmcp.CreateVariablesOutput{})

	value := func(path string) any {
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

	var saved polyformmcp.SaveProfileOutput
	callTool(t, session, "save_profile", map[string]any{"name": "Single"}, &saved)
	callTool(t, session, "save_profile", map[string]any{
		"name":      "Double",
		"variables": map[string]any{"Width": "1.8", "Leaves": "2"},
	}, &saved)
	assert.Equal(t, []string{"Double", "Single"}, saved.Profiles)
	assert.EqualValues(t, 0.9, value("Width"), "a snapshot with overrides leaves the live values alone")

	callTool(t, session, "apply_profile", map[string]any{"name": "Double"}, &polyformmcp.ApplyProfileOutput{})
	assert.EqualValues(t, 1.8, value("Width"))
	assert.EqualValues(t, 2, value("Leaves"))

	callTool(t, session, "apply_profile", map[string]any{"name": "Single"}, &polyformmcp.ApplyProfileOutput{})
	assert.EqualValues(t, 0.9, value("Width"))

	msg := callToolExpectingError(t, session, "apply_profile", map[string]any{"name": "Triple"})
	assert.Contains(t, msg, "Double, Single")

	var deleted polyformmcp.DeleteProfileOutput
	callTool(t, session, "delete_profile", map[string]any{"name": "Double"}, &deleted)
	require.Equal(t, []string{"Single"}, deleted.Profiles)
}

// A variable added after a profile was saved is not in it, so applying that
// profile leaves the variable wherever the last look put it. Silent drift
// is the failure; naming it is the fix.
func TestApplyProfileNamesTheVariablesItDoesNotCover(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Width", "type": "float64", "value": "0.9"}},
	}, &polyformmcp.CreateVariablesOutput{})
	callTool(t, session, "save_profile", map[string]any{"name": "Old"}, &polyformmcp.SaveProfileOutput{})

	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Finish", "type": "int", "value": "0"}},
	}, &polyformmcp.CreateVariablesOutput{})
	callTool(t, session, "save_profile", map[string]any{"name": "New"}, &polyformmcp.SaveProfileOutput{})

	var old polyformmcp.ApplyProfileOutput
	callTool(t, session, "apply_profile", map[string]any{"name": "Old"}, &old)
	assert.Equal(t, []string{"Finish"}, old.Unset)
	assert.Contains(t, old.Note, "save_profile")

	var recent polyformmcp.ApplyProfileOutput
	callTool(t, session, "apply_profile", map[string]any{"name": "New"}, &recent)
	assert.Empty(t, recent.Unset, "a profile saved with every variable covers them all")
	assert.Empty(t, recent.Note)
}
