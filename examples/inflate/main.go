package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/formats/pts"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
	"github.com/urfave/cli/v2"
)

func loadMesh(meshPath string) (*modeling.Mesh, error) {
	ext := strings.ToLower(path.Ext(meshPath))

	switch ext {
	case ".pts":
		f, err := os.Open(meshPath)
		if err != nil {
			return nil, err
		}
		return pts.ReadPointCloud(f)

	case ".ply":
		return ply.Load(meshPath)

	case ".obj":
		scene, err := obj.Load(meshPath)
		mesh := scene.ToMesh()
		return &mesh, err

	default:
		return nil, fmt.Errorf("unimplemented format to load: %s", ext)
	}
}

func main() {
	app := &cli.App{
		Name: "inflate",
		Authors: []*cli.Author{
			{
				Name: "Eli Davis",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "in",
				Aliases:  []string{"i"},
				Usage:    "object to inflate",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "out",
				Aliases:  []string{"o"},
				Usage:    "name of file of scaled object",
				Required: true,
			},
			&cli.Float64Flag{
				Name:    "threshold",
				Aliases: []string{"t"},
				Value:   .1,
			},
			&cli.Float64Flag{
				Name:    "radius",
				Aliases: []string{"r"},
				Value:   .1,
			},
			&cli.Float64Flag{
				Name:  "strength",
				Value: 10,
			},
			&cli.Float64Flag{
				Name:  "scale",
				Value: 12,
			},
			&cli.Float64Flag{
				Name:  "resolution",
				Value: 20,
			},
		},
		Action: func(c *cli.Context) error {
			loadedMesh, err := loadMesh(c.String("in"))
			if err != nil {
				return err
			}

			canvas := marching.NewMarchingCanvas(c.Float64("resolution"))

			startTime := time.Now()
			canvas.AddFieldParallel(marching.Mesh(
				loadedMesh.Transform(
					meshops.CenterAttribute3DTransformer{},
					meshops.ScaleAttribute3DTransformer{
						Amount: vector3.Fill(c.Float64("scale")),
					},
				),
				c.Float64("radius"),
				c.Float64("strength"),
			))
			log.Printf("Duration To add Field: %s\n", time.Since(startTime))

			// for i := 0; i < 10; i++ {
			// 	obj.Save(fmt.Sprintf("%d-%s", i, c.String("out")), canvas.March(float64(i)/10))
			// }

			return obj.SaveMesh(c.String("out"), canvas.MarchParallel(c.Float64("threshold")))
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
