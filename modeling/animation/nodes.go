package animation

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[UniformFramesNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[UniformFramesNode[quaternion.Quaternion]]](factory)

	generator.RegisterTypes(factory)
}

type UniformFramesNode[T any] struct {
	Data     nodes.Output[[]T]
	Duration nodes.Output[float64]
}

func (node UniformFramesNode[T]) Out(out *nodes.StructOutput[[]Frame[T]]) {
	out.Set(UniformFrames(
		nodes.TryGetOutputValue(out, node.Data, nil),
		nodes.TryGetOutputValue(out, node.Duration, 1),
	))
}
