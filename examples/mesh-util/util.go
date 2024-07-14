package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/formats/stl"
	"github.com/EliCDavis/polyform/modeling"
)

func readMesh(path string) (*modeling.Mesh, error) {
	ext := filepath.Ext(path)

	inFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()

	switch strings.ToLower(ext) {

	case ".obj":
		meshes, err := obj.Load(path)
		mesh := modeling.EmptyMesh(modeling.TriangleTopology)
		for _, m := range meshes {
			mesh = mesh.Append(m.Mesh)
		}
		return &mesh, err

	case ".ply":
		return ply.ReadMesh(bufio.NewReader(inFile))

	case ".stl":
		return stl.ReadMesh(bufio.NewReader(inFile))

	default:
		return nil, fmt.Errorf("unimplemented format: %s", ext)
	}
}
