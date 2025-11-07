package texturing

import (
	"image"

	"github.com/EliCDavis/polyform/drawing/coloring"
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
