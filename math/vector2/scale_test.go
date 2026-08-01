package vector2_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector2"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v2 "github.com/EliCDavis/vector/vector2"
)

func TestScaleNode(t *testing.T) {
	suite := nodetest.NewSuite(
		nodetest.NewTestCase(
			"empty empty",
			nodetest.NewNode(vector2.ScaleArray[float64]{}),
			nodetest.AssertOutput[[]v2.Float64]("Float 64", nil),
			nodetest.AssertOutput[[]v2.Int]("Int", nil),
		),

		nodetest.NewTestCase(
			"empty empty w/ amount",
			nodetest.NewNode(vector2.ScaleArray[float64]{
				Amount: nodetest.NewPortValue(1.),
			}),
			nodetest.AssertOutput[[]v2.Float64]("Float 64", nil),
			nodetest.AssertOutput[[]v2.Int]("Int", nil),
		),

		nodetest.NewTestCase(
			"arr w/o amount",
			nodetest.NewNode(vector2.ScaleArray[float64]{
				Vector: nodetest.NewPortValue([]v2.Float64{
					v2.New(1.1, 2.2),
				}),
			}),
			nodetest.AssertOutput("Float 64", []v2.Float64{
				v2.New(1.1, 2.2),
			}),
			nodetest.AssertOutput("Int", []v2.Int{
				v2.New(1, 2),
			}),
		),

		nodetest.NewTestCase(
			"arr w/ amount",
			nodetest.NewNode(vector2.ScaleArray[float64]{
				Vector: nodetest.NewPortValue([]v2.Float64{
					v2.New(1.1, 2.3),
				}),
				Amount: nodetest.NewPortValue(2.),
			}),
			nodetest.AssertOutput("Float 64", []v2.Float64{
				v2.New(2.2, 4.6),
			}),
			nodetest.AssertOutput("Int", []v2.Int{
				v2.New(2, 5),
			}),
		),
	)
	suite.Run(t)
}
