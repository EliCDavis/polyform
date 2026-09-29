package vector3_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector3"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v3 "github.com/EliCDavis/vector/vector3"
)

func TestSelectNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => 0,0,0",
			nodetest.NewNode(vector3.Select[float64]{}),
			nodetest.AssertOutput("X", 0.),
			nodetest.AssertOutput("Y", 0.),
			nodetest.AssertOutput("Z", 0.),
			nodetest.AssertLiftedInput("In", vec3Type),
		),
		nodetest.NewTestCase(
			"(1,2,3) => 1,2,3",
			nodetest.NewNode(vector3.Select[float64]{
				In: nodetest.NewPortValue(v3.New(1., 2., 3.)),
			}),
			nodetest.AssertOutput("X", 1.),
			nodetest.AssertOutput("Y", 2.),
			nodetest.AssertOutput("Z", 3.),
			nodetest.AssertOutputType("X", "float64"),
		),
		nodetest.NewTestCase(
			"an array splits into three parallel arrays",
			nodetest.NewNode(vector3.Select[float64]{
				In: nodetest.NewPortValue([]v3.Float64{v3.New(1., 2., 3.), v3.New(4., 5., 6.)}),
			}),
			nodetest.AssertOutput("X", []float64{1, 4}),
			nodetest.AssertOutput("Y", []float64{2, 5}),
			nodetest.AssertOutput("Z", []float64{3, 6}),
			nodetest.AssertOutputType("X", "[]float64"),
		),
		nodetest.NewTestCase(
			"an empty array splits into empty arrays",
			nodetest.NewNode(vector3.Select[float64]{
				In: nodetest.NewPortValue([]v3.Float64{}),
			}),
			nodetest.AssertOutput("X", []float64{}),
		),
	).Run(t)
}

func TestNewNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"three scalars make one vector",
			nodetest.NewNode(vector3.NewNode[float64]{
				X: nodetest.NewPortValue(1.),
				Y: nodetest.NewPortValue(2.),
				Z: nodetest.NewPortValue(3.),
			}),
			nodetest.AssertOutput("Out", v3.New(1., 2., 3.)),
		),
		nodetest.NewTestCase(
			"three arrays zip into an array of vectors",
			nodetest.NewNode(vector3.NewNode[float64]{
				X: nodetest.NewPortValue([]float64{1, 4}),
				Y: nodetest.NewPortValue([]float64{2, 5}),
				Z: nodetest.NewPortValue([]float64{3, 6}),
			}),
			nodetest.AssertOutput("Out", []v3.Float64{v3.New(1., 2., 3.), v3.New(4., 5., 6.)}),
			nodetest.AssertOutputType("Out", vec3ArrayType),
		),
		nodetest.NewTestCase(
			"one component can stay a single value",
			nodetest.NewNode(vector3.NewNode[float64]{
				X: nodetest.NewPortValue([]float64{1, 4}),
				Z: nodetest.NewPortValue(9.),
			}),
			nodetest.AssertOutput("Out", []v3.Float64{v3.New(1., 0., 9.), v3.New(4., 0., 9.)}),
		),
	).Run(t)
}
