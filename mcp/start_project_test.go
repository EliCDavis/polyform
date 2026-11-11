package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartProjectBeginsWithAnEmptyGraph(t *testing.T) {
	session := testSession(t)

	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &created)
	callTool(t, session, "create_variable", map[string]any{"path": "Size", "type": "float64", "value": "1"}, &polyformmcp.CreateVariableOutput{})

	var started polyformmcp.StartProjectOutput
	callTool(t, session, "start_project", map[string]any{"path": t.TempDir(), "outsideRoot": true}, &started)
	assert.True(t, started.Cleared)

	var listed polyformmcp.ListNodesOutput
	callTool(t, session, "list_nodes", map[string]any{}, &listed)
	assert.Empty(t, listed.Nodes)

	var variables polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &variables)
	assert.Empty(t, variables.Variables)

	var again polyformmcp.StartProjectOutput
	callTool(t, session, "start_project", map[string]any{"path": t.TempDir(), "outsideRoot": true}, &again)
	assert.False(t, again.Cleared, "an empty graph has nothing to clear")
	require.NotEqual(t, started.Path, again.Path)
}
