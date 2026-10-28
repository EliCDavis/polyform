package vector4_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector4"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v4 "github.com/EliCDavis/vector/vector4"
)

func TestBoundsNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => zero and an error",
			nodetest.NewNode(vector4.BoundsNode{}),
			nodetest.AssertOutput("Min", v4.Zero[float64]()),
			nodetest.AssertOutputError("Min", "no points to bound"),
		),
		nodetest.NewTestCase(
			"padding grows every side",
			nodetest.NewNode(vector4.BoundsNode{
				Points:  nodetest.NewPortValue([]v4.Float64{v4.New(1., 4., 0., 2.), v4.New(-3., 2., 2., 2.), v4.New(5., -2., 1., 2.)}),
				Padding: nodetest.NewPortValue(1.),
			}),
			nodetest.AssertOutput("Min", v4.New(-4., -3., -1., 1.)),
			nodetest.AssertOutput("Max", v4.New(6., 5., 3., 3.)),
			nodetest.AssertOutput("Center", v4.New(1., 1., 1., 2.)),
			nodetest.AssertOutput("Size", v4.New(10., 8., 4., 2.)),
		),
	).Run(t)
}
