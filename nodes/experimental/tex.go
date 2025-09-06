package experimental

import (
	"image"
	"image/draw"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/fogleman/gg"
)

type GridNode struct {
	HorizontalLines nodes.Output[int]
	VerticalLines   nodes.Output[int]
	Dimensions      nodes.Output[int]
	Color           nodes.Output[coloring.Color]
	LineColor       nodes.Output[coloring.Color]
	LineWidth       nodes.Output[float64]
}

func (gnd GridNode) Out(out *nodes.StructOutput[image.Image]) {
	dimensions := nodes.TryGetOutputValue(out, gnd.Dimensions, 256)
	img := image.NewRGBA(image.Rect(0, 0, dimensions, dimensions))

	draw.Draw(img, img.Bounds(), &image.Uniform{nodes.TryGetOutputValue(out, gnd.Color, coloring.Black())}, image.Point{}, draw.Src)

	ctx := gg.NewContextForImage(img)
	ctx.SetLineWidth(nodes.TryGetOutputValue(out, gnd.LineWidth, 1.))
	ctx.SetColor(nodes.TryGetOutputValue(out, gnd.LineColor, coloring.White()))

	horizontalLines := nodes.TryGetOutputValue(out, gnd.HorizontalLines, 10)
	horizontalSpacing := float64(dimensions) / float64(horizontalLines)
	for i := 0; i < horizontalLines; i++ {
		y := (horizontalSpacing * float64(i)) + (horizontalSpacing / 2)
		ctx.DrawLine(0, y, float64(dimensions), y)
	}

	verticalLines := nodes.TryGetOutputValue(out, gnd.VerticalLines, 10)
	verticalSpacing := float64(dimensions) / float64(verticalLines)
	for i := 0; i < verticalLines; i++ {
		x := (verticalSpacing * float64(i)) + (verticalSpacing / 2)
		ctx.DrawLine(x, 0, x, float64(dimensions))
	}
	ctx.Stroke()

	out.Set(ctx.Image())
}

type BrushedMetalNode struct {
	Dimensions nodes.Output[int]
	BaseColor  nodes.Output[coloring.Color]
	BrushColor nodes.Output[coloring.Color]
	BrushSize  nodes.Output[float64]
	Count      nodes.Output[int]
}

// func (gnd BrushedMetalNodeNode) Out(out *nodes.StructOutput[image.Image]) {
// func (gnd BrushedMetalNodeNode) Out(out *nodes.StructOutput[image.Image]) {

func (gnd BrushedMetalNode) Out(out *nodes.StructOutput[image.Image]) {
	dimensions := nodes.TryGetOutputValue(out, gnd.Dimensions, 512)
	img := image.NewRGBA(image.Rect(0, 0, dimensions, dimensions))

	baseColor := nodes.TryGetOutputValue(out, gnd.BaseColor, coloring.Grey(200))
	draw.Draw(img, img.Bounds(), &image.Uniform{baseColor}, image.Point{}, draw.Src)

	ctx := gg.NewContextForImage(img)

	ctx.SetLineWidth(nodes.TryGetOutputValue(out, gnd.BrushSize, 1.))

	brushColor := nodes.TryGetOutputValue(out, gnd.BrushColor, coloring.Grey(150))
	ctx.SetColor(brushColor)

	horizontalLines := nodes.TryGetOutputValue(out, gnd.Count, 10)
	horizontalSpacing := float64(dimensions) / float64(horizontalLines)
	for i := range horizontalLines {
		y := (horizontalSpacing * float64(i)) + (horizontalSpacing / 2)
		ctx.DrawLine(0, y, float64(dimensions), y)
	}

	ctx.Stroke()

	out.Set(ctx.Image())
}
