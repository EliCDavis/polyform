package obj

import (
	"bufio"
	"fmt"
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
		return nil, fmt.Errorf("failed to read mesh: %w", err)
	}

	loadedMaterials := make(map[string]*modeling.Material)
	for _, matPath := range matPaths {
		matFilePath := path.Join(path.Dir(objPath), matPath)
		matFile, err := os.Open(matFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open material file %q: %w", matFilePath, err)
		}
		defer matFile.Close()

		materials, err := ReadMaterials(matFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read materials: %w", err)
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
	if err := os.MkdirAll(path.Dir(objPath), os.ModeDir); err != nil {
		return fmt.Errorf("failed to create all dirs for path %q: %w", objPath, err)
	}

	objFile, err := os.Create(objPath)
	if err != nil {
		return fmt.Errorf("failed to create object file %q: %w", objPath, err)
	}
	defer objFile.Close()

	extension := filepath.Ext(objPath)
	mtlPath := ""
	if len(meshToSave.Materials()) > 0 {
		mtlName := objPath[0:len(objPath)-len(extension)] + ".mtl"
		mtlFile, err := os.Create(mtlName)
		if err != nil {
			return fmt.Errorf("failed to create material file %q: %w", mtlName, err)
		}
		defer mtlFile.Close()

		if err = WriteMaterialsFromMesh(meshToSave, mtlFile); err != nil {
			return fmt.Errorf("failed to write materials: %w", err)
		}
		mtlPath = path.Base(mtlName)
	}

	out := bufio.NewWriter(objFile)
	if err = WriteMesh(meshToSave, mtlPath, out); err != nil {
		return fmt.Errorf("failed to write mesh: %w", err)
	}

	if err = out.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	return nil
}

// SaveAll writes all provided meshes to the path specified in OBJ format, optionally writing
// an additional MTL file with materials are found within the modeling.
func SaveAll(objPath string, meshesToSave map[string]modeling.Mesh) error {
	if err := os.MkdirAll(path.Dir(objPath), os.ModeDir); err != nil {
		return fmt.Errorf("failed to create all dirs for path %q: %w", objPath, err)
	}

	objFile, err := os.Create(objPath)
	if err != nil {
		return fmt.Errorf("failed to create object file %q: %w", objPath, err)
	}
	defer objFile.Close()

	extension := filepath.Ext(objPath)
	var materials []modeling.MeshMaterial
	var objMeshes []ObjMesh
	for name, mesh := range meshesToSave {
		materials = append(materials, mesh.Materials()...)
		objMeshes = append(objMeshes, ObjMesh{
			Name: name, Mesh: mesh,
		})
	}

	mtlPath := ""
	if len(materials) > 0 {
		mtlName := objPath[0:len(objPath)-len(extension)] + ".mtl"
		mtlFile, err := os.Create(mtlName)
		if err != nil {
			return fmt.Errorf("failed to create material file %q: %w", mtlName, err)
		}
		defer mtlFile.Close()

		if err = WriteMaterials(materials, mtlFile); err != nil {
			return fmt.Errorf("failed to write materials: %w", err)
		}
		mtlPath = path.Base(mtlName)
	}

	out := bufio.NewWriter(objFile)

	if err = WriteMeshes(objMeshes, mtlPath, out); err != nil {
		return fmt.Errorf("failed to write mesh: %w", err)
	}

	if err = out.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	return nil
}
