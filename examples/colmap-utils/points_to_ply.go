package main

import (
	"github.com/EliCDavis/polyform/formats/colmap"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/urfave/cli/v2"
)

var PointsToPlyCommand = &cli.Command{
	Name:  "points-to-ply",
	Usage: "Convert a sparse reconstruction's points.bin file into a ply file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "points",
			Usage: "path to points.bin file to convert",
			Value: "points3D.bin",
		},
		&cli.StringFlag{
			Name:  "out",
			Usage: "path to write ply file",
			Value: "points3D.ply",
		},
	},
	Action: func(ctx *cli.Context) error {
		pointcloud, err := colmap.LoadSparsePointData(ctx.String("points"))
		if err != nil {
			return err
		}

		return ply.Save(ctx.String("out"), pointcloud, ply.BinaryLittleEndian)
	},
}
