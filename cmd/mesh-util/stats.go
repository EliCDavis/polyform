package main

import (
	"fmt"
	"os"

	"github.com/EliCDavis/mesh"
	"github.com/urfave/cli/v2"
)

func statsCommand() *cli.Command {
	return &cli.Command{
		Name:  "stats",
		Usage: "Get a summary of the mesh data",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "in",
				Aliases:  []string{"i"},
				Usage:    "object to scale",
				Required: true,
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

			_, err = fmt.Fprintf(c.App.Writer, "tris: %d", readMesh.TriCount())

			return err
		},
	}
}
