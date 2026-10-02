package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sdfSphereType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.SphereNode]"
	sdfUnionType  = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.UnionNode]"
	sdfMirrorType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.MirrorNode]"
)

func originPoint() []any {
	return []any{map[string]any{"x": 0.0, "y": 0.0, "z": 0.0}}
}

// A union's field output is called "Union", not "Field". Defaulting to the
// literal name "Field" made sample_field fail on exactly the combinator
// nodes a broken-geometry investigation wants to look at - it cost a real
// build two dead calls.
func TestSampleFieldFindsCombinatorOutputWithoutBeingTold(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var sphere polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   sdfSphereType,
		"inputs": map[string]any{"Radius": map[string]any{"value": "1"}},
	}, &sphere)

	var union polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": sdfUnionType,
		"inputs": map[string]any{
			"Fields": map[string]any{
				"elements": []any{map[string]any{"nodeId": sphere.NodeId, "port": "Field"}},
			},
		},
	}, &union)

	var out polyformmcp.SampleFieldOutput
	callTool(t, session, "sample_field", map[string]any{
		"nodeId": union.NodeId,
		"points": originPoint(),
	}, &out)

	assert.Equal(t, "Union", out.Port, "should have found the union's field output on its own")
	require.Len(t, out.Samples, 1)
	assert.Negative(t, out.Samples[0].Value, "origin is inside the unit sphere")
}

// A node with several field outputs can't be guessed at, so it should say
// so and name them rather than picking one.
func TestSampleFieldAsksWhichPortWhenSeveralFields(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var sphere polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": sdfSphereType}, &sphere)

	var mirror polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   sdfMirrorType,
		"inputs": map[string]any{"Field": map[string]any{"nodeId": sphere.NodeId, "port": "Field"}},
	}, &mirror)

	msg := callToolExpectingError(t, session, "sample_field", map[string]any{
		"nodeId": mirror.NodeId,
		"points": originPoint(),
	})
	assert.Contains(t, msg, "several field outputs")
	assert.Contains(t, msg, "XY")
}

// An explicitly named port that doesn't exist should list the real ones
// rather than leaving the caller to go run describe_graph.
func TestSampleFieldNamesTheRealPorts(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var sphere polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": sdfSphereType}, &sphere)

	msg := callToolExpectingError(t, session, "sample_field", map[string]any{
		"nodeId": sphere.NodeId,
		"port":   "Union",
		"points": originPoint(),
	})
	assert.Contains(t, msg, "Field")
}

// A mesh node has no field output at all; say that plainly.
func TestSampleFieldRejectsANonFieldNode(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": sphereNodeType}, &created)

	msg := callToolExpectingError(t, session, "sample_field", map[string]any{
		"nodeId": created.NodeId,
		"points": originPoint(),
	})
	assert.Contains(t, msg, "no SDF field output")
}
