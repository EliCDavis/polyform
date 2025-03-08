package experimental

import (
	"image"
	"image/draw"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/fogleman/gg"
)

type BrushedMetalNode = nodes.Struct[BrushedMetalNodeNodeData]

type BrushedMetalNodeNodeData struct {
	Dimensions nodes.Output[int]
	BaseColor  nodes.Output[coloring.WebColor]
	BrushColor nodes.Output[coloring.WebColor]
	BrushSize  nodes.Output[float64]
	Count      nodes.Output[int]
}

// func (gnd BrushedMetalNodeNodeData) Out() nodes.StructOutput[image.Image] {
// func (gnd BrushedMetalNodeNodeData) Out() nodes.StructOutput[image.Image] {

func (gnd BrushedMetalNodeNodeData) Out() nodes.StructOutput[image.Image] {
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
