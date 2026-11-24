package text

import (
	"strings"

	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[ConcatNode]](factory)
	refutil.RegisterType[nodes.Struct[EqualNode]](factory)
	refutil.RegisterType[nodes.Struct[SplitNode]](factory)

	registry.Register(factory)
}

type EqualNode struct {
	A nodes.LiftedPort[string]
	B nodes.LiftedPort[string]
}

func (EqualNode) Description() string {
	return "Whether two strings match exactly."
}

func (EqualNode) Keywords() []string {
	return []string{"compare"}
}

func (n EqualNode) Equal(out *nodes.Lifted[bool]) {
	nodes.Zip2(out, n.A, n.B, func(a, b string) bool { return a == b })
}

func (n EqualNode) NotEqual(out *nodes.Lifted[bool]) {
	nodes.Zip2(out, n.A, n.B, func(a, b string) bool { return a != b })
}

type SplitNode struct {
	Text      nodes.LiftedPort[string]
	Separator nodes.Output[string] `description:"Empty splits at runs of whitespace."`
}

func (SplitNode) Description() string {
	return "Cuts a string into the pieces between each separator."
}

func (SplitNode) Keywords() []string {
	return []string{"words", "tokenize"}
}

func (n SplitNode) Out(out *nodes.Lifted[[]string]) {
	separator := nodes.GetOutputValue(out, n.Separator)
	nodes.Zip1(out, n.Text, func(text string) []string {
		if separator == "" {
			return strings.Fields(text)
		}
		return strings.Split(text, separator)
	})
}

type ConcatNode struct {
	Strings   []nodes.LiftedPort[string] `description:"The strings to join, in order."`
	Separator nodes.Output[string]       `description:"Placed between each pair of strings."`
}

func (ConcatNode) Description() string {
	return "Joins strings end to end."
}

func (ConcatNode) Keywords() []string {
	return []string{"format"}
}

func (n ConcatNode) Out(out *nodes.Lifted[string]) {
	separator := nodes.GetOutputValue(out, n.Separator)
	nodes.ZipAll(out, n.Strings, func(values []string) string {
		return strings.Join(values, separator)
	})
}
