package basics

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[ImageNode]](factory)
	refutil.RegisterType[nodes.Struct[BinaryNode]](factory)
	refutil.RegisterType[nodes.Struct[IONode]](factory)
	refutil.RegisterType[nodes.Struct[TextNode]](factory)

	generator.RegisterTypes(factory)
}
