package vector3_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector3"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v3 "github.com/EliCDavis/vector/vector3"
)

func TestSubtractNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => 0",
			nodetest.NewNode(vector3.Subtract[float64]{}),
			nodetest.AssertOutput("Out", v3.Zero[float64]()),
			nodetest.AssertLiftedInput("A", vec3Type),
		),
		nodetest.NewTestCase(
			"(4,5,6) - (1,2,3)",
			nodetest.NewNode(vector3.Subtract[float64]{
				A: nodetest.NewPortValue(v3.New(4., 5., 6.)),
				B: nodetest.NewPortValue(v3.New(1., 2., 3.)),
			}),
			nodetest.AssertOutput("Out", v3.New(3., 3., 3.)),
		),
		nodetest.NewTestCase(
			"an array minus nothing is unchanged",
			nodetest.NewNode(vector3.Subtract[float64]{
				A: nodetest.NewPortValue([]v3.Float64{v3.New(1., 2., 3.)}),
			}),
			nodetest.AssertOutput("Out", []v3.Float64{v3.New(1., 2., 3.)}),
		),
		nodetest.NewTestCase(
			"one vector subtracted from every entry of an array",
			nodetest.NewNode(vector3.Subtract[float64]{
				A: nodetest.NewPortValue([]v3.Float64{v3.New(1., 1., 1.), v3.New(2., 2., 2.)}),
				B: nodetest.NewPortValue(v3.New(1., 2., 3.)),
			}),
			nodetest.AssertOutput("Out", []v3.Float64{v3.New(0., -1., -2.), v3.New(1., 0., -1.)}),
			nodetest.AssertOutputType("Out", vec3ArrayType),
		),
		nodetest.NewTestCase(
			"two arrays subtract pair by pair",
			nodetest.NewNode(vector3.Subtract[float64]{
				A: nodetest.NewPortValue([]v3.Float64{v3.New(1., 1., 1.), v3.New(2., 2., 2.)}),
				B: nodetest.NewPortValue([]v3.Float64{v3.New(1., 0., 0.), v3.New(0., 2., 0.)}),
			}),
			nodetest.AssertOutput("Out", []v3.Float64{v3.New(0., 1., 1.), v3.New(2., 0., 2.)}),
		),
	).Run(t)
}
