package coloring_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes/nodetest"
)

func TestInterpolateNodeOverArrays(t *testing.T) {
	black := coloring.Color{R: 0, G: 0, B: 0, A: 1}
	white := coloring.Color{R: 1, G: 1, B: 1, A: 1}
	red := coloring.Color{R: 1, G: 0, B: 0, A: 1}
	pink := coloring.Color{R: 1, G: 0.5, B: 0.5, A: 1}

	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired is opaque black",
			nodetest.NewNode(coloring.InterpolateNode{}),
			nodetest.AssertOutput("Out", black),
		),
		nodetest.NewTestCase(
			"one end alone is that end",
			nodetest.NewNode(coloring.InterpolateNode{
				B: nodetest.NewPortValue(red),
			}),
			nodetest.AssertOutput("Out", red),
		),
		nodetest.NewTestCase(
			"halfway between two colors",
			nodetest.NewNode(coloring.InterpolateNode{
				A: nodetest.NewPortValue(red),
				B: nodetest.NewPortValue(white),
			}),
			nodetest.AssertOutput("Out", pink),
		),
		nodetest.NewTestCase(
			"an array of times against two colors",
			nodetest.NewNode(coloring.InterpolateNode{
				A:    nodetest.NewPortValue(black),
				B:    nodetest.NewPortValue(white),
				Time: nodetest.NewPortValue([]float64{0, 1}),
			}),
			nodetest.AssertOutput("Out", []coloring.Color{black, white}),
			nodetest.AssertOutputType("Out", "[]github.com/EliCDavis/polyform/drawing/coloring.Color"),
		),
		nodetest.NewTestCase(
			"two arrays of colors blend pair by pair",
			nodetest.NewNode(coloring.InterpolateNode{
				A:    nodetest.NewPortValue([]coloring.Color{black, black, red}),
				B:    nodetest.NewPortValue([]coloring.Color{white, white, white}),
				Time: nodetest.NewPortValue([]float64{0, 1, 0.5}),
			}),
			nodetest.AssertOutput("Out", []coloring.Color{black, white, pink}),
		),
		nodetest.NewTestCase(
			"a single-entry array stretches to meet a longer one",
			nodetest.NewNode(coloring.InterpolateNode{
				A:    nodetest.NewPortValue([]coloring.Color{black, black}),
				B:    nodetest.NewPortValue([]coloring.Color{white}),
				Time: nodetest.NewPortValue([]float64{1, 1}),
			}),
			nodetest.AssertOutput("Out", []coloring.Color{white, white}),
		),
		nodetest.NewTestCase(
			"mismatched lengths are refused rather than truncated",
			nodetest.NewNode(coloring.InterpolateNode{
				A:    nodetest.NewPortValue([]coloring.Color{black, black}),
				B:    nodetest.NewPortValue([]coloring.Color{white, white, white}),
				Time: nodetest.NewPortValue([]float64{1, 1}),
			}),
			nodetest.AssertOutput("Out", []coloring.Color{}),
			nodetest.AssertOutputError("Out", "2 and 3"),
		),
	).Run(t)
}
