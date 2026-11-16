package math_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes/nodetest"
)

func TestCumulativeNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"running totals",
			nodetest.NewNode(math.CumulativeNode{
				Values: nodetest.NewPortValue([]float64{3, 1, 2}),
			}),
			nodetest.AssertOutput("Inclusive", []float64{3, 4, 6}),
			nodetest.AssertOutput("Exclusive", []float64{0, 3, 4}),
		),
		nodetest.NewTestCase(
			"one list of totals per list",
			nodetest.NewNode(math.CumulativeNode{
				Values: nodetest.NewPortValue([][]float64{{1, 1}, {2, 5}}),
			}),
			nodetest.AssertOutput("Inclusive", [][]float64{{1, 2}, {2, 7}}),
		),
		nodetest.NewTestCase(
			"empty",
			nodetest.NewNode(math.CumulativeNode{
				Values: nodetest.NewPortValue([]float64{}),
			}),
			nodetest.AssertOutput("Exclusive", []float64{}),
		),
	).Run(t)
}
