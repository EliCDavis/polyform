package experimental

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/nodes"
)

type SeamlessPerlinNode = nodes.Struct[SeamlessPerlinNodeData]

type SeamlessPerlinNodeData struct {
	Dimensions nodes.Output[int]
	Positive   nodes.Output[coloring.WebColor]
	Negative   nodes.Output[coloring.WebColor]
}

func (an SeamlessPerlinNodeData) Out() nodes.StructOutput[image.Image] {
	dim := nodes.TryGetOutputValue(an.Dimensions, 256)
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	// normals.Fill(img)

	n := noise.NewTilingNoise(dim, 1/64., 5)

	nR, nG, nB, _ := nodes.TryGetOutputValue(an.Negative, coloring.Black()).RGBA()
	pR, pG, pB, _ := nodes.TryGetOutputValue(an.Positive, coloring.White()).RGBA()

	rRange := float64(pR>>8) - float64(nR>>8)
	gRange := float64(pG>>8) - float64(nG>>8)
	bRange := float64(pB>>8) - float64(nB>>8)

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
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
	return nodes.NewStructOutput[image.Image](img)
}
