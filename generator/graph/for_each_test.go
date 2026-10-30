package graph_test

import (
	"fmt"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type floatListsSourceNode struct{}

func (floatListsSourceNode) Out(out *nodes.StructOutput[[][]float64]) {
	out.Set([][]float64{{1, 2}, {3, 4, 5}})
}

type floatPairSourceNode struct{}

func (floatPairSourceNode) Out(out *nodes.StructOutput[[]float64]) {
	out.Set([]float64{10, 20})
}

type floatSingletonSourceNode struct{}

func (floatSingletonSourceNode) Out(out *nodes.StructOutput[[]float64]) {
	out.Set([]float64{100})
}

type floatEmptySourceNode struct{}

func (floatEmptySourceNode) Out(out *nodes.StructOutput[[]float64]) {
	out.Set([]float64{})
}

type floatLimitNode struct {
	In nodes.Output[float64]
}

func (n floatLimitNode) Out(out *nodes.StructOutput[float64]) {
	v := nodes.TryGetOutputValue(out, n.In, 0.)
	if v > 2 {
		out.CaptureError(fmt.Errorf("%v is too big", v))
	}
	out.Set(v)
}

const (
	floatListsSourceType     = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.floatListsSourceNode]"
	floatPairSourceType      = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.floatPairSourceNode]"
	floatSingletonSourceType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.floatSingletonSourceNode]"
	floatEmptySourceType     = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.floatEmptySourceNode]"
	floatLimitType           = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.floatLimitNode]"
)

func testInstanceForForEach(t *testing.T) *graph.Instance {
	t.Helper()
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[floatArraySourceNode]](factory)
	refutil.RegisterType[nodes.Struct[liftedDoubleNode]](factory)
	refutil.RegisterType[nodes.Struct[floatArraySumNode]](factory)
	refutil.RegisterType[nodes.Struct[floatSourceNode]](factory)
	refutil.RegisterType[nodes.Struct[liftedTotalNode]](factory)
	refutil.RegisterType[nodes.Struct[floatConsumerNode]](factory)
	refutil.RegisterType[nodes.Struct[floatListsSourceNode]](factory)
	refutil.RegisterType[nodes.Struct[floatPairSourceNode]](factory)
	refutil.RegisterType[nodes.Struct[floatSingletonSourceNode]](factory)
	refutil.RegisterType[nodes.Struct[floatEmptySourceNode]](factory)
	refutil.RegisterType[nodes.Struct[floatLimitNode]](factory)
	factory.RegisterBuilder(subgraph.InputNodeTypeKey, func() any { return subgraph.NewInputNode("", "") })
	factory.RegisterBuilder(subgraph.OutputNodeTypeKey, func() any { return subgraph.NewOutputNode("", "") })
	factory.RegisterBuilder("Float64", func() any { return &parameter.Float64{} })

	instance := graph.New(graph.Config{TypeFactory: factory})
	nodes.DiscoverPortTypes(factory)
	return instance
}

type eachAdder struct {
	root       *graph.Instance
	definition *graph.Graph
	placement  string
	// Sum = A + B + Extra, A through a single-value node so a list cannot pass as one.
	a, b, sum, extra, total string
}

// Sum is A + B + Extra; A goes through a node taking one float64, so a whole list cannot pass for one.
func placeEachAdder(t *testing.T) eachAdder {
	t.Helper()
	root := testInstanceForForEach(t)
	require.NoError(t, root.CreateSubGraph("adder", "adder", ""))
	definition, err := root.SubGraphInstance("adder")
	require.NoError(t, err)

	a := boundary(t, definition, subgraph.InputNodeTypeKey, "float64", "A")
	b := boundary(t, definition, subgraph.InputNodeTypeKey, "float64", "B")
	sum := boundary(t, definition, subgraph.OutputNodeTypeKey, "float64", "Sum")
	_, single, err := definition.CreateNode(floatConsumerType)
	require.NoError(t, err)
	_, total, err := definition.CreateNode(liftedTotalType)
	require.NoError(t, err)
	_, extra, err := definition.CreateNode("Float64")
	require.NoError(t, err)
	require.NoError(t, definition.ConnectNodes(a, subgraph.ValuePortName, single, "In"))
	require.NoError(t, definition.ConnectNodes(single, "Out", total, "In.0"))
	require.NoError(t, definition.ConnectNodes(b, subgraph.ValuePortName, total, "In.1"))
	require.NoError(t, definition.ConnectNodes(extra, "Value", total, "In.2"))
	require.NoError(t, definition.ConnectNodes(total, "Out", sum, subgraph.ValuePortName))

	typePath, err := root.RegisterSubGraphNodeType("adder")
	require.NoError(t, err)
	_, placement, err := root.CreateNode(typePath)
	require.NoError(t, err)
	return eachAdder{root: root, definition: definition, placement: placement, a: a, b: b, sum: sum, extra: extra, total: total}
}

func (p eachAdder) feed(t *testing.T, input, sourceType string) string {
	t.Helper()
	_, source, err := p.root.CreateNode(sourceType)
	require.NoError(t, err)
	require.NoError(t, p.root.ConnectNodes(source, "Out", p.placement, input))
	return source
}

func (p eachAdder) each(t *testing.T, each bool) []graph.DroppedEdge {
	t.Helper()
	dropped, err := p.root.SetForEach(p.placement, each)
	require.NoError(t, err)
	return dropped
}

func (p eachAdder) node() *graph.SubgraphInstanceNode {
	return p.root.Node(p.placement).(*graph.SubgraphInstanceNode)
}

func (p eachAdder) sums(t *testing.T) []float64 {
	t.Helper()
	return nodes.GetNodeOutputPort[[]float64](p.node(), "Sum").Value()
}

func (p eachAdder) report(t *testing.T) nodes.ExecutionReport {
	t.Helper()
	port := p.node().Outputs()[p.sum]
	observable, ok := port.(nodes.ObservableExecution)
	require.True(t, ok, "%T", port)
	return observable.ExecutionReport()
}

func TestForEachIteratesAListAndBroadcastsASingleValue(t *testing.T) {
	p := placeEachAdder(t)
	p.each(t, true)
	p.feed(t, "A", floatArraySourceType)
	p.feed(t, "B", floatSourceType)

	assert.Equal(t, "[]float64", p.node().Outputs()[p.sum].(nodes.Typed).Type())
	assert.Equal(t, []float64{3, 4, 5}, p.sums(t))
	require.NoError(t, graph.CopiesMatchDefinitions(p.root))
}

func TestForEachIteratesAListOfListsFedToAListBoundary(t *testing.T) {
	root := testInstanceForForEach(t)
	require.NoError(t, root.CreateSubGraph("summer", "summer", ""))
	definition, err := root.SubGraphInstance("summer")
	require.NoError(t, err)
	in := boundary(t, definition, subgraph.InputNodeTypeKey, "[]float64", "Values")
	out := boundary(t, definition, subgraph.OutputNodeTypeKey, "float64", "Total")
	_, sum, err := definition.CreateNode(floatArraySumType)
	require.NoError(t, err)
	require.NoError(t, definition.ConnectNodes(in, subgraph.ValuePortName, sum, "In"))
	require.NoError(t, definition.ConnectNodes(sum, "Out", out, subgraph.ValuePortName))
	typePath, err := root.RegisterSubGraphNodeType("summer")
	require.NoError(t, err)
	_, placement, err := root.CreateNode(typePath)
	require.NoError(t, err)
	_, lists, err := root.CreateNode(floatListsSourceType)
	require.NoError(t, err)

	require.Error(t, root.ConnectNodes(lists, "Out", placement, "Values"), "a list boundary takes no list of lists while the placement runs once")
	require.NoError(t, errOf(root.SetForEach(placement, true)))
	require.NoError(t, root.ConnectNodes(lists, "Out", placement, "Values"))

	assert.Equal(t, []float64{3, 12}, nodes.GetNodeOutputPort[[]float64](root.Node(placement), "Total").Value())
}

func TestForEachRefusesIteratedListsOfDifferentLengths(t *testing.T) {
	p := placeEachAdder(t)
	p.each(t, true)
	p.feed(t, "A", floatArraySourceType)
	p.feed(t, "B", floatPairSourceType)

	assert.Empty(t, p.sums(t))
	report := p.report(t)
	require.Len(t, report.Errors, 1)
	assert.Contains(t, report.Errors[0], "3 and 2 cannot be combined")
}

func TestForEachBroadcastsAListOfOne(t *testing.T) {
	p := placeEachAdder(t)
	p.each(t, true)
	p.feed(t, "A", floatArraySourceType)
	p.feed(t, "B", floatSingletonSourceType)

	assert.Equal(t, []float64{101, 102, 103}, p.sums(t))
}

func TestForEachOverAnEmptyListRunsNothing(t *testing.T) {
	p := placeEachAdder(t)
	p.each(t, true)
	p.feed(t, "A", floatEmptySourceType)
	p.feed(t, "B", floatSourceType)

	assert.Empty(t, p.sums(t))
	assert.Equal(t, 0, graph.BuiltRuns(p.node()))
	assert.Empty(t, p.report(t).Errors)
}

func TestForEachWithNothingIteratedRunsOnce(t *testing.T) {
	p := placeEachAdder(t)
	p.each(t, true)
	p.feed(t, "A", floatSourceType)
	p.feed(t, "B", floatSourceType)

	assert.Equal(t, []float64{4}, p.sums(t))
	assert.Equal(t, 1, graph.BuiltRuns(p.node()))
}

func TestForEachTypesItsOutputsWithoutBuildingARun(t *testing.T) {
	p := placeEachAdder(t)
	p.each(t, true)
	p.feed(t, "A", floatArraySourceType)

	schema := p.root.NodeInstanceSchema(p.node())
	assert.True(t, schema.ForEach)
	assert.Equal(t, "[]float64", schema.Output["Sum"].Type)
	assert.Nil(t, p.node().BuiltGraph())
	assert.Equal(t, 0, graph.BuiltRuns(p.node()))
}

func TestTogglingForEachRetypesOutputsAndDropsEdgesThatNoLongerFit(t *testing.T) {
	p := placeEachAdder(t)
	p.feed(t, "A", floatSourceType)
	_, consumer, err := p.root.CreateNode(floatConsumerType)
	require.NoError(t, err)
	require.NoError(t, p.root.ConnectNodes(p.placement, "Sum", consumer, "In"))

	dropped := p.each(t, true)
	require.Len(t, dropped, 1, "%+v", dropped)
	assert.Equal(t, consumer, dropped[0].To)
	assert.Nil(t, p.root.Node(consumer).Inputs()["In"].(nodes.SingleValueInputPort).Value())

	_, sum, err := p.root.CreateNode(floatArraySumType)
	require.NoError(t, err)
	require.NoError(t, p.root.ConnectNodes(p.placement, "Sum", sum, "In"))
	assert.InDelta(t, 2, nodes.GetNodeOutputPort[float64](p.root.Node(sum), "Out").Value(), 1e-12)

	dropped = p.each(t, false)
	require.Len(t, dropped, 1, "%+v", dropped)
	assert.Equal(t, sum, dropped[0].To)
	assert.Equal(t, "float64", p.node().Outputs()[p.sum].(nodes.Typed).Type())
}

func TestAnEditInsideTheDefinitionReachesEveryRun(t *testing.T) {
	p := placeEachAdder(t)
	p.each(t, true)
	p.feed(t, "A", floatArraySourceType)
	p.feed(t, "B", floatSourceType)
	require.Equal(t, []float64{3, 4, 5}, p.sums(t))
	before, err := p.node().Runs()
	require.NoError(t, err)

	_, err = p.definition.UpdateParameter(p.extra, []byte(`10`))
	require.NoError(t, err)
	assert.Equal(t, []float64{13, 14, 15}, p.sums(t), "a parameter is shared with every run")
	after, err := p.node().Runs()
	require.NoError(t, err)
	assert.Equal(t, before, after, "nothing to rebuild for a parameter edit")

	require.NoError(t, errOf(p.definition.DeleteNodeInputConnection(p.total, "In.2")))
	assert.Equal(t, []float64{3, 4, 5}, p.sums(t))
	rebuilt, err := p.node().Runs()
	require.NoError(t, err)
	for i := range rebuilt {
		assert.NotSame(t, before[i], rebuilt[i], "run %d is rebuilt for a wiring edit", i)
	}
	require.NoError(t, graph.CopiesMatchDefinitions(p.root))
}

func TestGrowingAndShrinkingTheListKeepsExistingRuns(t *testing.T) {
	p := placeEachAdder(t)
	p.each(t, true)
	pair := p.feed(t, "A", floatPairSourceType)
	p.feed(t, "B", floatSourceType)
	require.Equal(t, []float64{12, 22}, p.sums(t))
	two, err := p.node().Runs()
	require.NoError(t, err)
	require.Len(t, two, 2)

	_, three, err := p.root.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	require.NoError(t, p.root.ConnectNodes(three, "Out", p.placement, "A"))
	require.Equal(t, []float64{3, 4, 5}, p.sums(t))
	grown, err := p.node().Runs()
	require.NoError(t, err)
	require.Len(t, grown, 3)
	assert.Same(t, two[0], grown[0])
	assert.Same(t, two[1], grown[1])

	require.NoError(t, p.root.ConnectNodes(pair, "Out", p.placement, "A"))
	require.Equal(t, []float64{12, 22}, p.sums(t))
	shrunk, err := p.node().Runs()
	require.NoError(t, err)
	require.Len(t, shrunk, 2)
	assert.Same(t, grown[0], shrunk[0])
	assert.Same(t, grown[1], shrunk[1])
}

func TestNestedForEachPeelsOneListLevelPerPlacement(t *testing.T) {
	root := testInstanceForForEach(t)
	doublerType, _, _ := liftedPassThroughSubgraphIn(t, root, "doubler")

	require.NoError(t, root.CreateSubGraph("doubleAll", "doubleAll", ""))
	outer, err := root.SubGraphInstance("doubleAll")
	require.NoError(t, err)
	in := boundary(t, outer, subgraph.InputNodeTypeKey, "[]float64", "Values")
	out := boundary(t, outer, subgraph.OutputNodeTypeKey, "[]float64", "Doubled")
	_, inner, err := outer.CreateNode(doublerType)
	require.NoError(t, err)
	require.NoError(t, errOf(outer.SetForEach(inner, true)))
	require.NoError(t, outer.ConnectNodes(in, subgraph.ValuePortName, inner, "Value In"))
	require.NoError(t, outer.ConnectNodes(inner, "Value Out", out, subgraph.ValuePortName))
	outerType, err := root.RegisterSubGraphNodeType("doubleAll")
	require.NoError(t, err)

	_, placement, err := root.CreateNode(outerType)
	require.NoError(t, err)
	require.NoError(t, errOf(root.SetForEach(placement, true)))
	_, lists, err := root.CreateNode(floatListsSourceType)
	require.NoError(t, err)
	require.NoError(t, root.ConnectNodes(lists, "Out", placement, "Values"))

	assert.Equal(t, "[][]float64", root.Node(placement).Outputs()[out].(nodes.Typed).Type())
	assert.Equal(t, [][]float64{{2, 4}, {6, 8, 10}}, nodes.GetNodeOutputPort[[][]float64](root.Node(placement), "Doubled").Value())
	require.NoError(t, graph.CopiesMatchDefinitions(root))
}

// liftedPassThroughSubgraph, in an instance holding the For Each test types.
func liftedPassThroughSubgraphIn(t *testing.T, inst *graph.Instance, id string) (typePath, inPort, outPort string) {
	t.Helper()
	require.NoError(t, inst.CreateSubGraph(id, id, ""))
	child, err := inst.SubGraphInstance(id)
	require.NoError(t, err)
	in := boundary(t, child, subgraph.InputNodeTypeKey, "float64", "Value In")
	out := boundary(t, child, subgraph.OutputNodeTypeKey, "float64", "Value Out")
	_, double, err := child.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	require.NoError(t, child.ConnectNodes(in, subgraph.ValuePortName, double, "In"))
	require.NoError(t, child.ConnectNodes(double, "Out", out, subgraph.ValuePortName))
	typePath, err = inst.RegisterSubGraphNodeType(id)
	require.NoError(t, err)
	return typePath, in, out
}

func TestARunsErrorNamesItsItem(t *testing.T) {
	root := testInstanceForForEach(t)
	require.NoError(t, root.CreateSubGraph("limited", "limited", ""))
	definition, err := root.SubGraphInstance("limited")
	require.NoError(t, err)
	in := boundary(t, definition, subgraph.InputNodeTypeKey, "float64", "In")
	out := boundary(t, definition, subgraph.OutputNodeTypeKey, "float64", "Out")
	_, limit, err := definition.CreateNode(floatLimitType)
	require.NoError(t, err)
	require.NoError(t, definition.ConnectNodes(in, subgraph.ValuePortName, limit, "In"))
	require.NoError(t, definition.ConnectNodes(limit, "Out", out, subgraph.ValuePortName))
	typePath, err := root.RegisterSubGraphNodeType("limited")
	require.NoError(t, err)
	_, placement, err := root.CreateNode(typePath)
	require.NoError(t, err)
	require.NoError(t, errOf(root.SetForEach(placement, true)))
	_, source, err := root.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	require.NoError(t, root.ConnectNodes(source, "Out", placement, "In"))

	node := root.Node(placement)
	assert.Equal(t, []float64{1, 2, 3}, nodes.GetNodeOutputPort[[]float64](node, "Out").Value())
	report := node.Outputs()[out].(nodes.ObservableExecution).ExecutionReport()
	assert.Equal(t, []string{"item 2: 3 is too big"}, report.Errors)
	assert.Len(t, report.Steps, 3)
}

func TestForEachSurvivesASaveAndLoad(t *testing.T) {
	p := placeEachAdder(t)
	p.feed(t, "B", floatSourceType)
	off, err := p.root.EncodeToAppSchema()
	require.NoError(t, err)

	p.each(t, true)
	p.each(t, false)
	toggled, err := p.root.EncodeToAppSchema()
	require.NoError(t, err)
	assert.Equal(t, string(off), string(toggled), "a placement that runs once saves as it always did")

	p.each(t, true)
	p.feed(t, "A", floatArraySourceType)
	on, err := p.root.EncodeToAppSchema()
	require.NoError(t, err)

	reloaded := testInstanceForForEach(t)
	require.NoError(t, reloaded.ApplyAppSchema(on))
	placement := reloaded.Node(p.placement).(*graph.SubgraphInstanceNode)
	assert.True(t, placement.ForEach())
	assert.Equal(t, []float64{3, 4, 5}, nodes.GetNodeOutputPort[[]float64](placement, "Sum").Value())
}

func TestADownstreamNodeRecomputesWhenAnIteratedListChanges(t *testing.T) {
	p := placeEachAdder(t)
	p.each(t, true)
	p.feed(t, "A", floatPairSourceType)
	p.feed(t, "B", floatSourceType)
	_, sum, err := p.root.CreateNode(floatArraySumType)
	require.NoError(t, err)
	require.NoError(t, p.root.ConnectNodes(p.placement, "Sum", sum, "In"))
	total := nodes.GetNodeOutputPort[float64](p.root.Node(sum), "Out")
	require.InDelta(t, 34, total.Value(), 1e-12)

	_, three, err := p.root.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	require.NoError(t, p.root.ConnectNodes(three, "Out", p.placement, "A"))
	assert.InDelta(t, 12, total.Value(), 1e-12)

	_, err = p.definition.UpdateParameter(p.extra, []byte(`1`))
	require.NoError(t, err)
	assert.InDelta(t, 15, total.Value(), 1e-12, "a parameter edit inside reaches a consumer outside")
}
