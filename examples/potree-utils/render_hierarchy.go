package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/potree"
	"github.com/urfave/cli/v2"
)

var RenderHierarchyCommand = &cli.Command{
	Name:  "render",
	Usage: "Renders the hierarchy point count data to an image",
	Flags: []cli.Flag{
		metadataFlag,
		hierarchyFlag,

		//
		&cli.IntFlag{
			Name:  "row-count",
			Value: 100,
		},

		&cli.StringFlag{
			Name:  "out",
			Value: "image.png",
		},

		&cli.StringFlag{
			Name:  "type",
			Value: "point-count",
			Usage: "[point-count, file-position]",
		},
	},
	Action: func(ctx *cli.Context) error {
		_, hierarchy, err := loadHierarchy(ctx)
		if err != nil {
			return err
		}
		organizedNodes := make(map[int][]*potree.OctreeNode)

		minPoints := math.MaxInt
		maxPoints := 0
		numNodes := 0
		var largestFileStart uint64 = 0
		hierarchy.Walk(func(o *potree.OctreeNode) bool {
			if o.NumPoints == 0 {
				return true
			}
			if int(o.NumPoints) > maxPoints {
				maxPoints = int(o.NumPoints)
			}

			if _, ok := organizedNodes[o.Level]; !ok {
				organizedNodes[o.Level] = make([]*potree.OctreeNode, 0, 1)
			}
			organizedNodes[o.Level] = append(organizedNodes[o.Level], o)

			largestFileStart = max(largestFileStart, o.ByteOffset)
			minPoints = min(minPoints, int(o.NumPoints))
			numNodes++
			return true
		})
		pointRange := maxPoints - minPoints

		rows := ctx.Int("row-count")
		columns := numNodes / rows

		columns += (len(organizedNodes) - 1) * 2 // Add spacing

		fmt.Fprintf(ctx.App.Writer, "Column Count: %d", columns)

		img := image.NewRGBA(image.Rectangle{
			Min: image.Point{},
			Max: image.Point{
				X: columns,
				Y: rows,
			},
		})
		draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

		f, err := os.Create(ctx.String("out"))
		if err != nil {
			return err
		}
		defer f.Close()

		stack := coloring.NewColorStack(
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{0, 0, 255, 255}),
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{0, 255, 255, 255}),
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{0, 255, 0, 255}),
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{255, 255, 0, 255}),
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{255, 0, 0, 255}),
		)

		renderType := ctx.String("type")

		depth := 0
		count := 0
		offset := 0
		for i := 0; i < numNodes; i++ {
			y := (i % rows) - offset
			x := int(math.Floor(float64(i)/float64(rows))) + (depth * 2)

			var v float64

			switch renderType {
			case "point-count":
				v = float64(int(organizedNodes[depth][count].NumPoints)-minPoints) / float64(pointRange)

			case "file-position":
				v = float64(organizedNodes[depth][count].ByteOffset) / float64(largestFileStart)

			}

			img.Set(x, y, stack.LinearSample(v))

			count++

			if count == len(organizedNodes[depth]) {
				count = 0
				depth++
				// offset = rows - (y + rows)
			}
		}

		return png.Encode(f, img)
	},
}
