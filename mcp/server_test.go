package mcp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
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
	_ "github.com/EliCDavis/polyform/formats/gltf"
	_ "github.com/EliCDavis/polyform/generator/manifest/basics"
	_ "github.com/EliCDavis/polyform/generator/parameter"
	_ "github.com/EliCDavis/polyform/generator/subgraph/register"
	_ "github.com/EliCDavis/polyform/math"
	_ "github.com/EliCDavis/polyform/math/curves"
	_ "github.com/EliCDavis/polyform/math/quaternion"
	_ "github.com/EliCDavis/polyform/math/sdf"
	_ "github.com/EliCDavis/polyform/math/sequence"
	_ "github.com/EliCDavis/polyform/math/trs"
	_ "github.com/EliCDavis/polyform/math/vector3"
	_ "github.com/EliCDavis/polyform/modeling"
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

func TestSearchNodeTypesRegexCaseSensitiveByDefault(t *testing.T) {
	session := testSession(t)

	var upper polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{"query": "CUBENODE", "regex": true}, &upper)
	require.Empty(t, upper.Results, "regex mode should be case-sensitive by default")

	var insensitive polyformmcp.SearchNodeTypesOutput
	callTool(t, session, "search_node_types", map[string]any{"query": "(?i)CUBENODE", "regex": true}, &insensitive)
	found := false
	for _, r := range insensitive.Results {
		if r.Type == cubeNodeType {
			found = true
		}
	}
	require.True(t, found, "(?i) prefix should make regex mode case-insensitive")
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

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_equation_subgraph",
		Arguments: map[string]any{
			"id":       "bad",
			"equation": "y = sin(x)",
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError, "sin() has no backing node and should be rejected, not silently approximated")
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

	require.Len(t, out.Values, 3)
	require.InDelta(t, -0.5, out.Values[0], 1e-9)
	require.InDelta(t, 0.0, out.Values[1], 1e-9)
	require.InDelta(t, 0.5, out.Values[2], 1e-9)
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
			"Mesh": map[string]any{"nodeId": cube.NodeId, "port": "Out"},
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
