package mcp_test

import (
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildScopedSphere makes a subgraph whose sphere radius comes from a
// boundary input, plus one instance of it wired to a real radius. It
// returns the interior sphere node's id and the instance's id.
func buildScopedSphere(t *testing.T, session *mcpsdk.ClientSession, radius string) (string, string) {
	t.Helper()

	var sg polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{"id": "part", "name": "Part"}, &sg)

	var radiusIn polyformmcp.CreateBoundaryNodeOutput
	callTool(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "part", "kind": "input", "portType": "float64", "name": "Radius",
	}, &radiusIn)

	var sphere polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"scope": "part", "type": sdfSphereNodeType,
		"inputs": map[string]any{
			"Radius": map[string]any{"nodeId": radiusIn.NodeId, "port": radiusIn.Port},
		},
	}, &sphere)

	var fieldOut polyformmcp.CreateBoundaryNodeOutput
	callTool(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "part", "kind": "output",
		"portType": "github.com/EliCDavis/polyform/math/sample.Vec3ToFloat", "name": "Field",
	}, &fieldOut)

	var c polyformmcp.ConnectNodesOutput
	callTool(t, session, "connect_nodes", map[string]any{
		"scope": "part", "outNodeId": sphere.NodeId, "outPort": "Field",
		"inNodeId": fieldOut.NodeId, "inPort": fieldOut.Port,
	}, &c)

	var instance polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "part",
		"inputs":     map[string]any{"Radius": map[string]any{"value": radius}},
	}, &instance)

	return sphere.NodeId, instance.NodeId
}

// Probing a subgraph by scope alone reads the shared definition, where
// nothing is wired into the boundary inputs, so everything downstream
// evaluates to a zero value. The danger is that those numbers look real -
// a distance of 0 reads as "on the surface", not as "no input" - and it
// has twice convinced a build that a working node was broken. It must say
// what it is showing.
func TestSampleFieldWarnsWhenProbingADefinition(t *testing.T) {
	session, _ := testSessionWithInstance(t)
	sphereID, instanceID := buildScopedSphere(t, session, "3")

	var scoped polyformmcp.SampleFieldOutput
	callTool(t, session, "sample_field", map[string]any{
		"scope": "part", "nodeId": sphereID, "points": []any{vec(0, 0, 0)},
	}, &scoped)

	require.NotEmpty(t, scoped.EvaluationContext,
		"a definition-scoped sample must say the values aren't from a live instance")
	assert.Contains(t, scoped.EvaluationContext, instanceID,
		"and should name the instance that can be measured instead")
}

// Naming the instance measures its own copy, with its inputs applied.
func TestSampleFieldThroughAnInstanceSeesItsWiredValues(t *testing.T) {
	session, _ := testSessionWithInstance(t)
	sphereID, instanceID := buildScopedSphere(t, session, "3")

	var live polyformmcp.SampleFieldOutput
	callTool(t, session, "sample_field", map[string]any{
		"instanceId": instanceID, "nodeId": sphereID, "points": []any{vec(0, 0, 0)},
	}, &live)

	require.Len(t, live.Samples, 1)
	assert.InDelta(t, -3, live.Samples[0].Value, 1e-9,
		"the centre of a radius-3 sphere is 3 inside it")
	assert.Empty(t, live.EvaluationContext, "a live instance needs no caveat")
}

// The same guarantee for raycast, which shares the resolution path.
func TestRaycastFieldThroughAnInstanceSeesItsWiredValues(t *testing.T) {
	session, _ := testSessionWithInstance(t)
	sphereID, instanceID := buildScopedSphere(t, session, "2")

	var hit polyformmcp.RaycastFieldOutput
	callTool(t, session, "raycast_field", map[string]any{
		"instanceId": instanceID, "nodeId": sphereID,
		"rays": []any{map[string]any{"origin": vec(0, 0, 0), "direction": vec(0, 1, 0)}},
	}, &hit)

	require.Len(t, hit.Hits, 1)
	require.True(t, hit.Hits[0].Hit)
	assert.InDelta(t, 2, hit.Hits[0].Point.Y, 0.001, "the surface of a radius-2 sphere")
	assert.Empty(t, hit.Hits[0].Note)
}

// A wrong instanceId has to be named plainly rather than silently falling
// back to the definition, which would reintroduce the whole problem.
func TestFieldToolsRejectABadInstanceId(t *testing.T) {
	session, _ := testSessionWithInstance(t)
	sphereID, _ := buildScopedSphere(t, session, "3")

	msg := callToolExpectingError(t, session, "sample_field", map[string]any{
		"instanceId": "Node-does-not-exist", "nodeId": sphereID, "points": []any{vec(0, 0, 0)},
	})
	assert.Contains(t, msg, "no node exists")

	// A node that exists but isn't a subgraph instance.
	var plain polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": sdfSphereNodeType}, &plain)
	msg = callToolExpectingError(t, session, "sample_field", map[string]any{
		"instanceId": plain.NodeId, "nodeId": sphereID, "points": []any{vec(0, 0, 0)},
	})
	assert.Contains(t, msg, "not a subgraph instance")
}
