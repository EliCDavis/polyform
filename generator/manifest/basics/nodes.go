package basics

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[ImageNode](factory)
	refutil.RegisterType[BinaryNode](factory)
	refutil.RegisterType[IONode](factory)
	refutil.RegisterType[TextNode](factory)

	generator.RegisterTypes(factory)
}
