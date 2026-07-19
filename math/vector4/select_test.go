package vector4_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector4"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v4 "github.com/EliCDavis/vector/vector4"
)

func TestSelectNode(t *testing.T) {
	suite := nodetest.NewSuite(
		nodetest.NewTestCase(
			"(nil) => 0,0,0,0",
			nodetest.NewNode(vector4.Select[float64]{}),
			nodetest.AssertOutput("X", 0.),
			nodetest.AssertOutput("Y", 0.),
			nodetest.AssertOutput("Z", 0.),
			nodetest.AssertOutput("W", 0.),
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
	)
	suite.Run(t)
}

func TestSelectArrayNode(t *testing.T) {
	suite := nodetest.NewSuite(
		nodetest.NewTestCase(
			"(nil) => empty",
			nodetest.NewNode(vector4.SelectArray[float64]{}),
			nodetest.AssertOutput("X", []float64{}),
			nodetest.AssertOutput("Y", []float64{}),
			nodetest.AssertOutput("Z", []float64{}),
			nodetest.AssertOutput("W", []float64{}),
		),
		nodetest.NewTestCase(
			"(1,2,3,4) => []1, []2, []3, []4",
			nodetest.NewNode(vector4.SelectArray[float64]{
				In: nodetest.NewPortValue([]v4.Vector[float64]{
					v4.New(1., 2., 3., 4.),
				}),
			}),
			nodetest.AssertOutput("X", []float64{1.}),
			nodetest.AssertOutput("Y", []float64{2.}),
			nodetest.AssertOutput("Z", []float64{3.}),
			nodetest.AssertOutput("W", []float64{4.}),
		),
	)
	suite.Run(t)
}
