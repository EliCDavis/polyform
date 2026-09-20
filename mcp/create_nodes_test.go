package mcp_test

import (
	"context"
	"strings"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sphereNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/primitives.UvSphereNode]"
	combineType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/meshops.CombineNode]"
)

// TestCreateNodesResolvesForwardAliases is the property the whole tool
// exists for: an entry may reference another entry that appears later in
// the list. Without it the caller has to topologically sort the batch by
// hand, and any node it got wrong is back to a separate round trip.
func TestCreateNodesResolvesForwardAliases(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			// "combine" references "sphere", which is declared after it.
			map[string]any{
				"alias": "combine",
				"type":  combineType,
				"inputs": map[string]any{
					"Meshes": map[string]any{
						"elements": []any{map[string]any{"nodeId": "sphere", "port": "Out"}},
					},
				},
			},
			map[string]any{
				"alias":  "sphere",
				"type":   sphereNodeType,
				"inputs": map[string]any{"Radius": map[string]any{"value": "2.5"}},
			},
		},
	}, &out)

	require.Empty(t, out.Errors)
	require.Len(t, out.Nodes, 2)
	require.Contains(t, out.Nodes, "combine")
	require.Contains(t, out.Nodes, "sphere")

	// The wiring really happened: the combined mesh has the sphere in it.
	port := nodes.GetNodeOutputPort[modeling.Mesh](inst.Node(out.Nodes["combine"]), "Out")
	assert.Greater(t, port.Value().PrimitiveCount(), 0, "combine should have received the sphere")
}

// TestCreateNodesReferencesPreexistingNodes pins that a nodeId matching no
// alias is still treated as a real id, so a batch can wire into whatever
// was already built.
func TestCreateNodesReferencesPreexistingNodes(t *testing.T) {
	session := testSession(t)

	var first polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": sphereNodeType}, &first)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{
				"alias": "combine",
				"type":  combineType,
				"inputs": map[string]any{
					"Meshes": map[string]any{
						"elements": []any{map[string]any{"nodeId": first.NodeId, "port": "Out"}},
					},
				},
			},
		},
	}, &out)

	require.Empty(t, out.Errors)
	require.Contains(t, out.Nodes, "combine")
}

// TestCreateNodesPartialFailureKeepsTheRest covers the failure mode that
// would otherwise cost back every turn the batch saved: one bad entry must
// not discard the ids of the entries that worked.
func TestCreateNodesPartialFailureKeepsTheRest(t *testing.T) {
	session := testSession(t)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "good", "type": sphereNodeType},
			map[string]any{"alias": "bogus", "type": "not.A.Real.Type"},
			map[string]any{"alias": "alsoGood", "type": sphereNodeType},
		},
	}, &out)

	require.Len(t, out.Errors, 1)
	assert.Contains(t, out.Errors[0], "bogus", "the error should name the entry that failed")
	assert.Contains(t, out.Nodes, "good")
	assert.Contains(t, out.Nodes, "alsoGood")
	assert.NotContains(t, out.Nodes, "bogus")
}

// TestCreateNodesRejectsDuplicateAliases fails the whole batch, because a
// duplicate alias makes every reference to it ambiguous.
func TestCreateNodesRejectsDuplicateAliases(t *testing.T) {
	session := testSession(t)

	err := callToolExpectingError(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "dup", "type": sphereNodeType},
			map[string]any{"alias": "dup", "type": sphereNodeType},
		},
	})
	assert.Contains(t, strings.ToLower(err), "unique")
}

// TestConnectNodesBatch pins the batch form of connect_nodes, and that a
// bad edge is reported per-index without dropping the good ones.
func TestConnectNodesBatch(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "a", "type": sphereNodeType},
			map[string]any{"alias": "b", "type": sphereNodeType},
			map[string]any{"alias": "combine", "type": combineType},
		},
	}, &made)
	require.Empty(t, made.Errors)

	var out polyformmcp.ConnectNodesOutput
	callTool(t, session, "connect_nodes", map[string]any{
		"connections": []any{
			map[string]any{"outNodeId": made.Nodes["a"], "outPort": "Out", "inNodeId": made.Nodes["combine"], "inPort": "Meshes"},
			map[string]any{"outNodeId": made.Nodes["b"], "outPort": "NotAPort", "inNodeId": made.Nodes["combine"], "inPort": "Meshes"},
			map[string]any{"outNodeId": made.Nodes["b"], "outPort": "Out", "inNodeId": made.Nodes["combine"], "inPort": "Meshes"},
		},
	}, &out)

	assert.False(t, out.Connected, "a batch with a bad edge is not fully connected")
	assert.Equal(t, 2, out.Made)
	require.Len(t, out.Errors, 1)
	assert.Contains(t, out.Errors[0], "connection 1", "the error should name which edge failed")

	port := nodes.GetNodeOutputPort[modeling.Mesh](inst.Node(made.Nodes["combine"]), "Out")
	assert.Greater(t, port.Value().PrimitiveCount(), 0)
}

// callToolExpectingError returns the tool-error text from a call that is
// supposed to fail.
func callToolExpectingError(t *testing.T, session *mcpsdk.ClientSession, name string, args map[string]any) string {
	t.Helper()

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{Name: name, Arguments: args})
	require.NoError(t, err)
	require.True(t, res.IsError, "expected tool %q to report an error", name)

	var msg string
	for _, c := range res.Content {
		if tc, ok := c.(*mcpsdk.TextContent); ok {
			msg += tc.Text
		}
	}
	return msg
}

// Replays the real failure from an anglerfish build: a batch referenced an
// alias declared in an EARLIER create_nodes call. Aliases are batch-local,
// so that is a genuine mistake - but wiring reaches inst.ConnectNodes,
// which panics rather than erroring, and the unrecovered panic failed the
// whole tool call. That discarded the structured content, so every id the
// batch had just created was lost and had to be re-derived with
// describe_graph, which is exactly what the errors field exists to avoid.
func TestCreateNodesStaleAliasKeepsTheBatch(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var first polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{map[string]any{"alias": "earlier", "type": sphereNodeType}},
	}, &first)
	require.Empty(t, first.Errors)

	var second polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{
				"alias": "combine",
				"type":  combineType,
				"inputs": map[string]any{
					// "earlier" belongs to the previous call, not this one.
					"Meshes": map[string]any{
						"elements": []any{map[string]any{"nodeId": "earlier", "port": "Out"}},
					},
				},
			},
			map[string]any{"alias": "fresh", "type": sphereNodeType},
		},
	}, &second)

	require.Len(t, second.Errors, 1)
	assert.Contains(t, second.Errors[0], "earlier")
	assert.Contains(t, second.Errors[0], "alias", "the message should explain aliases are batch-local")

	// The point of the fix: both nodes still exist and are still reported.
	require.Len(t, second.Nodes, 2)
	assert.NotNil(t, inst.Node(second.Nodes["combine"]))
	assert.NotNil(t, inst.Node(second.Nodes["fresh"]))
}

// An entry that fails to be created leaves its alias dangling. Saying that
// beats reporting the alias as an unknown node id, which reads as though
// aliases are broken rather than pointing at the entry that actually failed.
func TestCreateNodesNamesTheFailedEntryBehindADanglingAlias(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "broken", "type": "not.A.Real.NodeType"},
			map[string]any{
				"alias": "combine",
				"type":  combineType,
				"inputs": map[string]any{
					"Meshes": map[string]any{
						"elements": []any{map[string]any{"nodeId": "broken", "port": "Out"}},
					},
				},
			},
		},
	}, &out)

	require.Len(t, out.Errors, 2)
	joined := strings.Join(out.Errors, " | ")
	assert.Contains(t, joined, "entry broken")
	assert.Contains(t, joined, "entry combine")
	assert.Contains(t, joined, "failed to be created")
}

// connect_nodes takes scope at the top level, but every entry of a batch
// looks like it should carry its own - the anglerfish run put it there and
// lost a call to a schema rejection. Accepting it per entry costs nothing
// and makes cross-scope batches expressible.
func TestConnectNodesPerConnectionScope(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var sg polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "part", "name": "Part", "description": "a part",
	}, &sg)

	var inner polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"scope": "part",
		"nodes": []any{
			map[string]any{"alias": "sphere", "type": sphereNodeType},
			map[string]any{"alias": "combine", "type": combineType},
		},
	}, &inner)
	require.Empty(t, inner.Errors)

	var out polyformmcp.ConnectNodesOutput
	callTool(t, session, "connect_nodes", map[string]any{
		"connections": []any{
			map[string]any{
				"outNodeId": inner.Nodes["sphere"], "outPort": "Out",
				"inNodeId": inner.Nodes["combine"], "inPort": "Meshes",
				"scope": "part",
			},
		},
	}, &out)

	require.Empty(t, out.Errors)
	assert.True(t, out.Connected)
	assert.Equal(t, 1, out.Made)
}

// A subgraph isn't a registered node type, but reaching for create_nodes
// with a subgraph id is natural once every other node in the batch is
// created that way. The bare "no factory registered with ID x" gave no
// hint that the id was right and only the tool was wrong.
func TestCreateNodesNamesInstantiateSubgraphForASubgraphId(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var sg polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id":   "fin",
		"name": "Fin",
	}, &sg)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "sphere", "type": sphereNodeType},
			map[string]any{"alias": "theFin", "type": "fin"},
		},
	}, &out)

	require.Len(t, out.Errors, 1)
	assert.Contains(t, out.Errors[0], "instantiate_subgraph")
	assert.Contains(t, out.Errors[0], "subgraph")
	assert.Contains(t, out.Nodes, "sphere", "the rest of the batch still landed")
}

// The same guidance on the single-node tool, where it is a tool error.
func TestCreateNodeNamesInstantiateSubgraphForASubgraphId(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var sg polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{"id": "fin", "name": "Fin"}, &sg)

	msg := callToolExpectingError(t, session, "create_node", map[string]any{"type": "fin"})
	assert.Contains(t, msg, "instantiate_subgraph")
}

// One bad reference used to abort the rest of the entry's wiring, and Go's
// random map iteration meant which ports survived changed between runs. A
// real build lost Mesh and Rotation on seven fin ModelNodes from a single
// stale alias, and the fins simply didn't appear - the error named only
// the bad reference, never the ports it took down with it.
func TestCreateNodesWiresGoodInputsDespiteOneBadReference(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "sphere", "type": sphereNodeType,
				"inputs": map[string]any{"Radius": map[string]any{"value": "2"}}},
			map[string]any{
				"alias": "model",
				"type":  gltfModelType,
				"inputs": map[string]any{
					"Mesh":        map[string]any{"nodeId": "sphere", "port": "Out"},
					"Translation": map[string]any{"value": `{"x":1,"y":2,"z":3}`},
					"Material":    map[string]any{"nodeId": "materialThatDoesNotExist", "port": "Out"},
				},
			},
		},
	}, &out)

	require.Len(t, out.Errors, 1, "the bad reference should be reported")
	assert.Contains(t, out.Errors[0], "Material")
	assert.Contains(t, out.Errors[0], "every other input on this node was wired")

	// The point of the fix: the good inputs actually landed.
	model := inst.Node(out.Nodes["model"])
	require.NotNil(t, model)
	mesh, ok := model.Inputs()["Mesh"].(nodes.SingleValueInputPort)
	require.True(t, ok)
	assert.NotNil(t, mesh.Value(), "Mesh must still be wired despite the bad Material reference")

	translation, ok := model.Inputs()["Translation"].(nodes.SingleValueInputPort)
	require.True(t, ok)
	assert.NotNil(t, translation.Value(), "the literal must still be wired")
}

// Several bad inputs on one node are all named, so fixing them doesn't
// take one round trip each.
func TestCreateNodesReportsEveryBadInputAtOnce(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{
				"alias": "model",
				"type":  gltfModelType,
				"inputs": map[string]any{
					"Mesh":     map[string]any{"nodeId": "missingA", "port": "Out"},
					"Material": map[string]any{"nodeId": "missingB", "port": "Out"},
				},
			},
		},
	}, &out)

	require.Len(t, out.Errors, 1)
	assert.Contains(t, out.Errors[0], "Mesh")
	assert.Contains(t, out.Errors[0], "Material")
}
