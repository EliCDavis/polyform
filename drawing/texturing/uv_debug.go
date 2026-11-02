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
	ImageResolution      nodes.Output[int]            `default:"256"`
	BoardResolution      nodes.Output[int]            `default:"10"`
	PositiveCheckerColor nodes.Output[coloring.Color] `default:"#00ff00"`
	NegativeCheckerColor nodes.Output[coloring.Color] `default:"#000000"`
	XColorScale          nodes.Output[coloring.Color] `default:"#ff0000"`
	YColorScale          nodes.Output[coloring.Color] `default:"#00ffff"`
}

func (n DebugUVNode) Description() string {
	return "Checkerboard test texture for inspecting how UVs are laid out on a mesh."
}

func (n DebugUVNode) Result(out *nodes.StructOutput[image.Image]) {
	out.Set(DebugUV{
		ImageResolution:      nodes.GetOutputValue(out, n.ImageResolution),
		BoardResolution:      nodes.GetOutputValue(out, n.BoardResolution),
		PositiveCheckerColor: color.Color(nodes.GetOutputValue(out, n.PositiveCheckerColor)),
		NegativeCheckerColor: color.Color(nodes.GetOutputValue(out, n.NegativeCheckerColor)),
		XColorScale:          color.Color(nodes.GetOutputValue(out, n.XColorScale)),
		YColorScale:          color.Color(nodes.GetOutputValue(out, n.YColorScale)),
	}.Image())
}
