package main

import (
	"os"
	"path/filepath"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/urfave/cli/v2"
)

var ToGLTFCommand = &cli.Command{
	Name:  "to-gltf",
	Usage: "converts a PLY file to gltf",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "out",
			Usage:   "Path to write the GLTF to",
			Aliases: []string{"o"},
			Value:   "out.glb",
		},
	},
	Action: func(ctx *cli.Context) error {
		mesh, err := getPlyFile()
		if err != nil {
			return err
		}

		meshPath := ctx.String("out")
		err = os.MkdirAll(filepath.Dir(meshPath), 0777)
		if err != nil {
			return err
		}

		cleanedMesh := mesh.Transform(
			meshops.VertexColorSpaceTransformer{
				Transformation:         meshops.VertexColorSpaceSRGBToLinear,
				SkipOnMissingAttribute: true,
			},
		)

		return gltf.Save(meshPath, gltf.PolyformScene{
			Models: []gltf.PolyformModel{
				{Name: "PLY", Mesh: &cleanedMesh},
			},
		})
	},
}
