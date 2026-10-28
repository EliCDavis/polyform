package math_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes/nodetest"
)

func TestBoundsNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => 0 and an error",
			nodetest.NewNode(math.BoundsNode{}),
			nodetest.AssertOutput("Min", 0.),
			nodetest.AssertOutput("Size", 0.),
			nodetest.AssertOutputError("Min", "no values to bound"),
		),
		nodetest.NewTestCase(
			"3, -1, 5 => -1..5",
			nodetest.NewNode(math.BoundsNode{
				Values: nodetest.NewPortValue([]float64{3, -1, 5}),
			}),
			nodetest.AssertOutput("Min", -1.),
			nodetest.AssertOutput("Max", 5.),
			nodetest.AssertOutput("Center", 2.),
			nodetest.AssertOutput("Size", 6.),
		),
		nodetest.NewTestCase(
			"padding grows both ends",
			nodetest.NewNode(math.BoundsNode{
				Values:  nodetest.NewPortValue([]float64{3, -1, 5}),
				Padding: nodetest.NewPortValue(0.5),
			}),
			nodetest.AssertOutput("Min", -1.5),
			nodetest.AssertOutput("Max", 5.5),
			nodetest.AssertOutput("Center", 2.),
			nodetest.AssertOutput("Size", 7.),
		),
	).Run(t)
}
