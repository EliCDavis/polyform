package main

import (
	"encoding/json"
	"os"

	"github.com/urfave/cli/v2"
)

var HierarhcyToJsonCommand = &cli.Command{
	Name: "hierarchy-to-json",
	Flags: []cli.Flag{
		metadataFlag,
		hierarchyFlag,
		&cli.StringFlag{
			Name:  "out",
			Value: "hierarchy.json",
			Usage: "Name of JSON file to write hierarchy data too",
		},
	},
	Action: func(ctx *cli.Context) error {
		_, hierarchy, err := loadHierarchy(ctx)
		if err != nil {
			return err
		}

		data, err := json.Marshal(hierarchy)
		if err != nil {
			return err
		}

		f, err := os.Create(ctx.String("out"))
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.Write(data)

		return err
	},
}
