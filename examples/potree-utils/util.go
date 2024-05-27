package main

import (
	"os"
	"path/filepath"

	"github.com/EliCDavis/polyform/formats/potree"
	"github.com/urfave/cli/v2"
)

var metadataFlag = &cli.StringFlag{
	Name:  "metadata",
	Value: "metadata.json",
}

var hierarchyFlag = &cli.StringFlag{
	Name:  "hierarchy",
	Value: "",
	Usage: "If blank, it will assume the file named 'hierarchy.bin' located in the same folder as the metadata.json.",
}

var octreeFlag = &cli.StringFlag{
	Name:  "octree",
	Value: "",
	Usage: "If blank, it will assume the file named 'octree.bin' located in the same folder as the metadata.json.",
}

func loadHierarchy(ctx *cli.Context) (*potree.Metadata, *potree.OctreeNode, error) {
	metadataPath := ctx.String("metadata")
	hierarchyPath := ctx.String("hierarchy")
	if hierarchyPath == "" {
		hierarchyPath = filepath.Join(filepath.Dir(metadataPath), "hierarchy.bin")
	}

	metadata, err := potree.LoadMetadata(metadataPath)
	if err != nil {
		return nil, nil, err
	}

	hierarchy, err := metadata.LoadHierarchy(hierarchyPath)
	return metadata, hierarchy, err
}

func openOctreeFile(ctx *cli.Context) (*os.File, error) {
	metadataPath := ctx.String("metadata")
	octreePath := ctx.String("octree")
	if octreePath == "" {
		octreePath = filepath.Join(filepath.Dir(metadataPath), "octree.bin")
	}

	return os.Open(octreePath)
}
