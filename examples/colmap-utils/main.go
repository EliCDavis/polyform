package main

import (
	"os"

	"github.com/EliCDavis/polyform/examples/colmap-utils/images"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.App{
		Name:        "colmap-utils",
		Description: "Utils around dealing with colmap data",
		Commands: []*cli.Command{
			PointsToPlyCommand,
			CamerasCommand,
			images.Command,
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
