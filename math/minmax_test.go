package math_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
)

func TestMinMaxNodes(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"Min: nothing wired => 0",
			nodetest.NewNode(math.MinNode[float64]{}),
			nodetest.AssertOutput("Float 64", 0.),
			nodetest.AssertOutputType("Float 64", "float64"),
			nodetest.AssertLiftedInput("In", "float64"),
		),
		nodetest.NewTestCase(
			"Min: 1 => 1",
			nodetest.NewNode(math.MinNode[float64]{In: nodetest.NewPortValues(1.)}),
			nodetest.AssertOutput("Float 64", 1.),
		),
		nodetest.NewTestCase(
			"Min: 2, 1, 3 => 1",
			nodetest.NewNode(math.MinNode[float64]{In: nodetest.NewPortValues(2., 1., 3.)}),
			nodetest.AssertOutput("Float 64", 1.),
			nodetest.AssertOutput("Int", 1),
		),
		nodetest.NewTestCase(
			"Min: []{4,1,6} against 3 => []{3,1,3}",
			nodetest.NewNode(math.MinNode[float64]{In: []nodes.LiftedPort[float64]{
				nodetest.NewPortValue([]float64{4, 1, 6}),
				nodetest.NewPortValue(3.),
			}}),
			nodetest.AssertOutput("Float 64", []float64{3, 1, 3}),
			nodetest.AssertOutputType("Float 64", "[]float64"),
		),
		nodetest.NewTestCase(
			"Min: []{4,1,6} against []{5,5,5} => []{4,1,5}",
			nodetest.NewNode(math.MinNode[float64]{In: []nodes.LiftedPort[float64]{
				nodetest.NewPortValue([]float64{4, 1, 6}),
				nodetest.NewPortValue([]float64{5, 5, 5}),
			}}),
			nodetest.AssertOutput("Float 64", []float64{4, 1, 5}),
		),
		nodetest.NewTestCase(
			"Min: mismatched lengths are refused",
			nodetest.NewNode(math.MinNode[float64]{In: []nodes.LiftedPort[float64]{
				nodetest.NewPortValue([]float64{1, 2, 3}),
				nodetest.NewPortValue([]float64{1, 2}),
			}}),
			nodetest.AssertOutput("Float 64", []float64{}),
			nodetest.AssertOutputError("Float 64", "3 and 2"),
		),

		nodetest.NewTestCase(
			"Max: nothing wired => 0",
			nodetest.NewNode(math.MaxNode[float64]{}),
			nodetest.AssertOutput("Float 64", 0.),
		),
		nodetest.NewTestCase(
			"Max: 1, 2 => 2",
			nodetest.NewNode(math.MaxNode[float64]{In: nodetest.NewPortValues(1., 2.)}),
			nodetest.AssertOutput("Float 64", 2.),
		),
		nodetest.NewTestCase(
			"Max: []{4,1,6} against 3 => []{4,3,6}",
			nodetest.NewNode(math.MaxNode[float64]{In: []nodes.LiftedPort[float64]{
				nodetest.NewPortValue([]float64{4, 1, 6}),
				nodetest.NewPortValue(3.),
			}}),
			nodetest.AssertOutput("Float 64", []float64{4, 3, 6}),
			nodetest.AssertOutputType("Float 64", "[]float64"),
		),

		nodetest.NewTestCase(
			"Min Array: reduces to one value",
			nodetest.NewNode(math.MinArrayNode[float64]{In: nodetest.NewPortValue([]float64{-1, 1})}),
			nodetest.AssertOutput("Float 64", -1.),
			nodetest.AssertOutputType("Float 64", "float64"),
		),
		nodetest.NewTestCase(
			"Min Array: empty => 0",
			nodetest.NewNode(math.MinArrayNode[float64]{In: nodetest.NewPortValue([]float64{})}),
			nodetest.AssertOutput("Float 64", 0.),
		),
		nodetest.NewTestCase(
			"Max Array: reduces to one value",
			nodetest.NewNode(math.MaxArrayNode[float64]{In: nodetest.NewPortValue([]float64{-1, 1})}),
			nodetest.AssertOutput("Float 64", 1.),
		),
		nodetest.NewTestCase(
			"Max Array: empty => 0",
			nodetest.NewNode(math.MaxArrayNode[float64]{In: nodetest.NewPortValue([]float64{})}),
			nodetest.AssertOutput("Float 64", 0.),
		),
	).Run(t)
}
