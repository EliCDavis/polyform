package experimental

import (
	"image"
	"image/draw"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/fogleman/gg"
)

type BrushedMetalNode struct {
	Dimensions nodes.Output[int]            `default:"512"`
	BaseColor  nodes.Output[coloring.Color] `default:"#c8c8c8"`
	BrushColor nodes.Output[coloring.Color] `default:"#969696"`
	BrushSize  nodes.Output[float64]        `default:"1"`
	Count      nodes.Output[int]            `default:"10"`
}

// func (gnd BrushedMetalNodeNode) Out(out *nodes.StructOutput[image.Image]) {
// func (gnd BrushedMetalNodeNode) Out(out *nodes.StructOutput[image.Image]) {

func (gnd BrushedMetalNode) Description() string {
	return "Generates a brushed metal image."
}

func (gnd BrushedMetalNode) Out(out *nodes.StructOutput[image.Image]) {
	dimensions := nodes.GetOutputValue(out, gnd.Dimensions)
	img := image.NewRGBA(image.Rect(0, 0, dimensions, dimensions))

	baseColor := nodes.GetOutputValue(out, gnd.BaseColor)
	draw.Draw(img, img.Bounds(), &image.Uniform{baseColor}, image.Point{}, draw.Src)

	ctx := gg.NewContextForImage(img)

	ctx.SetLineWidth(nodes.GetOutputValue(out, gnd.BrushSize))

	brushColor := nodes.GetOutputValue(out, gnd.BrushColor)
	ctx.SetColor(brushColor)

	horizontalLines := nodes.GetOutputValue(out, gnd.Count)
	horizontalSpacing := float64(dimensions) / float64(horizontalLines)
	for i := range horizontalLines {
		y := (horizontalSpacing * float64(i)) + (horizontalSpacing / 2)
		ctx.DrawLine(0, y, float64(dimensions), y)
	}

	ctx.Stroke()

	out.Set(ctx.Image())
}
