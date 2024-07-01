package main

import (
	"image"
	"image/draw"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/fogleman/gg"
)

type GridNode = nodes.StructNode[image.Image, GridNodeData]

type GridNodeData struct {
	HorizontalLines nodes.NodeOutput[int]
	VerticalLines   nodes.NodeOutput[int]
	Dimensions      nodes.NodeOutput[int]
	Color           nodes.NodeOutput[coloring.WebColor]
	LineColor       nodes.NodeOutput[coloring.WebColor]
	LineWidth       nodes.NodeOutput[float64]
}

func (gnd GridNodeData) Process() (image.Image, error) {
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

	return ctx.Image(), nil
}

func main() {
	lines := &parameter.Int{Name: "Lines", DefaultValue: 10}

	app := generator.App{
		Name:    "Grid Texture",
		Version: "1.0.0",
		Authors: []generator.Author{{Name: "Eli C Davis"}},
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"grid.png": artifact.NewImageNode(&GridNode{
				Data: GridNodeData{
					HorizontalLines: lines,
					VerticalLines:   lines,
					Dimensions: &parameter.Int{
						Name:         "Dimensions",
						DefaultValue: 512,
					},
					Color: &parameter.Color{
						Name:         "Color",
						DefaultValue: coloring.WebColor{R: 0x19, G: 0x1c, B: 0x1c, A: 255},
					},
					LineColor: &parameter.Color{
						Name:         "Line Color",
						DefaultValue: coloring.WebColor{R: 0x5c, G: 0xdb, B: 0xdb, A: 255},
					},
					LineWidth: &parameter.Float64{
						Name:         "Line Width",
						DefaultValue: 4,
					},
				},
			}),
		},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
