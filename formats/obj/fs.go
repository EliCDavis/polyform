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
func Load(objPath string) (*Scene, error) {
	inFile, err := os.Open(objPath)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()

	scene, matPaths, err := ReadMesh(bufio.NewReader(inFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read mesh: %w", err)
	}

	loadedMaterials := make(map[string]*Material)
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

	for _, o := range scene.Objects {
		for i, e := range o.Entries {
			o.Entries[i].Material = loadedMaterials[e.Material.Name]
		}
	}

	return scene, nil
}

func LoadMesh(objPath string) (*modeling.Mesh, error) {
	scene, err := Load(objPath)
	if err != nil {
		return nil, err
	}
	mesh := scene.ToMesh()
	return &mesh, nil
}

// SaveMesh writes the mesh to the path specified in OBJ format, optionally writing
// an additional MTL file with all materials that are found within the modeling.
func SaveMesh(objPath string, meshToSave modeling.Mesh) error {
	return Save(objPath, Scene{
		Objects: []Object{
			{
				Entries: []Entry{
					{
						Mesh: meshToSave,
					},
				},
			},
		},
	})
}

// Save writes all provided meshes to the path specified in OBJ format, optionally writing
// an additional MTL file with all materials that are found across all meshes.
func Save(objPath string, scene Scene) error {
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
	if scene.containsMaterials() {
		mtlName := objPath[0:len(objPath)-len(extension)] + ".mtl"
		mtlFile, err := os.Create(mtlName)
		if err != nil {
			return fmt.Errorf("failed to create material file %q: %w", mtlName, err)
		}
		defer mtlFile.Close()

		if err = WriteMaterials(scene, mtlFile); err != nil {
			return fmt.Errorf("failed to write materials: %w", err)
		}
		mtlPath = path.Base(mtlName)
	}

	out := bufio.NewWriter(objFile)

	if err = Write(scene, mtlPath, out); err != nil {
		return fmt.Errorf("failed to write mesh: %w", err)
	}

	if err = out.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	return nil
}
