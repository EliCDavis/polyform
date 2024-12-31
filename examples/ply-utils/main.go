package main

import (
	"os"

	"github.com/EliCDavis/polyform/examples/ply-utils/properties"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/urfave/cli/v2"
)

var inFilePath string

func openPlyFile() (*os.File, error) {
	return os.Open(inFilePath)
}

func getPlyFile() (*modeling.Mesh, error) {
	return ply.Load(inFilePath)
}

func main() {

	cmd := cli.App{
		Name:    "ply-utils",
		Usage:   "Different utilities for inspecting and processing ply files",
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
			HeaderCommand,
			ToGLTFCommand,
			properties.PropertiesCommand,
		},
	}

	if err := cmd.Run(os.Args); err != nil {
		panic(err)
	}
}
