package subgraph

import (
	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	factory.RegisterBuilder(InputNodeTypeKey, func() any { return NewInputNode("", "") })
	factory.RegisterBuilder(OutputNodeTypeKey, func() any { return NewOutputNode("", "") })
	registry.Register(factory)
}
