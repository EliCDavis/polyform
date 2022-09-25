package main

import (
	"os"

	"github.com/EliCDavis/mesh"
	"github.com/urfave/cli/v2"
)

func smoothCommand() *cli.Command {
	return &cli.Command{
		Name:  "smooth",
		Usage: "apply laplacian smoothing to a mesh",
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
			&cli.IntFlag{
				Name:    "iterations",
				Aliases: []string{"it"},
				Usage:   "Number of times to run the smoothing",
				Value:   5,
			},
			&cli.Float64Flag{
				Name:    "smoothing-weight",
				Aliases: []string{"sw"},
				Usage:   "Weight to apply to each smoothing iteration",
				Value:   0.5,
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

			return readMesh.
				SmoothLaplacian(c.Int("iterations"), c.Float64("smoothing-weight")).
				CalculateSmoothNormals().
				WriteObj(outFile)
		},
	}
}
