package vector2_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector2"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v2 "github.com/EliCDavis/vector/vector2"
)

func TestSelectNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => 0,0",
			nodetest.NewNode(vector2.Select[float64]{}),
			nodetest.AssertOutput("X", 0.),
			nodetest.AssertOutput("Y", 0.),
			nodetest.AssertLiftedInput("In", vec2Type),
		),
		nodetest.NewTestCase(
			"(1,2) => 1,2",
			nodetest.NewNode(vector2.Select[float64]{
				In: nodetest.NewPortValue(v2.New(1., 2.)),
			}),
			nodetest.AssertOutput("X", 1.),
			nodetest.AssertOutput("Y", 2.),
		),
		nodetest.NewTestCase(
			"an array splits into two parallel arrays",
			nodetest.NewNode(vector2.Select[float64]{
				In: nodetest.NewPortValue([]v2.Float64{v2.New(1., 2.), v2.New(3., 4.)}),
			}),
			nodetest.AssertOutput("X", []float64{1, 3}),
			nodetest.AssertOutput("Y", []float64{2, 4}),
			nodetest.AssertOutputType("X", "[]float64"),
		),
	).Run(t)
}

func TestNewNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"two scalars make one vector",
			nodetest.NewNode(vector2.NewNode[float64]{
				X: nodetest.NewPortValue(1.),
				Y: nodetest.NewPortValue(2.),
			}),
			nodetest.AssertOutput("Out", v2.New(1., 2.)),
		),
		nodetest.NewTestCase(
			"two arrays zip into an array of vectors",
			nodetest.NewNode(vector2.NewNode[float64]{
				X: nodetest.NewPortValue([]float64{1, 3}),
				Y: nodetest.NewPortValue([]float64{2, 4}),
			}),
			nodetest.AssertOutput("Out", []v2.Float64{v2.New(1., 2.), v2.New(3., 4.)}),
			nodetest.AssertOutputType("Out", vec2ArrayType),
		),
	).Run(t)
}
