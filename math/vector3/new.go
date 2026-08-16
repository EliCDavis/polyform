package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type NewNode[T vector.Number] struct {
	X nodes.Output[T] `description:"Defaults to 0."`
	Y nodes.Output[T] `description:"Defaults to 0."`
	Z nodes.Output[T] `description:"Defaults to 0."`
}

func (cn NewNode[T]) Description() string {
	return "Builds a vector3 from its X/Y/Z scalar components."
}

func (cn NewNode[T]) Out(out *nodes.StructOutput[vector3.Vector[T]]) {
	out.Set(vector3.New(
		nodes.TryGetOutputValue(out, cn.X, 0),
		nodes.TryGetOutputValue(out, cn.Y, 0),
		nodes.TryGetOutputValue(out, cn.Z, 0),
	))
}

type ArrayFromComponentsNode[T vector.Number] struct {
	X nodes.Output[[]T] `description:"X values, one per output vector. Missing/shorter entries default to 0."`
	Y nodes.Output[[]T] `description:"Y values, one per output vector. Missing/shorter entries default to 0."`
	Z nodes.Output[[]T] `description:"Z values, one per output vector. Missing/shorter entries default to 0."`
}

func (snd ArrayFromComponentsNode[T]) Description() string {
	return "Builds an array of vector3s by zipping together three parallel arrays of X/Y/Z components. The output length is the longest of the three input arrays."
}

func (snd ArrayFromComponentsNode[T]) Out(out *nodes.StructOutput[[]vector3.Vector[T]]) {
	xArr := nodes.TryGetOutputValue(out, snd.X, nil)
	yArr := nodes.TryGetOutputValue(out, snd.Y, nil)
	zArr := nodes.TryGetOutputValue(out, snd.Z, nil)

	arr := make([]vector3.Vector[T], max(len(xArr), len(yArr), len(zArr)))
	for i := range arr {
		var x T
		var y T
		var z T

		if i < len(xArr) {
			x = xArr[i]
		}

		if i < len(yArr) {
			y = yArr[i]
		}

		if i < len(zArr) {
			z = zArr[i]
		}

		arr[i] = vector3.New(x, y, z)
	}

	out.Set(arr)
}

type ArrayFromNodesNode[T vector.Number] struct {
	In []nodes.Output[vector3.Vector[T]] `description:"The vectors to collect into an array, in order."`
}

func (node ArrayFromNodesNode[T]) Description() string {
	return "Collects individual vector3 values into one array, in order."
}

func (node ArrayFromNodesNode[T]) Out(out *nodes.StructOutput[[]vector3.Vector[T]]) {
	out.Set(nodes.GetOutputValues(out, node.In))
}
