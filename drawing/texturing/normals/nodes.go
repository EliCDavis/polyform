package normals

import (
	"image"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[NewNode]](factory)

	refutil.RegisterType[nodes.Struct[FromImageNode]](factory)
	refutil.RegisterType[nodes.Struct[FromNormalMapNode]](factory)
	refutil.RegisterType[nodes.Struct[FromHeightMapNode]](factory)

	refutil.RegisterType[nodes.Struct[DrawSpheresNode]](factory)

	generator.RegisterTypes(factory)
}

// ============================================================================

type FromImageNode struct {
	In    nodes.Output[image.Image]
	Scale nodes.Output[float64]
}

func (n FromImageNode) Heightmap(out *nodes.StructOutput[HeightMap]) {
	img := nodes.TryGetOutputValue(out, n.In, nil)
	if img == nil {
		return
	}

	scale := nodes.TryGetOutputValue(out, n.Scale, 1.)
	out.Set(ImageToHeightmap(img, scale))
}

func (n FromImageNode) Normalmap(out *nodes.StructOutput[NormalMap]) {
	img := nodes.TryGetOutputValue(out, n.In, nil)
	if img == nil {
		return
	}

	scale := nodes.TryGetOutputValue(out, n.Scale, 1.)
	heightmap := ImageToHeightmap(img, scale)
	out.Set(FromHeightmap(heightmap, 1))
}

func (n FromImageNode) NormalMapImage(out *nodes.StructOutput[image.Image]) {
	img := nodes.TryGetOutputValue(out, n.In, nil)
	if img == nil {
		return
	}

	scale := nodes.TryGetOutputValue(out, n.Scale, 1.)
	heightmap := ImageToHeightmap(img, scale)
	normalMap := FromHeightmap(heightmap, 1)
	out.Set(image.Image(RasterizeNormalmap(normalMap)))
}

// ============================================================================

type FromHeightMapNode struct {
	In    nodes.Output[HeightMap]
	Scale nodes.Output[float64]
}

func (n FromHeightMapNode) NormalMapImage(out *nodes.StructOutput[image.Image]) {
	if n.In == nil {
		return
	}

	out.Set(RasterizeNormalmap(FromHeightmap(
		nodes.GetOutputValue(out, n.In),
		nodes.TryGetOutputValue(out, n.Scale, 1.),
	)))
}

func (n FromHeightMapNode) NormalMap(out *nodes.StructOutput[NormalMap]) {
	if n.In == nil {
		return
	}

	out.Set(FromHeightmap(
		nodes.GetOutputValue(out, n.In),
		nodes.TryGetOutputValue(out, n.Scale, 1.),
	))
}

// ============================================================================

type FromNormalMapNode struct {
	Normals nodes.Output[NormalMap]
}

func (n FromNormalMapNode) Image(out *nodes.StructOutput[image.Image]) {
	if n.Normals == nil {
		return
	}

	out.Set(RasterizeNormalmap(nodes.GetOutputValue(out, n.Normals)))
}
