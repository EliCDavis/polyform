package opearations

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[SeperateNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[SeperateNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[SeperateNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[SeperateNode[trs.TRS]]](factory)
	refutil.RegisterType[nodes.Struct[SeperateNode[quaternion.Quaternion]]](factory)

	generator.RegisterTypes(factory)
}
