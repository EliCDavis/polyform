package mcp_test

import (
	"path/filepath"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
)

func TestAPortNamedWithADotAndDigitsIsOnePortNotAnElement(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var sub polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "switch", "name": "Switch",
		"inputs":  []any{map[string]any{"name": "Tact Switch 6x3.5", "type": "float64"}},
		"outputs": []any{map[string]any{"name": "Size", "nodeId": "Tact Switch 6x3.5", "port": "Value"}},
	}, &sub)

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "switch",
		"inputs":     map[string]any{"Tact Switch 6x3.5": map[string]any{"value": "3.5"}},
	}, &placed)
	inputByLabel(t, inst.Node(placed.NodeId), "Tact Switch 6x3.5")

	path := filepath.Join(t.TempDir(), "graph.json")
	callTool(t, session, "save_graph", map[string]any{"path": path}, nil)
	reloaded := testSession(t)
	callTool(t, reloaded, "load_graph", map[string]any{"path": path}, nil)

	var evaluated polyformmcp.EvaluateNodeOutput
	callTool(t, reloaded, "evaluate_node", map[string]any{"nodeId": placed.NodeId, "port": "Size"}, &evaluated)
	assert.Equal(t, 3.5, evaluated.Value)

	var disconnected polyformmcp.DisconnectOutput
	callTool(t, reloaded, "disconnect", map[string]any{"nodeId": placed.NodeId, "port": "Tact Switch 6x3.5"}, &disconnected)
	assert.True(t, disconnected.Disconnected)
}
