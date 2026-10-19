package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	roundNodeType          = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.RoundNode]"
	varyingRadiusLinesType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.VaryingRadiusLinesNode]"
)

// array -> round (lifted) -> lines.Radii, which only takes []float64.
func liftedIntoArrayOnlyInput(t *testing.T) (session *mcpsdk.ClientSession, array, round, lines string) {
	t.Helper()
	s := testSession(t)
	var made polyformmcp.CreateNodesOutput
	callTool(t, s, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "array", "type": floatArrayParamType},
		map[string]any{"alias": "round", "type": roundNodeType, "inputs": map[string]any{
			"In": map[string]any{"nodeId": "array", "port": "Value"},
		}},
		map[string]any{"alias": "lines", "type": varyingRadiusLinesType, "inputs": map[string]any{
			"Radii": map[string]any{"nodeId": "round", "port": "Float"},
		}},
	}}, &made)
	require.Empty(t, made.Errors)
	return s, made.Nodes["array"], made.Nodes["round"], made.Nodes["lines"]
}

func TestDisconnectReportsAnEdgeItsChangeOfTypeDropped(t *testing.T) {
	session, _, round, lines := liftedIntoArrayOnlyInput(t)

	var out polyformmcp.DisconnectOutput
	callTool(t, session, "disconnect", map[string]any{"nodeId": round, "port": "In"}, &out)

	require.Len(t, out.Dropped, 1, "%+v", out.Dropped)
	assert.Equal(t, round+".Float", out.Dropped[0].From)
	assert.Equal(t, lines+".Radii", out.Dropped[0].To)
}

func TestDeleteNodeReportsAnEdgeItsChangeOfTypeDropped(t *testing.T) {
	session, array, round, lines := liftedIntoArrayOnlyInput(t)

	var out polyformmcp.DeleteNodeOutput
	callTool(t, session, "delete_node", map[string]any{"nodeId": array}, &out)

	require.Len(t, out.Dropped, 1, "%+v", out.Dropped)
	assert.Equal(t, round+".Float", out.Dropped[0].From)
	assert.Equal(t, lines+".Radii", out.Dropped[0].To)
}

func TestDeleteNodeLeavesOutEdgesIntoNodesItAlsoDeleted(t *testing.T) {
	session, array, _, lines := liftedIntoArrayOnlyInput(t)

	var out polyformmcp.DeleteNodeOutput
	callTool(t, session, "delete_node", map[string]any{"nodes": []any{
		map[string]any{"nodeId": array},
		map[string]any{"nodeId": lines},
	}}, &out)

	assert.Empty(t, out.Dropped)
}
