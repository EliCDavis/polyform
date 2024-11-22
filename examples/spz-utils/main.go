package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

var inFilePath string

func openSPZFile() (*os.File, error) {
	return os.Open(inFilePath)
}

func main() {

	cmd := cli.App{
		Name:    "spz-utils",
		Usage:   "Different utilities for inspecting and processing spz files",
		Version: "1.0.0",
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
			ToPlyCommand,
		},
	}

	if err := cmd.Run(os.Args); err != nil {
		panic(err)
	}
}
