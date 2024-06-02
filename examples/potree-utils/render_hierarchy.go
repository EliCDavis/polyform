package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"github.com/EliCDavis/polyform/formats/potree"
	"github.com/urfave/cli/v2"
)

func GetPointCounts(depth int, node *potree.OctreeNode, out map[int][]int) {
	if _, ok := out[depth]; !ok {
		out[depth] = make([]int, 0, 1)
	}

	out[depth] = append(out[depth], int(node.NumPoints))

	for _, c := range node.Children {
		GetPointCounts(depth+1, c, out)
	}
}

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
	},
	Action: func(ctx *cli.Context) error {
		_, hierarchy, err := loadHierarchy(ctx)
		if err != nil {
			return err
		}
		counts := make(map[int][]int)
		GetPointCounts(0, hierarchy, counts)

		maxPoints := 0
		hierarchy.Walk(func(o *potree.OctreeNode) {
			if int(o.NumPoints) > maxPoints {
				maxPoints = int(o.NumPoints)
			}
		})

		rows := ctx.Int("row-count")
		numNodes := hierarchy.DescendentCount() + 1
		columns := numNodes / rows

		columns += (len(counts) - 1) * 2 // Add spacing

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

		depth := 0
		count := 0
		offset := 0
		for i := 0; i < numNodes; i++ {
			y := (i % rows) - offset
			x := int(math.Floor(float64(i)/float64(rows))) + (depth * 2)

			v := byte((float64(counts[depth][count]) / float64(maxPoints)) * 255)
			img.Set(x, y, color.RGBA{
				R: v,
				A: 255,
			})

			count++

			if count == len(counts[depth]) {
				count = 0
				depth++
				// offset = rows - (y + rows)
			}
		}

		return png.Encode(f, img)
	},
}
