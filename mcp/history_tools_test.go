package mcp_test

import (
	"context"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUndoTakesBackTheLastCall(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &cube)
	require.Contains(t, inst.Schema().Nodes, cube.NodeId)

	var out polyformmcp.HistoryOutput
	callTool(t, session, "undo", map[string]any{}, &out)

	assert.Equal(t, []string{"create_node"}, out.Undone, "the step is named by the tool that made it")
	assert.NotContains(t, inst.Schema().Nodes, cube.NodeId)
	assert.Equal(t, []string{"create_node"}, out.CanRedo)
}

func TestRedoPutsItBack(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &cube)
	callTool(t, session, "undo", map[string]any{}, &polyformmcp.HistoryOutput{})

	var out polyformmcp.HistoryOutput
	callTool(t, session, "redo", map[string]any{}, &out)

	assert.Equal(t, []string{"create_node"}, out.Redone)
	assert.Contains(t, inst.Schema().Nodes, cube.NodeId)
}

func TestUndoWalksSeveralStepsAtOnce(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	for range 3 {
		callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &polyformmcp.CreateNodeOutput{})
	}
	require.Len(t, inst.Schema().Nodes, 3)

	var out polyformmcp.HistoryOutput
	callTool(t, session, "undo", map[string]any{"steps": 2}, &out)

	assert.Len(t, out.Undone, 2)
	assert.Len(t, inst.Schema().Nodes, 1)
}

func TestUndoOnAFreshGraphSaysSoRatherThanFailing(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var out polyformmcp.HistoryOutput
	callTool(t, session, "undo", map[string]any{}, &out)

	assert.Empty(t, out.Undone)
	assert.Empty(t, out.CanUndo)
}

// The bug this is for: a call that failed partway used to leave its
// earlier mutations applied, with describe_graph and render_preview then
// disagreeing about what the graph was.
func TestAFailedCallLeavesNothingBehind(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_nodes",
		Arguments: map[string]any{
			"nodes": []map[string]any{
				{"alias": "a", "type": cubeNodeType},
			},
		},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)
	before := len(inst.Schema().Nodes)

	// Connecting a node to itself is refused by the graph layer, after the
	// call has already created the node it was going to wire.
	res, err = session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_node",
		Arguments: map[string]any{
			"type":   cubeNodeType,
			"inputs": map[string]any{"Height": map[string]any{"value": "not valid json at all"}},
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError)

	assert.Equal(t, before, len(inst.Schema().Nodes),
		"the node the failed call created is gone too")
}

func TestListHistoryChangesNothing(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &polyformmcp.CreateNodeOutput{})

	var out polyformmcp.HistoryOutput
	callTool(t, session, "list_history", map[string]any{}, &out)

	assert.Equal(t, []string{"create_node"}, out.CanUndo)
	assert.Len(t, inst.Schema().Nodes, 1)
}
