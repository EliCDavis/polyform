package texturing

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/fogleman/gg"
)

type DebugUVTexture struct {
	ImageResolution int
	BoardResolution int

	PositiveCheckerColor color.Color
	NegativeCheckerColor color.Color

	XColorScale color.Color
	YColorScale color.Color
}

func (duvt DebugUVTexture) Image() image.Image {
	img := gg.NewContext(duvt.ImageResolution, duvt.ImageResolution)

	checkerSize := float64(duvt.ImageResolution) / float64(duvt.BoardResolution)
	for x := 0; x < duvt.BoardResolution; x++ {
		xShift := 0
		if x%2 == 0 {
			xShift = 1
		}
		xPercent := float64(x) / float64(duvt.BoardResolution)
		for y := 0; y < duvt.BoardResolution; y++ {
			yPercent := float64(y) / float64(duvt.BoardResolution)
			c := duvt.NegativeCheckerColor
			if (y+xShift)%2 == 0 {
				c = duvt.PositiveCheckerColor

				if duvt.XColorScale != nil && duvt.YColorScale != nil {
					c = coloring.AddRGB(
						coloring.MultiplyRGBByConstant(duvt.PositiveCheckerColor, 1-xPercent),
						coloring.MultiplyRGBByConstant(duvt.XColorScale, xPercent),
						coloring.MultiplyRGBByConstant(duvt.YColorScale, yPercent),
					)
				}
			}

			img.SetColor(c)
			img.DrawRectangle(
				float64(x)*checkerSize,
				float64(y)*checkerSize,
				checkerSize,
				checkerSize,
			)
			img.Fill()
		}
	}

	return img.Image()
}
