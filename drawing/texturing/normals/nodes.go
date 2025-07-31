package normals

import (
	"image"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[FromHeightmapNode]](factory)

	generator.RegisterTypes(factory)
}

type FromHeightmapNode struct {
	In    nodes.Output[image.Image]
	Scale nodes.Output[float64]
}

func (n FromHeightmapNode) Out() nodes.StructOutput[image.Image] {
	out := nodes.NewStructOutput[image.Image](nil)
	img := nodes.TryGetOutputValue(&out, n.In, nil)
	if img == nil {
		return out
	}

	scale := nodes.TryGetOutputValue(&out, n.Scale, 1.)
	out.Set(image.Image(FromHeightmap(img, scale)))
	return out
}
