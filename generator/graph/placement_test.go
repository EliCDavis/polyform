package graph_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type placedDoubler struct {
	root        *graph.Instance
	child       *graph.Graph
	placement   string
	in, out     string
	boundaryIn  string
	boundaryOut string
	double      string
}

// The root places a subgraph that doubles what it is fed.
func placeDoubler(t *testing.T) placedDoubler {
	t.Helper()
	root := testInstanceWithLiftedAndBoundaryTypes(t)
	typePath, boundaryIn, boundaryOut := liftedPassThroughSubgraph(t, root, "doubler")
	_, placement, err := root.CreateNode(typePath)
	require.NoError(t, err)

	child, err := root.SubGraphInstance("doubler")
	require.NoError(t, err)
	var double string
	for _, id := range child.NodeIds() {
		if id != boundaryIn && id != boundaryOut {
			double = id
		}
	}

	in, out := instancePorts(t, root, placement)
	return placedDoubler{root: root, child: child, placement: placement, in: in, out: out,
		boundaryIn: boundaryIn, boundaryOut: boundaryOut, double: double}
}

func (p placedDoubler) feedArrayIntoAnArrayReader(t *testing.T) (sum string) {
	t.Helper()
	_, source, err := p.root.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, sum, err = p.root.CreateNode(floatArraySumType)
	require.NoError(t, err)
	require.NoError(t, p.root.ConnectNodes(source, "Out", p.placement, p.in))
	require.NoError(t, p.root.ConnectNodes(p.placement, p.out, sum, "In"))
	require.InDelta(t, 12, nodes.GetNodeOutputPort[float64](p.root.Node(sum), "Out").Value(), 1e-12)
	return sum
}

func TestRemovingAWiredSubgraphOutputDropsAndReportsTheEdgeReadingIt(t *testing.T) {
	p := placeDoubler(t)
	_, source, err := p.root.CreateNode(floatSourceType)
	require.NoError(t, err)
	_, consumer, err := p.root.CreateNode(floatConsumerType)
	require.NoError(t, err)
	require.NoError(t, p.root.ConnectNodes(source, "Out", p.placement, p.in))
	require.NoError(t, p.root.ConnectNodes(p.placement, p.out, consumer, "In"))

	dropped, err := p.child.DeleteNodeById(p.boundaryOut)
	require.NoError(t, err)

	require.Len(t, dropped, 1, "%+v", dropped)
	assert.Equal(t, "", dropped[0].Scope)
	assert.Equal(t, p.placement, dropped[0].From)
	assert.Equal(t, consumer, dropped[0].To)
	assert.Nil(t, p.root.Node(consumer).Inputs()["In"].(nodes.SingleValueInputPort).Value())
}

func TestRemovingAWiredSubgraphInputDropsAndReportsTheEdgeFeedingIt(t *testing.T) {
	p := placeDoubler(t)
	_, source, err := p.root.CreateNode(floatSourceType)
	require.NoError(t, err)
	require.NoError(t, p.root.ConnectNodes(source, "Out", p.placement, p.in))

	dropped, err := p.child.DeleteNodeById(p.boundaryIn)
	require.NoError(t, err)

	require.Len(t, dropped, 1, "%+v", dropped)
	assert.Equal(t, graph.DroppedEdge{
		From: source, FromPort: "Out", To: p.placement, ToPort: p.in, Reason: dropped[0].Reason,
	}, dropped[0])
}

func TestAConnectInsideASubgraphThatWouldBreakAnEdgeWhereItIsPlacedIsRefused(t *testing.T) {
	p := placeDoubler(t)
	sum := p.feedArrayIntoAnArrayReader(t)
	_, scalar, err := p.child.CreateNode(floatSourceType)
	require.NoError(t, err)

	require.Error(t, p.child.ConnectNodes(scalar, "Out", p.double, "In"), "a scalar into the doubler makes its output a scalar, which the root's sum cannot read")

	held := p.child.Node(p.double).Inputs()["In"].(nodes.SingleValueInputPort).Value()
	assert.Equal(t, p.boundaryIn, p.child.NodeId(held.Node()), "the subgraph is left as it was")
	assert.InDelta(t, 12, nodes.GetNodeOutputPort[float64](p.root.Node(sum), "Out").Value(), 1e-12,
		"and so is the root")
}

func TestAConnectInsideASubgraphThatAPlacementCannotBeRebuiltAroundIsRefused(t *testing.T) {
	p := placeDoubler(t)
	p.feedArrayIntoAnArrayReader(t)
	_, consumer, err := p.child.CreateNode(floatConsumerType)
	require.NoError(t, err)

	require.Error(t, p.child.ConnectNodes(p.boundaryIn, subgraph.ValuePortName, consumer, "In"), "the placement feeds an array in, which a float64 input cannot take")

	assert.Nil(t, p.child.Node(consumer).Inputs()["In"].(nodes.SingleValueInputPort).Value())
}

func TestAnEdgeDroppedInsideAPlacedSubgraphIsGoneFromItsPlacementsToo(t *testing.T) {
	p := placeDoubler(t)
	require.NoError(t, p.root.CreateSubGraph("summer", "summer", ""))
	middle, err := p.root.SubGraphInstance("summer")
	require.NoError(t, err)
	doublerType, err := p.root.RegisterSubGraphNodeType("doubler")
	require.NoError(t, err)
	_, source, err := middle.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, doubler, err := middle.CreateNode(doublerType)
	require.NoError(t, err)
	_, sum, err := middle.CreateNode(floatArraySumType)
	require.NoError(t, err)
	in, out := instancePorts(t, middle, doubler)
	require.NoError(t, middle.ConnectNodes(source, "Out", doubler, in))
	require.NoError(t, middle.ConnectNodes(doubler, out, sum, "In"))

	summerType, err := p.root.RegisterSubGraphNodeType("summer")
	require.NoError(t, err)
	_, placed, err := p.root.CreateNode(summerType)
	require.NoError(t, err)

	dropped, err := p.child.DeleteNodeInputConnection(p.double, "In")
	require.NoError(t, err)

	require.Len(t, dropped, 1, "%+v", dropped)
	assert.Equal(t, "summer", dropped[0].Scope)
	assert.Equal(t, sum, dropped[0].To)
	inside := p.root.Node(placed).(*graph.SubgraphInstanceNode).LiveGraph()
	assert.Nil(t, inside.Node(sum).Inputs()["In"].(nodes.SingleValueInputPort).Value(),
		"the placement's copy matches its definition")
}

func TestTakingAnArrayOffAPlacementRebuildsItsInside(t *testing.T) {
	root := testInstanceWithLiftedAndBoundaryTypes(t)
	require.NoError(t, root.CreateSubGraph("plusTwo", "plusTwo", ""))
	child, err := root.SubGraphInstance("plusTwo")
	require.NoError(t, err)
	_, boundaryIn, err := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	_, two, err := child.CreateNode(floatSourceType)
	require.NoError(t, err)
	_, total, err := child.CreateNode(liftedTotalType)
	require.NoError(t, err)
	_, boundaryOut, err := child.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(boundaryIn, "Value In"))
	require.NoError(t, child.SetBoundaryNodeInfo(boundaryOut, "Value Out"))
	require.NoError(t, child.ConnectNodes(boundaryIn, subgraph.ValuePortName, total, "In.0"))
	require.NoError(t, child.ConnectNodes(two, "Out", total, "In.1"))
	require.NoError(t, child.ConnectNodes(total, "Out", boundaryOut, subgraph.ValuePortName))
	typePath, err := root.RegisterSubGraphNodeType("plusTwo")
	require.NoError(t, err)

	_, fed, err := root.CreateNode(typePath)
	require.NoError(t, err)
	_, unfed, err := root.CreateNode(typePath)
	require.NoError(t, err)
	_, source, err := root.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	in, out := instancePorts(t, root, fed)
	require.NoError(t, root.ConnectNodes(source, "Out", fed, in))
	require.Equal(t, []float64{3, 4, 5}, nodes.GetNodeOutputPort[[]float64](root.Node(fed), out).Value())

	require.NoError(t, errOf(root.DeleteNodeInputConnection(fed, in)))

	assert.Equal(t,
		nodes.GetNodeOutputPort[float64](root.Node(unfed), out).Value(),
		nodes.GetNodeOutputPort[float64](root.Node(fed), out).Value(),
		"a placement that lost its array computes what one never fed does")
}

func TestABatchReportsTheEdgesItHadToDrop(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)
	_, source, err := instance.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, double, err := instance.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, sum, err := instance.CreateNode(floatArraySumType)
	require.NoError(t, err)

	err = instance.Batch(func() error {
		require.NoError(t, instance.ConnectNodes(source, "Out", double, "In"))
		require.NoError(t, instance.ConnectNodes(double, "Out", sum, "In"))
		require.NoError(t, errOf(instance.DeleteNodeInputConnection(double, "In")))
		return nil
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot take")
	assert.Nil(t, instance.Node(sum).Inputs()["In"].(nodes.SingleValueInputPort).Value())
}

func TestLoadingAGraphWhereANodeReadsItsOwnOutputIsRefused(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)
	_, first, err := instance.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, second, err := instance.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	require.NoError(t, instance.ConnectNodes(first, "Out", second, "In"))
	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)

	var app map[string]any
	require.NoError(t, json.Unmarshal(saved, &app))
	data := app["data"].(map[string]any)
	firstNode := data["nodes"].(map[string]any)[first].(map[string]any)
	firstNode["assignedInput"] = map[string]any{"In": map[string]any{"id": second, "port": "Out"}}
	cyclic, err := json.Marshal(app)
	require.NoError(t, err)

	err = testInstanceWithLiftedNodes(t).ApplyAppSchema(cyclic)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reads its own output")
}

func TestConnectingAMissingPortNamesThePortRatherThanACycle(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)
	_, double, err := instance.CreateNode(liftedDoubleType)
	require.NoError(t, err)

	assert.EqualError(t, instance.ConnectNodes(double, "Nope", double, "In"), `node "`+double+`" contains no out-port "Nope"`)
}
