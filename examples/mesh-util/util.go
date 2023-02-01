package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/formats/ply"
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
		return obj.Load(path)

	case ".ply":
		return ply.ReadMesh(inFile)

	default:
		return nil, fmt.Errorf("unimplemented format: %s", ext)
	}
}
