package texturing

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func DotProduct(texture Texture[vector3.Float64], v vector3.Float64) Texture[float64] {
	result := Empty[float64](texture.width, texture.height)

	for y := range texture.height {
		for x := range texture.width {
			result.Set(x, y, v.Dot(texture.Get(x, y).Normalized()))
		}
	}

	return result
}

type DotProductNode struct {
	Texture nodes.Output[Texture[vector3.Float64]]
	Vector  nodes.Output[vector3.Float64]
}

func (n DotProductNode) DotProduct(out *nodes.StructOutput[Texture[float64]]) {
	if n.Texture == nil {
		return
	}

	out.Set(DotProduct(
		nodes.GetOutputValue(out, n.Texture),
		nodes.TryGetOutputValue(out, n.Vector, vector3.Up[float64]()),
	))
}
