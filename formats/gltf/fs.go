package gltf

import (
	"bufio"
	"os"
	"path"

	"github.com/EliCDavis/polyform/modeling"
)

// Save writes the mesh to the path specified in OBJ format, optionally writing
// an additional MTL file with materials are found within the modeling.
func SaveText(gltfPath string, meshToSave modeling.Mesh) error {
	err := os.MkdirAll(path.Dir(gltfPath), os.ModeDir)
	if err != nil {
		return err
	}

	gltfFile, err := os.Create(gltfPath)
	if err != nil {
		return err
	}
	defer gltfFile.Close()

	// extension := filepath.Ext(gltfPath)
	// mtlName := gltfPath[0:len(gltfPath)-len(extension)] + ".bin"
	// if len(meshToSave.Materials()) > 0 {
	// 	mtlFile, err := os.Create(mtlName)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer mtlFile.Close()

	// 	err = WriteMaterials(meshToSave, mtlFile)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	out := bufio.NewWriter(gltfFile)
	err = WriteText(meshToSave, out)
	if err != nil {
		return err
	}
	return out.Flush()
}
