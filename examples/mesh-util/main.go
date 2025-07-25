package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "mesh-util",
		Authors: []*cli.Author{
			{
				Name: "Eli Davis",
			},
		},
		Commands: []*cli.Command{
			scaleCommand(),
			statsCommand(),
			smoothCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
