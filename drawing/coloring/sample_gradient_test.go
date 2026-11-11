package coloring_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes/nodetest"
)

func TestSampleGradientNode(t *testing.T) {
	ramp := coloring.NewGradient1D(
		coloring.GradientKey[float64]{Time: 0, Value: 2},
		coloring.GradientKey[float64]{Time: 0.5, Value: 4},
		coloring.GradientKey[float64]{Time: 1, Value: 0},
	)

	black := coloring.Color{R: 0, G: 0, B: 0, A: 1}
	white := coloring.Color{R: 1, G: 1, B: 1, A: 1}
	grey := coloring.Color{R: 0.5, G: 0.5, B: 0.5, A: 1}
	colors := coloring.NewGradientColor(
		coloring.GradientKey[coloring.Color]{Time: 0, Value: black},
		coloring.GradientKey[coloring.Color]{Time: 1, Value: white},
	)

	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired is zero",
			nodetest.NewNode(coloring.SampleGradientNode[float64]{}),
			nodetest.AssertOutput("Out", 0.),
		),
		nodetest.NewTestCase(
			"one time",
			nodetest.NewNode(coloring.SampleGradientNode[float64]{
				Gradient: nodetest.NewPortValue(ramp),
				Time:     nodetest.NewPortValue(0.25),
			}),
			nodetest.AssertOutput("Out", 3.),
		),
		nodetest.NewTestCase(
			"times outside 0 to 1 hold the end keys",
			nodetest.NewNode(coloring.SampleGradientNode[float64]{
				Gradient: nodetest.NewPortValue(ramp),
				Time:     nodetest.NewPortValue([]float64{-1, 0.5, 0.75, 2}),
			}),
			nodetest.AssertOutput("Out", []float64{2, 4, 2, 0}),
			nodetest.AssertOutputType("Out", "[]float64"),
		),
		nodetest.NewTestCase(
			"an array of times through a color ramp",
			nodetest.NewNode(coloring.SampleGradientNode[coloring.Color]{
				Gradient: nodetest.NewPortValue(colors),
				Time:     nodetest.NewPortValue([]float64{0, 0.5, 1}),
			}),
			nodetest.AssertOutput("Out", []coloring.Color{black, grey, white}),
			nodetest.AssertOutputType("Out", "[]github.com/EliCDavis/polyform/drawing/coloring.Color"),
		),
	).Run(t)
}
