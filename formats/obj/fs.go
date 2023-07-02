package obj

import (
	"bufio"
	"os"
	"path"
	"path/filepath"

	"github.com/EliCDavis/polyform/modeling"
)

// Load reads an obj file from the path specified, and optionally loads all
// associated metadata files the obj file might reference.
func Load(objPath string) ([]ObjMesh, error) {
	inFile, err := os.Open(objPath)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()

	buf := bufio.NewReader(inFile)
	meshes, matPaths, err := ReadMesh(buf)
	if err != nil {
		return nil, err
	}

	loadedMaterials := make(map[string]*modeling.Material)
	for _, matPath := range matPaths {
		matFile, err := os.Open(path.Join(path.Dir(objPath), matPath))
		if err != nil {
			return nil, err
		}
		defer matFile.Close()

		materials, err := ReadMaterials(matFile)
		if err != nil {
			return nil, err
		}
		for matI, mat := range materials {
			loadedMaterials[mat.Name] = &materials[matI]
		}
	}

	for meshI, mesh := range meshes {
		for matI, mat := range mesh.Mesh.Materials() {
			meshes[meshI].Mesh.Materials()[matI].Material = loadedMaterials[mat.Material.Name]
		}
	}

	return meshes, nil
}

// Save writes the mesh to the path specified in OBJ format, optionally writing
// an additional MTL file with materials are found within the modeling.
func Save(objPath string, meshToSave modeling.Mesh) error {
	err := os.MkdirAll(path.Dir(objPath), os.ModeDir)
	if err != nil {
		return err
	}

	objFile, err := os.Create(objPath)
	if err != nil {
		return err
	}
	defer objFile.Close()

	extension := filepath.Ext(objPath)
	mtlPath := ""
	if len(meshToSave.Materials()) > 0 {
		mtlName := objPath[0:len(objPath)-len(extension)] + ".mtl"
		mtlFile, err := os.Create(mtlName)
		if err != nil {
			return err
		}
		defer mtlFile.Close()

		err = WriteMaterials(meshToSave, mtlFile)
		if err != nil {
			return err
		}
		mtlPath = path.Base(mtlName)
	}

	out := bufio.NewWriter(objFile)
	err = WriteMesh(meshToSave, mtlPath, out)
	if err != nil {
		return err
	}
	return out.Flush()
}
