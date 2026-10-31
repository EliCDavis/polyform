package basics

import (
	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[ImageNode]](factory)
	refutil.RegisterType[nodes.Struct[BinaryNode]](factory)
	refutil.RegisterType[nodes.Struct[IONode]](factory)
	refutil.RegisterType[nodes.Struct[TextNode]](factory)

	registry.Register(factory)
}
