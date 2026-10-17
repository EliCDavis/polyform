package text

import (
	"strings"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[ConcatNode]](factory)

	generator.RegisterTypes(factory)
}

type ConcatNode struct {
	Strings   []nodes.LiftedPort[string] `description:"The strings to join, in order. Lists are joined element by element."`
	Separator nodes.Output[string]       `description:"Placed between each pair of strings. Defaults to nothing."`
}

func (ConcatNode) Description() string {
	return "Joins strings end to end."
}

func (ConcatNode) Keywords() []string {
	return []string{"format"}
}

func (n ConcatNode) Out(out *nodes.Lifted[string]) {
	separator := nodes.TryGetOutputValue(out, n.Separator, "")
	nodes.ZipAll(out, n.Strings, func(values []string) string {
		return strings.Join(values, separator)
	})
}
