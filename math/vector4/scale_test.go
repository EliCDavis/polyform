package vector4_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector4"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v4 "github.com/EliCDavis/vector/vector4"
)

func TestScaleNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired",
			nodetest.NewNode(vector4.Scale[float64]{}),
			nodetest.AssertOutput("Float 64", v4.Zero[float64]()),
			nodetest.AssertLiftedInput("Vector", vec4Type),
			nodetest.AssertLiftedInput("Amount", "float64"),
		),
		nodetest.NewTestCase(
			"an array without an amount is unchanged",
			nodetest.NewNode(vector4.Scale[float64]{
				Vector: nodetest.NewPortValue([]v4.Float64{v4.New(1.1, 2.2, 3.3, 4.4)}),
			}),
			nodetest.AssertOutput("Float 64", []v4.Float64{v4.New(1.1, 2.2, 3.3, 4.4)}),
			nodetest.AssertOutputType("Float 64", vec4ArrayType),
		),
		nodetest.NewTestCase(
			"one amount scales every vector in an array",
			nodetest.NewNode(vector4.Scale[float64]{
				Vector: nodetest.NewPortValue([]v4.Float64{v4.New(1., 2., 3., 4.)}),
				Amount: nodetest.NewPortValue(2.),
			}),
			nodetest.AssertOutput("Float 64", []v4.Float64{v4.New(2., 4., 6., 8.)}),
			nodetest.AssertOutput("Int", []v4.Int{v4.New(2, 4, 6, 8)}),
		),
		nodetest.NewTestCase(
			"an array of amounts scales index for index",
			nodetest.NewNode(vector4.Scale[float64]{
				Vector: nodetest.NewPortValue([]v4.Float64{
					v4.New(1., 1., 1., 1.), v4.New(2., 2., 2., 2.),
				}),
				Amount: nodetest.NewPortValue([]float64{0.5, 2}),
			}),
			nodetest.AssertOutput("Float 64", []v4.Float64{
				v4.New(0.5, 0.5, 0.5, 0.5), v4.New(4., 4., 4., 4.),
			}),
		),
	).Run(t)
}
