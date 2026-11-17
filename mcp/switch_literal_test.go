package mcp_test

import (
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/math/norm"
	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestASwitchOfLiteralCasesFeedsANormInput(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "index", "type": intParamType, "value": "1"},
		map[string]any{"alias": "rough", "type": switchNodeType, "inputs": map[string]any{
			"Index": map[string]any{"nodeId": "index", "port": "Value"},
			"Cases": map[string]any{"elements": []any{
				map[string]any{"value": "0.42"},
				map[string]any{"value": "0.62"},
			}},
		}},
		map[string]any{"alias": "material", "type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.MaterialNode]", "inputs": map[string]any{
			"Roughness Factor": map[string]any{"nodeId": "rough", "port": "Out"},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	material := inst.Node(created.Nodes["material"]).(*nodes.Struct[gltf.MaterialNode])
	require.NotNil(t, material.Data.RoughnessFactor)
	assert.InDelta(t, 0.62, float64(material.Data.RoughnessFactor.Value()), 1e-9)
	var _ nodes.Output[norm.UFloat64] = material.Data.RoughnessFactor
}

func TestASwitchOfLiteralCasesInsideANewSubgraphFeedsANormInput(t *testing.T) {
	session, _ := testSessionWithInstance(t)
	var out polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id":     "finish",
		"name":   "Finish",
		"inputs": []any{map[string]any{"name": "Index", "type": "int"}},
		"nodes": []any{
			map[string]any{"alias": "rough", "type": switchNodeType, "inputs": map[string]any{
				"Index": map[string]any{"nodeId": "Index", "port": "Value"},
				"Cases": map[string]any{"elements": []any{
					map[string]any{"value": "0.42"},
					map[string]any{"value": "0.62"},
				}},
			}},
			map[string]any{"alias": "material", "type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.MaterialNode]", "inputs": map[string]any{
				"Roughness Factor": map[string]any{"nodeId": "rough", "port": "Out"},
			}},
		},
	}, &out)
	assert.Empty(t, out.Errors)
}
