package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/mesh/obj"
	"github.com/EliCDavis/mesh/ply"
)

func readMesh(path string) (*mesh.Mesh, error) {
	ext := filepath.Ext(path)

	inFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()

	switch strings.ToLower(ext) {

	case ".obj":
		mesh, _, err := obj.ReadMesh(inFile)
		return mesh, err

	case ".ply":
		return ply.ToMesh(inFile)

	default:
		return nil, fmt.Errorf("unimplemented format: %s", ext)
	}
}
