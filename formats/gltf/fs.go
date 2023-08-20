package gltf

import (
	"bufio"
	"io"
	"os"
	"path"
)

func save(gltfPath string, scene PolyformScene, saveFunc func(scene PolyformScene, out io.Writer) error) error {
	err := os.MkdirAll(path.Dir(gltfPath), os.ModeDir)
	if err != nil {
		return err
	}

	gltfFile, err := os.Create(gltfPath)
	if err != nil {
		return err
	}
	defer gltfFile.Close()

	out := bufio.NewWriter(gltfFile)
	err = saveFunc(scene, out)
	if err != nil {
		return err
	}
	return out.Flush()
}

// Save writes the mesh to the path specified in GLTF format
func SaveText(gltfPath string, scene PolyformScene) error {
	return save(gltfPath, scene, WriteText)
}

// Save writes the mesh to the path specified in GLB format
func SaveBinary(gltfPath string, scene PolyformScene) error {
	return save(gltfPath, scene, WriteBinary)
}
