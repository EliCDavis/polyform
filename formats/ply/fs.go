// Package ply implements the PLY file format and provides utilities for
// interacting with the data within the rest of polyform.
package ply

import (
	"bufio"
	"os"
	"path"

	"github.com/EliCDavis/polyform/modeling"
)

// Save writes the mesh to the path specified in PLY format
func Save(plyPath string, meshToSave modeling.Mesh, format Format) error {
	err := os.MkdirAll(path.Dir(plyPath), os.ModeDir)
	if err != nil {
		return err
	}

	plyFile, err := os.Create(plyPath)
	if err != nil {
		return err
	}
	defer plyFile.Close()

	out := bufio.NewWriter(plyFile)
	err = Write(out, meshToSave, format, "")
	if err != nil {
		return err
	}
	return out.Flush()
}

// Build a polyform mesh from the contents of the ply file specied at the
// filepath
func Load(filepath string) (*modeling.Mesh, error) {
	in, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	return ReadMesh(bufio.NewReader(in))
}
