package vector4_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector4"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v4 "github.com/EliCDavis/vector/vector4"
)

func TestSumNode(t *testing.T) {
	suite := nodetest.NewSuite(
		nodetest.NewTestCase(
			"nil => 0",
			nodetest.NewNode(vector4.SumNode[float64]{}),
			nodetest.AssertOutput("Out", v4.Zero[float64]()),
		),
		nodetest.NewTestCase(
			"[(1,2,3,4)] => (1,2,3,4)",
			nodetest.NewNode(vector4.SumNode[float64]{
				Values: []nodes.Output[v4.Vector[float64]]{
					nodetest.NewPortValue(v4.New(1., 2., 3., 4.)),
				},
			}),
			nodetest.AssertOutput("Out", v4.New(1., 2., 3., 4.)),
		),
		nodetest.NewTestCase(
			"[(1,2,3,4), (5,6,7,8)] => (6,8,10,12)",
			nodetest.NewNode(vector4.SumNode[float64]{
				Values: []nodes.Output[v4.Vector[float64]]{
					nodetest.NewPortValue(v4.New(1., 2., 3., 4.)),
					nodetest.NewPortValue(v4.New(5., 6., 7., 8.)),
				},
			}),
			nodetest.AssertOutput("Out", v4.New(6., 8., 10., 12.)),
		),
		nodetest.NewTestCase(
			"[(1,2,3,4), nil] => (1,2,3,4)",
			nodetest.NewNode(vector4.SumNode[float64]{
				Values: []nodes.Output[v4.Vector[float64]]{
					nodetest.NewPortValue(v4.New(1., 2., 3., 4.)),
					nil,
				},
			}),
			nodetest.AssertOutput("Out", v4.New(1., 2., 3., 4.)),
		),
	)
	suite.Run(t)
}

func TestAddToArrayNode(t *testing.T) {
	suite := nodetest.NewSuite(
		nodetest.NewTestCase(
			"(nil + nil) => nil",
			nodetest.NewNode(vector4.AddToArrayNode[float64]{}),
			nodetest.AssertOutput[[]v4.Float64]("Out", nil),
		),
		nodetest.NewTestCase(
			"((1,2,3,4) + nil) => nil",
			nodetest.NewNode(vector4.AddToArrayNode[float64]{
				Amount: nodetest.NewPortValue(v4.New(1., 2., 3., 4.)),
			}),
			nodetest.AssertOutput[[]v4.Float64]("Out", nil),
		),
		nodetest.NewTestCase(
			"(nil + [(1,2,3,4)]) => [(1,2,3,4)]",
			nodetest.NewNode(vector4.AddToArrayNode[float64]{
				Array: nodetest.NewPortValue([]v4.Float64{
					v4.New(1., 2., 3., 4.),
				}),
			}),
			nodetest.AssertOutput("Out", []v4.Float64{
				v4.New(1., 2., 3., 4.),
			}),
		),
		nodetest.NewTestCase(
			"((1,2,3,4) + [(1,1,1,1), (2,2,2,2)]) => [(2,3,4,5), (3,4,5,6)]",
			nodetest.NewNode(vector4.AddToArrayNode[float64]{
				Amount: nodetest.NewPortValue(v4.New(1., 2., 3., 4.)),
				Array: nodetest.NewPortValue([]v4.Float64{
					v4.New(1., 1., 1., 1.),
					v4.New(2., 2., 2., 2.),
				}),
			}),
			nodetest.AssertOutput("Out", []v4.Float64{
				v4.New(2., 3., 4., 5.),
				v4.New(3., 4., 5., 6.),
			}),
		),
	)
	suite.Run(t)
}
