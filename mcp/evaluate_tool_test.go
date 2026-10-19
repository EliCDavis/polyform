package mcp_test

import (
	"encoding/json"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	transformPointType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trs.TransformPointNode]"
	linearSeqType      = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sequence.LinearNode]"
)

func remarshal(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}

func TestEvaluateNodeReturnsAComputedVector(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "trs", "type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trs.NewNode]", "inputs": map[string]any{
			"Position": map[string]any{"value": `{"x":10,"y":0,"z":0}`},
		}},
		map[string]any{"alias": "moved", "type": transformPointType, "inputs": map[string]any{
			"TRS":   map[string]any{"nodeId": "trs", "port": "Out"},
			"Point": map[string]any{"value": `{"x":1,"y":2,"z":3}`},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var out polyformmcp.EvaluateNodeOutput
	callTool(t, session, "evaluate_node", map[string]any{"nodeId": created.Nodes["moved"]}, &out)
	assert.Equal(t, "Out", out.Port)
	var v struct{ X, Y, Z float64 }
	require.NoError(t, json.Unmarshal(remarshal(t, out.Value), &v))
	assert.InDelta(t, 11, v.X, 1e-9)
	assert.InDelta(t, 2, v.Y, 1e-9)
	assert.InDelta(t, 3, v.Z, 1e-9)
}

func TestEvaluateNodeTruncatesArraysButReportsLength(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": linearSeqType, "inputs": map[string]any{
		"Start": map[string]any{"value": "0"}, "End": map[string]any{"value": "1"}, "Samples": map[string]any{"value": "100"},
	}}, &created)

	var out polyformmcp.EvaluateNodeOutput
	callTool(t, session, "evaluate_node", map[string]any{"nodeId": created.NodeId, "maxItems": 3}, &out)
	assert.Equal(t, 100, out.Length)
	var arr []float64
	require.NoError(t, json.Unmarshal(remarshal(t, out.Value), &arr))
	assert.Len(t, arr, 3)
}

func TestEvaluateNodeSummarizesAFieldInsteadOfDumpingIt(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": sdfSphereType}, &created)

	var out polyformmcp.EvaluateNodeOutput
	callTool(t, session, "evaluate_node", map[string]any{"nodeId": created.NodeId}, &out)
	assert.Empty(t, out.Value)
	assert.Contains(t, out.Summary, "sample_field")
}

func TestEvaluateNodeNamesThePortsWhenAmbiguous(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": vec3ValueType}, &created)

	var out polyformmcp.EvaluateNodeOutput
	callTool(t, session, "evaluate_node", map[string]any{"nodeId": created.NodeId, "port": "Value"}, &out)
	assert.NotEmpty(t, out.Value)

	msg := callToolExpectingError(t, session, "evaluate_node", map[string]any{"nodeId": created.NodeId, "port": "Nope"})
	assert.Contains(t, msg, "Value")
}
