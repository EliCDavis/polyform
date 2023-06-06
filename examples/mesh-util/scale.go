package main

import (
	"os"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
	"github.com/urfave/cli/v2"
)

func scaleCommand() *cli.Command {
	return &cli.Command{
		Name:  "scale",
		Usage: "scale a mesh by some vector",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "in",
				Aliases:  []string{"i"},
				Usage:    "object to scale",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "out",
				Aliases:  []string{"o"},
				Usage:    "name of file of scaled object",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "smooth-normals",
				Usage: "Whether or not to calculate smoothed normals for the mesh",
			},
			&cli.Float64Flag{
				Name:  "x",
				Usage: "value to scale mesh in x direction",
				Value: 1,
			},
			&cli.Float64Flag{
				Name:  "y",
				Usage: "value to scale mesh in y direction",
				Value: 1,
			},
			&cli.Float64Flag{
				Name:  "z",
				Usage: "value to scale mesh in z direction",
				Value: 1,
			},
		},
		Action: func(c *cli.Context) error {
			loadedMesh, err := readMesh(c.String("in"))
			if err != nil {
				return err
			}

			outFile, err := os.Create(c.String("out"))
			if err != nil {
				return err
			}

			scaledMesh := loadedMesh.
				Scale(
					vector3.Zero[float64](),
					vector3.New(c.Float64("x"), c.Float64("y"), c.Float64("z")),
				)

			if c.IsSet("smooth-normals") && c.Bool("smooth-normals") {
				scaledMesh = scaledMesh.Transform(
					meshops.SmoothNormalsTransformer{},
				)
			}

			return obj.WriteMesh(scaledMesh, "", outFile)
		},
	}
}
