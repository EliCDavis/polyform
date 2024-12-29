package images

import (
	"github.com/EliCDavis/polyform/formats/colmap"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/urfave/cli/v2"
)

var toPlyCommand = &cli.Command{
	Name:  "to-ply",
	Usage: "Convert a spart reconstruction's images.bin file into a ply file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "out",
			Usage: "path to write ply file",
			Value: "images.ply",
		},
	},
	Action: func(ctx *cli.Context) error {
		pointcloud, err := colmap.LoadImageData(ctx.String(imagesPathFlagName))
		if err != nil {
			return err
		}

		return ply.Save(ctx.String("out"), pointcloud, ply.BinaryLittleEndian)
	},
}
