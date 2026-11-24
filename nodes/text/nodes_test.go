package text_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	"github.com/EliCDavis/polyform/nodes/text"
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

func TestEqualNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"same strings",
			nodetest.NewNode(text.EqualNode{A: nodetest.NewPortValue("Esc"), B: nodetest.NewPortValue("Esc")}),
			nodetest.AssertOutput("Equal", true),
			nodetest.AssertOutput("Not Equal", false),
		),
		nodetest.NewTestCase(
			"a list compares element by element",
			nodetest.NewNode(text.EqualNode{A: nodetest.NewPortValue([]string{"Esc", "Tab", "esc"}), B: nodetest.NewPortValue("Esc")}),
			nodetest.AssertOutput("Equal", []bool{true, false, false}),
		),
	).Run(t)
}

func TestSplitNode(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"splits at whitespace by default",
			nodetest.NewNode(text.SplitNode{Text: nodetest.NewPortValue("  POLY FORM\tKEYS ")}),
			nodetest.AssertOutput("Out", []string{"POLY", "FORM", "KEYS"}),
		),
		nodetest.NewTestCase(
			"a list splits each string",
			nodetest.NewNode(text.SplitNode{
				Text:      nodetest.NewPortValue([]string{"a,b", "c"}),
				Separator: nodetest.NewPortValue(","),
			}),
			nodetest.AssertOutput("Out", [][]string{{"a", "b"}, {"c"}}),
		),
	).Run(t)
}
