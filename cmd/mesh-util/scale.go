package main

import (
	"os"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/vector"
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
			inFile, err := os.Open(c.String("in"))
			if err != nil {
				return err
			}

			readMesh, err := mesh.FromObj(inFile)
			if err != nil {
				return err
			}

			outFile, err := os.Create(c.String("out"))
			if err != nil {
				return err
			}

			scaledMesh := readMesh.
				Scale(
					vector.Vector3Zero(),
					vector.NewVector3(c.Float64("x"), c.Float64("y"), c.Float64("z")),
				)

			if c.IsSet("smooth-normals") && c.Bool("smooth-normals") {
				scaledMesh = scaledMesh.CalculateSmoothNormals()
			}

			return scaledMesh.WriteObj(outFile)
		},
	}
}
