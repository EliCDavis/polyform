package mcp_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/EliCDavis/polyform/generator/variable"
	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileToolsKeepFileVariablesAsReferences(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Splat", "type": "file"}},
	}, &polyformmcp.CreateVariablesOutput{})
	v, err := inst.GetVariable("Splat")
	require.NoError(t, err)
	file := v.(*variable.FileVariable)
	content := bytes.Repeat([]byte("ply"), 200)
	file.SetValue(content)

	callTool(t, session, "save_profile", map[string]any{"name": "Default"}, &polyformmcp.SaveProfileOutput{})

	var got polyformmcp.GetProfileOutput
	callTool(t, session, "get_profile", map[string]any{"name": "Default"}, &got)
	var ref struct {
		SHA256 string `json:"sha256"`
	}
	require.NoError(t, json.Unmarshal([]byte(got.Variables["Splat"]), &ref))
	assert.Len(t, ref.SHA256, 64, "the entry names the content instead of inlining it")

	callTool(t, session, "save_profile", map[string]any{"name": "Empty"}, &polyformmcp.SaveProfileOutput{})
	var updated polyformmcp.UpdateProfileOutput
	callTool(t, session, "update_profile", map[string]any{"name": "Empty", "variables": map[string]any{"Splat": got.Variables["Splat"]}}, &updated)
	assert.Equal(t, got.Variables["Splat"], updated.Variables["Splat"])

	path := filepath.Join(t.TempDir(), "graph.json")
	callTool(t, session, "save_graph", map[string]any{"path": path}, &polyformmcp.SaveGraphOutput{})
	payload, err := os.ReadFile(path)
	require.NoError(t, err)
	var saved struct {
		Buffers []struct {
			ByteLength int `json:"byteLength"`
		} `json:"buffers"`
		BufferViews []json.RawMessage `json:"bufferViews"`
	}
	require.NoError(t, json.Unmarshal(payload, &saved))
	require.Len(t, saved.Buffers, 1)
	assert.Equal(t, len(content), saved.Buffers[0].ByteLength)
	assert.Len(t, saved.BufferViews, 1)
}
