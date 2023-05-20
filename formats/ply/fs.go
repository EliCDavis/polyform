package ply

import (
	"bufio"
	"os"
	"path"

	"github.com/EliCDavis/polyform/modeling"
)

// Save writes the mesh to the path specified in PLY format
func Save(plyPath string, meshToSave modeling.Mesh) error {
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
	err = WriteASCII(out, meshToSave)
	if err != nil {
		return err
	}
	return out.Flush()
}

func SaveBinary(plyPath string, meshToSave modeling.Mesh) error {
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
	err = WriteBinary(out, meshToSave)
	if err != nil {
		return err
	}
	return out.Flush()
}
