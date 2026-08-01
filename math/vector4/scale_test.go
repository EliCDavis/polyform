package vector4_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector4"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v4 "github.com/EliCDavis/vector/vector4"
)

func TestScaleNode(t *testing.T) {
	suite := nodetest.NewSuite(
		nodetest.NewTestCase(
			"empty empty",
			nodetest.NewNode(vector4.ScaleArray[float64]{}),
			nodetest.AssertOutput[[]v4.Float64]("Float 64", nil),
			nodetest.AssertOutput[[]v4.Int]("Int", nil),
		),

		nodetest.NewTestCase(
			"empty empty w/ amount",
			nodetest.NewNode(vector4.ScaleArray[float64]{
				Amount: nodetest.NewPortValue(1.),
			}),
			nodetest.AssertOutput[[]v4.Float64]("Float 64", nil),
			nodetest.AssertOutput[[]v4.Int]("Int", nil),
		),

		nodetest.NewTestCase(
			"arr w/o amount",
			nodetest.NewNode(vector4.ScaleArray[float64]{
				Vector: nodetest.NewPortValue([]v4.Float64{
					v4.New(1.1, 2.2, 3.7, 4.8),
				}),
			}),
			nodetest.AssertOutput("Float 64", []v4.Float64{
				v4.New(1.1, 2.2, 3.7, 4.8),
			}),
			nodetest.AssertOutput("Int", []v4.Int{
				v4.New(1, 2, 4, 5),
			}),
		),

		nodetest.NewTestCase(
			"arr w/ amount",
			nodetest.NewNode(vector4.ScaleArray[float64]{
				Vector: nodetest.NewPortValue([]v4.Float64{
					v4.New(1.1, 2.3, 3.7, 4.8),
				}),
				Amount: nodetest.NewPortValue(2.),
			}),
			nodetest.AssertOutput("Float 64", []v4.Float64{
				v4.New(2.2, 4.6, 7.4, 9.6),
			}),
			nodetest.AssertOutput("Int", []v4.Int{
				v4.New(2, 5, 7, 10),
			}),
		),
	)
	suite.Run(t)
}
