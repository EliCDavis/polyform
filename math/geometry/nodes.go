package geometry

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[LinesFromPoints3DNode]](factory)
	refutil.RegisterType[nodes.Struct[LineLengths3DNode]](factory)
	refutil.RegisterType[nodes.Struct[PositionsOnLinesAtTime3DNode]](factory)
	refutil.RegisterType[nodes.Struct[PositionsOnLineAtTimes3DNode]](factory)
	refutil.RegisterType[nodes.Struct[TrsFromLines3DNode]](factory)
	generator.RegisterTypes(factory)
}
