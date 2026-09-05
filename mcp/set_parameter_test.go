package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const floatValueType = "github.com/EliCDavis/polyform/generator/parameter.Value[float64]"

// sphereRadius reads back the radius actually wired into a UvSphereNode.
func sphereRadius(t *testing.T, node nodes.Node) float64 {
	t.Helper()
	sphere, ok := node.(*nodes.Struct[primitives.UvSphereNode])
	require.True(t, ok, "expected a UvSphereNode")
	require.NotNil(t, sphere.Data.Radius, "Radius port is unwired")
	return sphere.Data.Radius.Value()
}

// TestSetParameterByPort is the whole point of port addressing: create_node
// invents a literal parameter node for an input given as a value and never
// reports its id, so without this the only way to change that value is a
// describe_graph to go hunting for the id first.
func TestSetParameterByPort(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   sphereNodeType,
		"inputs": map[string]any{"Radius": map[string]any{"value": "2.5"}},
	}, &created)

	var out polyformmcp.SetParameterOutput
	callTool(t, session, "set_parameter", map[string]any{
		"nodeId": created.NodeId,
		"port":   "Radius",
		"value":  "7.25",
	}, &out)
	assert.True(t, out.Updated)

	assert.Equal(t, 7.25, sphereRadius(t, inst.Node(created.NodeId)))
}

// An unwired port has no literal to update yet. Creating one is the same
// thing create_node's inputs would have done, and is far friendlier than
// erroring on a port the caller can plainly see exists.
func TestSetParameterByPortCreatesMissingLiteral(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": sphereNodeType}, &created)

	var out polyformmcp.SetParameterOutput
	callTool(t, session, "set_parameter", map[string]any{
		"nodeId": created.NodeId,
		"port":   "Radius",
		"value":  "3.5",
	}, &out)
	assert.True(t, out.Updated)

	assert.Equal(t, 3.5, sphereRadius(t, inst.Node(created.NodeId)))
}

// Port names are the space-cased form of a Go field, and callers guess the
// raw field name constantly. set_parameter goes through the same tolerant
// resolution the connect tools use.
func TestSetParameterByPortToleratesPortSpelling(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   cylinderNodeType,
		"inputs": map[string]any{"Height": map[string]any{"value": "1"}},
	}, &created)

	var out polyformmcp.SetParameterOutput
	callTool(t, session, "set_parameter", map[string]any{
		"nodeId": created.NodeId,
		"port":   "height",
		"value":  "4",
	}, &out)
	assert.True(t, out.Updated)

	cylinder, ok := inst.Node(created.NodeId).(*nodes.Struct[primitives.CylinderNode])
	require.True(t, ok)
	assert.Equal(t, 4.0, cylinder.Data.Height.Value())
}

// A port fed by a computed node has no literal behind it. Saying so beats
// the bare "not a parameter" panic text, which named an opaque node id the
// caller never chose and gave no next step.
func TestSetParameterByPortRejectsComputedInput(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "inner", "type": sphereNodeType},
			map[string]any{
				"alias": "combine",
				"type":  combineType,
				"inputs": map[string]any{
					"Meshes": map[string]any{
						"elements": []any{map[string]any{"nodeId": "inner", "port": "Out"}},
					},
				},
			},
		},
	}, &out)
	require.Empty(t, out.Errors)

	msg := callToolExpectingError(t, session, "set_parameter", map[string]any{
		"nodeId": out.Nodes["combine"],
		"port":   "Meshes",
		"value":  "1",
	})
	assert.Contains(t, msg, "array")
}

// A misspelled port should list the real ones rather than making the
// caller spend a get_node_types to find out.
func TestSetParameterByPortNamesTheRealPorts(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": sphereNodeType}, &created)

	msg := callToolExpectingError(t, session, "set_parameter", map[string]any{
		"nodeId": created.NodeId,
		"port":   "Radiuz",
		"value":  "1",
	})
	assert.Contains(t, msg, "Radius")
}

// The run that motivated batching spent 28 separate calls setting values,
// most of them in clusters tweaking one part's dimensions together.
func TestSetParameterBatch(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "a", "type": sphereNodeType, "inputs": map[string]any{"Radius": map[string]any{"value": "1"}}},
			map[string]any{"alias": "b", "type": sphereNodeType, "inputs": map[string]any{"Radius": map[string]any{"value": "1"}}},
		},
	}, &out)
	require.Empty(t, out.Errors)

	var set polyformmcp.SetParameterOutput
	callTool(t, session, "set_parameter", map[string]any{
		"parameters": []any{
			map[string]any{"nodeId": out.Nodes["a"], "port": "Radius", "value": "9"},
			map[string]any{"nodeId": out.Nodes["b"], "port": "Radius", "value": "11"},
		},
	}, &set)

	require.Empty(t, set.Errors)
	assert.True(t, set.Updated)
	assert.Equal(t, 2, set.Set)
	assert.Equal(t, 9.0, sphereRadius(t, inst.Node(out.Nodes["a"])))
	assert.Equal(t, 11.0, sphereRadius(t, inst.Node(out.Nodes["b"])))
}

// A bad entry mid-batch must not discard the assignments around it, or the
// caller has to re-derive which ones landed.
func TestSetParameterBatchPartialFailureKeepsTheRest(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": sphereNodeType}, &created)

	var set polyformmcp.SetParameterOutput
	callTool(t, session, "set_parameter", map[string]any{
		"parameters": []any{
			map[string]any{"nodeId": "Node-does-not-exist", "port": "Radius", "value": "1"},
			map[string]any{"nodeId": created.NodeId, "port": "Radius", "value": "6"},
		},
	}, &set)

	require.Len(t, set.Errors, 1)
	assert.Contains(t, set.Errors[0], "parameter 0")
	assert.False(t, set.Updated)
	assert.Equal(t, 1, set.Set)
	assert.Equal(t, 6.0, sphereRadius(t, inst.Node(created.NodeId)), "the good assignment still applied")
}

// disconnect accepts "Port.N" to remove one array element, but
// set_parameter used to reject the same spelling, so changing one
// element's literal meant disconnect + create + reconnect.
func TestSetParameterAddressesOneArrayElement(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "a", "type": sphereNodeType, "inputs": map[string]any{"Radius": map[string]any{"value": "1"}}},
			map[string]any{"alias": "b", "type": sphereNodeType, "inputs": map[string]any{"Radius": map[string]any{"value": "1"}}},
			map[string]any{
				"alias": "combine",
				"type":  combineType,
				"inputs": map[string]any{
					"Meshes": map[string]any{"elements": []any{
						map[string]any{"nodeId": "a", "port": "Out"},
						map[string]any{"nodeId": "b", "port": "Out"},
					}},
				},
			},
		},
	}, &out)
	require.Empty(t, out.Errors)

	// Element 1 of the combine's Meshes is the second sphere, which is fed
	// by a node, not a literal - so this must say so rather than silently
	// doing nothing.
	msg := callToolExpectingError(t, session, "set_parameter", map[string]any{
		"nodeId": out.Nodes["combine"], "port": "Meshes.1", "value": "2",
	})
	assert.Contains(t, msg, "computes its value")

	// Out of range is named precisely rather than reported as a bad port.
	msg = callToolExpectingError(t, session, "set_parameter", map[string]any{
		"nodeId": out.Nodes["combine"], "port": "Meshes.9", "value": "2",
	})
	assert.Contains(t, msg, "no index 9")

	// And an array port addressed without an index points at the syntax.
	msg = callToolExpectingError(t, session, "set_parameter", map[string]any{
		"nodeId": out.Nodes["combine"], "port": "Meshes", "value": "2",
	})
	assert.Contains(t, msg, "Meshes.0")

	_ = inst
}

// The literal case that motivated it: an array of values whose elements
// are literals, one of which needs changing.
func TestSetParameterUpdatesAnArrayElementLiteral(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "a", "type": floatValueType, "inputs": map[string]any{}},
			map[string]any{"alias": "b", "type": floatValueType, "inputs": map[string]any{}},
			map[string]any{
				"alias": "sum",
				"type":  "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.MultiplyNode[float64]]",
				"inputs": map[string]any{
					"Values": map[string]any{"elements": []any{
						map[string]any{"nodeId": "a", "port": "Value"},
						map[string]any{"nodeId": "b", "port": "Value"},
					}},
				},
			},
		},
	}, &out)
	require.Empty(t, out.Errors)

	var set polyformmcp.SetParameterOutput
	callTool(t, session, "set_parameter", map[string]any{
		"nodeId": out.Nodes["sum"], "port": "Values.1", "value": "7.5",
	}, &set)
	assert.True(t, set.Updated)

	port := nodes.GetNodeOutputPort[float64](inst.Node(out.Nodes["b"]), "Value")
	assert.Equal(t, 7.5, port.Value(), "the second element's literal should have changed")
}
