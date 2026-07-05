package graph_test

import (
	"testing"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testInstanceWithSubGraphTypes(t *testing.T) *graph.Instance {
	t.Helper()
	factory := &refutil.TypeFactory{}
	factory.RegisterBuilder(subgraph.InputNodeTypeKey, func() any {
		return subgraph.NewInputNode("", "")
	})
	factory.RegisterBuilder(subgraph.OutputNodeTypeKey, func() any {
		return subgraph.NewOutputNode("", "")
	})
	return graph.New(graph.Config{
		TypeFactory: factory,
	})
}

func testInstanceWithSubGraphTypesExtended(t *testing.T) *graph.Instance {
	t.Helper()
	factory := &refutil.TypeFactory{}
	factory.RegisterBuilder(subgraph.InputNodeTypeKey, func() any {
		return subgraph.NewInputNode("", "")
	})
	factory.RegisterBuilder(subgraph.OutputNodeTypeKey, func() any {
		return subgraph.NewOutputNode("", "")
	})
	factory.RegisterBuilder("Float64", func() any {
		return &parameter.Float64{CurrentValue: 2}
	})
	factory.RegisterBuilder("Sum", func() any {
		return &nodes.Struct[math.AddNode[float64]]{}
	})
	return graph.New(graph.Config{
		TypeFactory: factory,
	})
}

func TestSubGraphCreateAndBoundaryPorts(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)

	err := inst.CreateSubGraph("test-sub", "Test SubGraph", "desc")
	require.NoError(t, err)

	child, err := inst.SubGraphInstance("test-sub")
	require.NoError(t, err)

	inputNode, inputID, err := child.CreateNode(subgraph.InputNodeTypeKey)
	require.NoError(t, err)
	require.NotEmpty(t, inputID)

	outputNode, outputID, err := child.CreateNode(subgraph.OutputNodeTypeKey)
	require.NoError(t, err)
	require.NotEmpty(t, outputID)

	input, ok := inputNode.(*subgraph.InputNode)
	require.True(t, ok)
	err = child.SetBoundaryNodeInfo(inputID, "Position", "github.com/EliCDavis/vector/vector3.Vector[float64]")
	require.NoError(t, err)
	require.Equal(t, "Position", input.BoundaryPortName())

	output, ok := outputNode.(*subgraph.OutputNode)
	require.True(t, ok)
	err = child.SetBoundaryNodeInfo(outputID, "Result", "float64")
	require.NoError(t, err)
	require.Equal(t, "Result", output.BoundaryPortName())

	ports, err := inst.CollectBoundaryPorts("test-sub")
	require.NoError(t, err)
	require.Len(t, ports, 2)

	runtimeNode, _, err := inst.CreateNode(subgraph.RuntimeTypePath("test-sub"))
	require.NoError(t, err)

	inputs := runtimeNode.Inputs()
	require.Contains(t, inputs, "Position")

	outputs := runtimeNode.Outputs()
	require.Contains(t, outputs, "Result")
}

func TestUnconfiguredBoundaryExcludedFromParentPorts(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)

	require.NoError(t, inst.CreateSubGraph("test-sub", "Test SubGraph", ""))
	child, err := inst.SubGraphInstance("test-sub")
	require.NoError(t, err)

	_, _, err = child.CreateNode(subgraph.InputNodeTypeKey)
	require.NoError(t, err)
	_, _, err = child.CreateNode(subgraph.OutputNodeTypeKey)
	require.NoError(t, err)

	ports, err := inst.CollectBoundaryPorts("test-sub")
	require.NoError(t, err)
	assert.Empty(t, ports)

	runtimeNode, _, err := inst.CreateNode(subgraph.RuntimeTypePath("test-sub"))
	require.NoError(t, err)
	assert.Empty(t, runtimeNode.Inputs())
	assert.Empty(t, runtimeNode.Outputs())
}

func TestSetBoundaryNodeInfoRequiresPortType(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("test-sub", "Test SubGraph", ""))
	child, err := inst.SubGraphInstance("test-sub")
	require.NoError(t, err)

	_, inputID, err := child.CreateNode(subgraph.InputNodeTypeKey)
	require.NoError(t, err)

	err = child.SetBoundaryNodeInfo(inputID, "Width", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port type")

	err = child.SetBoundaryNodeInfo(inputID, "", "float64")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port name")
}

func TestSubGraphCreateDuplicate(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("dup", "One", ""))
	err := inst.CreateSubGraph("dup", "Two", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestSubGraphDeleteNotFound(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	err := inst.DeleteSubGraph("missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestSubGraphDeleteGuard(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)

	err := inst.CreateSubGraph("guarded", "Guarded", "")
	require.NoError(t, err)

	_, _, err = inst.CreateNode(subgraph.RuntimeTypePath("guarded"))
	require.NoError(t, err)

	err = inst.DeleteSubGraph("guarded")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "still referenced")
}

func TestSubGraphDeleteAfterRemovingInstances(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("temp", "Temp", ""))

	runtimeNode, nodeID, err := inst.CreateNode(subgraph.RuntimeTypePath("temp"))
	require.NoError(t, err)
	require.NotNil(t, runtimeNode)

	inst.DeleteNodeById(nodeID)
	require.NoError(t, inst.DeleteSubGraph("temp"))

	_, err = inst.SubGraphInstance("temp")
	require.Error(t, err)
}

func TestSubGraphSetInfo(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("info", "Old", "old desc"))
	require.NoError(t, inst.SetSubGraphInfo("info", "New Name", "new desc"))

	runtimeNode := graph.NewRuntimeNode(inst, "info")
	assert.Equal(t, "New Name", runtimeNode.Name())
	assert.Equal(t, "new desc", runtimeNode.Description())
}

func TestSubGraphSetInfoNotFound(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	err := inst.SetSubGraphInfo("nope", "X", "")
	require.Error(t, err)
}

func TestSubGraphInstanceNotFound(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	_, err := inst.SubGraphInstance("missing")
	require.Error(t, err)
}

func TestScopeResolveInstance(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("scoped", "Scoped", ""))

	root, err := graph.RootScope.ResolveInstance(inst)
	require.NoError(t, err)
	assert.Equal(t, inst, root)

	root, err = graph.Scope("").ResolveInstance(inst)
	require.NoError(t, err)
	assert.Equal(t, inst, root)

	child, err := graph.SubGraphScope("scoped").ResolveInstance(inst)
	require.NoError(t, err)
	assert.True(t, child.IsSubGraphScope())

	_, err = graph.SubGraphScope("missing").ResolveInstance(inst)
	require.Error(t, err)

	_, err = graph.Scope("invalid/scope").ResolveInstance(inst)
	require.Error(t, err)
}

func TestSubGraphScopeID(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("scope-id", "Scope", ""))

	child, err := inst.SubGraphInstance("scope-id")
	require.NoError(t, err)
	assert.Equal(t, "scope-id", child.SubGraphScopeID())
	assert.Equal(t, "", inst.SubGraphScopeID())
}

func TestBoundaryDuplicatePortNameRejected(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("dup-ports", "Dup", ""))

	child, err := inst.SubGraphInstance("dup-ports")
	require.NoError(t, err)

	_, id1, err := child.CreateNode(subgraph.InputNodeTypeKey)
	require.NoError(t, err)
	_, id2, err := child.CreateNode(subgraph.InputNodeTypeKey)
	require.NoError(t, err)

	require.NoError(t, child.SetBoundaryNodeInfo(id1, "Shared", "float64"))
	err = child.SetBoundaryNodeInfo(id2, "Shared", "int")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already used")
}

func TestSetBoundaryNodeInfoNotBoundaryNode(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, inst.CreateSubGraph("bad-boundary", "Bad", ""))

	child, err := inst.SubGraphInstance("bad-boundary")
	require.NoError(t, err)

	_, paramID, err := child.CreateNode("Float64")
	require.NoError(t, err)

	err = child.SetBoundaryNodeInfo(paramID, "X", "float64")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a sub-graph boundary node")
}

func TestRuntimeInputSyncsToBoundaryInput(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, inst.CreateSubGraph("sync", "Sync", ""))

	child, err := inst.SubGraphInstance("sync")
	require.NoError(t, err)

	inputNode, inputID, err := child.CreateNode(subgraph.InputNodeTypeKey)
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(inputID, "A", "float64"))

	boundary := inputNode.(*subgraph.InputNode)

	extParam, extID, err := inst.CreateNode("Float64")
	require.NoError(t, err)

	runtimeNode, runtimeID, err := inst.CreateNode(subgraph.RuntimeTypePath("sync"))
	require.NoError(t, err)

	extOut := extParam.Outputs()["Value"]
	runtimeIn := runtimeNode.Inputs()["A"].(nodes.SingleValueInputPort)
	require.NoError(t, runtimeIn.Set(extOut))

	assert.Equal(t, extOut, boundary.ExternalSource())

	_ = runtimeID
	_ = extID
}

func TestSubGraphInnerConnections(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, inst.CreateSubGraph("wired", "Wired", ""))

	child, err := inst.SubGraphInstance("wired")
	require.NoError(t, err)

	_, inputID, err := child.CreateNode(subgraph.InputNodeTypeKey)
	require.NoError(t, err)
	_, outputID, err := child.CreateNode(subgraph.OutputNodeTypeKey)
	require.NoError(t, err)
	_, sumID, err := child.CreateNode("Sum")
	require.NoError(t, err)
	_, param1ID, err := child.CreateNode("Float64")
	require.NoError(t, err)
	_, param2ID, err := child.CreateNode("Float64")
	require.NoError(t, err)

	require.NoError(t, child.SetBoundaryNodeInfo(inputID, "A", "float64"))
	require.NoError(t, child.SetBoundaryNodeInfo(outputID, "Sum", "float64"))

	child.ConnectNodes(inputID, subgraph.ValuePortName, sumID, "Values.0")
	child.ConnectNodes(param1ID, "Value", sumID, "Values.1")
	child.ConnectNodes(param2ID, "Value", sumID, "Values.2")
	child.ConnectNodes(sumID, "Float", outputID, subgraph.ValuePortName)

	ports, err := inst.CollectBoundaryPorts("wired")
	require.NoError(t, err)
	require.Len(t, ports, 2)

	schema := inst.Schema()
	require.Contains(t, schema.SubGraphs, "wired")
	require.Contains(t, schema.SubGraphs["wired"].Nodes, sumID)
}

func TestSubGraphSchemaIncludesBoundaryMetadata(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("meta", "Meta", ""))

	child, err := inst.SubGraphInstance("meta")
	require.NoError(t, err)

	_, inputID, err := child.CreateNode(subgraph.InputNodeTypeKey)
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(inputID, "Width", "float64"))

	schema := inst.Schema()
	nodeInst := schema.SubGraphs["meta"].Nodes[inputID]
	require.NotNil(t, nodeInst.SubGraphInputBoundary)
	assert.Equal(t, "Width", nodeInst.SubGraphInputBoundary.PortName)
	assert.Equal(t, "float64", nodeInst.SubGraphInputBoundary.PortType)
}

func TestRuntimeNodeSchemaIncludesSubGraphId(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("runtime-meta", "Runtime", ""))

	_, runtimeID, err := inst.CreateNode(subgraph.RuntimeTypePath("runtime-meta"))
	require.NoError(t, err)

	schema := inst.Schema()
	nodeInst := schema.Nodes[runtimeID]
	assert.Equal(t, "runtime-meta", nodeInst.SubGraphId)
	assert.Equal(t, "Runtime", nodeInst.Name)
}

func TestCollectBoundaryPortsUnknownSubGraph(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	_, err := inst.CollectBoundaryPorts("missing")
	require.Error(t, err)
}

func TestCollectAllPortTypes(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	nodeTypes := inst.BuildSchemaForAllNodeTypes()
	types := graph.CollectAllPortTypes(nodeTypes)
	require.NotEmpty(t, types)
}

func TestSubGraphEncodeDecodeRoundtrip(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, inst.CreateSubGraph("roundtrip", "Roundtrip", ""))

	child, err := inst.SubGraphInstance("roundtrip")
	require.NoError(t, err)

	_, inputID, err := child.CreateNode(subgraph.InputNodeTypeKey)
	require.NoError(t, err)
	_, outputID, err := child.CreateNode(subgraph.OutputNodeTypeKey)
	require.NoError(t, err)

	require.NoError(t, child.SetBoundaryNodeInfo(inputID, "In", "float64"))
	require.NoError(t, child.SetBoundaryNodeInfo(outputID, "Out", "float64"))
	child.ConnectNodes(inputID, subgraph.ValuePortName, outputID, subgraph.ValuePortName)

	_, _, err = inst.CreateNode(subgraph.RuntimeTypePath("roundtrip"))
	require.NoError(t, err)

	payload, err := inst.EncodeToAppSchema()
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	fresh := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, fresh.ApplyAppSchema(payload))

	_, err = fresh.SubGraphInstance("roundtrip")
	require.NoError(t, err)

	ports, err := fresh.CollectBoundaryPorts("roundtrip")
	require.NoError(t, err)
	require.Len(t, ports, 2)

	freshChild, err := fresh.SubGraphInstance("roundtrip")
	require.NoError(t, err)

	var foundInput, foundOutput bool
	for _, node := range freshChild.Schema().Nodes {
		if node.SubGraphInputBoundary != nil && node.SubGraphInputBoundary.PortName == "In" {
			foundInput = true
		}
		if node.SubGraphOutputBoundary != nil && node.SubGraphOutputBoundary.PortName == "Out" {
			foundOutput = true
		}
	}
	assert.True(t, foundInput)
	assert.True(t, foundOutput)

	var restoredOutput *subgraph.OutputNode
	for nodeID, nodeInst := range freshChild.Schema().Nodes {
		if nodeInst.SubGraphOutputBoundary == nil || nodeInst.SubGraphOutputBoundary.PortName != "Out" {
			continue
		}
		out, ok := freshChild.Node(nodeID).(*subgraph.OutputNode)
		require.True(t, ok)
		restoredOutput = out
		break
	}
	require.NotNil(t, restoredOutput)
	assert.NotNil(t, restoredOutput.ConnectedSource())
}

func TestSubGraphDefinitionInAppSchema(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("json-check", "JSON", "desc"))

	payload, err := inst.EncodeToAppSchema()
	require.NoError(t, err)

	app, err := jbtf.Unmarshal[persistence.App](payload)
	require.NoError(t, err)
	require.Contains(t, app.SubGraphs, "json-check")
	assert.Equal(t, "JSON", app.SubGraphs["json-check"].Name)
	assert.Equal(t, "desc", app.SubGraphs["json-check"].Description)
}

func TestSubGraphEncodeDecodeWithMetadata(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("notes", "Notes", ""))

	child, err := inst.SubGraphInstance("notes")
	require.NoError(t, err)
	child.SetMetadata("notes", map[string]any{"sticky1": map[string]any{"text": "hello"}})

	payload, err := inst.EncodeToAppSchema()
	require.NoError(t, err)

	fresh := testInstanceWithSubGraphTypes(t)
	require.NoError(t, fresh.ApplyAppSchema(payload))

	freshChild, err := fresh.SubGraphInstance("notes")
	require.NoError(t, err)

	notes := freshChild.Schema().Notes
	require.NotNil(t, notes)
}

func TestSubGraphModelVersionBubblesFromChild(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("version", "Version", ""))

	v0 := inst.ModelVersion()
	child, err := inst.SubGraphInstance("version")
	require.NoError(t, err)

	_, _, err = child.CreateNode(subgraph.InputNodeTypeKey)
	require.NoError(t, err)

	assert.Greater(t, inst.ModelVersion(), v0)
}

func TestRegisterSubGraphNodeType(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("registered", "Registered", ""))

	typePath, _ := inst.RegisterSubGraphNodeType("registered")
	assert.Equal(t, subgraph.RuntimeTypePath("registered"), typePath)

	node, _, err := inst.CreateNode(typePath)
	require.NoError(t, err)
	runtime, ok := node.(*graph.SubgraphInstanceNode)
	require.True(t, ok)
	assert.Equal(t, "registered", runtime.SubGraphID())
}

func TestRuntimeNodeOutputsReflectBoundaryTypes(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("types", "Types", ""))

	child, err := inst.SubGraphInstance("types")
	require.NoError(t, err)

	_, inID, err := child.CreateNode(subgraph.InputNodeTypeKey)
	require.NoError(t, err)
	_, outID, err := child.CreateNode(subgraph.OutputNodeTypeKey)
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(inID, "Scale", "float64"))
	require.NoError(t, child.SetBoundaryNodeInfo(outID, "Image", "image.Image"))

	runtimeNode := graph.NewRuntimeNode(inst, "types")

	in := runtimeNode.Inputs()["Scale"]
	require.NotNil(t, in)
	typedIn, ok := in.(interface{ Type() string })
	require.True(t, ok)
	assert.Equal(t, "float64", typedIn.Type())

	out := runtimeNode.Outputs()["Image"]
	require.NotNil(t, out)

	typed, ok := out.(interface{ Type() string })
	require.True(t, ok)
	assert.Equal(t, "image.Image", typed.Type())
}

var _ nodes.Node = (*graph.SubgraphInstanceNode)(nil)

func TestRuntimeNodeInputPersistedThroughConnectNodes(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, inst.CreateSubGraph("persist", "Persist", ""))

	child, err := inst.SubGraphInstance("persist")
	require.NoError(t, err)

	_, inputID, err := child.CreateNode(subgraph.InputNodeTypeKey)
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(inputID, "A", "float64"))

	extParam, extID, err := inst.CreateNode("Float64")
	require.NoError(t, err)

	runtimeNode, runtimeID, err := inst.CreateNode(subgraph.RuntimeTypePath("persist"))
	require.NoError(t, err)

	inst.ConnectNodes(extID, "Value", runtimeID, "A")

	runtimeIn := runtimeNode.Inputs()["A"].(nodes.SingleValueInputPort)
	assert.Equal(t, extParam.Outputs()["Value"], runtimeIn.Value())

	schema := inst.NodeInstanceSchema(runtimeNode)
	require.Contains(t, schema.AssignedInput, "A")
	assert.Equal(t, extID, schema.AssignedInput["A"].NodeId)
}

func TestCollectBoundaryPortsAllowsSameNameAcrossKinds(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	require.NoError(t, inst.CreateSubGraph("shared-name", "Shared", ""))

	child, err := inst.SubGraphInstance("shared-name")
	require.NoError(t, err)

	_, inID, err := child.CreateNode(subgraph.InputNodeTypeKey)
	require.NoError(t, err)
	_, outID, err := child.CreateNode(subgraph.OutputNodeTypeKey)
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(inID, "Value", "float64"))
	require.NoError(t, child.SetBoundaryNodeInfo(outID, "Value", "float64"))

	ports, err := inst.CollectBoundaryPorts("shared-name")
	require.NoError(t, err)
	require.Len(t, ports, 2)

	runtimeNode := graph.NewRuntimeNode(inst, "shared-name")
	require.Contains(t, runtimeNode.Inputs(), "Value")
	require.Contains(t, runtimeNode.Outputs(), "Value")
}

func TestNestedRuntimeTypeLoadsFromSubGraph(t *testing.T) {
	root := testInstanceWithSubGraphTypes(t)
	require.NoError(t, root.CreateSubGraph("inner", "Inner", ""))
	require.NoError(t, root.CreateSubGraph("outer", "Outer", ""))

	outer, err := root.SubGraphInstance("outer")
	require.NoError(t, err)

	_, _, err = outer.CreateNode(subgraph.RuntimeTypePath("inner"))
	require.NoError(t, err)

	payload, err := root.EncodeToAppSchema()
	require.NoError(t, err)

	fresh := testInstanceWithSubGraphTypes(t)
	require.NoError(t, fresh.ApplyAppSchema(payload))

	freshOuter, err := fresh.SubGraphInstance("outer")
	require.NoError(t, err)
	assert.Len(t, freshOuter.Schema().Nodes, 1)
}
