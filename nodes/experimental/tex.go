package experimental

import (
	"image"
	"image/draw"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/fogleman/gg"
)

type GridNode = nodes.Struct[GridNodeData]

type GridNodeData struct {
	HorizontalLines nodes.Output[int]
	VerticalLines   nodes.Output[int]
	Dimensions      nodes.Output[int]
	Color           nodes.Output[coloring.WebColor]
	LineColor       nodes.Output[coloring.WebColor]
	LineWidth       nodes.Output[float64]
}

func (gnd GridNodeData) Out() nodes.StructOutput[image.Image] {
	dimensions := gnd.Dimensions.Value()
	img := image.NewRGBA(image.Rect(0, 0, dimensions, dimensions))

	draw.Draw(img, img.Bounds(), &image.Uniform{gnd.Color.Value()}, image.Point{}, draw.Src)

	ctx := gg.NewContextForImage(img)
	ctx.SetLineWidth(gnd.LineWidth.Value())
	ctx.SetColor(gnd.LineColor.Value())

	horizontalLines := gnd.HorizontalLines.Value()
	horizontalSpacing := float64(dimensions) / float64(horizontalLines)
	for i := 0; i < horizontalLines; i++ {
		y := (horizontalSpacing * float64(i)) + (horizontalSpacing / 2)
		ctx.DrawLine(0, y, float64(dimensions), y)
	}

	verticalLines := gnd.VerticalLines.Value()
	verticalSpacing := float64(dimensions) / float64(verticalLines)
	for i := 0; i < verticalLines; i++ {
		x := (verticalSpacing * float64(i)) + (verticalSpacing / 2)
		ctx.DrawLine(x, 0, x, float64(dimensions))
	}
	ctx.Stroke()

	return nodes.NewStructOutput(ctx.Image())
}

type BrushedMetalNode = nodes.Struct[BrushedMetalNodeData]

type BrushedMetalNodeData struct {
	Dimensions nodes.Output[int]
	BaseColor  nodes.Output[coloring.WebColor]
	BrushColor nodes.Output[coloring.WebColor]
	BrushSize  nodes.Output[float64]
	Count      nodes.Output[int]
}

// func (gnd BrushedMetalNodeNodeData) Out() nodes.StructOutput[image.Image] {
// func (gnd BrushedMetalNodeNodeData) Out() nodes.StructOutput[image.Image] {

func (gnd BrushedMetalNodeData) Out() nodes.StructOutput[image.Image] {
	dimensions := 512
	if gnd.Dimensions != nil {
		dimensions = gnd.Dimensions.Value()
	}
	img := image.NewRGBA(image.Rect(0, 0, dimensions, dimensions))

	baseColor := coloring.Grey(200)
	if gnd.BaseColor != nil {
		baseColor = gnd.BaseColor.Value()
	}
	draw.Draw(img, img.Bounds(), &image.Uniform{baseColor}, image.Point{}, draw.Src)

	ctx := gg.NewContextForImage(img)

	ctx.SetLineWidth(nodes.TryGetOutputValue(gnd.BrushSize, 1.))

	bruchColor := coloring.Grey(150)
	if gnd.BrushColor != nil {
		bruchColor = gnd.BrushColor.Value()
	}
	ctx.SetColor(bruchColor)

	horizontalLines := nodes.TryGetOutputValue(gnd.Count, 10)
	horizontalSpacing := float64(dimensions) / float64(horizontalLines)
	for i := 0; i < horizontalLines; i++ {
		y := (horizontalSpacing * float64(i)) + (horizontalSpacing / 2)
		ctx.DrawLine(0, y, float64(dimensions), y)
	}

	ctx.Stroke()

	return nodes.NewStructOutput(ctx.Image())
}
