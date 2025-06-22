package main

import (
	"os"

	"github.com/EliCDavis/polyform/examples/ply-utils/properties"
	"github.com/urfave/cli/v2"
)

func main() {

	cmd := cli.App{
		Name:    "ply-utils",
		Usage:   "Different utilities for inspecting and processing ply files",
		Version: "0.0.1",
		Authors: []*cli.Author{
			{Name: "Eli Davis"},
		},
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			HeaderCommand,
			ToGLTFCommand,
			properties.PropertiesCommand,
			FromCSVCommand,
		},
	}

	if err := cmd.Run(os.Args); err != nil {
		panic(err)
	}
}
