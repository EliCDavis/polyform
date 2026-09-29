package vector3_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector3"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v3 "github.com/EliCDavis/vector/vector3"
)

func TestMultiplyNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"mirror across yz",
			nodetest.NewNode(vector3.Multiply[float64]{
				A: nodetest.NewPortValue(v3.New(0.055, 0.165, 0.078)),
				B: nodetest.NewPortValue(v3.New(-1., 1., 1.)),
			}),
			nodetest.AssertOutput("Float 64", v3.New(-0.055, 0.165, 0.078)),
		),
		nodetest.NewTestCase(
			"missing B is identity",
			nodetest.NewNode(vector3.Multiply[float64]{
				A: nodetest.NewPortValue(v3.New(1., 2., 3.)),
			}),
			nodetest.AssertOutput("Float 64", v3.New(1., 2., 3.)),
			nodetest.AssertOutput("Int", v3.New(1, 2, 3)),
		),
		nodetest.NewTestCase(
			"mirror a point list",
			nodetest.NewNode(vector3.Multiply[float64]{
				A: nodetest.NewPortValue([]v3.Float64{v3.New(1., 2., 3.), v3.New(-1., 0., 0.)}),
				B: nodetest.NewPortValue(v3.New(-1., 1., 1.)),
			}),
			nodetest.AssertOutput("Float 64", []v3.Float64{v3.New(-1., 2., 3.), v3.New(1., 0., 0.)}),
			nodetest.AssertOutputType("Float 64", vec3ArrayType),
		),
		nodetest.NewTestCase(
			"two point lists multiply pair by pair",
			nodetest.NewNode(vector3.Multiply[float64]{
				A: nodetest.NewPortValue([]v3.Float64{v3.New(1., 2., 3.), v3.New(4., 5., 6.)}),
				B: nodetest.NewPortValue([]v3.Float64{v3.New(2., 2., 2.), v3.New(0., 1., 0.)}),
			}),
			nodetest.AssertOutput("Float 64", []v3.Float64{v3.New(2., 4., 6.), v3.New(0., 5., 0.)}),
		),
		nodetest.NewTestCase(
			"nothing wired",
			nodetest.NewNode(vector3.Multiply[float64]{}),
			nodetest.AssertOutput("Float 64", v3.Zero[float64]()),
		),
	).Run(t)
}
