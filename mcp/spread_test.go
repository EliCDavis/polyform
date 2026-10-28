package mcp_test

import (
	"path/filepath"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	spreadNodeType       = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/arrays.SpreadNode]"
	fromElementsNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/arrays.FromElementsNode]"
	sumNodeType          = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.AddNode[float64]]"
	combineNodeType      = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/meshops.CombineNode]"
	repeatArrayNodeType  = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/arrays.RepeatNode]"
	appendNodeType       = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/arrays.AppendNode]"
	concatNodeType       = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/arrays.ConcatNode]"
	switchNodeType       = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/logic.SwitchNode]"
	csgUnionNodeType     = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/csg.UnionNode]"
)

func spreadOf(alias, source string) map[string]any {
	return map[string]any{"alias": alias, "type": spreadNodeType, "inputs": map[string]any{
		"Array": map[string]any{"nodeId": source, "port": "Out"},
	}}
}

func TestSpreadIntoAddSumsTheArray(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "values", "type": floatArrayParamType, "value": "[1, 2, 3]"},
		map[string]any{"alias": "spread", "type": spreadNodeType, "inputs": map[string]any{
			"Array": map[string]any{"nodeId": "values", "port": "Value"},
		}},
		map[string]any{"alias": "sum", "type": sumNodeType, "inputs": map[string]any{
			"Values": map[string]any{"elements": []any{
				map[string]any{"nodeId": "spread", "port": "Out"},
				map[string]any{"value": "10"},
			}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	assert.InDelta(t, 16, evalFloat64Output(t, inst, created.Nodes["sum"], "Float"), 1e-9)
}

func TestSpreadSurvivesSaveAndLoad(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "values", "type": floatArrayParamType, "value": "[1, 2, 3]"},
		map[string]any{"alias": "spread", "type": spreadNodeType, "inputs": map[string]any{
			"Array": map[string]any{"nodeId": "values", "port": "Value"},
		}},
		map[string]any{"alias": "sum", "type": sumNodeType, "inputs": map[string]any{
			"Values": map[string]any{"elements": []any{map[string]any{"nodeId": "spread", "port": "Out"}}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)
	path := filepath.Join(t.TempDir(), "graph.json")
	callTool(t, session, "save_graph", map[string]any{"path": path}, nil)

	reloaded, inst := testSessionWithInstance(t)
	callTool(t, reloaded, "load_graph", map[string]any{"path": path}, nil)
	assert.InDelta(t, 6, evalFloat64Output(t, inst, created.Nodes["sum"], "Float"), 1e-9)
}

func TestSpreadOfArraysAddsThemElementByElement(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "a", "type": floatArrayParamType, "value": "[1, 2, 3]"},
		map[string]any{"alias": "b", "type": floatArrayParamType, "value": "[10, 20, 30]"},
		map[string]any{"alias": "both", "type": fromElementsNodeType, "inputs": map[string]any{
			"Elements": map[string]any{"elements": []any{
				map[string]any{"nodeId": "a", "port": "Value"},
				map[string]any{"nodeId": "b", "port": "Value"},
			}},
		}},
		spreadOf("spread", "both"),
		map[string]any{"alias": "sum", "type": sumNodeType, "inputs": map[string]any{
			"Values": map[string]any{"elements": []any{map[string]any{"nodeId": "spread", "port": "Out"}}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	sum := nodes.GetNodeOutputPort[[]float64](inst.Node(created.Nodes["sum"]), "Float").Value()
	assert.Equal(t, []float64{11, 22, 33}, sum)
}

func meshSpread(t *testing.T) []any {
	t.Helper()
	return []any{
		map[string]any{"alias": "small", "type": cubeNodeType},
		map[string]any{"alias": "big", "type": cubeNodeType, "inputs": map[string]any{
			"Width": map[string]any{"value": "3"}, "Height": map[string]any{"value": "3"}, "Depth": map[string]any{"value": "3"},
		}},
		map[string]any{"alias": "meshes", "type": fromElementsNodeType, "inputs": map[string]any{
			"Elements": map[string]any{"elements": []any{
				map[string]any{"nodeId": "small", "port": "Out"},
				map[string]any{"nodeId": "big", "port": "Out"},
			}},
		}},
		spreadOf("spread", "meshes"),
	}
}

func TestSpreadIntoCombineJoinsEveryMesh(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": append(meshSpread(t),
		map[string]any{"alias": "combine", "type": combineNodeType, "inputs": map[string]any{
			"Meshes": map[string]any{"elements": []any{map[string]any{"nodeId": "spread", "port": "Out"}}},
		}},
	)}, &created)
	require.Empty(t, created.Errors)

	combined := nodes.GetNodeOutputPort[modeling.Mesh](inst.Node(created.Nodes["combine"]), "Out").Value()
	assert.Equal(t, 24, combined.PrimitiveCount())
}

func TestSpreadIntoCSGUnionMergesEverySolid(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": append(meshSpread(t),
		map[string]any{"alias": "union", "type": csgUnionNodeType, "inputs": map[string]any{
			"Meshes": map[string]any{"elements": []any{map[string]any{"nodeId": "spread", "port": "Out"}}},
		}},
	)}, &created)
	require.Empty(t, created.Errors)

	union := nodes.GetNodeOutputPort[modeling.Mesh](inst.Node(created.Nodes["union"]), "Out")
	bounds := union.Value().BoundingBox(modeling.PositionAttribute)
	assert.InDelta(t, 3, bounds.Size().X(), 1e-9, "the wide cube made it into the union")
	assert.Empty(t, union.(nodes.ObservableExecution).ExecutionReport().Errors)
}

func TestSpreadIntoManifestExportsEveryModel(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType},
		map[string]any{"alias": "a", "type": gltfModelType, "inputs": map[string]any{
			"Name": map[string]any{"value": `"A"`}, "Mesh": map[string]any{"nodeId": "cube", "port": "Out"},
		}},
		map[string]any{"alias": "b", "type": gltfModelType, "inputs": map[string]any{
			"Name": map[string]any{"value": `"B"`}, "Mesh": map[string]any{"nodeId": "cube", "port": "Out"},
		}},
		map[string]any{"alias": "models", "type": fromElementsNodeType, "inputs": map[string]any{
			"Elements": map[string]any{"elements": []any{
				map[string]any{"nodeId": "a", "port": "Out"},
				map[string]any{"nodeId": "b", "port": "Out"},
			}},
		}},
		spreadOf("spread", "models"),
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"elements": []any{map[string]any{"nodeId": "spread", "port": "Out"}}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var out polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": created.Nodes["manifest"]}, &out)
	assert.Equal(t, 2, out.ModelCount)
}

func TestSpreadIntoTheArrayNodesListInputs(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "values", "type": floatArrayParamType, "value": "[1, 2, 3]"},
		map[string]any{"alias": "tail", "type": floatArrayParamType, "value": "[9]"},
		map[string]any{"alias": "spread", "type": spreadNodeType, "inputs": map[string]any{
			"Array": map[string]any{"nodeId": "values", "port": "Value"},
		}},
		map[string]any{"alias": "collect", "type": fromElementsNodeType, "inputs": map[string]any{
			"Elements": map[string]any{"elements": []any{map[string]any{"nodeId": "spread", "port": "Out"}}},
		}},
		map[string]any{"alias": "append", "type": appendNodeType, "inputs": map[string]any{
			"Array":    map[string]any{"nodeId": "tail", "port": "Value"},
			"Elements": map[string]any{"elements": []any{map[string]any{"nodeId": "spread", "port": "Out"}}},
		}},
		map[string]any{"alias": "lists", "type": fromElementsNodeType, "inputs": map[string]any{
			"Elements": map[string]any{"elements": []any{
				map[string]any{"nodeId": "values", "port": "Value"},
				map[string]any{"nodeId": "tail", "port": "Value"},
			}},
		}},
		spreadOf("spreadLists", "lists"),
		map[string]any{"alias": "concat", "type": concatNodeType, "inputs": map[string]any{
			"Arrays": map[string]any{"elements": []any{map[string]any{"nodeId": "spreadLists", "port": "Out"}}},
		}},
		map[string]any{"alias": "hundred", "type": "github.com/EliCDavis/polyform/generator/parameter.Value[float64]", "value": "100"},
		map[string]any{"alias": "pick", "type": switchNodeType, "inputs": map[string]any{
			"Index": map[string]any{"value": "2"},
			"Cases": map[string]any{"elements": []any{
				map[string]any{"nodeId": "hundred", "port": "Value"},
				map[string]any{"nodeId": "spread", "port": "Out"},
			}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	array := func(alias string) []float64 {
		return nodes.GetNodeOutputPort[[]float64](inst.Node(created.Nodes[alias]), "Out").Value()
	}
	assert.Equal(t, []float64{1, 2, 3}, array("collect"))
	assert.Equal(t, []float64{9, 1, 2, 3}, array("append"))
	assert.Equal(t, []float64{1, 2, 3, 9}, array("concat"))
	assert.InDelta(t, 2, evalFloat64Output(t, inst, created.Nodes["pick"], "Out"), 1e-9,
		"case 0 is the literal, cases 1-3 the spread's elements")
}

func TestSpreadIsRefusedByInputsThatTakeOneValue(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "values", "type": floatArrayParamType, "value": "[1, 2, 3]"},
		map[string]any{"alias": "spread", "type": spreadNodeType, "inputs": map[string]any{
			"Array": map[string]any{"nodeId": "values", "port": "Value"},
		}},
		map[string]any{"alias": "cube", "type": cubeNodeType},
		map[string]any{"alias": "round", "type": roundNodeType},
		map[string]any{"alias": "repeat", "type": repeatArrayNodeType},
	}}, &created)
	require.Empty(t, created.Errors)

	for _, target := range []struct{ node, port string }{
		{created.Nodes["cube"], "Width"},
		{created.Nodes["round"], "In"},
		{created.Nodes["repeat"], "Value"},
	} {
		msg := callToolExpectingError(t, session, "connect_nodes", map[string]any{
			"outNodeId": created.Nodes["spread"], "outPort": "Out", "inNodeId": target.node, "inPort": target.port,
		})
		assert.Contains(t, msg, "not a spread", target.port)
	}
}
