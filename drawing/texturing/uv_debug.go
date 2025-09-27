package texturing

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/fogleman/gg"
)

type DebugUV struct {
	ImageResolution int
	BoardResolution int

	PositiveCheckerColor color.Color
	NegativeCheckerColor color.Color

	XColorScale color.Color
	YColorScale color.Color
}

func (duvt DebugUV) Image() image.Image {
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
						coloring.ScaleRGB(duvt.PositiveCheckerColor, 1-xPercent),
						coloring.ScaleRGB(duvt.XColorScale, xPercent),
						coloring.ScaleRGB(duvt.YColorScale, yPercent),
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

type DebugUVNode struct {
	ImageResolution      nodes.Output[int]
	BoardResolution      nodes.Output[int]
	PositiveCheckerColor nodes.Output[coloring.Color]
	NegativeCheckerColor nodes.Output[coloring.Color]
	XColorScale          nodes.Output[coloring.Color]
	YColorScale          nodes.Output[coloring.Color]
}

func (n DebugUVNode) Result(out *nodes.StructOutput[image.Image]) {
	out.Set(DebugUV{
		ImageResolution:      nodes.TryGetOutputValue(out, n.ImageResolution, 256),
		BoardResolution:      nodes.TryGetOutputValue(out, n.BoardResolution, 10),
		PositiveCheckerColor: color.Color(nodes.TryGetOutputValue(out, n.PositiveCheckerColor, coloring.Color{R: 0, G: 1, B: 0, A: 1})),
		NegativeCheckerColor: color.Color(nodes.TryGetOutputValue(out, n.PositiveCheckerColor, coloring.Color{R: 0, G: 0, B: 0, A: 1})),
		XColorScale:          color.Color(nodes.TryGetOutputValue(out, n.PositiveCheckerColor, coloring.Color{R: 1, G: 0, B: 0, A: 1})),
		YColorScale:          color.Color(nodes.TryGetOutputValue(out, n.PositiveCheckerColor, coloring.Color{R: 0, G: 1, B: 1, A: 1})),
	}.Image())
}
