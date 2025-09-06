package normals

import (
	"image"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[HeightmapFromImageNode]](factory)
	refutil.RegisterType[nodes.Struct[ToNormalMapNode]](factory)
	refutil.RegisterType[nodes.Struct[TextureFromHeightmapNode]](factory)

	generator.RegisterTypes(factory)
}

// ============================================================================

type HeightmapFromImageNode struct {
	In    nodes.Output[image.Image]
	Scale nodes.Output[float64]
}

func (n HeightmapFromImageNode) Out(out *nodes.StructOutput[image.Image]) {
	img := nodes.TryGetOutputValue(out, n.In, nil)
	if img == nil {
		return
	}

	scale := nodes.TryGetOutputValue(out, n.Scale, 1.)
	out.Set(image.Image(ImageFromHeightmap(img, scale)))
}

// ============================================================================

type TextureFromHeightmapNode struct {
	In    nodes.Output[texturing.Texture[float64]]
	Scale nodes.Output[float64]
}

func (n TextureFromHeightmapNode) Image(out *nodes.StructOutput[image.Image]) {
	if n.In == nil {
		return
	}

	out.Set(ToNormalmap(FromHeightmap(
		nodes.GetOutputValue(out, n.In),
		nodes.TryGetOutputValue(out, n.Scale, 1.),
	)))
}

func (n TextureFromHeightmapNode) Texture(out *nodes.StructOutput[texturing.Texture[vector3.Float64]]) {
	if n.In == nil {
		return
	}

	out.Set(FromHeightmap(
		nodes.GetOutputValue(out, n.In),
		nodes.TryGetOutputValue(out, n.Scale, 1.),
	))
}

// ============================================================================

type ToNormalMapNode struct {
	Normals nodes.Output[texturing.Texture[vector3.Float64]]
}

func (n ToNormalMapNode) NormalMap(out *nodes.StructOutput[image.Image]) {
	if n.Normals == nil {
		return
	}

	out.Set(ToNormalmap(nodes.GetOutputValue(out, n.Normals)))
}
