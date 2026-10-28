package vector2_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector2"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v2 "github.com/EliCDavis/vector/vector2"
)

func TestBoundsNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => zero and an error",
			nodetest.NewNode(vector2.BoundsNode{}),
			nodetest.AssertOutput("Min", v2.Zero[float64]()),
			nodetest.AssertOutputError("Min", "no points to bound"),
		),
		nodetest.NewTestCase(
			"padding grows every side",
			nodetest.NewNode(vector2.BoundsNode{
				Points:  nodetest.NewPortValue([]v2.Float64{v2.New(1., 4.), v2.New(-3., 2.), v2.New(5., -2.)}),
				Padding: nodetest.NewPortValue(1.),
			}),
			nodetest.AssertOutput("Min", v2.New(-4., -3.)),
			nodetest.AssertOutput("Max", v2.New(6., 5.)),
			nodetest.AssertOutput("Center", v2.New(1., 1.)),
			nodetest.AssertOutput("Size", v2.New(10., 8.)),
		),
	).Run(t)
}
