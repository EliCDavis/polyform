package main

import (
	"fmt"
	"io"
	"os"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/formats/potree"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/urfave/cli/v2"
)

func buildModel(octreeFile *os.File, node *potree.OctreeNode, metadata *potree.Metadata, includeChildren bool) (*modeling.Mesh, error) {
	_, err := octreeFile.Seek(int64(node.ByteOffset), 0)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, node.ByteSize)
	_, err = io.ReadFull(octreeFile, buf)
	if err != nil {
		return nil, err
	}

	mesh := potree.LoadNode(*node, *metadata, buf)
	if includeChildren {
		for _, c := range node.Children {
			childMesh, err := buildModel(octreeFile, c, metadata, true)
			if err != nil {
				return nil, err
			}
			mesh = mesh.Append(*childMesh)
		}
	}
	return &mesh, nil
}

var ExtractPointcloudCommand = &cli.Command{
	Name: "extract-pointcloud",
	Flags: []cli.Flag{
		metadataFlag,
		hierarchyFlag,
		octreeFlag,
		&cli.StringFlag{
			Name:  "node",
			Value: "r",
			Usage: "Name of node to extract point data from",
		},
		&cli.BoolFlag{
			Name:  "include-children",
			Value: false,
			Usage: "Whether or not to include children data",
		},
		&cli.StringFlag{
			Name:  "out",
			Value: "out.ply",
			Usage: "Name of ply file to write pointcloud data too",
		},
	},
	Action: func(ctx *cli.Context) error {
		metadata, hierarchy, err := loadHierarchy(ctx)
		if err != nil {
			return err
		}

		octreeFile, err := openOctreeFile(ctx)
		if err != nil {
			return err
		}
		defer octreeFile.Close()

		mesh, err := buildModel(octreeFile, hierarchy, metadata, ctx.Bool("include-children"))
		if err != nil {
			return err
		}

		fmt.Fprintf(ctx.App.Writer, "Writing pointcloud with %d points to %s", mesh.Indices().Len(), ctx.String("out"))

		return ply.SaveBinary(ctx.String("out"), *mesh)
	},
}
