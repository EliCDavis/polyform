package mcp_test

import (
	"encoding/json"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cubeWithVariableWidth(t *testing.T, session *mcpsdk.ClientSession) string {
	t.Helper()
	callTool(t, session, "create_variable", map[string]any{
		"path": "Width", "type": "float64", "value": "1",
	}, &polyformmcp.CreateVariableOutput{})
	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType, "inputs": map[string]any{
		"Width": map[string]any{"variable": "Width"},
	}}, &cube)
	return cube.NodeId
}

func TestDescribeMeshMeasuresAVariantWithoutKeepingIt(t *testing.T) {
	session := testSession(t)
	cube := cubeWithVariableWidth(t, session)

	var wide polyformmcp.DescribeMeshOutput
	callTool(t, session, "describe_mesh", map[string]any{
		"nodeId": cube, "variables": map[string]any{"Width": "4"},
	}, &wide)
	assert.InDelta(t, 4, wide.Extent.Size.X, 1e-9)

	var back polyformmcp.DescribeMeshOutput
	callTool(t, session, "describe_mesh", map[string]any{"nodeId": cube}, &back)
	assert.InDelta(t, 1, back.Extent.Size.X, 1e-9)
}

func TestEvaluateNodeReadsAVariantWithoutKeepingIt(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variable", map[string]any{
		"path": "Scale", "type": "float64", "value": "2",
	}, &polyformmcp.CreateVariableOutput{})
	var mul polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": multiplyNodeType, "inputs": map[string]any{
		"Values": map[string]any{"elements": []any{
			map[string]any{"variable": "Scale"},
			map[string]any{"value": "3"},
		}},
	}}, &mul)

	read := func(args map[string]any) float64 {
		var out polyformmcp.EvaluateNodeOutput
		callTool(t, session, "evaluate_node", args, &out)
		var v float64
		require.NoError(t, json.Unmarshal(remarshal(t, out.Value), &v))
		return v
	}
	assert.InDelta(t, 30, read(map[string]any{
		"nodeId": mul.NodeId, "port": "Float", "variables": map[string]any{"Scale": "10"},
	}), 1e-9)
	assert.InDelta(t, 6, read(map[string]any{"nodeId": mul.NodeId, "port": "Float"}), 1e-9)
}
