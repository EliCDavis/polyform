package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.App{
		Name:    "PolyWASM",
		Version: "0.0.1",
		Authors: []*cli.Author{
			{
				Name: "Eli C Davis",
			},
		},
		Commands: []*cli.Command{
			buildCommand(),
			serverCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
