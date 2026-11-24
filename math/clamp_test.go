package math_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
)

func TestClampNode(t *testing.T) {
	clamp := func(value nodes.LiftedPort[float64], lo, hi float64) nodes.Node {
		return nodetest.NewNode(math.ClampNode[float64]{Value: value, Min: nodetest.NewPortValue(lo), Max: nodetest.NewPortValue(hi)})
	}
	nodetest.NewSuite(
		nodetest.NewTestCase("inside", clamp(nodetest.NewPortValue(0.5), 0, 1), nodetest.AssertOutput("Out", 0.5)),
		nodetest.NewTestCase("below", clamp(nodetest.NewPortValue(-2.), 0, 1), nodetest.AssertOutput("Out", 0.)),
		nodetest.NewTestCase("above", clamp(nodetest.NewPortValue(3.), 0, 1.3), nodetest.AssertOutput("Out", 1.3)),
		nodetest.NewTestCase("reversed bounds", clamp(nodetest.NewPortValue(3.), 1, 0), nodetest.AssertOutput("Out", 1.)),
		nodetest.NewTestCase("unwired max is one",
			nodetest.NewNode(math.ClampNode[float64]{Value: nodetest.NewPortValue(5.)}),
			nodetest.AssertOutput("Out", 1.),
		),
		nodetest.NewTestCase("list",
			nodetest.NewNode(math.ClampNode[int]{Value: nodetest.NewPortValue([]int{-1, 3, 9}), Max: nodetest.NewPortValue(5)}),
			nodetest.AssertOutput("Out", []int{0, 3, 5}),
		),
	).Run(t)
}

func TestSmoothStepNode(t *testing.T) {
	step := func(value, from, to float64) nodes.Node {
		return nodetest.NewNode(math.SmoothStepNode{Value: nodetest.NewPortValue(value), From: nodetest.NewPortValue(from), To: nodetest.NewPortValue(to)})
	}
	nodetest.NewSuite(
		nodetest.NewTestCase("before from", step(-1, 0, 1), nodetest.AssertOutput("Out", 0.)),
		nodetest.NewTestCase("after to", step(2, 0, 1), nodetest.AssertOutput("Out", 1.)),
		nodetest.NewTestCase("midpoint", step(0.5, 0, 1), nodetest.AssertOutput("Out", 0.5)),
		nodetest.NewTestCase("quarter eases in", step(0.25, 0, 1), nodetest.AssertOutput("Out", 0.15625)),
		nodetest.NewTestCase("descending edges", step(0.75, 1, 0), nodetest.AssertOutput("Out", 0.15625)),
		nodetest.NewTestCase("equal edges step", step(0.5, 0.5, 0.5), nodetest.AssertOutput("Out", 1.)),
		nodetest.NewTestCase("equal edges below", step(0.4, 0.5, 0.5), nodetest.AssertOutput("Out", 0.)),
		nodetest.NewTestCase("list",
			nodetest.NewNode(math.SmoothStepNode{Value: nodetest.NewPortValue([]float64{0, 0.5, 1}), To: nodetest.NewPortValue(1.)}),
			nodetest.AssertOutput("Out", []float64{0, 0.5, 1}),
		),
	).Run(t)
}
