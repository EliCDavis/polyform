package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func main() {

	cmd := cli.App{
		Name:    "Potree Utils",
		Version: "0.0.1",
		Authors: []*cli.Author{
			{Name: "Eli Davis"},
		},
		Description: "Different utilities for inspecting potree files",
		Commands: []*cli.Command{
			RenderHierarchyCommand,
			ExtractPointcloudCommand,
			HierarhcyToJsonCommand,
		},
	}

	if err := cmd.Run(os.Args); err != nil {
		panic(err)
	}
}
