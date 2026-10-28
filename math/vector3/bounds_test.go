package vector3_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/vector3"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v3 "github.com/EliCDavis/vector/vector3"
)

func TestBoundsNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => zero and an error",
			nodetest.NewNode(vector3.BoundsNode{}),
			nodetest.AssertOutput("Min", v3.Zero[float64]()),
			nodetest.AssertOutputError("Min", "no points to bound"),
		),
		nodetest.NewTestCase(
			"padding grows every side",
			nodetest.NewNode(vector3.BoundsNode{
				Points:  nodetest.NewPortValue([]v3.Float64{v3.New(1., 4., 0.), v3.New(-3., 2., 2.), v3.New(5., -2., 1.)}),
				Padding: nodetest.NewPortValue(1.),
			}),
			nodetest.AssertOutput("Min", v3.New(-4., -3., -1.)),
			nodetest.AssertOutput("Max", v3.New(6., 5., 3.)),
			nodetest.AssertOutput("Center", v3.New(1., 1., 1.)),
			nodetest.AssertOutput("Size", v3.New(10., 8., 4.)),
			nodetest.AssertOutput("AABB", geometry.NewAABB(v3.New(1., 1., 1.), v3.New(10., 8., 4.))),
		),
	).Run(t)
}
