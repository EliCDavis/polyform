package vector4_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector4"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v4 "github.com/EliCDavis/vector/vector4"
)

const (
	vec4Type      = "github.com/EliCDavis/vector/vector4.Vector[float64]"
	vec4ArrayType = "[]" + vec4Type
)

func TestSumNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => 0",
			nodetest.NewNode(vector4.SumNode[float64]{}),
			nodetest.AssertOutput("Out", v4.Zero[float64]()),
			nodetest.AssertLiftedInput("Values", vec4Type),
		),
		nodetest.NewTestCase(
			"two vectors add",
			nodetest.NewNode(vector4.SumNode[float64]{
				Values: nodetest.NewPortValues(v4.New(1., 2., 3., 4.), v4.New(5., 6., 7., 8.)),
			}),
			nodetest.AssertOutput("Out", v4.New(6., 8., 10., 12.)),
		),
		nodetest.NewTestCase(
			"an unwired connection is skipped",
			nodetest.NewNode(vector4.SumNode[float64]{
				Values: []nodes.LiftedPort[v4.Vector[float64]]{
					nodetest.NewPortValue(v4.New(1., 2., 3., 4.)),
					nil,
				},
			}),
			nodetest.AssertOutput("Out", v4.New(1., 2., 3., 4.)),
		),
		nodetest.NewTestCase(
			"one vector added to every entry of an array",
			nodetest.NewNode(vector4.SumNode[float64]{
				Values: []nodes.LiftedPort[v4.Vector[float64]]{
					nodetest.NewPortValue([]v4.Float64{
						v4.New(1., 1., 1., 1.), v4.New(2., 2., 2., 2.),
					}),
					nodetest.NewPortValue(v4.New(1., 2., 3., 4.)),
				},
			}),
			nodetest.AssertOutput("Out", []v4.Float64{
				v4.New(2., 3., 4., 5.), v4.New(3., 4., 5., 6.),
			}),
			nodetest.AssertOutputType("Out", vec4ArrayType),
		),
	).Run(t)
}

func TestSubtractNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"(5,6,7,8) - (1,2,3,4)",
			nodetest.NewNode(vector4.Subtract[float64]{
				A: nodetest.NewPortValue(v4.New(5., 6., 7., 8.)),
				B: nodetest.NewPortValue(v4.New(1., 2., 3., 4.)),
			}),
			nodetest.AssertOutput("Out", v4.New(4., 4., 4., 4.)),
		),
		nodetest.NewTestCase(
			"one vector subtracted from every entry of an array",
			nodetest.NewNode(vector4.Subtract[float64]{
				A: nodetest.NewPortValue([]v4.Float64{
					v4.New(1., 1., 1., 1.), v4.New(2., 2., 2., 2.),
				}),
				B: nodetest.NewPortValue(v4.New(1., 2., 3., 4.)),
			}),
			nodetest.AssertOutput("Out", []v4.Float64{
				v4.New(0., -1., -2., -3.), v4.New(1., 0., -1., -2.),
			}),
			nodetest.AssertOutputType("Out", vec4ArrayType),
		),
	).Run(t)
}
