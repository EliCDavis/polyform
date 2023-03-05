package gltf

import (
	"bufio"
	"os"
	"path"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/animation"
)

// Save writes the mesh to the path specified in GLTF format
func SaveText(gltfPath string, meshToSave modeling.Mesh) error {
	return SaveTextWithAnimations(gltfPath, meshToSave, nil, nil)
}

func SaveTextWithAnimations(gltfPath string, meshToSave modeling.Mesh, joints *animation.Skeleton, animations []animation.Sequence) error {
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
	err = WriteTextWithAnimations(meshToSave, out, joints, animations)
	if err != nil {
		return err
	}
	return out.Flush()
}
