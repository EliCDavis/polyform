package texturing

import (
	"image"
	"image/color"

	"github.com/EliCDavis/vector/vector3"
)

func averageColorComponents(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	r8 := r >> 8
	g8 := g >> 8
	b8 := b >> 8

	return (float64(r8+g8+b8) / 3.) / 255.
}

// https://stackoverflow.com/questions/5281261/generating-a-normal-map-from-a-height-map
func ToNormal(src image.Image) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	scale := 1.
	Convolve(src, func(x, y int, vals []color.Color) {
		// float s[9] contains above samples
		n := vector3.New(0, 0, scale)
		s0 := averageColorComponents(vals[0])
		s1 := averageColorComponents(vals[1])
		s2 := averageColorComponents(vals[2])
		s3 := averageColorComponents(vals[3])
		s5 := averageColorComponents(vals[5])
		s6 := averageColorComponents(vals[6])
		s7 := averageColorComponents(vals[7])
		s8 := averageColorComponents(vals[8])

		n = n.SetX(scale * -((s2 - s0) + (2 * (s5 - s3)) + (s8 - s6)))
		n = n.SetY(scale * -((s6 - s0) + (2 * (s7 - s1)) + (s8 - s2)))
		n = n.Normalized()

		dst.Set(x, y, color.RGBA{
			R: uint8((0.5 + (n.X() / 2.)) * 255),
			G: uint8((0.5 + (n.Y() / 2.)) * 255),
			B: uint8((0.5 + (n.Z() / 2.)) * 255),
			A: 255,
		})
	})
	return dst
}
