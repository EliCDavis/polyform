package main

import (
	"fmt"
	"os"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/urfave/cli/v2"
)

var inFilePath string

func openPlyFile() (*os.File, error) {
	return os.Open(inFilePath)
}

func main() {

	cmd := cli.App{
		Name:    "PLY Utils",
		Usage:   "Different utilities for inspecting ply files",
		Version: "0.0.1",
		Authors: []*cli.Author{
			{Name: "Eli Davis"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "in",
				Required:    true,
				Aliases:     []string{"i", "f", "file"},
				Destination: &inFilePath,
			},
		},
		Commands: []*cli.Command{
			{
				Name: "header",
				Action: func(ctx *cli.Context) error {
					f, err := openPlyFile()
					if err != nil {
						return err
					}

					header, err := ply.ReadHeader(f)
					if err != nil {
						return err
					}

					fmt.Fprintf(ctx.App.Writer, "Format: %s\n", header.Format.String())
					fmt.Fprintln(ctx.App.Writer, "Elements")
					for _, ele := range header.Elements {
						fmt.Fprintf(ctx.App.Writer, "\t%s: %d entries\n", ele.Name, ele.Count)
						for _, prop := range ele.Properties {
							if scalar, ok := prop.(ply.ScalarProperty); ok {
								fmt.Fprintf(ctx.App.Writer, "\t\t%s (%s)\n", prop.Name(), scalar.Type)
							} else if arr, ok := prop.(ply.ListProperty); ok {
								fmt.Fprintf(ctx.App.Writer, "\t\t%s (%s, %s)\n", prop.Name(), arr.CountType, arr.ListType)
							}
						}
					}

					return nil
				},
			},
		},
	}

	if err := cmd.Run(os.Args); err != nil {
		panic(err)
	}
}
