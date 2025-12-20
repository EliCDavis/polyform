package texturing

import (
	"image"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
)

func FromImage(img image.Image) Texture[coloring.Color] {
	bounds := img.Bounds()
	tex := Empty[coloring.Color](bounds.Dx(), bounds.Dy())

	for y := range tex.Height() {
		for x := range tex.Width() {
			r, g, b, a := img.At(x, y).RGBA()
			tex.Set(x, y, coloring.Color{
				R: float64(r>>8) / 255.,
				G: float64(g>>8) / 255.,
				B: float64(b>>8) / 255.,
				A: float64(a>>8) / 255.,
			})
		}
	}

	return tex
}

type FromImageNode struct {
	Image nodes.Output[image.Image]
}

func (n FromImageNode) Texture(out *nodes.StructOutput[Texture[coloring.Color]]) {
	if n.Image == nil {
		return
	}

	img := nodes.GetOutputValue(out, n.Image)
	if img == nil {
		out.CaptureError(nodes.NilInputError{Input: n.Image})
		return
	}

	out.Set(FromImage(img))
}
