package vector2_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector2"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v2 "github.com/EliCDavis/vector/vector2"
)

func TestScaleNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired",
			nodetest.NewNode(vector2.Scale[float64]{}),
			nodetest.AssertOutput("Float 64", v2.Zero[float64]()),
			nodetest.AssertLiftedInput("Vector", vec2Type),
			nodetest.AssertLiftedInput("Amount", "float64"),
		),
		nodetest.NewTestCase(
			"an array without an amount is unchanged",
			nodetest.NewNode(vector2.Scale[float64]{
				Vector: nodetest.NewPortValue([]v2.Float64{v2.New(1.1, 2.2)}),
			}),
			nodetest.AssertOutput("Float 64", []v2.Float64{v2.New(1.1, 2.2)}),
			nodetest.AssertOutputType("Float 64", vec2ArrayType),
		),
		nodetest.NewTestCase(
			"one amount scales every vector in an array",
			nodetest.NewNode(vector2.Scale[float64]{
				Vector: nodetest.NewPortValue([]v2.Float64{v2.New(1.1, 2.3)}),
				Amount: nodetest.NewPortValue(2.),
			}),
			nodetest.AssertOutput("Float 64", []v2.Float64{v2.New(2.2, 4.6)}),
			nodetest.AssertOutput("Int", []v2.Int{v2.New(2, 5)}),
		),
		nodetest.NewTestCase(
			"an array of amounts scales index for index",
			nodetest.NewNode(vector2.Scale[float64]{
				Vector: nodetest.NewPortValue([]v2.Float64{v2.New(1., 1.), v2.New(2., 2.)}),
				Amount: nodetest.NewPortValue([]float64{0.5, 2}),
			}),
			nodetest.AssertOutput("Float 64", []v2.Float64{v2.New(0.5, 0.5), v2.New(4., 4.)}),
		),
	).Run(t)
}
