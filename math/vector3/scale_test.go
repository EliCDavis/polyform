package vector3_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector3"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v3 "github.com/EliCDavis/vector/vector3"
)

func TestScaleNode(t *testing.T) {
	suite := nodetest.NewSuite(
		nodetest.NewTestCase(
			"empty empty",
			nodetest.NewNode(vector3.ScaleArray[float64]{}),
			nodetest.AssertOutput[[]v3.Float64]("Float 64", nil),
			nodetest.AssertOutput[[]v3.Int]("Int", nil),
		),

		nodetest.NewTestCase(
			"empty empty w/ amount",
			nodetest.NewNode(vector3.ScaleArray[float64]{
				Amount: nodetest.NewPortValue(1.),
			}),
			nodetest.AssertOutput[[]v3.Float64]("Float 64", nil),
			nodetest.AssertOutput[[]v3.Int]("Int", nil),
		),

		nodetest.NewTestCase(
			"arr w/o amount",
			nodetest.NewNode(vector3.ScaleArray[float64]{
				Vector: nodetest.NewPortValue([]v3.Float64{
					v3.New(1.1, 2.2, 3.7),
				}),
			}),
			nodetest.AssertOutput("Float 64", []v3.Float64{
				v3.New(1.1, 2.2, 3.7),
			}),
			nodetest.AssertOutput("Int", []v3.Int{
				v3.New(1, 2, 4),
			}),
		),

		nodetest.NewTestCase(
			"arr w/ amount",
			nodetest.NewNode(vector3.ScaleArray[float64]{
				Vector: nodetest.NewPortValue([]v3.Float64{
					v3.New(1.1, 2.3, 3.7),
				}),
				Amount: nodetest.NewPortValue(2.),
			}),
			nodetest.AssertOutput("Float 64", []v3.Float64{
				v3.New(2.2, 4.6, 7.4),
			}),
			nodetest.AssertOutput("Int", []v3.Int{
				v3.New(2, 5, 7),
			}),
		),
	)
	suite.Run(t)
}
