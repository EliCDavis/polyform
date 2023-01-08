package main

import (
	"log"
	"os"
	"time"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/vector"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "inflate",
		Authors: []*cli.Author{
			{
				Name:  "Eli Davis",
				Email: "eli@recolude.com",
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
				Value:   .5,
			},
			&cli.Float64Flag{
				Name:    "radius",
				Aliases: []string{"r"},
				Value:   .1,
			},
			&cli.Float64Flag{
				Name:    "strength",
				Aliases: []string{"s"},
				Value:   10,
			},
		},
		Action: func(c *cli.Context) error {
			loadedMesh, err := obj.Load(c.String("in"))
			if err != nil {
				return err
			}

			cubesPerUnit := 10.

			canvas := marching.NewMarchingCanvas(cubesPerUnit)

			startTime := time.Now()
			canvas.AddFieldParallel(marching.Mesh(
				loadedMesh.
					CenterFloat3Attribute(modeling.PositionAttribute).
					Scale(vector.Vector3Zero(), vector.Vector3(vector.NewVector3(12, 12, 12))),
				c.Float64("radius"),
				c.Float64("strength"),
			))
			log.Printf("Duration To add Field: %s\n", time.Now().Sub(startTime))

			// for i := 0; i < 10; i++ {
			// 	obj.Save(fmt.Sprintf("%d-%s", i, c.String("out")), canvas.March(float64(i)/10))
			// }

			return obj.Save(c.String("out"), canvas.MarchParallel(c.Float64("threshold")))
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}