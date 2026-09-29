package vector3_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector3"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v3 "github.com/EliCDavis/vector/vector3"
)

const (
	vec3Type      = "github.com/EliCDavis/vector/vector3.Vector[float64]"
	vec3ArrayType = "[]" + vec3Type
)

func TestSumNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => 0",
			nodetest.NewNode(vector3.SumNode[float64]{}),
			nodetest.AssertOutput("Out", v3.Zero[float64]()),
			nodetest.AssertLiftedInput("Values", vec3Type),
		),
		nodetest.NewTestCase(
			"one vector is itself",
			nodetest.NewNode(vector3.SumNode[float64]{
				Values: nodetest.NewPortValues(v3.New(1., 2., 3.)),
			}),
			nodetest.AssertOutput("Out", v3.New(1., 2., 3.)),
		),
		nodetest.NewTestCase(
			"two vectors add",
			nodetest.NewNode(vector3.SumNode[float64]{
				Values: nodetest.NewPortValues(v3.New(1., 2., 3.), v3.New(4., 5., 6.)),
			}),
			nodetest.AssertOutput("Out", v3.New(5., 7., 9.)),
		),
		nodetest.NewTestCase(
			"an unwired connection is skipped",
			nodetest.NewNode(vector3.SumNode[float64]{
				Values: []nodes.LiftedPort[v3.Vector[float64]]{
					nodetest.NewPortValue(v3.New(1., 2., 3.)),
					nil,
				},
			}),
			nodetest.AssertOutput("Out", v3.New(1., 2., 3.)),
		),
		nodetest.NewTestCase(
			"one vector added to every entry of an array",
			nodetest.NewNode(vector3.SumNode[float64]{
				Values: []nodes.LiftedPort[v3.Vector[float64]]{
					nodetest.NewPortValue([]v3.Float64{v3.New(1., 1., 1.), v3.New(2., 2., 2.)}),
					nodetest.NewPortValue(v3.New(1., 2., 3.)),
				},
			}),
			nodetest.AssertOutput("Out", []v3.Float64{v3.New(2., 3., 4.), v3.New(3., 4., 5.)}),
			nodetest.AssertOutputType("Out", vec3ArrayType),
		),
		nodetest.NewTestCase(
			"two arrays add pair by pair",
			nodetest.NewNode(vector3.SumNode[float64]{
				Values: []nodes.LiftedPort[v3.Vector[float64]]{
					nodetest.NewPortValue([]v3.Float64{v3.New(1., 1., 1.), v3.New(2., 2., 2.)}),
					nodetest.NewPortValue([]v3.Float64{v3.New(0., 1., 0.), v3.New(1., 0., 1.)}),
				},
			}),
			nodetest.AssertOutput("Out", []v3.Float64{v3.New(1., 2., 1.), v3.New(3., 2., 3.)}),
		),
	).Run(t)
}
