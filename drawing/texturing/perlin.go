package texturing

import (
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/nodes"
)

type SeamlessPerlinNode struct {
	Dimensions nodes.Output[int]
	Positive   nodes.Output[float64]
	Negative   nodes.Output[float64]
}

func (an SeamlessPerlinNode) Out(out *nodes.StructOutput[Texture[float64]]) {
	dim := nodes.TryGetOutputValue(out, an.Dimensions, 256)
	n := noise.NewTilingNoise(dim, 1/64., 5)

	tex := NewTexture[float64](dim, dim)
	negative := nodes.TryGetOutputValue(out, an.Negative, 0)
	positive := nodes.TryGetOutputValue(out, an.Positive, 1)
	valueRange := positive - negative

	for y := range dim {
		for x := range dim {
			p := (n.Noise(x, y) * 0.5) + 0.5

			tex.Set(x, y, negative+(valueRange*p))
		}
	}
	out.Set(tex)
}
