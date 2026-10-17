package text_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	"github.com/EliCDavis/polyform/text"
)

func TestConcatNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => empty",
			nodetest.NewNode(text.ConcatNode{}),
			nodetest.AssertOutput("Out", ""),
			nodetest.AssertOutputType("Out", "string"),
			nodetest.AssertLiftedInput("Strings", "string"),
		),
		nodetest.NewTestCase(
			"joined in order",
			nodetest.NewNode(text.ConcatNode{Strings: nodetest.NewPortValues("Passive", " ", "Body")}),
			nodetest.AssertOutput("Out", "Passive Body"),
		),
		nodetest.NewTestCase(
			"separator between each",
			nodetest.NewNode(text.ConcatNode{
				Strings:   nodetest.NewPortValues("a", "b", "c"),
				Separator: nodetest.NewPortValue(", "),
			}),
			nodetest.AssertOutput("Out", "a, b, c"),
		),
		nodetest.NewTestCase(
			"a list joins element by element",
			nodetest.NewNode(text.ConcatNode{Strings: []nodes.LiftedPort[string]{
				nodetest.NewPortValue([]string{"R1", "R2"}),
				nodetest.NewPortValue(" Body"),
			}}),
			nodetest.AssertOutput("Out", []string{"R1 Body", "R2 Body"}),
			nodetest.AssertOutputType("Out", "[]string"),
		),
	).Run(t)
}
