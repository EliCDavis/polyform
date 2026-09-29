package vector2_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector2"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v2 "github.com/EliCDavis/vector/vector2"
)

const (
	vec2Type      = "github.com/EliCDavis/vector/vector2.Vector[float64]"
	vec2ArrayType = "[]" + vec2Type
)

func TestSumNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => 0",
			nodetest.NewNode(vector2.SumNode[float64]{}),
			nodetest.AssertOutput("Out", v2.Zero[float64]()),
			nodetest.AssertLiftedInput("Values", vec2Type),
		),
		nodetest.NewTestCase(
			"two vectors add",
			nodetest.NewNode(vector2.SumNode[float64]{
				Values: nodetest.NewPortValues(v2.New(1., 2.), v2.New(4., 5.)),
			}),
			nodetest.AssertOutput("Out", v2.New(5., 7.)),
		),
		nodetest.NewTestCase(
			"an unwired connection is skipped",
			nodetest.NewNode(vector2.SumNode[float64]{
				Values: []nodes.LiftedPort[v2.Vector[float64]]{
					nodetest.NewPortValue(v2.New(1., 2.)),
					nil,
				},
			}),
			nodetest.AssertOutput("Out", v2.New(1., 2.)),
		),
		nodetest.NewTestCase(
			"one vector added to every entry of an array",
			nodetest.NewNode(vector2.SumNode[float64]{
				Values: []nodes.LiftedPort[v2.Vector[float64]]{
					nodetest.NewPortValue([]v2.Float64{v2.New(1., 1.), v2.New(2., 2.)}),
					nodetest.NewPortValue(v2.New(1., 2.)),
				},
			}),
			nodetest.AssertOutput("Out", []v2.Float64{v2.New(2., 3.), v2.New(3., 4.)}),
			nodetest.AssertOutputType("Out", vec2ArrayType),
		),
	).Run(t)
}

func TestSubtractNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"(4,5) - (1,2)",
			nodetest.NewNode(vector2.Subtract[float64]{
				A: nodetest.NewPortValue(v2.New(4., 5.)),
				B: nodetest.NewPortValue(v2.New(1., 2.)),
			}),
			nodetest.AssertOutput("Out", v2.New(3., 3.)),
		),
		nodetest.NewTestCase(
			"one vector subtracted from every entry of an array",
			nodetest.NewNode(vector2.Subtract[float64]{
				A: nodetest.NewPortValue([]v2.Float64{v2.New(1., 1.), v2.New(2., 2.)}),
				B: nodetest.NewPortValue(v2.New(1., 2.)),
			}),
			nodetest.AssertOutput("Out", []v2.Float64{v2.New(0., -1.), v2.New(1., 0.)}),
			nodetest.AssertOutputType("Out", vec2ArrayType),
		),
	).Run(t)
}
