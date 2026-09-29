package vector3_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector3"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v3 "github.com/EliCDavis/vector/vector3"
)

func TestScaleNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired",
			nodetest.NewNode(vector3.Scale[float64]{}),
			nodetest.AssertOutput("Float 64", v3.Zero[float64]()),
			nodetest.AssertOutput("Int", v3.Zero[int]()),
			nodetest.AssertLiftedInput("Vector", vec3Type),
			nodetest.AssertLiftedInput("Amount", "float64"),
		),

		nodetest.NewTestCase(
			"a single vector without an amount is unchanged",
			nodetest.NewNode(vector3.Scale[float64]{
				Vector: nodetest.NewPortValue(v3.New(1.1, 2.2, 3.7)),
			}),
			nodetest.AssertOutput("Float 64", v3.New(1.1, 2.2, 3.7)),
			nodetest.AssertOutput("Int", v3.New(1, 2, 4)),
			nodetest.AssertOutputType("Float 64", vec3Type),
		),

		nodetest.NewTestCase(
			"an array without an amount is unchanged",
			nodetest.NewNode(vector3.Scale[float64]{
				Vector: nodetest.NewPortValue([]v3.Float64{v3.New(1.1, 2.2, 3.7)}),
			}),
			nodetest.AssertOutput("Float 64", []v3.Float64{v3.New(1.1, 2.2, 3.7)}),
			nodetest.AssertOutput("Int", []v3.Int{v3.New(1, 2, 4)}),
			nodetest.AssertOutputType("Float 64", vec3ArrayType),
		),

		nodetest.NewTestCase(
			"one amount scales every vector in an array",
			nodetest.NewNode(vector3.Scale[float64]{
				Vector: nodetest.NewPortValue([]v3.Float64{v3.New(1.1, 2.3, 3.7)}),
				Amount: nodetest.NewPortValue(2.),
			}),
			nodetest.AssertOutput("Float 64", []v3.Float64{v3.New(2.2, 4.6, 7.4)}),
			nodetest.AssertOutput("Int", []v3.Int{v3.New(2, 5, 7)}),
		),

		nodetest.NewTestCase(
			"an array of amounts scales index for index",
			nodetest.NewNode(vector3.Scale[float64]{
				Vector: nodetest.NewPortValue([]v3.Float64{
					v3.New(1., 1., 1.), v3.New(2., 2., 2.),
				}),
				Amount: nodetest.NewPortValue([]float64{0.5, 2}),
			}),
			nodetest.AssertOutput("Float 64", []v3.Float64{
				v3.New(0.5, 0.5, 0.5), v3.New(4., 4., 4.),
			}),
		),

		nodetest.NewTestCase(
			"mismatched lengths are refused",
			nodetest.NewNode(vector3.Scale[float64]{
				Vector: nodetest.NewPortValue([]v3.Float64{
					v3.New(1., 1., 1.), v3.New(2., 2., 2.), v3.New(3., 3., 3.),
				}),
				Amount: nodetest.NewPortValue([]float64{0.5, 2}),
			}),
			nodetest.AssertOutput("Float 64", []v3.Float64{}),
			nodetest.AssertOutputError("Float 64", "3 and 2"),
		),
	).Run(t)
}
