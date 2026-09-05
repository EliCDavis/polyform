package mcp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/graph"
	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/EliCDavis/polyform/nodes"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	// Blank-imported so their node types are registered for these tests.
	_ "github.com/EliCDavis/polyform/drawing/coloring"
	_ "github.com/EliCDavis/polyform/drawing/texturing"
	_ "github.com/EliCDavis/polyform/formats/gltf"
	_ "github.com/EliCDavis/polyform/generator/manifest/basics"
	_ "github.com/EliCDavis/polyform/generator/parameter"
	_ "github.com/EliCDavis/polyform/generator/subgraph/register"
	_ "github.com/EliCDavis/polyform/math"
	_ "github.com/EliCDavis/polyform/math/curves"
	_ "github.com/EliCDavis/polyform/math/geometry"
	_ "github.com/EliCDavis/polyform/math/quaternion"
	_ "github.com/EliCDavis/polyform/math/sdf"
	_ "github.com/EliCDavis/polyform/math/sequence"
	_ "github.com/EliCDavis/polyform/math/trig"
	_ "github.com/EliCDavis/polyform/math/trs"
	_ "github.com/EliCDavis/polyform/math/vector3"
	_ "github.com/EliCDavis/polyform/modeling"
	_ "github.com/EliCDavis/polyform/modeling/extrude"
	_ "github.com/EliCDavis/polyform/modeling/marching"
	_ "github.com/EliCDavis/polyform/modeling/primitives"
	_ "github.com/EliCDavis/polyform/modeling/repeat"
)

const (
	cubeNodeType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/primitives.CubeNode]"
	floatParamType = "github.com/EliCDavis/polyform/generator/parameter.Value[float64]"
)

// testSession spins up a polyform-mcp server backed by a fresh graph
// instance and an in-process client connected to it over an in-memory
// transport, so tools can be exercised without going through real stdio.
func testSession(t *testing.T) *mcpsdk.ClientSession {
	t.Helper()
	session, _ := testSessionWithInstance(t)
	return session
}

// testSessionWithInstance is like testSession, but also returns the
// underlying graph.Instance directly — for tests that need to verify
// computed values (e.g. evaluating an equation subgraph numerically),
// which isn't something exposed over the MCP tool surface itself.
func testSessionWithInstance(t *testing.T) (*mcpsdk.ClientSession, *graph.Instance) {
	t.Helper()

	inst := graph.New(graph.Config{
		TypeFactory:     generator.Types(),
		VariableFactory: polyformmcp.NewTypedVariable,
	})
	server := polyformmcp.NewServer(inst)

	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()

	ctx := context.Background()
	_, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)

	t.Cleanup(func() { _ = session.Close() })

	return session, inst
}

// callTool calls a tool and decodes its structured output into out. It
// fails the test if the call errored or the tool itself reported failure.
func callTool(t *testing.T, session *mcpsdk.ClientSession, name string, args map[string]any, out any) {
	t.Helper()

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	require.NoError(t, err)
	if res.IsError {
		var msg string
		for _, c := range res.Content {
			if tc, ok := c.(*mcpsdk.TextContent); ok {
				msg += tc.Text
			}
		}
		t.Fatalf("tool %q reported an error: %s", name, msg)
	}

	if out == nil {
		return
	}

	// StructuredContent may already be json.RawMessage (server-side) or a
	// generically-decoded value (after a real client/server round trip);
	// re-marshaling and unmarshaling handles either case.
	require.NotNil(t, res.StructuredContent, "expected structured content for tool %q", name)
	data, err := json.Marshal(res.StructuredContent)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, out))
}

func TestSearchNodeTypes(t *testing.T) {
	session := testSession(t)

	var out polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{
		"query": "cube",
	}, &out)

	require.NotEmpty(t, out.Results)
	found := false
	for _, r := range out.Results {
		if r.Type == cubeNodeType {
			found = true
		}
	}
	require.True(t, found, "expected search results to include %s", cubeNodeType)
}

func TestSearchNodeTypesMatchesPortNames(t *testing.T) {
	session := testSession(t)

	// "Depth" is one of CubeNode's input port names, not something that
	// appears in a generic display name/path — this only matches at all
	// because search_node_types now searches port names too.
	var out polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{"query": "depth"}, &out)

	found := false
	for _, r := range out.Results {
		if r.Type == cubeNodeType {
			found = true
		}
	}
	require.True(t, found, "expected searching a port name (\"depth\") to surface %s", cubeNodeType)
}

// TestSearchNodeTypesFallsBackToAnyTerm covers the dominant real-world
// search failure: a list of synonyms, hoping one lands. Requiring every
// term makes that a guaranteed zero (no one node contains all four
// words), so the search retries on any term. Query taken verbatim from a
// real session where it returned nothing.
func TestSearchNodeTypesFallsBackToAnyTerm(t *testing.T) {
	session := testSession(t)

	// "wedge" matches nothing in the library (there is no wedge primitive),
	// so no node can satisfy every term and the fallback has to carry this.
	// Deliberately paired with a word that does exist, so the test keeps
	// exercising the fallback even as node descriptions improve.
	var out polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{
		"query": "wedge cylinder",
	}, &out)

	require.NotEmpty(t, out.Results, "a synonym list should find candidates instead of dead-ending")
	require.Equal(t, "any-term", out.MatchMode, "caller must be told the results are looser than asked for")

	// Best-first: the top hit must match a query term in its own name, not
	// merely mention one somewhere in its description or a port name.
	top := strings.ToLower(out.Results[0].Type + " " + out.Results[0].DisplayName)
	require.Contains(t, top, "cylinder",
		"expected the term that actually exists to rank first, got %q", out.Results[0].Type)
}

// TestSearchNodeTypesFindsTrig is the other verbatim query from that
// session - it wanted an arctangent node and got nothing, both because
// "angle" never co-occurred with "atan" anywhere and because no scalar
// arctangent node existed to find. Either route (all-terms now that the
// scalar nodes describe themselves in both spellings, or the any-term
// fallback) is fine; what matters is that atan2 surfaces.
func TestSearchNodeTypesFindsTrig(t *testing.T) {
	session := testSession(t)

	var out polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{
		"query":      "atan arctan angle",
		"pathPrefix": "math",
	}, &out)

	found := map[string]bool{}
	for _, r := range out.Results {
		found[r.Type] = true
	}
	require.True(t, found["github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trig.ArcTan2Node]"],
		"expected the scalar atan2 node to surface for this query, got %+v", out.Results)
}

// TestSearchNodeTypesPrefersAllTerms confirms the fallback is only a
// fallback - a query whose terms all match one node keeps the tighter
// result set rather than being widened.
func TestSearchNodeTypesPrefersAllTerms(t *testing.T) {
	session := testSession(t)

	var out polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{"query": "cube"}, &out)

	require.NotEmpty(t, out.Results)
	require.Empty(t, out.MatchMode, "an ordinary match must not be reported as a widened one")
}

func TestSearchNodeTypesSingleTermMissStillReturnsNothing(t *testing.T) {
	session := testSession(t)

	var out polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{"query": "zzzznotarealnode"}, &out)

	require.Empty(t, out.Results, "one term that matches nothing has nothing looser to fall back to")
	require.Equal(t, 0, out.TotalMatches)
	require.Empty(t, out.MatchMode)
}

// TestSearchFindsNodesByIntent replays queries that came back empty in real
// sessions. Each one is a case where the capability existed but was
// reachable only by guessing its literal type name - the node library was
// two thirds undescribed, so search had nothing but type keys to match on.
func TestSearchFindsNodesByIntent(t *testing.T) {
	session := testSession(t)

	cases := []struct {
		query string
		want  string
	}{
		// Guitar session: searched four times for a way to move a mesh.
		{"transform mesh", "modeling/meshops.TransformNode"},
		// Guitar session: "Append" and "Combine merge" both came up short.
		{"append meshes", "modeling/meshops.CombineNode"},
		// Flashlight session: needed normals on a cone, found nothing.
		{"faceted hard normals", "modeling/meshops.FlatNormalsNode"},
		// The lathe that was invisible because it's named Screw.
		{"lathe", "modeling/extrude.ScrewNode"},
		// Tank session: wanted a wedge; a low-sided cone is the closest thing.
		{"pyramid spike", "modeling/primitives.ConeNode"},
		{"dome bowl", "modeling/primitives.HemisphereNode"},
		{"doughnut ring", "modeling/primitives.TorusNode"},
		// Diving helmet session: wanted a washer and fell back to a
		// Cylinder with both caps off, which has no radial thickness and
		// renders as a hairline edge-on.
		{"washer annulus", "modeling/primitives.TubeNode"},
		{"hollow cylinder pipe", "modeling/primitives.TubeNode"},
		// Armchair session: three separate searches for a way to derive a
		// darker shade from a color variable, all of which came back empty.
		{"lighten darken shade tint", "drawing/coloring.AdjustHSVNode"},
		{"hsv", "drawing/coloring.AdjustHSVNode"},
		{"color adjust brightness", "drawing/coloring.AdjustHSVNode"},
		{"brightness", "drawing/coloring.BrightnessNode"},
		// Armchair session: needed a March domain that tracked its part's
		// size variables. Searching "aabb" found only the literal parameter
		// node, so the log showed a successful search and the gap was
		// invisible until the agent reported it.
		{"aabb from center size", "math/geometry.AABBNode"},
		{"bounding box construct", "math/geometry.AABBNode"},
	}

	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			var out polyformmcp.SearchNodeTypesOutput
			callTool(t, session, "search_node_types", map[string]any{"query": tc.query}, &out)

			for _, r := range out.Results {
				if strings.Contains(r.Type, tc.want) {
					return
				}
			}
			t.Fatalf("searching %q did not surface %s; got %d results", tc.query, tc.want, len(out.Results))
		})
	}
}

// TestSearchNodeTypesRegexIsCaseInsensitive pins the diving-helmet
// session's dead end. Every string a regex matches against is a Go
// identifier, so a lowercase pattern used to match nothing at all: both
// "torus" and "multiply" returned zero under regex:true while the very
// same word returned results as a plain term.
func TestSearchNodeTypesRegexIsCaseInsensitive(t *testing.T) {
	session := testSession(t)

	for _, tc := range []struct{ query, want string }{
		{"torus", "modeling/primitives.TorusNode"},
		{"multiply", "math.MultiplyNode"},
	} {
		t.Run(tc.query, func(t *testing.T) {
			var out polyformmcp.SearchNodeTypesOutput
			callTool(t, session, "search_node_types", map[string]any{
				"query": tc.query,
				"regex": true,
			}, &out)

			require.Empty(t, out.MatchMode, "a lowercase regex should match directly, not fall back")
			for _, r := range out.Results {
				if strings.Contains(r.Type, tc.want) {
					return
				}
			}
			t.Fatalf("regex %q did not surface %s; got %d results", tc.query, tc.want, len(out.Results))
		})
	}
}

// TestSearchNodeTypesRegexFallsBackToTerms covers the other half of that
// session: "torus disc" sent with regex:true is a literal with a space in
// it and can never match, even case-insensitively. Rather than dead-end,
// the query is retried as ordinary terms.
func TestSearchNodeTypesRegexFallsBackToTerms(t *testing.T) {
	session := testSession(t)

	var out polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{
		"query": "torus disc",
		"regex": true,
	}, &out)

	require.NotEmpty(t, out.Results, "a regex that matches nothing should retry as terms")
	require.Contains(t, []string{"substring", "any-term"}, out.MatchMode,
		"the loosened match must be reported so the caller knows to check it")

	for _, r := range out.Results {
		if strings.Contains(r.Type, "modeling/primitives.TorusNode") {
			return
		}
	}
	t.Fatalf("fallback did not surface TorusNode; got %d results", len(out.Results))
}

// TestSearchNodeTypesPathPrefixMath guards the filter the diving-helmet
// report suspected. It was not at fault - the zero result there came from
// regex case sensitivity - but nothing pinned math's one-segment path.
func TestSearchNodeTypesPathPrefixMath(t *testing.T) {
	session := testSession(t)

	var out polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{
		"query":      "multiply",
		"pathPrefix": "math",
	}, &out)

	for _, r := range out.Results {
		if strings.Contains(r.Type, "math.MultiplyNode") {
			return
		}
	}
	t.Fatalf("pathPrefix 'math' hid math.MultiplyNode; got %d results", len(out.Results))
}

const cylinderNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/primitives.CylinderNode]"

func TestSearchNodeTypesRegexAlternation(t *testing.T) {
	session := testSession(t)

	var out polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{
		"query": "CubeNode|CylinderNode",
		"regex": true,
	}, &out)

	found := map[string]bool{}
	for _, r := range out.Results {
		found[r.Type] = true
	}
	require.True(t, found[cubeNodeType], "expected regex alternation to match %s", cubeNodeType)
	require.True(t, found[cylinderNodeType], "expected regex alternation to match %s", cylinderNodeType)
}

func TestSearchNodeTypesRegexMatchesTypeKey(t *testing.T) {
	session := testSession(t)

	// Only findable because the type key itself is part of the search
	// haystack — "primitives.CubeNode]" appears nowhere else (display
	// name/path/description/port names don't spell out the Go type name
	// with its package qualifier and closing bracket like this).
	var out polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{
		"query": `primitives\.CubeNode\]`,
		"regex": true,
	}, &out)

	require.Len(t, out.Results, 1)
	require.Equal(t, cubeNodeType, out.Results[0].Type)
}

func TestSearchNodeTypesRegexCaseInsensitiveByDefault(t *testing.T) {
	session := testSession(t)

	for _, query := range []string{"CUBENODE", "cubenode", "(?i)CubeNode"} {
		var out polyformmcp.SearchNodeTypesOutput
		callTool(t, session, "search_node_types", map[string]any{"query": query, "regex": true}, &out)

		found := false
		for _, r := range out.Results {
			if r.Type == cubeNodeType {
				found = true
			}
		}
		require.True(t, found, "regex %q should match CubeNode regardless of case", query)
		require.Empty(t, out.MatchMode, "regex %q should match directly, not fall back", query)
	}

	var sensitive polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{"query": "(?-i)CUBENODE", "regex": true}, &sensitive)
	for _, r := range sensitive.Results {
		require.NotEqual(t, cubeNodeType, r.Type, "(?-i) should restore case sensitivity")
	}
}

func TestSearchNodeTypesRegexInvalidPatternIsToolError(t *testing.T) {
	session := testSession(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "search_node_types",
		Arguments: map[string]any{
			"query": "(unclosed",
			"regex": true,
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError, "an invalid regex should surface as a tool error, not crash the server")
}

func TestSearchNodeTypesResultsAreLightweight(t *testing.T) {
	session := testSession(t)

	var out polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{"query": "cube"}, &out)

	require.NotEmpty(t, out.Results)
	for _, r := range out.Results {
		if r.Type == cubeNodeType {
			require.Greater(t, r.InputCount, 0)
		}
	}

	// Results should decode with no "inputs"/"outputs"/"outputCount" keys
	// at all now — confirm by round-tripping through raw JSON rather than
	// just relying on the Go struct shape (which wouldn't catch a stray
	// field left in by an unrelated future edit).
	raw, err := json.Marshal(out.Results)
	require.NoError(t, err)
	require.NotContains(t, string(raw), `"outputCount"`)
	require.NotContains(t, string(raw), `"inputs"`)
	require.NotContains(t, string(raw), `"outputs"`)
}

func TestGetNodeTypes(t *testing.T) {
	session := testSession(t)

	var out polyformmcp.GetNodeTypesOutput
	callTool(t, session, "get_node_types", map[string]any{
		"types": []string{cubeNodeType},
	}, &out)

	require.Empty(t, out.NotFound)
	require.Len(t, out.Results, 1)

	detail := out.Results[0]
	require.Equal(t, cubeNodeType, detail.Type)

	names := map[string]bool{}
	for _, p := range detail.Inputs {
		names[p.Name] = true
	}
	require.True(t, names["Width"])
	require.True(t, names["Height"])
	require.True(t, names["Depth"])

	outNames := map[string]bool{}
	for _, p := range detail.Outputs {
		outNames[p.Name] = true
	}
	require.True(t, outNames["Out"])
}

func TestGetNodeTypesReportsNotFound(t *testing.T) {
	session := testSession(t)

	var out polyformmcp.GetNodeTypesOutput
	callTool(t, session, "get_node_types", map[string]any{
		"types": []string{cubeNodeType, "not-a-real-type-key"},
	}, &out)

	require.Len(t, out.Results, 1)
	require.Equal(t, []string{"not-a-real-type-key"}, out.NotFound)
}

func TestCreateAndConnectNodes(t *testing.T) {
	session := testSession(t)

	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &cube)
	require.NotEmpty(t, cube.NodeId)

	var width polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": floatParamType}, &width)
	require.NotEmpty(t, width.NodeId)

	var setParam polyformmcp.SetParameterOutput
	callTool(t, session, "set_parameter", map[string]any{
		"nodeId": width.NodeId,
		"value":  "3",
	}, &setParam)
	require.True(t, setParam.Updated)

	var connect polyformmcp.ConnectNodesOutput
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": width.NodeId,
		"outPort":   "Value",
		"inNodeId":  cube.NodeId,
		"inPort":    "Width",
	}, &connect)
	require.True(t, connect.Connected)

	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{}, &desc)
	require.Len(t, desc.Nodes, 2)

	var disc polyformmcp.DisconnectOutput
	callTool(t, session, "disconnect", map[string]any{
		"nodeId": cube.NodeId,
		"port":   "Width",
	}, &disc)
	require.True(t, disc.Disconnected)

	var del polyformmcp.DeleteNodeOutput
	callTool(t, session, "delete_node", map[string]any{"nodeId": cube.NodeId}, &del)
	require.True(t, del.Deleted)
}

// TestConnectNodesRejectsSelfCycle covers a silent footgun: connecting a
// node's own output back into its own array input succeeded and created a
// loop with no error at all.
func TestConnectNodesRejectsSelfCycle(t *testing.T) {
	session := testSession(t)

	var mul polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.MultiplyNode[float64]]",
	}, &mul)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "connect_nodes",
		Arguments: map[string]any{
			"outNodeId": mul.NodeId, "outPort": "Float",
			"inNodeId": mul.NodeId, "inPort": "Values",
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError, "a node feeding its own input should be refused, not silently accepted")
}

// TestConnectNodesRejectsIndirectCycle covers the same thing one hop out:
// A feeds B, so B must not be allowed to feed A.
func TestConnectNodesRejectsIndirectCycle(t *testing.T) {
	session := testSession(t)

	var a, b polyformmcp.CreateNodeOutput
	mulType := "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.MultiplyNode[float64]]"
	callTool(t, session, "create_node", map[string]any{"type": mulType}, &a)
	callTool(t, session, "create_node", map[string]any{"type": mulType}, &b)

	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": a.NodeId, "outPort": "Float",
		"inNodeId": b.NodeId, "inPort": "Values",
	}, &polyformmcp.ConnectNodesOutput{})

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "connect_nodes",
		Arguments: map[string]any{
			"outNodeId": b.NodeId, "outPort": "Float",
			"inNodeId": a.NodeId, "inPort": "Values",
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError, "closing a loop back to an upstream node should be refused")
}

// TestCreateNodeWiresArrayElementsInOneCall covers the ergonomics gap that
// made every "literal times variable" multiply cost a create_node plus a
// follow-up connect_nodes.
func TestCreateNodeWiresArrayElementsInOneCall(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	callTool(t, session, "create_variable", map[string]any{
		"path": "Radius", "type": "float64", "value": "4",
	}, &polyformmcp.CreateVariableOutput{})

	var mul polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.MultiplyNode[float64]]",
		"inputs": map[string]any{
			"Values": map[string]any{
				"elements": []map[string]any{
					{"variable": "Radius"},
					{"value": "0.25"},
				},
			},
		},
	}, &mul)

	require.InDelta(t, 1.0, evalFloat64Output(t, inst, mul.NodeId, "Float"), 1e-9,
		"both array elements should have been wired: 4 * 0.25")
}

func TestCreateNodeElementsRejectsNonArrayPort(t *testing.T) {
	session := testSession(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_node",
		Arguments: map[string]any{
			"type": subtractNodeType,
			"inputs": map[string]any{
				"A": map[string]any{"elements": []map[string]any{{"value": "1"}}},
			},
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError, "elements on a single-value port should be rejected")
}

func TestConnectNodesInvalidPortReturnsToolError(t *testing.T) {
	session := testSession(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "connect_nodes",
		Arguments: map[string]any{
			"outNodeId": "does-not-exist",
			"outPort":   "Out",
			"inNodeId":  "also-missing",
			"inPort":    "In",
		},
	})
	require.NoError(t, err) // transport-level call succeeds
	require.True(t, res.IsError, "expected a bad connect_nodes call to surface as a tool error, not crash the server")
}

func TestSubgraphComposition(t *testing.T) {
	session := testSession(t)

	var sg polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id":          "wheel",
		"name":        "Wheel",
		"description": "A single wheel",
	}, &sg)
	require.Equal(t, "wheel", sg.Id)

	var inputBoundary polyformmcp.CreateBoundaryNodeOutput
	callTool(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "wheel",
		"kind":       "input",
		"portType":   "float64",
		"name":       "Radius",
	}, &inputBoundary)
	require.NotEmpty(t, inputBoundary.NodeId)

	var outputBoundary polyformmcp.CreateBoundaryNodeOutput
	callTool(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "wheel",
		"kind":       "output",
		"portType":   "float64",
		"name":       "Result",
	}, &outputBoundary)
	require.NotEmpty(t, outputBoundary.NodeId)

	// Wire the subgraph's own input straight through to its output.
	var wire polyformmcp.ConnectNodesOutput
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": inputBoundary.NodeId,
		"outPort":   "Value",
		"inNodeId":  outputBoundary.NodeId,
		"inPort":    "Value",
		"scope":     "wheel",
	}, &wire)
	require.True(t, wire.Connected)

	var list polyformmcp.ListSubgraphsOutput
	callTool(t, session, "list_subgraphs", map[string]any{}, &list)
	require.Len(t, list.Subgraphs, 1)
	require.Equal(t, "wheel", list.Subgraphs[0].Id)

	// Place two instances in the root graph, the way "4 wheels on a car"
	// would be composed from repeated subgraph instances.
	var inst1, inst2 polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "wheel"}, &inst1)
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "wheel"}, &inst2)
	require.NotEmpty(t, inst1.NodeId)
	require.NotEmpty(t, inst2.NodeId)
	require.NotEqual(t, inst1.NodeId, inst2.NodeId)

	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{}, &desc)
	require.Len(t, desc.Nodes, 2)
}

func TestSetGraphInfo(t *testing.T) {
	session := testSession(t)

	var set polyformmcp.SetGraphInfoOutput
	callTool(t, session, "set_graph_info", map[string]any{
		"name":        "Cat",
		"description": "A small procedural cat",
		"version":     "0.1.0",
	}, &set)
	require.Equal(t, "Cat", set.Name)
	require.Equal(t, "A small procedural cat", set.Description)
	require.Equal(t, "0.1.0", set.Version)

	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{}, &desc)
	require.Equal(t, "Cat", desc.Name)
	require.Equal(t, "A small procedural cat", desc.Description)
	require.Equal(t, "0.1.0", desc.Version)

	// Omitted fields leave the existing value alone.
	var partial polyformmcp.SetGraphInfoOutput
	callTool(t, session, "set_graph_info", map[string]any{"version": "0.2.0"}, &partial)
	require.Equal(t, "Cat", partial.Name)
	require.Equal(t, "A small procedural cat", partial.Description)
	require.Equal(t, "0.2.0", partial.Version)
}

func TestSaveAndLoadGraph(t *testing.T) {
	session := testSession(t)

	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &cube)

	callTool(t, session, "set_graph_info", map[string]any{"name": "Test Graph"}, &polyformmcp.SetGraphInfoOutput{})

	path := filepath.Join(t.TempDir(), "graph.json")

	var save polyformmcp.SaveGraphOutput
	callTool(t, session, "save_graph", map[string]any{"path": path}, &save)
	require.Equal(t, path, save.Path)

	_, err := os.Stat(path)
	require.NoError(t, err)

	// Loading into the same running server should replace its graph
	// without error.
	var load polyformmcp.LoadGraphOutput
	callTool(t, session, "load_graph", map[string]any{"path": path}, &load)
	require.True(t, load.Loaded)

	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{}, &desc)
	require.Len(t, desc.Nodes, 1)
	require.Equal(t, "Test Graph", desc.Name)
}

func TestSetProducerAndGenerate(t *testing.T) {
	session := testSession(t)

	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &cube)

	var model polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ModelNode]",
	}, &model)

	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ManifestNode]",
	}, &manifest)

	var connectMesh polyformmcp.ConnectNodesOutput
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": cube.NodeId, "outPort": "Out",
		"inNodeId": model.NodeId, "inPort": "Mesh",
	}, &connectMesh)
	require.True(t, connectMesh.Connected)

	var connectModel polyformmcp.ConnectNodesOutput
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": model.NodeId, "outPort": "Out",
		"inNodeId": manifest.NodeId, "inPort": "Models.0",
	}, &connectModel)
	require.True(t, connectModel.Connected)

	var producer polyformmcp.SetProducerOutput
	callTool(t, session, "set_producer", map[string]any{
		"nodeId": manifest.NodeId,
		"port":   "Out",
		"name":   "cube.glb",
	}, &producer)
	require.Equal(t, "cube.glb", producer.Name)

	outDir := t.TempDir()
	var gen polyformmcp.GenerateOutput
	callTool(t, session, "generate", map[string]any{"outputDir": outDir}, &gen)
	require.NotEmpty(t, gen.Files)

	var mermaid polyformmcp.RenderMermaidOutput
	callTool(t, session, "render_mermaid", map[string]any{}, &mermaid)
	require.NotEmpty(t, mermaid.Mermaid)
}

func TestVariables(t *testing.T) {
	session := testSession(t)

	var created polyformmcp.CreateVariableOutput
	callTool(t, session, "create_variable", map[string]any{
		"path":        "Radius",
		"type":        "float64",
		"description": "Radius of the wheel",
		"value":       "2.5",
	}, &created)
	require.Equal(t, "Radius", created.Path)

	var list polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &list)
	require.Len(t, list.Variables, 1)
	require.Equal(t, "Radius", list.Variables[0].Path)
	require.Equal(t, "Radius of the wheel", list.Variables[0].Description)
	require.Equal(t, 2.5, list.Variables[0].Value)

	// A variable's path doubles as a node type key: create_node places a
	// live reference to it anywhere in the graph.
	var ref polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": "Radius"}, &ref)
	require.NotEmpty(t, ref.NodeId)

	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &cube)

	var connect polyformmcp.ConnectNodesOutput
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": ref.NodeId, "outPort": "Value",
		"inNodeId": cube.NodeId, "inPort": "Width",
	}, &connect)
	require.True(t, connect.Connected)

	var update polyformmcp.UpdateVariableOutput
	callTool(t, session, "update_variable", map[string]any{"path": "Radius", "value": "9"}, &update)
	require.True(t, update.Updated)

	list = polyformmcp.ListVariablesOutput{}
	callTool(t, session, "list_variables", map[string]any{}, &list)
	require.Equal(t, 9.0, list.Variables[0].Value)

	var renamed polyformmcp.RenameVariableOutput
	callTool(t, session, "rename_variable", map[string]any{"path": "Radius", "newPath": "WheelRadius"}, &renamed)
	require.Equal(t, "WheelRadius", renamed.Path)

	list = polyformmcp.ListVariablesOutput{}
	callTool(t, session, "list_variables", map[string]any{}, &list)
	require.Equal(t, "WheelRadius", list.Variables[0].Path)
	require.Equal(t, "Radius of the wheel", list.Variables[0].Description, "rename should preserve description when not explicitly overridden")

	var deleted polyformmcp.DeleteVariableOutput
	callTool(t, session, "delete_variable", map[string]any{"path": "WheelRadius"}, &deleted)
	require.True(t, deleted.Deleted)

	// Deleting the variable should also remove the node that referenced it,
	// leaving just the cube.
	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{}, &desc)
	require.Len(t, desc.Nodes, 1)
}

func TestSaveAndLoadGraphWithVariable(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_variable", map[string]any{
		"path": "Height", "type": "int", "value": "5",
	}, &polyformmcp.CreateVariableOutput{})

	path := filepath.Join(t.TempDir(), "graph-with-variable.json")
	var save polyformmcp.SaveGraphOutput
	callTool(t, session, "save_graph", map[string]any{"path": path}, &save)

	var load polyformmcp.LoadGraphOutput
	callTool(t, session, "load_graph", map[string]any{"path": path}, &load)
	require.True(t, load.Loaded)

	var list polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &list)
	require.Len(t, list.Variables, 1)
	require.Equal(t, "Height", list.Variables[0].Path)
	require.Equal(t, 5.0, list.Variables[0].Value)
}

func TestLoadGraphWithRealExampleVariables(t *testing.T) {
	session := testSession(t)

	var load polyformmcp.LoadGraphOutput
	callTool(t, session, "load_graph", map[string]any{
		"path": "../generator/edit/examples/tutorial.json",
	}, &load)
	require.True(t, load.Loaded)

	var list polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &list)

	byPath := map[string]polyformmcp.VariableSummary{}
	for _, v := range list.Variables {
		byPath[v.Path] = v
	}
	require.Equal(t, 15.0, byPath["Brick Count"].Value)
	require.Equal(t, 5.0, byPath["Height"].Value)
	require.Equal(t, 5.0, byPath["Radius"].Value)
}

// evalFloat64Output reads the live computed value of a node's output port
// directly through the graph API — there's no MCP tool to evaluate an
// arbitrary node, so tests that need a numeric answer (not just "the tool
// call didn't error") reach into the instance returned by
// testSessionWithInstance.
func evalFloat64Output(t *testing.T, inst *graph.Instance, nodeID, port string) float64 {
	t.Helper()
	node := inst.Node(nodeID)
	require.NotNil(t, node, "no node with id %q", nodeID)
	out, ok := node.Outputs()[port]
	require.True(t, ok, "node %q has no output port %q", nodeID, port)
	valued, ok := out.(nodes.Output[float64])
	require.True(t, ok, "node %q port %q is not a float64 output", nodeID, port)
	return valued.Value()
}

// literalFloat64Node creates a parameter.Value[float64] node preset to v,
// via the same MCP tools an agent would use.
func literalFloat64Node(t *testing.T, session *mcpsdk.ClientSession, v float64) string {
	t.Helper()
	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": floatParamType}, &created)
	var set polyformmcp.SetParameterOutput
	callTool(t, session, "set_parameter", map[string]any{
		"nodeId": created.NodeId,
		"value":  fmt.Sprintf("%v", v),
	}, &set)
	require.True(t, set.Updated)
	return created.NodeId
}

func TestCreateEquationSubgraphHypotenuse(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var eq polyformmcp.CreateEquationSubgraphOutput
	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id":       "hypotenuse",
		"equation": "c = sqrt(a^2 + b^2)",
	}, &eq)
	require.Equal(t, "hypotenuse", eq.SubgraphId)
	require.Equal(t, []string{"a", "b"}, eq.Inputs)
	require.Equal(t, "c", eq.Output)

	var instantiated polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "hypotenuse"}, &instantiated)
	require.NotEmpty(t, instantiated.NodeId)

	aID := literalFloat64Node(t, session, 3)
	bID := literalFloat64Node(t, session, 4)

	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": aID, "outPort": "Value",
		"inNodeId": instantiated.NodeId, "inPort": "a",
	}, &polyformmcp.ConnectNodesOutput{})
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": bID, "outPort": "Value",
		"inNodeId": instantiated.NodeId, "inPort": "b",
	}, &polyformmcp.ConnectNodesOutput{})

	got := evalFloat64Output(t, inst, instantiated.NodeId, "c")
	require.InDelta(t, 5.0, got, 1e-9)
}

func TestCreateEquationSubgraphDedupesRepeatedVariable(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	// x appears twice; a, b, c once each. Expect exactly one boundary
	// input per distinct name, in first-appearance order.
	var eq polyformmcp.CreateEquationSubgraphOutput
	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id":       "quadratic_term",
		"equation": "y = a*x^2 + b*x + c",
	}, &eq)
	require.Equal(t, []string{"a", "x", "b", "c"}, eq.Inputs)

	var instantiated polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "quadratic_term"}, &instantiated)

	inputs := map[string]float64{"a": 2, "x": 3, "b": 4, "c": 5}
	for name, v := range inputs {
		id := literalFloat64Node(t, session, v)
		callTool(t, session, "connect_nodes", map[string]any{
			"outNodeId": id, "outPort": "Value",
			"inNodeId": instantiated.NodeId, "inPort": name,
		}, &polyformmcp.ConnectNodesOutput{})
	}

	got := evalFloat64Output(t, inst, instantiated.NodeId, "y")
	require.InDelta(t, 2*3*3+4*3+5, got, 1e-9) // a*x^2 + b*x + c = 2*9+12+5 = 35
}

func TestCreateEquationSubgraphNegativeAndFractionalPowers(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var eq polyformmcp.CreateEquationSubgraphOutput
	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id":       "inv_sqrt",
		"equation": "y = x^-1 + x^0.5",
	}, &eq)
	require.Equal(t, []string{"x"}, eq.Inputs)

	var instantiated polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "inv_sqrt"}, &instantiated)

	xID := literalFloat64Node(t, session, 4)
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": xID, "outPort": "Value",
		"inNodeId": instantiated.NodeId, "inPort": "x",
	}, &polyformmcp.ConnectNodesOutput{})

	got := evalFloat64Output(t, inst, instantiated.NodeId, "y")
	require.InDelta(t, 0.25+2.0, got, 1e-9) // 4^-1 + 4^0.5 = 0.25 + 2
}

func TestCreateEquationSubgraphRejectsUnsupportedFunction(t *testing.T) {
	session := testSession(t)

	// sin/cos/tan/atan2 are supported now; abs still has no backing node.
	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_equation_subgraph",
		Arguments: map[string]any{
			"id":       "bad",
			"equation": "y = abs(x)",
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError, "abs() has no backing node and should be rejected, not silently approximated")
}

func TestCreateEquationSubgraphRejectsVariableExponent(t *testing.T) {
	session := testSession(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_equation_subgraph",
		Arguments: map[string]any{
			"id":       "bad_exponent",
			"equation": "y = a^b",
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError, "polyform has no general pow(base,exponent) node; a variable exponent should error, not silently misbehave")
}

func TestCreateEquationSubgraphRejectsMalformedEquation(t *testing.T) {
	session := testSession(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_equation_subgraph",
		Arguments: map[string]any{
			"id":       "bad_syntax",
			"equation": "a + b", // missing "output = "
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError)
}

func TestCreateEquationSubgraphPrecedenceAndAssociativity(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	cases := []struct {
		name     string
		equation string
		inputs   map[string]float64
		want     float64
	}{
		{"mul_before_add", "y = a + b * c", map[string]float64{"a": 2, "b": 3, "c": 4}, 14},    // 2 + 3*4
		{"parens_override", "y = (a + b) * c", map[string]float64{"a": 2, "b": 3, "c": 4}, 20}, // (2+3)*4
		{"pow_right_assoc", "y = a ^ 2 ^ 3", map[string]float64{"a": 2}, 256},                  // right-assoc: a^(2^3) = 2^8 = 256, not (a^2)^3 = 64
		{"sub_left_assoc", "y = a - b - c", map[string]float64{"a": 10, "b": 3, "c": 2}, 5},    // (10-3)-2
		{"div_left_assoc", "y = a / b / c", map[string]float64{"a": 100, "b": 5, "c": 2}, 10},  // (100/5)/2
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := fmt.Sprintf("precedence_%d", i)
			var eq polyformmcp.CreateEquationSubgraphOutput
			callTool(t, session, "create_equation_subgraph", map[string]any{
				"id": id, "equation": tc.equation,
			}, &eq)

			var instantiated polyformmcp.InstantiateSubgraphOutput
			callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": id}, &instantiated)

			for name, v := range tc.inputs {
				litID := literalFloat64Node(t, session, v)
				callTool(t, session, "connect_nodes", map[string]any{
					"outNodeId": litID, "outPort": "Value",
					"inNodeId": instantiated.NodeId, "inPort": name,
				}, &polyformmcp.ConnectNodesOutput{})
			}

			got := evalFloat64Output(t, inst, instantiated.NodeId, "y")
			require.InDeltaf(t, tc.want, got, 1e-9, "equation %q", tc.equation)
		})
	}
}

func TestCreateEquationSubgraphMinMaxAndConstants(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var eq polyformmcp.CreateEquationSubgraphOutput
	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id":       "circle_stuff",
		"equation": "y = max(a, b, c) - min(a, b, c) + pi",
	}, &eq)
	require.ElementsMatch(t, []string{"a", "b", "c"}, eq.Inputs)

	var instantiated polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "circle_stuff"}, &instantiated)

	for name, v := range map[string]float64{"a": 5, "b": 1, "c": 9} {
		litID := literalFloat64Node(t, session, v)
		callTool(t, session, "connect_nodes", map[string]any{
			"outNodeId": litID, "outPort": "Value",
			"inNodeId": instantiated.NodeId, "inPort": name,
		}, &polyformmcp.ConnectNodesOutput{})
	}

	got := evalFloat64Output(t, inst, instantiated.NodeId, "y")
	require.InDelta(t, (9.0-1.0)+math.Pi, got, 1e-9)
}

func TestCreateEquationSubgraphUppercaseEIsAVariableNotEuler(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	// Lowercase "e" is Euler's number; "E" is not — it should be treated
	// as an ordinary variable (e.g. Young's modulus), not silently
	// collide with the constant.
	var eq polyformmcp.CreateEquationSubgraphOutput
	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id":       "stress",
		"equation": "y = E * strain",
	}, &eq)
	require.ElementsMatch(t, []string{"E", "strain"}, eq.Inputs)

	var instantiated polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "stress"}, &instantiated)

	for name, v := range map[string]float64{"E": 200, "strain": 0.5} {
		litID := literalFloat64Node(t, session, v)
		callTool(t, session, "connect_nodes", map[string]any{
			"outNodeId": litID, "outPort": "Value",
			"inNodeId": instantiated.NodeId, "inPort": name,
		}, &polyformmcp.ConnectNodesOutput{})
	}

	got := evalFloat64Output(t, inst, instantiated.NodeId, "y")
	require.InDelta(t, 100.0, got, 1e-9)
}

// TestCreateEquationSubgraphAtan2 covers the case this was added for: a
// sloped plate (a tank's glacis, a roof pitch, a ramp) whose length was
// already parametric via hypot() while its matching angle had to be
// hand-computed and frozen as a radian literal, so the two silently
// disagreed as soon as a dimension changed.
func TestCreateEquationSubgraphAtan2(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var eq polyformmcp.CreateEquationSubgraphOutput
	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id":       "glacis_angle",
		"equation": "angle = atan2(rise, run)",
	}, &eq)
	require.Equal(t, []string{"rise", "run"}, eq.Inputs)

	var instantiated polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "glacis_angle"}, &instantiated)

	riseID := literalFloat64Node(t, session, 1)
	runID := literalFloat64Node(t, session, 1)
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": riseID, "outPort": "Value",
		"inNodeId": instantiated.NodeId, "inPort": "rise",
	}, &polyformmcp.ConnectNodesOutput{})
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": runID, "outPort": "Value",
		"inNodeId": instantiated.NodeId, "inPort": "run",
	}, &polyformmcp.ConnectNodesOutput{})

	require.InDelta(t, math.Pi/4, evalFloat64Output(t, inst, instantiated.NodeId, "angle"), 1e-9)
}

// TestCreateEquationSubgraphAtan2KeepsQuadrant confirms atan2 resolves the
// full circle rather than collapsing to atan's [-pi/2, pi/2] - the whole
// reason a rise/run angle needs two arguments instead of one ratio.
func TestCreateEquationSubgraphAtan2KeepsQuadrant(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id":       "back_slope",
		"equation": "angle = atan2(rise, run)",
	}, &polyformmcp.CreateEquationSubgraphOutput{})

	var instantiated polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "back_slope"}, &instantiated)

	riseID := literalFloat64Node(t, session, 1)
	runID := literalFloat64Node(t, session, -1)
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": riseID, "outPort": "Value",
		"inNodeId": instantiated.NodeId, "inPort": "rise",
	}, &polyformmcp.ConnectNodesOutput{})
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": runID, "outPort": "Value",
		"inNodeId": instantiated.NodeId, "inPort": "run",
	}, &polyformmcp.ConnectNodesOutput{})

	// atan(1/-1) would be -pi/4; atan2(1,-1) is 3pi/4.
	require.InDelta(t, 3*math.Pi/4, evalFloat64Output(t, inst, instantiated.NodeId, "angle"), 1e-9)
}

func TestCreateEquationSubgraphTrigFunctions(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	cases := []struct {
		name     string
		equation string
		input    float64
		want     float64
	}{
		{"sin", "y = sin(x)", math.Pi / 2, 1},
		{"cos", "y = cos(x)", 0, 1},
		{"tan", "y = tan(x)", math.Pi / 4, 1},
		{"asin", "y = asin(x)", 1, math.Pi / 2},
		{"acos", "y = acos(x)", 1, 0},
		{"atan", "y = atan(x)", 1, math.Pi / 4},
		{"arctan_alias", "y = arctan(x)", 1, math.Pi / 4},
		{"radians", "y = radians(x)", 180, math.Pi},
		{"degrees", "y = degrees(x)", math.Pi, 180},
		{"composed", "y = degrees(atan(x))", 1, 45},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := fmt.Sprintf("trig_%d", i)
			callTool(t, session, "create_equation_subgraph", map[string]any{
				"id": id, "equation": tc.equation,
			}, &polyformmcp.CreateEquationSubgraphOutput{})

			var instantiated polyformmcp.InstantiateSubgraphOutput
			callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": id}, &instantiated)

			xID := literalFloat64Node(t, session, tc.input)
			callTool(t, session, "connect_nodes", map[string]any{
				"outNodeId": xID, "outPort": "Value",
				"inNodeId": instantiated.NodeId, "inPort": "x",
			}, &polyformmcp.ConnectNodesOutput{})

			require.InDeltaf(t, tc.want, evalFloat64Output(t, inst, instantiated.NodeId, "y"), 1e-9, "equation %q", tc.equation)
		})
	}
}

const subtractNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.SubtractNode[float64]]"

func TestCreateNodeWithLiteralInputs(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var sub polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": subtractNodeType,
		"inputs": map[string]any{
			"A": map[string]any{"value": "10"},
			"B": map[string]any{"value": "3"},
		},
	}, &sub)
	require.NotEmpty(t, sub.NodeId)

	// One node call should have produced 3 nodes total: the subtract node
	// plus one literal parameter node per input.
	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{}, &desc)
	require.Len(t, desc.Nodes, 3)

	got := evalFloat64Output(t, inst, sub.NodeId, "Float")
	require.InDelta(t, 7.0, got, 1e-9)
}

func TestCreateNodeWithMixedReferenceAndLiteralInputs(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	aID := literalFloat64Node(t, session, 20)

	var sub polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": subtractNodeType,
		"inputs": map[string]any{
			"A": map[string]any{"nodeId": aID, "port": "Value"},
			"B": map[string]any{"value": "8"},
		},
	}, &sub)

	got := evalFloat64Output(t, inst, sub.NodeId, "Float")
	require.InDelta(t, 12.0, got, 1e-9)
}

func TestCreateNodeInputsRejectsBothNodeIdAndValue(t *testing.T) {
	session := testSession(t)
	aID := literalFloat64Node(t, session, 1)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_node",
		Arguments: map[string]any{
			"type": subtractNodeType,
			"inputs": map[string]any{
				"A": map[string]any{"nodeId": aID, "port": "Value", "value": "5"},
			},
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError, "specifying both nodeId and value for one input should be rejected")
}

func TestCreateNodeInputsRejectsUnknownPort(t *testing.T) {
	session := testSession(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_node",
		Arguments: map[string]any{
			"type": subtractNodeType,
			"inputs": map[string]any{
				"NotAPort": map[string]any{"value": "1"},
			},
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError)
}

func TestCreateNodeInputsRejectsUnsupportedLiteralType(t *testing.T) {
	session := testSession(t)

	// gltf.ModelNode's Rotation input is a quaternion.Quaternion, which
	// has no registered generator/parameter.Value[T] instantiation — the
	// tool should fail clearly rather than silently skip the wiring.
	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_node",
		Arguments: map[string]any{
			"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ModelNode]",
			"inputs": map[string]any{
				"Rotation": map[string]any{"value": `{"x":0,"y":0,"z":0,"w":1}`},
			},
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError)
}

func TestInstantiateSubgraphWithInputs(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id":       "double",
		"equation": "y = x * 2",
	}, &polyformmcp.CreateEquationSubgraphOutput{})

	var instantiated polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "double",
		"inputs": map[string]any{
			"x": map[string]any{"value": "21"},
		},
	}, &instantiated)
	require.NotEmpty(t, instantiated.NodeId)

	// One instantiate_subgraph call should have produced 2 nodes: the
	// subgraph instance plus the literal parameter node for x.
	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{}, &desc)
	require.Len(t, desc.Nodes, 2)

	got := evalFloat64Output(t, inst, instantiated.NodeId, "y")
	require.InDelta(t, 42.0, got, 1e-9)
}

func TestCreateNodeWithVariableInput(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	callTool(t, session, "create_variable", map[string]any{
		"path": "Radius", "type": "float64", "value": "9",
	}, &polyformmcp.CreateVariableOutput{})

	var sub polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": subtractNodeType,
		"inputs": map[string]any{
			"A": map[string]any{"variable": "Radius"},
			"B": map[string]any{"value": "4"},
		},
	}, &sub)

	// One node call should have produced 2 nodes: the subtract node plus
	// the variable-reference node — not a third node for a literal.
	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{}, &desc)
	require.Len(t, desc.Nodes, 3) // subtract + variable reference + literal "B"

	got := evalFloat64Output(t, inst, sub.NodeId, "Float")
	require.InDelta(t, 5.0, got, 1e-9)

	// Updating the variable should flow through to the already-wired node.
	callTool(t, session, "update_variable", map[string]any{"path": "Radius", "value": "20"}, &polyformmcp.UpdateVariableOutput{})
	got = evalFloat64Output(t, inst, sub.NodeId, "Float")
	require.InDelta(t, 16.0, got, 1e-9)
}

func TestCreateNodeVariableInputRejectsUnknownPath(t *testing.T) {
	session := testSession(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_node",
		Arguments: map[string]any{
			"type": subtractNodeType,
			"inputs": map[string]any{
				"A": map[string]any{"variable": "DoesNotExist"},
			},
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError)
}

func TestCreateVariablesBatch(t *testing.T) {
	session := testSession(t)

	var out polyformmcp.CreateVariablesOutput
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{
			{"path": "Body Color", "type": "coloring.color", "description": "Paint color", "value": `"#cc3333"`},
			{"path": "Wheel Radius", "type": "float64", "value": "0.35"},
			{"path": "Length", "type": "float64"},
		},
	}, &out)
	require.Equal(t, []string{"Body Color", "Wheel Radius", "Length"}, out.Paths)

	var list polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &list)
	require.Len(t, list.Variables, 3)

	byPath := map[string]polyformmcp.VariableSummary{}
	for _, v := range list.Variables {
		byPath[v.Path] = v
	}
	require.Equal(t, "Paint color", byPath["Body Color"].Description)
	require.Equal(t, 0.35, byPath["Wheel Radius"].Value)
	require.Equal(t, 0.0, byPath["Length"].Value) // zero value when omitted
}

func TestCreateVariablesBatchStopsAtFirstError(t *testing.T) {
	session := testSession(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_variables",
		Arguments: map[string]any{
			"variables": []map[string]any{
				{"path": "Good", "type": "float64", "value": "1"},
				{"path": "Bad", "type": "not-a-real-type"},
				{"path": "Never Reached", "type": "float64"},
			},
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError)

	// "Good" should still have been created even though the batch call as
	// a whole reported an error, matching plain sequential create_variable
	// semantics (no fake transaction).
	var list polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &list)
	require.Len(t, list.Variables, 1)
	require.Equal(t, "Good", list.Variables[0].Path)
}

func TestCreateAndListVariantSet(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_variable", map[string]any{"path": "Wheel Count", "type": "int", "value": "4"}, &polyformmcp.CreateVariableOutput{})
	callTool(t, session, "create_variable", map[string]any{"path": "Body Color", "type": "coloring.color", "value": `"#ff0000"`}, &polyformmcp.CreateVariableOutput{})

	var created polyformmcp.CreateVariantSetOutput
	callTool(t, session, "create_variant_set", map[string]any{
		"name": "Fleet",
		"dimensions": []map[string]any{
			{"path": "Wheel Count", "type": "numericRange", "data": `{"min":2,"max":8,"samples":4}`},
			{"path": "Body Color", "type": "discrete", "data": `{"values":["#ff0000","#00ff00","#0000ff"]}`},
		},
	}, &created)
	require.Equal(t, "Fleet", created.Name)
	require.Equal(t, 12, created.TotalCombinations) // 4 * 3

	var list polyformmcp.ListVariantSetsOutput
	callTool(t, session, "list_variant_sets", map[string]any{}, &list)
	require.Len(t, list.VariantSets, 1)

	set := list.VariantSets[0]
	require.Equal(t, "Fleet", set.Name)
	require.Equal(t, 12, set.TotalCombinations)
	require.Len(t, set.Dimensions, 2)

	byPath := map[string]int{}
	for _, d := range set.Dimensions {
		byPath[d.Path] = d.Count
	}
	require.Equal(t, 4, byPath["Wheel Count"])
	require.Equal(t, 3, byPath["Body Color"])
}

func TestCreateVariantSetReplacesExisting(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_variant_set", map[string]any{
		"name":       "Sizes",
		"dimensions": []map[string]any{{"path": "Scale", "type": "numericRange", "data": `{"min":0,"max":1,"samples":5}`}},
	}, &polyformmcp.CreateVariantSetOutput{})

	var replaced polyformmcp.CreateVariantSetOutput
	callTool(t, session, "create_variant_set", map[string]any{
		"name":       "Sizes",
		"dimensions": []map[string]any{{"path": "Scale", "type": "numericRange", "data": `{"min":0,"max":1,"samples":2}`}},
	}, &replaced)
	require.Equal(t, 2, replaced.TotalCombinations)

	var list polyformmcp.ListVariantSetsOutput
	callTool(t, session, "list_variant_sets", map[string]any{}, &list)
	require.Len(t, list.VariantSets, 1, "creating a variant set with an existing name should replace it, not add a second one")
}

func TestVariantSetVectorAndColorDimensions(t *testing.T) {
	session := testSession(t)

	var created polyformmcp.CreateVariantSetOutput
	callTool(t, session, "create_variant_set", map[string]any{
		"name": "Placement",
		"dimensions": []map[string]any{
			{"path": "Position", "type": "vector3Range", "data": `{"min":{"x":0,"y":0,"z":0},"max":{"x":1,"y":2,"z":3},"samples":5}`},
			{"path": "Tint", "type": "hsvRange", "data": `{"min":{"h":0,"s":0.5,"v":0.5},"max":{"h":360,"s":1,"v":1},"samples":3}`},
		},
	}, &created)
	require.Equal(t, 15, created.TotalCombinations) // 5 * 3

	var list polyformmcp.ListVariantSetsOutput
	callTool(t, session, "list_variant_sets", map[string]any{}, &list)
	require.Len(t, list.VariantSets, 1)
	byPath := map[string]int{}
	for _, d := range list.VariantSets[0].Dimensions {
		byPath[d.Path] = d.Count
	}
	require.Equal(t, 5, byPath["Position"])
	require.Equal(t, 3, byPath["Tint"])
}

func TestRenameAndDeleteVariantSet(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_variant_set", map[string]any{
		"name":       "Old Name",
		"dimensions": []map[string]any{{"path": "Scale", "type": "numericRange", "data": `{"min":0,"max":1,"samples":3}`}},
	}, &polyformmcp.CreateVariantSetOutput{})

	var renamed polyformmcp.RenameVariantSetOutput
	callTool(t, session, "rename_variant_set", map[string]any{"name": "Old Name", "newName": "New Name"}, &renamed)
	require.Equal(t, "New Name", renamed.Name)

	var list polyformmcp.ListVariantSetsOutput
	callTool(t, session, "list_variant_sets", map[string]any{}, &list)
	require.Len(t, list.VariantSets, 1)
	require.Equal(t, "New Name", list.VariantSets[0].Name)

	var deleted polyformmcp.DeleteVariantSetOutput
	callTool(t, session, "delete_variant_set", map[string]any{"name": "New Name"}, &deleted)
	require.True(t, deleted.Deleted)

	list = polyformmcp.ListVariantSetsOutput{}
	callTool(t, session, "list_variant_sets", map[string]any{}, &list)
	require.Empty(t, list.VariantSets)
}

func TestCreateVariantSetRejectsUnknownType(t *testing.T) {
	session := testSession(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_variant_set",
		Arguments: map[string]any{
			"name": "Bad",
			"dimensions": []map[string]any{
				{"path": "Scale", "type": "not-a-real-type", "data": `{}`},
			},
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError)

	var list polyformmcp.ListVariantSetsOutput
	callTool(t, session, "list_variant_sets", map[string]any{}, &list)
	require.Empty(t, list.VariantSets, "an invalid dimension should prevent the set from being created at all")
}

func TestCreateVariantSetRejectsMalformedData(t *testing.T) {
	session := testSession(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_variant_set",
		Arguments: map[string]any{
			"name": "Bad",
			"dimensions": []map[string]any{
				{"path": "Scale", "type": "numericRange", "data": "not json"},
			},
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError)
}

func TestSaveAndLoadGraphWithVariantSet(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_variant_set", map[string]any{
		"name": "Fleet",
		"dimensions": []map[string]any{
			{"path": "Wheel Count", "type": "intRange", "data": `{"min":2,"max":8,"samples":4}`},
			{"path": "Body Color", "type": "discrete", "data": `{"values":["#ff0000","#00ff00"]}`},
		},
	}, &polyformmcp.CreateVariantSetOutput{})

	path := filepath.Join(t.TempDir(), "graph-with-variant-set.json")
	callTool(t, session, "save_graph", map[string]any{"path": path}, &polyformmcp.SaveGraphOutput{})

	var load polyformmcp.LoadGraphOutput
	callTool(t, session, "load_graph", map[string]any{"path": path}, &load)
	require.True(t, load.Loaded)

	var list polyformmcp.ListVariantSetsOutput
	callTool(t, session, "list_variant_sets", map[string]any{}, &list)
	require.Len(t, list.VariantSets, 1)
	require.Equal(t, "Fleet", list.VariantSets[0].Name)
	require.Equal(t, 8, list.VariantSets[0].TotalCombinations) // 4 * 2
}

const textNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/manifest/basics.TextNode]"

func TestRunVariantSweep(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_variable", map[string]any{"path": "Message", "type": "string", "value": `"unset"`}, &polyformmcp.CreateVariableOutput{})

	var ref polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": "Message"}, &ref)

	var text polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   textNodeType,
		"inputs": map[string]any{"In": map[string]any{"nodeId": ref.NodeId, "port": "Value"}},
	}, &text)

	callTool(t, session, "set_producer", map[string]any{"nodeId": text.NodeId, "port": "Out", "name": "output"}, &polyformmcp.SetProducerOutput{})

	callTool(t, session, "create_variant_set", map[string]any{
		"name": "Messages",
		"dimensions": []map[string]any{
			{"path": "Message", "type": "discrete", "data": `{"values":["Hello","World"]}`},
		},
	}, &polyformmcp.CreateVariantSetOutput{})

	outDir := t.TempDir()
	var run polyformmcp.RunVariantSweepOutput
	callTool(t, session, "run_variant_sweep", map[string]any{
		"name":      "Messages",
		"outputDir": outDir,
	}, &run)
	require.Equal(t, 2, run.TotalCombinations)
	require.Len(t, run.Folders, 2)

	hello, err := os.ReadFile(filepath.Join(outDir, "variant-0000", "output", "text.txt"))
	require.NoError(t, err)
	require.Equal(t, "Hello", string(hello))

	world, err := os.ReadFile(filepath.Join(outDir, "variant-0001", "output", "text.txt"))
	require.NoError(t, err)
	require.Equal(t, "World", string(world))
}

// TestRunVariantSweepLeavesGraphAtLastCombination confirms the documented
// caveat in run_variant_sweep's tool description - it applies each profile
// straight to the live graph as it sweeps and doesn't restore the original
// values afterward, so a caller relying on the graph's state post-sweep
// needs to know it ends up at the last combination, not back to normal.
func TestRunVariantSweepLeavesGraphAtLastCombination(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_variable", map[string]any{"path": "Message", "type": "string", "value": `"unset"`}, &polyformmcp.CreateVariableOutput{})
	callTool(t, session, "create_variant_set", map[string]any{
		"name":       "Messages",
		"dimensions": []map[string]any{{"path": "Message", "type": "discrete", "data": `{"values":["Hello","World"]}`}},
	}, &polyformmcp.CreateVariantSetOutput{})

	callTool(t, session, "run_variant_sweep", map[string]any{"name": "Messages", "outputDir": t.TempDir()}, &polyformmcp.RunVariantSweepOutput{})

	var list polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &list)
	require.Equal(t, "World", list.Variables[0].Value)
}

func TestRunVariantSweepOverThresholdRequiresConfirm(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_variable", map[string]any{"path": "Value", "type": "float64"}, &polyformmcp.CreateVariableOutput{})
	callTool(t, session, "create_variant_set", map[string]any{
		"name":       "Huge",
		"dimensions": []map[string]any{{"path": "Value", "type": "numericRange", "data": `{"min":0,"max":1,"samples":1001}`}},
	}, &polyformmcp.CreateVariantSetOutput{})

	outDir := t.TempDir()
	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "run_variant_sweep",
		Arguments: map[string]any{"name": "Huge", "outputDir": outDir},
	})
	require.NoError(t, err)
	require.True(t, res.IsError)

	entries, _ := os.ReadDir(outDir)
	require.Empty(t, entries, "no variants should have been written without confirm")
}

func TestRunVariantSweepCustomWarnThreshold(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_variable", map[string]any{"path": "Value", "type": "float64"}, &polyformmcp.CreateVariableOutput{})
	callTool(t, session, "create_variant_set", map[string]any{
		"name":       "Small",
		"dimensions": []map[string]any{{"path": "Value", "type": "numericRange", "data": `{"min":0,"max":1,"samples":3}`}},
	}, &polyformmcp.CreateVariantSetOutput{})

	outDir := t.TempDir()
	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "run_variant_sweep",
		Arguments: map[string]any{"name": "Small", "outputDir": outDir, "warnThreshold": 2},
	})
	require.NoError(t, err)
	require.True(t, res.IsError, "3 combinations should exceed a warnThreshold of 2")

	var run polyformmcp.RunVariantSweepOutput
	callTool(t, session, "run_variant_sweep", map[string]any{
		"name": "Small", "outputDir": outDir, "warnThreshold": 2, "confirm": true,
	}, &run)
	require.Equal(t, 3, run.TotalCombinations)
}

func TestRunVariantSweepUnknownSetIsToolError(t *testing.T) {
	session := testSession(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "run_variant_sweep",
		Arguments: map[string]any{"name": "does-not-exist", "outputDir": t.TempDir()},
	})
	require.NoError(t, err)
	require.True(t, res.IsError)
}

func decodePNGConfig(path string) (image.Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return image.Config{}, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	return cfg, err
}

// buildCubeManifest builds the same minimal cube -> ModelNode -> ManifestNode
// graph as buildAndGenerateCubeGlb, but returns the ManifestNode's id
// directly instead of running generate — for tools like render_preview that
// rasterize straight from the graph's Models array.
func buildCubeManifest(t *testing.T, session *mcpsdk.ClientSession) string {
	t.Helper()

	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &cube)

	var model polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ModelNode]",
		"inputs": map[string]any{"Mesh": map[string]any{"nodeId": cube.NodeId, "port": "Out"}},
	}, &model)

	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ManifestNode]",
		"inputs": map[string]any{"Models": map[string]any{"nodeId": model.NodeId, "port": "Out"}},
	}, &manifest)

	return manifest.NodeId
}

// TestRenderPreviewExpandsGpuInstances confirms render_preview draws a
// model once per GpuInstances entry instead of once total at the model's
// own TRS - a real gltf.ModelNode feature (and the cheapest way to place
// many copies of one part, per the "never hand-repeat a node structure"
// standing rule) that this rasterizer used to silently ignore entirely,
// rendering only the single un-instanced mesh no matter how many
// instances were set.
func TestRenderPreviewExpandsGpuInstances(t *testing.T) {
	session := testSession(t)

	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &cube)

	var circle polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/repeat.CircleNode]",
		"inputs": map[string]any{"Radius": map[string]any{"value": "2"}, "Times": map[string]any{"value": "6"}},
	}, &circle)

	var model polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ModelNode]",
		"inputs": map[string]any{
			"Mesh":          map[string]any{"nodeId": cube.NodeId, "port": "Out"},
			"Gpu Instances": map[string]any{"nodeId": circle.NodeId, "port": "Out"},
		},
	}, &model)

	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ManifestNode]",
		"inputs": map[string]any{"Models": map[string]any{"nodeId": model.NodeId, "port": "Out"}},
	}, &manifest)

	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId":     manifest.NodeId,
		"outputPath": filepath.Join(t.TempDir(), "gpu-instances.png"),
	}, &out)

	// A cube is 12 triangles; 6 GpuInstances should draw 6 full copies,
	// not 1 - this was the exact bug (silently rendered only 1x12).
	require.Equal(t, 72, out.TriangleCount)
}

// TestRenderPreviewExclude confirms the exclude param drops a specific
// ModelNode's contribution from a single render call without mutating the
// graph - the non-destructive alternative to disconnect/render/reconnect
// for isolating which part is responsible for a visual defect.
func TestRenderPreviewExclude(t *testing.T) {
	session := testSession(t)

	newCubeModel := func() polyformmcp.CreateNodeOutput {
		var cube polyformmcp.CreateNodeOutput
		callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &cube)
		var model polyformmcp.CreateNodeOutput
		callTool(t, session, "create_node", map[string]any{
			"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ModelNode]",
			"inputs": map[string]any{"Mesh": map[string]any{"nodeId": cube.NodeId, "port": "Out"}},
		}, &model)
		return model
	}

	modelA := newCubeModel()
	modelB := newCubeModel()

	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ManifestNode]",
		"inputs": map[string]any{"Models": map[string]any{
			"nodeId": modelA.NodeId, "port": "Out",
		}},
	}, &manifest)
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": modelB.NodeId, "outPort": "Out",
		"inNodeId": manifest.NodeId, "inPort": "Models",
	}, nil)

	var baseline polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId":     manifest.NodeId,
		"outputPath": filepath.Join(t.TempDir(), "both.png"),
	}, &baseline)
	require.Equal(t, 24, baseline.TriangleCount) // two cubes, 12 triangles each

	var excluded polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId":     manifest.NodeId,
		"outputPath": filepath.Join(t.TempDir(), "one-excluded.png"),
		"exclude":    []string{modelB.NodeId},
	}, &excluded)
	require.Equal(t, 12, excluded.TriangleCount) // only modelA's cube left

	// The graph itself must be untouched by exclude - a normal render
	// afterward should still show both cubes.
	var again polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId":     manifest.NodeId,
		"outputPath": filepath.Join(t.TempDir(), "both-again.png"),
	}, &again)
	require.Equal(t, 24, again.TriangleCount)
}

// TestSampleField confirms sample_field evaluates a math/sdf node's field
// at explicit points and returns the real signed distance - inside
// (negative), on the surface (zero), and outside (positive) - without any
// marching or rendering.
func TestSampleField(t *testing.T) {
	session := testSession(t)

	var sphere polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.SphereNode]",
	}, &sphere)

	var out polyformmcp.SampleFieldOutput
	callTool(t, session, "sample_field", map[string]any{
		"nodeId": sphere.NodeId,
		"points": []map[string]any{
			{"x": 0, "y": 0, "z": 0},   // center - default radius 0.5, so 0.5 inside
			{"x": 0.5, "y": 0, "z": 0}, // exactly on the default-radius surface
			{"x": 1, "y": 0, "z": 0},   // outside
		},
	}, &out)

	require.Len(t, out.Samples, 3)
	require.InDelta(t, -0.5, out.Samples[0].Value, 1e-9)
	require.InDelta(t, 0.0, out.Samples[1].Value, 1e-9)
	require.InDelta(t, 0.5, out.Samples[2].Value, 1e-9)
	require.Empty(t, out.Warning, "a healthy field should report no warning")
	for i, s := range out.Samples {
		require.Emptyf(t, s.NonFinite, "sample %d should be a real number", i)
	}
}

// TestSampleFieldReportsNaNInsteadOfFailing covers a real dead end: a field
// that evaluated to NaN made the whole call fail with "json: unsupported
// value: NaN", so the agent investigating broken geometry got nothing back
// and abandoned the investigation. NaN is the finding, not an obstacle to
// reporting one.
func TestSampleFieldReportsNaNInsteadOfFailing(t *testing.T) {
	session := testSession(t)

	// A sphere with a NaN radius poisons every sample taken from it.
	var sphere polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.SphereNode]",
		"inputs": map[string]any{
			"Radius": map[string]any{"nodeId": nanFloatNode(t, session), "port": "Out"},
		},
	}, &sphere)

	var out polyformmcp.SampleFieldOutput
	callTool(t, session, "sample_field", map[string]any{
		"nodeId": sphere.NodeId,
		"points": []map[string]any{{"x": 0, "y": 0, "z": 0}},
	}, &out)

	require.Len(t, out.Samples, 1)
	require.Equal(t, "NaN", out.Samples[0].NonFinite, "expected the NaN to be reported, not swallowed")
	require.NotEmpty(t, out.Warning, "a non-finite result should come with an explanation of what it means")
}

// nanFloatNode builds a node whose float output is NaN. Division by zero
// won't do it - DivideNode guards against that and returns 0 - but the
// square root of a negative number will.
func nanFloatNode(t *testing.T, session *mcpsdk.ClientSession) string {
	t.Helper()

	var sqrt polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.SquareRootNode]",
		"inputs": map[string]any{"In": map[string]any{"value": "-1"}},
	}, &sqrt)
	return sqrt.NodeId
}

// TestRenderPreviewReadsVertexColor confirms render_preview shades a mesh
// from its per-vertex "Color" attribute when present, instead of always
// falling back to the material's flat BaseColorFactor - the standard
// coloring technique for SDF/marched meshes (which have no UVs to
// texture), previously invisible in render_preview's own output.
func TestRenderPreviewReadsVertexColor(t *testing.T) {
	session := testSession(t)

	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": cubeNodeType,
		"inputs": map[string]any{
			"Width": map[string]any{"value": "2"}, "Height": map[string]any{"value": "2"}, "Depth": map[string]any{"value": "2"},
		},
	}, &cube)

	var selectPos polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling.SelectFromMeshNode]",
		"inputs": map[string]any{"Mesh": map[string]any{"nodeId": cube.NodeId, "port": "Out"}},
	}, &selectPos)

	var selectY polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/vector3.SelectArray[float64]]",
		"inputs": map[string]any{"In": map[string]any{"nodeId": selectPos.NodeId, "port": "Position"}},
	}, &selectY)

	var remap polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.RemapToArrayNode[float64]]",
		"inputs": map[string]any{
			"Value": map[string]any{"nodeId": selectY.NodeId, "port": "Y"},
			"InMin": map[string]any{"value": "-1"}, "InMax": map[string]any{"value": "1"},
			"OutMin": map[string]any{"value": "0"}, "OutMax": map[string]any{"value": "1"},
		},
	}, &remap)

	var interp polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/drawing/coloring.InterpolateToArrayNode]",
		"inputs": map[string]any{
			"A": map[string]any{"value": `"#ff0000"`}, "B": map[string]any{"value": `"#0000ff"`},
			"Time": map[string]any{"nodeId": remap.NodeId, "port": "Out"},
		},
	}, &interp)

	var toVec polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/drawing/coloring.ToVectorArrayNode]",
		"inputs": map[string]any{"In": map[string]any{"nodeId": interp.NodeId, "port": "Out"}},
	}, &toVec)

	var setAttr polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling.SetAttribute3DNode]",
		"inputs": map[string]any{
			"Mesh":      map[string]any{"nodeId": cube.NodeId, "port": "Out"},
			"Attribute": map[string]any{"value": `"Color"`},
			"Data":      map[string]any{"nodeId": toVec.NodeId, "port": "Vector 3"},
		},
	}, &setAttr)

	var model polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ModelNode]",
		"inputs": map[string]any{"Mesh": map[string]any{"nodeId": setAttr.NodeId, "port": "Out"}},
	}, &model)

	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ManifestNode]",
		"inputs": map[string]any{"Models": map[string]any{"nodeId": model.NodeId, "port": "Out"}},
	}, &manifest)

	outPath := filepath.Join(t.TempDir(), "vertex-color.png")
	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId":     manifest.NodeId,
		"outputPath": outPath,
		"views":      []map[string]any{{"azimuth": 0, "elevation": 0}}, // straight-on front view
	}, &out)

	f, err := os.Open(outPath)
	require.NoError(t, err)
	defer f.Close()
	img, _, err := image.Decode(f)
	require.NoError(t, err)

	bounds := img.Bounds()
	topR, topG, topB, _ := img.At(bounds.Dx()/2, bounds.Min.Y+5).RGBA()
	bottomR, bottomG, bottomB, _ := img.At(bounds.Dx()/2, bounds.Max.Y-5).RGBA()

	// Top of the cube is blue-shaded (B), bottom is red-shaded (A) - if
	// this were still using the flat material fallback, top and bottom
	// would be identical gray instead of visibly different colors.
	require.Greater(t, bottomR, topR, "bottom of the gradient should be redder than the top")
	require.Greater(t, topB, bottomB, "top of the gradient should be bluer than the bottom")
	_ = topG
	_ = bottomG
}

// TestRenderPreviewSamplesColorTexture confirms render_preview actually
// samples a UV-mapped glTF ColorTexture per-pixel rather than falling back
// to the material's flat BaseColorFactor - the gap flagged in
// topics/texturing-and-color.md before texture support was added.
func TestRenderPreviewSamplesColorTexture(t *testing.T) {
	session := testSession(t)

	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": cubeNodeType,
		"inputs": map[string]any{
			"Width": map[string]any{"value": "2"}, "Height": map[string]any{"value": "2"}, "Depth": map[string]any{"value": "2"},
		},
	}, &cube)

	var uvImage polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/drawing/texturing.DebugUVNode]",
		"inputs": map[string]any{
			"Image Resolution":       map[string]any{"value": "64"},
			"Board Resolution":       map[string]any{"value": "2"},
			"Positive Checker Color": map[string]any{"value": `"#ff0000"`},
			"Negative Checker Color": map[string]any{"value": `"#0000ff"`},
		},
	}, &uvImage)

	var texture polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.TextureNode]",
		"inputs": map[string]any{"Image": map[string]any{"nodeId": uvImage.NodeId, "port": "Result"}},
	}, &texture)

	var material polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.MaterialNode]",
		"inputs": map[string]any{"Color Texture": map[string]any{"nodeId": texture.NodeId, "port": "Out"}},
	}, &material)

	var model polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ModelNode]",
		"inputs": map[string]any{
			"Mesh":     map[string]any{"nodeId": cube.NodeId, "port": "Out"},
			"Material": map[string]any{"nodeId": material.NodeId, "port": "Out"},
		},
	}, &model)

	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ManifestNode]",
		"inputs": map[string]any{"Models": map[string]any{"nodeId": model.NodeId, "port": "Out"}},
	}, &manifest)

	outPath := filepath.Join(t.TempDir(), "textured.png")
	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId":     manifest.NodeId,
		"outputPath": outPath,
		"width":      200,
		"height":     200,
	}, &out)
	require.Equal(t, 12, out.TriangleCount) // one cube

	f, err := os.Open(outPath)
	require.NoError(t, err)
	defer f.Close()
	img, _, err := image.Decode(f)
	require.NoError(t, err)

	// Sample a grid of pixels across the rendered image and count distinct
	// colors. A flat BaseColorFactor fallback (or a solid vertex color)
	// would render the visible cube faces as a small handful of smoothly
	// lit, closely related colors; a sampled checkerboard texture must
	// produce sharply different colors next to each other.
	bounds := img.Bounds()
	seen := map[color.RGBA]bool{}
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 4 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 4 {
			r, g, b, a := img.At(x, y).RGBA()
			seen[color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}] = true
		}
	}
	require.Greater(t, len(seen), 10, "expected a checkerboard-textured surface to show many distinct colors, not a flat fallback")
}

func TestRenderPreview(t *testing.T) {
	session := testSession(t)
	manifestID := buildCubeManifest(t, session)

	outPath := filepath.Join(t.TempDir(), "preview.png")

	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId":     manifestID,
		"outputPath": outPath,
		"width":      100,
		"height":     80,
	}, &out)

	require.Equal(t, outPath, out.Path)
	require.Equal(t, 100, out.Width)
	require.Equal(t, 80, out.Height)
	require.Equal(t, 12, out.TriangleCount) // a cube is 12 triangles
	require.Zero(t, out.Views)              // no views requested -> single default render

	cfg, err := decodePNGConfig(outPath)
	require.NoError(t, err)
	require.Equal(t, 100, cfg.Width)
	require.Equal(t, 80, cfg.Height)
}

func TestRenderPreviewMultiView(t *testing.T) {
	session := testSession(t)
	manifestID := buildCubeManifest(t, session)

	outPath := filepath.Join(t.TempDir(), "preview-grid.png")

	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId":     manifestID,
		"outputPath": outPath,
		"width":      64,
		"height":     48,
		"views": []map[string]any{
			{"name": "front", "azimuth": 0, "elevation": 0},
			{"name": "top", "azimuth": 0, "elevation": 89},
			{"name": "closeup", "azimuth": 30, "elevation": 20, "zoom": 0.2, "target": map[string]any{"x": 0, "y": 0, "z": 0}},
		},
	}, &out)

	require.Equal(t, 3, out.Views)
	require.Equal(t, 2, out.Columns) // ceil(sqrt(3)) = 2
	require.Equal(t, 2, out.Rows)    // ceil(3/2) = 2
	require.Equal(t, 12, out.TriangleCount)

	// Captions are on (every view has a name) -> +20px per row.
	require.Equal(t, 2*64, out.Width)
	require.Equal(t, 2*(48+20), out.Height)

	cfg, err := decodePNGConfig(outPath)
	require.NoError(t, err)
	require.Equal(t, out.Width, cfg.Width)
	require.Equal(t, out.Height, cfg.Height)
}

// TestPortNameResolutionTolerance exercises the real bug found by
// analyzing an actual orchestrator run's call log: port names are the
// CamelCase-to-space-case form of the Go struct field (e.g. the field
// Radius2 is port "Radius 2", ColorTexture is "Color Texture"), and an
// agent guessing the raw field name instead of confirming it via
// get_node_types fails with "no such input port" / "contains no in-port".
// create_node's inputs shortcut, connect_nodes, and disconnect now all
// tolerate this specific class of mistake.
func TestPortNameResolutionTolerance(t *testing.T) {
	session := testSession(t)

	var cube, material, texture polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &cube)
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.MaterialNode]",
	}, &material)
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.TextureNode]",
	}, &texture)

	// connect_nodes: "ColorTexture" (no space) for the real port "Color Texture".
	var connectOut polyformmcp.ConnectNodesOutput
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": texture.NodeId, "outPort": "Out",
		"inNodeId": material.NodeId, "inPort": "ColorTexture",
	}, &connectOut)
	require.True(t, connectOut.Connected)

	// create_node's inputs shortcut: "ColorTexture" again, this time as a
	// map key, plus "Radius2" (real port "Radius 2") on a fresh CylinderNode.
	var model polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ModelNode]",
		"inputs": map[string]any{
			"Mesh":     map[string]any{"nodeId": cube.NodeId, "port": "Out"},
			"Material": map[string]any{"nodeId": material.NodeId, "port": "Out"},
		},
	}, &model)
	require.NotEmpty(t, model.NodeId)

	var cylinder polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/primitives.CylinderNode]",
		"inputs": map[string]any{
			"Radius":  map[string]any{"value": "1"},
			"Radius2": map[string]any{"value": "0.5"},
		},
	}, &cylinder)
	require.NotEmpty(t, cylinder.NodeId)

	// disconnect: same "ColorTexture" spelling.
	var disconnectOut polyformmcp.DisconnectOutput
	callTool(t, session, "disconnect", map[string]any{
		"nodeId": material.NodeId, "port": "ColorTexture",
	}, &disconnectOut)
	require.True(t, disconnectOut.Disconnected)
}

// TestPortNameResolutionStillRejectsBogusNames confirms the tolerance
// above doesn't turn into silently accepting a genuinely wrong port name
// - only a whitespace/case difference from a real port is forgiven.
func TestPortNameResolutionStillRejectsBogusNames(t *testing.T) {
	session := testSession(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_node",
		Arguments: map[string]any{
			"type":   cubeNodeType,
			"inputs": map[string]any{"NotARealPort": map[string]any{"value": "1"}},
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError)
}

func TestEnableCallLog(t *testing.T) {
	inst := graph.New(graph.Config{
		TypeFactory:     generator.Types(),
		VariableFactory: polyformmcp.NewTypedVariable,
	})
	server := polyformmcp.NewServer(inst)

	logPath := filepath.Join(t.TempDir(), "calls.jsonl")
	closeLog, err := server.EnableCallLog(logPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = closeLog() })

	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()
	ctx := context.Background()
	_, connErr := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, connErr)
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, sessErr := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, sessErr)
	t.Cleanup(func() { _ = session.Close() })

	// A real regex search, and a deliberately-broken call, to confirm both
	// the exact arguments (including "regex":true) and error status make
	// it into the log.
	callTool(t, session, "search_node_types", map[string]any{"query": "sphere|cylinder", "regex": true}, &polyformmcp.SearchNodeTypesOutput{})

	_, err = session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      "create_node",
		Arguments: map[string]any{"type": "does-not-exist"},
	})
	require.NoError(t, err) // tool-level failure, not a transport error

	content, err := os.ReadFile(logPath)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	require.Len(t, lines, 2)

	var searchEntry map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &searchEntry))
	require.Equal(t, "search_node_types", searchEntry["tool"])
	args, ok := searchEntry["arguments"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "sphere|cylinder", args["query"])
	require.Equal(t, true, args["regex"])
	require.NotEmpty(t, searchEntry["time"])

	var errorEntry map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &errorEntry))
	require.Equal(t, "create_node", errorEntry["tool"])
	require.Equal(t, true, errorEntry["isError"])
}

func TestRenderPreviewMultiViewNoCaptions(t *testing.T) {
	session := testSession(t)
	manifestID := buildCubeManifest(t, session)

	outPath := filepath.Join(t.TempDir(), "preview-grid-nocaption.png")

	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId":     manifestID,
		"outputPath": outPath,
		"width":      64,
		"height":     48,
		"views": []map[string]any{
			{"azimuth": 0, "elevation": 0},
			{"azimuth": 180, "elevation": 0},
		},
	}, &out)

	require.Equal(t, 2, out.Views)
	require.Equal(t, 2, out.Columns) // ceil(sqrt(2)) = 2
	require.Equal(t, 1, out.Rows)    // ceil(2/2) = 1
	// No view has a name -> no caption strip, exact cell height.
	require.Equal(t, 2*64, out.Width)
	require.Equal(t, 48, out.Height)
}

func autosavePath(dir string) string {
	return filepath.Join(dir, "autosave.json")
}

func TestStartProject_OmittedPathAutoGeneratesUniqueDir(t *testing.T) {
	sessionA := testSession(t)
	sessionB := testSession(t)

	var outA, outB polyformmcp.StartProjectOutput
	callTool(t, sessionA, "start_project", map[string]any{}, &outA)
	callTool(t, sessionB, "start_project", map[string]any{}, &outB)
	t.Cleanup(func() {
		os.RemoveAll(outA.Path)
		os.RemoveAll(outB.Path)
	})

	require.NotEmpty(t, outA.Path)
	require.NotEmpty(t, outB.Path)
	require.NotEqual(t, outA.Path, outB.Path, "two sessions that both omit path must never land in the same directory")
	require.True(t, strings.HasPrefix(outA.Path, polyformmcp.DefaultOutputRoot()), "auto-generated projects should live under DefaultOutputRoot(), got %q", outA.Path)
	require.Empty(t, outA.RecoveredFrom, "a freshly generated unique directory can never have a pre-existing autosave")

	cwd, err := os.Getwd()
	require.NoError(t, err)
	require.False(t, strings.HasPrefix(outA.Path, cwd), "an auto-generated project must never land inside the repo/working directory, got %q under cwd %q", outA.Path, cwd)

	_, statErr := os.Stat(outA.Path)
	require.NoError(t, statErr)
}

func TestStartProject_NoExistingAutosaveIsCleanStart(t *testing.T) {
	session := testSession(t)
	dir := filepath.Join(t.TempDir(), "project")

	var out polyformmcp.StartProjectOutput
	callTool(t, session, "start_project", map[string]any{"path": dir}, &out)

	require.Equal(t, dir, out.Path)
	require.Equal(t, autosavePath(dir), out.AutosavePath)
	require.Empty(t, out.RecoveredFrom)

	_, statErr := os.Stat(dir)
	require.NoError(t, statErr, "start_project should create the directory")
}

func TestAutosave_WritesAfterEveryCall(t *testing.T) {
	session := testSession(t)
	dir := t.TempDir()

	callTool(t, session, "start_project", map[string]any{"path": dir}, &polyformmcp.StartProjectOutput{})

	// start_project's own successful call should already have produced an
	// autosave - the mechanism runs through atomic() unconditionally, not
	// just for tools that obviously mutate the graph.
	_, statErr := os.Stat(autosavePath(dir))
	require.NoError(t, statErr, "expected an autosave right after start_project")

	callTool(t, session, "create_variable", map[string]any{
		"path": "Radius", "type": "float64", "value": "3",
	}, &polyformmcp.CreateVariableOutput{})

	data, err := os.ReadFile(autosavePath(dir))
	require.NoError(t, err)

	restored := graph.New(graph.Config{TypeFactory: generator.Types(), VariableFactory: polyformmcp.NewTypedVariable})
	require.NoError(t, restored.ApplyAppSchema(data))
	require.NotPanics(t, func() { restored.GetVariable("Radius") }, "expected the autosaved graph to contain the variable created before it")
}

func TestAutosave_RecoversAfterSimulatedCrash(t *testing.T) {
	dir := t.TempDir()

	// Session A: does some work, then "crashes" (no save_graph, no clean
	// shutdown - we just stop using it, same as a killed process).
	sessionA := testSession(t)
	callTool(t, sessionA, "start_project", map[string]any{"path": dir}, &polyformmcp.StartProjectOutput{})
	callTool(t, sessionA, "create_variable", map[string]any{
		"path": "Body Color", "type": "coloring.color", "value": `"#cc3333"`,
	}, &polyformmcp.CreateVariableOutput{})

	// Session B: a fresh server/graph, as if the process restarted.
	sessionB := testSession(t)
	var recovered polyformmcp.StartProjectOutput
	callTool(t, sessionB, "start_project", map[string]any{"path": dir}, &recovered)

	require.NotEmpty(t, recovered.RecoveredFrom, "expected session B to find session A's autosave")
	require.NotEmpty(t, recovered.RecoveredModifiedAt)

	// The recovered file must be untouched by session B's own start_project
	// autosave (which would land at the live autosave.json path instead).
	require.NotEqual(t, autosavePath(dir), recovered.RecoveredFrom)

	var load polyformmcp.LoadGraphOutput
	callTool(t, sessionB, "load_graph", map[string]any{"path": recovered.RecoveredFrom}, &load)
	require.True(t, load.Loaded)

	var list polyformmcp.ListVariablesOutput
	callTool(t, sessionB, "list_variables", map[string]any{}, &list)
	require.Len(t, list.Variables, 1)
	require.Equal(t, "Body Color", list.Variables[0].Path)
}

func TestSaveGraph_DeletesAutosaveThenNextCallRecreatesIt(t *testing.T) {
	session := testSession(t)
	dir := t.TempDir()

	callTool(t, session, "start_project", map[string]any{"path": dir}, &polyformmcp.StartProjectOutput{})
	callTool(t, session, "create_variable", map[string]any{
		"path": "Radius", "type": "float64", "value": "3",
	}, &polyformmcp.CreateVariableOutput{})

	_, statErr := os.Stat(autosavePath(dir))
	require.NoError(t, statErr, "sanity check: autosave should exist before save_graph")

	realSavePath := filepath.Join(t.TempDir(), "graph.json")
	callTool(t, session, "save_graph", map[string]any{"path": realSavePath}, &polyformmcp.SaveGraphOutput{})

	_, statErr = os.Stat(autosavePath(dir))
	require.True(t, os.IsNotExist(statErr), "autosave should be gone immediately after a deliberate save_graph")

	// Any further call resumes normal autosaving - the deletion only
	// covers the gap up to the last deliberate save, not forever.
	callTool(t, session, "list_variables", map[string]any{}, &polyformmcp.ListVariablesOutput{})

	_, statErr = os.Stat(autosavePath(dir))
	require.NoError(t, statErr, "autosave should be recreated by the next call after save_graph")
}
