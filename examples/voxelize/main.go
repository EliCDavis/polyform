package main

import (
	"log"
	"os"
	"time"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/voxelize"
	"github.com/EliCDavis/vector/vector3"
	"github.com/urfave/cli/v2"
)

func main() {
	const inFlag = "in"
	const outFlag = "out"
	const voxelSizeFlag = "voxel-size"

	app := &cli.App{
		Name:  "voxelize",
		Usage: "Iterates through each vertex of the mesh passed in and computes the corresponding voxel the vertex belongs to. The output mesh is a visualization of those voxels.",
		Authors: []*cli.Author{
			{
				Name:  "Eli Davis",
				Email: "eli@recolude.com",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     inFlag,
				Required: true,
			},
			&cli.StringFlag{
				Name:     outFlag,
				Required: true,
			},
			&cli.Float64Flag{
				Name:     voxelSizeFlag,
				Required: true,
			},
		},
		Action: func(ctx *cli.Context) error {
			meshes, err := obj.Load(ctx.String(inFlag))
			if err != nil {
				return err
			}

			mesh := modeling.EmptyMesh(modeling.TriangleTopology)
			for _, m := range meshes {
				mesh = mesh.Append(m.Mesh)
			}

			voxelSize := ctx.Float64(voxelSizeFlag)

			startVoxel := time.Now()
			voxels := voxelize.Surface(mesh, modeling.PositionAttribute, voxelSize)
			log.Printf("Time to voxelize: %s", time.Since(startVoxel))

			startMesh := time.Now()
			voxelizedMesh := modeling.EmptyMesh(modeling.TriangleTopology)
			for _, v := range voxels {
				cube := primitives.Cube().
					Scale(vector3.Fill(voxelSize)).
					Translate(v)
				voxelizedMesh = voxelizedMesh.Append(cube)
			}
			log.Printf("Time to mesh: %s", time.Since(startMesh))

			return obj.Save(ctx.String(outFlag), voxelizedMesh)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
