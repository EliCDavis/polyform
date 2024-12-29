package images

import "github.com/urfave/cli/v2"

const (
	imagesPathFlagName = "images"
)

var Command = &cli.Command{
	Name:  "images",
	Usage: "functionality pertaining to images data",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    imagesPathFlagName,
			Aliases: []string{"i"},
			Usage:   "path to images.bin file",
			Value:   "images.bin",
		},
	},
	Subcommands: []*cli.Command{
		toPlyCommand,
		infoCommand,
	},
}
