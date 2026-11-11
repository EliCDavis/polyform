package mcp_test

import (
	"path/filepath"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	groupNodeType       = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/nodes/arrays.GroupNode]"
	spreadArrayNodeType      = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/nodes/arrays.SpreadNode]"
	modelNodeType       = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ModelNode]"
	manifestNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ManifestNode]"
	trsListParamType    = "github.com/EliCDavis/polyform/generator/parameter.Value[[]github.com/EliCDavis/polyform/math/trs.TRS]"
	meshPortType        = "github.com/EliCDavis/polyform/modeling.Mesh"
	trsListPortType     = "[]github.com/EliCDavis/polyform/math/trs.TRS"
	twoTransformsAtOnce = "[{\"position\":{\"x\":1,\"y\":0,\"z\":0}},{\"position\":{\"x\":2,\"y\":0,\"z\":0}}]"
)

func TestAnEmptyInstanceGroupRunForEachDrawsNothing(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_subgraph", map[string]any{
		"id": "instanced", "name": "Instanced",
		"inputs": []map[string]any{
			{"name": "Mesh", "type": meshPortType},
			{"name": "Transforms", "type": trsListPortType},
		},
		"nodes": []map[string]any{
			{"alias": "model", "type": modelNodeType, "inputs": map[string]any{
				"Mesh":          map[string]any{"nodeId": "Mesh", "port": "Value"},
				"Gpu Instances": map[string]any{"nodeId": "Transforms", "port": "Value"},
			}},
		},
		"outputs": []map[string]any{{"name": "Model", "nodeId": "model", "port": "Out"}},
	}, &polyformmcp.CreateSubgraphOutput{})

	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []map[string]any{
		{"alias": "cube", "type": cubeNodeType},
		{"alias": "all", "type": trsListParamType, "value": twoTransformsAtOnce},
		// Both transforms land in group 1; group 0 stays empty.
		{"alias": "groups", "type": groupNodeType, "inputs": map[string]any{
			"Array":  map[string]any{"nodeId": "all", "port": "Value"},
			"Groups": map[string]any{"value": "[1,1]"},
			"Count":  map[string]any{"value": "2"},
		}},
		{"alias": "runs", "subgraphId": "instanced", "forEach": true, "inputs": map[string]any{
			"Mesh":       map[string]any{"nodeId": "cube", "port": "Out"},
			"Transforms": map[string]any{"nodeId": "groups", "port": "Out"},
		}},
		{"alias": "spread", "type": spreadArrayNodeType, "inputs": map[string]any{"Array": map[string]any{"nodeId": "runs", "port": "Model"}}},
		{"alias": "group", "type": modelNodeType, "inputs": map[string]any{
			"Children": map[string]any{"elements": []map[string]any{{"nodeId": "spread", "port": "Out"}}},
		}},
		{"alias": "manifest", "type": manifestNodeType, "inputs": map[string]any{
			"Models": map[string]any{"elements": []map[string]any{{"nodeId": "group", "port": "Out"}}},
		}},
	}}, &created)

	var described polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": created.Nodes["manifest"]}, &described)

	drawn := 0
	for _, model := range described.Models {
		drawn += model.Triangles
	}
	require.Len(t, described.Models, 3, "the group plus one model per kind, even the empty one")
	assert.Equal(t, 24, drawn, "two instanced cubes and nothing from the empty group; a stray cube at the origin adds 12")

	var rendered polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": created.Nodes["manifest"], "outputPath": filepath.Join(t.TempDir(), "empty-group.png"),
	}, &rendered)
	assert.Equal(t, 24, rendered.TriangleCount, "the preview draws the same as the file")
}
