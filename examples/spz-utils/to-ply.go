package main

import (
	"bufio"
	"os"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/formats/spz"
	"github.com/urfave/cli/v2"
)

var ToPlyCommand = &cli.Command{
	Name: "to-ply",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "out",
			Usage:    "path to write ply data to",
			Aliases:  []string{"o"},
			Required: true,
		},
	},
	Action: func(ctx *cli.Context) error {
		cloud, err := spz.Load(inFilePath)
		if err != nil {
			return err
		}

		plyFile, err := os.Create(ctx.String("out"))
		if err != nil {
			return err
		}
		defer plyFile.Close()
		out := bufio.NewWriter(plyFile)

		header := ply.Header{
			Format: ply.BinaryLittleEndian,
			Elements: []ply.Element{
				{
					Name:  "vertex",
					Count: int64(cloud.Header.NumPoints),
					Properties: []ply.Property{
						ply.ScalarProperty{PropertyName: "x", Type: ply.Float},
						ply.ScalarProperty{PropertyName: "y", Type: ply.Float},
						ply.ScalarProperty{PropertyName: "z", Type: ply.Float},
						ply.ScalarProperty{PropertyName: "f_dc_0", Type: ply.UChar},
						ply.ScalarProperty{PropertyName: "f_dc_1", Type: ply.UChar},
						ply.ScalarProperty{PropertyName: "f_dc_2", Type: ply.UChar},
						ply.ScalarProperty{PropertyName: "scale_0", Type: ply.UChar},
						ply.ScalarProperty{PropertyName: "scale_1", Type: ply.UChar},
						ply.ScalarProperty{PropertyName: "scale_2", Type: ply.UChar},
						ply.ScalarProperty{PropertyName: "rot_0", Type: ply.UChar},
						ply.ScalarProperty{PropertyName: "rot_1", Type: ply.UChar},
						ply.ScalarProperty{PropertyName: "rot_2", Type: ply.UChar},
						ply.ScalarProperty{PropertyName: "opacity", Type: ply.UChar},
					},
				},
			},
		}

		err = header.Write(out)
		if err != nil {
			return err
		}

		return nil
	},
}
