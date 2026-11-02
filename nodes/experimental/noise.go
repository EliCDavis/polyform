package experimental

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/nodes"
)

type SeamlessPerlinNode struct {
	Dimensions nodes.Output[int]            `default:"256"`
	Positive   nodes.Output[coloring.Color] `default:"#ffffff"`
	Negative   nodes.Output[coloring.Color] `default:"#000000"`
}

func (an SeamlessPerlinNode) Out(out *nodes.StructOutput[image.Image]) {
	dim := nodes.GetOutputValue(out, an.Dimensions)
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	// normals.Fill(img)

	n := noise.NewTilingNoise(dim, 1/64., 5)

	nR, nG, nB, _ := nodes.GetOutputValue(out, an.Negative).RGBA()
	pR, pG, pB, _ := nodes.GetOutputValue(out, an.Positive).RGBA()

	rRange := float64(pR>>8) - float64(nR>>8)
	gRange := float64(pG>>8) - float64(nG>>8)
	bRange := float64(pB>>8) - float64(nB>>8)

	for x := range dim {
		for y := range dim {
			val := n.Noise(x, y)
			p := (val * 0.5) + 0.5

			r := uint32(float64(nR) + (rRange * p))
			g := uint32(float64(nG) + (gRange * p))
			b := uint32(float64(nB) + (bRange * p))

			img.Set(x, y, color.RGBA{
				R: byte(r), // byte(len * 255),
				G: byte(g),
				B: byte(b),
				A: 255,
			})
		}
	}
	out.Set(img)
}
