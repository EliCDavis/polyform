package coloring_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes/nodetest"
)

func TestGradientKeyNodeOverArrays(t *testing.T) {
	tan := coloring.Color{R: 0.95, G: 0.89, B: 0.77, A: 1}
	brown := coloring.Color{R: 0.4, G: 0.25, B: 0.1, A: 1}

	nodetest.NewSuite(
		nodetest.NewTestCase(
			"single time and value make one key",
			nodetest.NewNode(coloring.GradientKeyNode[coloring.Color]{
				Time:  nodetest.NewPortValue(0.35),
				Value: nodetest.NewPortValue(tan),
			}),
			nodetest.AssertOutput("Gradient", coloring.GradientKey[coloring.Color]{Time: 0.35, Value: tan}),
		),
		nodetest.NewTestCase(
			"lists of times and colors make a list of keys",
			nodetest.NewNode(coloring.GradientKeyNode[coloring.Color]{
				Time:  nodetest.NewPortValue([]float64{0, 0.35}),
				Value: nodetest.NewPortValue([]coloring.Color{tan, brown}),
			}),
			nodetest.AssertOutput("Gradient", []coloring.GradientKey[coloring.Color]{
				{Time: 0, Value: tan},
				{Time: 0.35, Value: brown},
			}),
		),
		nodetest.NewTestCase(
			"one color broadcasts across a list of times",
			nodetest.NewNode(coloring.GradientKeyNode[coloring.Color]{
				Time:  nodetest.NewPortValue([]float64{0, 1}),
				Value: nodetest.NewPortValue(brown),
			}),
			nodetest.AssertOutput("Gradient", []coloring.GradientKey[coloring.Color]{
				{Time: 0, Value: brown},
				{Time: 1, Value: brown},
			}),
		),
	).Run(t)
}
