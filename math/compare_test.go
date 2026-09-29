package math_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes/nodetest"
)

func TestCompareNode(t *testing.T) {
	two := nodetest.NewNode(math.CompareNode[int]{A: nodetest.NewPortValue(2), B: nodetest.NewPortValue(2)})
	oneTwo := nodetest.NewNode(math.CompareNode[int]{A: nodetest.NewPortValue(1), B: nodetest.NewPortValue(2)})
	suite := nodetest.NewSuite(
		nodetest.NewTestCase("equal", two,
			nodetest.AssertOutput("Equal", true),
			nodetest.AssertOutput("Not Equal", false),
			nodetest.AssertOutput("Greater", false),
			nodetest.AssertOutput("Greater Or Equal", true),
			nodetest.AssertOutput("Less", false),
			nodetest.AssertOutput("Less Or Equal", true),
		),
		nodetest.NewTestCase("less", oneTwo,
			nodetest.AssertOutput("Equal", false),
			nodetest.AssertOutput("Less", true),
			nodetest.AssertOutput("Greater Or Equal", false),
		),
		nodetest.NewTestCase("unwired reads as zero",
			nodetest.NewNode(math.CompareNode[float64]{B: nodetest.NewPortValue(0.)}),
			nodetest.AssertOutput("Equal", true),
		),
	)
	suite.Run(t)
}

func TestBoolToNumberAndNot(t *testing.T) {
	suite := nodetest.NewSuite(
		nodetest.NewTestCase("true",
			nodetest.NewNode(math.BoolToNumberNode{In: nodetest.NewPortValue(true)}),
			nodetest.AssertOutput("Int", 1),
			nodetest.AssertOutput("Float 64", 1.),
		),
		nodetest.NewTestCase("false",
			nodetest.NewNode(math.BoolToNumberNode{In: nodetest.NewPortValue(false)}),
			nodetest.AssertOutput("Int", 0),
			nodetest.AssertOutput("Float 64", 0.),
		),
		nodetest.NewTestCase("not",
			nodetest.NewNode(math.NotNode{In: nodetest.NewPortValue(true)}),
			nodetest.AssertOutput("Out", false),
		),
	)
	suite.Run(t)
}
