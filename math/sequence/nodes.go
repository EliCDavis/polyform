package sequence

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[LinearNode]](factory)
	refutil.RegisterType[nodes.Struct[RandomFloatNode]](factory)
	refutil.RegisterType[nodes.Struct[RandomBoolNode]](factory)

	generator.RegisterTypes(factory)
}
