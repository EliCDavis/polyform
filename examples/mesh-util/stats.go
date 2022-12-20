package main

import (
	"fmt"

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
			loadedMesh, err := readMesh(c.String("in"))
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(c.App.Writer, "tris: %d", loadedMesh.TriCount())

			return err
		},
	}
}
