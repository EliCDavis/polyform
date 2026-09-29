package vector4_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector4"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v4 "github.com/EliCDavis/vector/vector4"
)

func TestSelectNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => 0,0,0,0",
			nodetest.NewNode(vector4.Select[float64]{}),
			nodetest.AssertOutput("X", 0.),
			nodetest.AssertOutput("Y", 0.),
			nodetest.AssertOutput("Z", 0.),
			nodetest.AssertOutput("W", 0.),
			nodetest.AssertLiftedInput("In", vec4Type),
		),
		nodetest.NewTestCase(
			"(1,2,3,4) => 1,2,3,4",
			nodetest.NewNode(vector4.Select[float64]{
				In: nodetest.NewPortValue(v4.New(1., 2., 3., 4.)),
			}),
			nodetest.AssertOutput("X", 1.),
			nodetest.AssertOutput("Y", 2.),
			nodetest.AssertOutput("Z", 3.),
			nodetest.AssertOutput("W", 4.),
		),
		nodetest.NewTestCase(
			"an array splits into four parallel arrays",
			nodetest.NewNode(vector4.Select[float64]{
				In: nodetest.NewPortValue([]v4.Float64{
					v4.New(1., 2., 3., 4.), v4.New(5., 6., 7., 8.),
				}),
			}),
			nodetest.AssertOutput("X", []float64{1, 5}),
			nodetest.AssertOutput("W", []float64{4, 8}),
			nodetest.AssertOutputType("X", "[]float64"),
		),
	).Run(t)
}

func TestNewNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"four scalars make one vector",
			nodetest.NewNode(vector4.NewNode[float64]{
				X: nodetest.NewPortValue(1.),
				Y: nodetest.NewPortValue(2.),
				Z: nodetest.NewPortValue(3.),
				W: nodetest.NewPortValue(4.),
			}),
			nodetest.AssertOutput("Out", v4.New(1., 2., 3., 4.)),
		),
		nodetest.NewTestCase(
			"four arrays zip into an array of vectors",
			nodetest.NewNode(vector4.NewNode[float64]{
				X: nodetest.NewPortValue([]float64{1, 5}),
				Y: nodetest.NewPortValue([]float64{2, 6}),
				Z: nodetest.NewPortValue([]float64{3, 7}),
				W: nodetest.NewPortValue([]float64{4, 8}),
			}),
			nodetest.AssertOutput("Out", []v4.Float64{
				v4.New(1., 2., 3., 4.), v4.New(5., 6., 7., 8.),
			}),
			nodetest.AssertOutputType("Out", vec4ArrayType),
		),
		nodetest.NewTestCase(
			"one component can stay a single value",
			nodetest.NewNode(vector4.NewNode[float64]{
				X: nodetest.NewPortValue([]float64{1, 5}),
				W: nodetest.NewPortValue(9.),
			}),
			nodetest.AssertOutput("Out", []v4.Float64{
				v4.New(1., 0., 0., 9.), v4.New(5., 0., 0., 9.),
			}),
		),
	).Run(t)
}
