package normals

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func FillImage(img draw.Image) {
	blue := color.RGBA{128, 128, 255, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{blue}, image.Point{}, draw.Src)
}

type NewNode struct {
	Dimensions nodes.Output[vector2.Int]
}

func (n NewNode) NormalMap(out *nodes.StructOutput[NormalMap]) {
	dim := nodes.TryGetOutputValue(out, n.Dimensions, vector2.New(256, 256))
	if dim.MinComponent() <= 0 {
		out.CaptureError(texturing.InvalidDimension(dim))
		return
	}

	tex := texturing.Empty[vector3.Float64](dim.X(), dim.Y())
	tex.Fill(vector3.New(0, 0, 1.))
	out.Set(tex)
}
