package gltf

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
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

// SaveText writes the mesh to the path specified in GLTF format
func SaveText(gltfPath string, scene PolyformScene) error {
	return save(gltfPath, scene, WriteText)
}

// SaveBinary writes the mesh to the path specified in GLB format
func SaveBinary(gltfPath string, scene PolyformScene) error {
	return save(gltfPath, scene, WriteBinary)
}

// Save writes the mesh to the path in the format dictated by the extension in
// the path
func Save(modelPath string, scene PolyformScene) error {
	ext := filepath.Ext(modelPath)
	switch ext {
	case ".glb":
		return SaveBinary(modelPath, scene)

	case ".gltf":
		return SaveText(modelPath, scene)

	default:
		panic(fmt.Errorf("don't know how to save file with extension: %s", ext))
	}
}
