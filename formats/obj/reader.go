package obj

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func parseObjVectorLine(components []string) (vector3.Float64, error) {
	parsedX, err := strconv.ParseFloat(strings.TrimSpace(components[1]), 32)
	if err != nil {
		return vector3.Zero[float64](), fmt.Errorf("unable to parse x component '%s': %w", components[1], err)
	}

	parsedY, err := strconv.ParseFloat(strings.TrimSpace(components[2]), 32)
	if err != nil {
		return vector3.Zero[float64](), fmt.Errorf("unable to parse y component '%s': %w", components[2], err)
	}

	parsedZ, err := strconv.ParseFloat(strings.TrimSpace(components[3]), 32)
	if err != nil {
		return vector3.Zero[float64](), fmt.Errorf("unable to parse z component '%s': %w", components[3], err)
	}

	return vector3.New(parsedX, parsedY, parsedZ), nil
}

func parseObjTextureLine(components []string) (vector2.Float64, error) {
	parsedX, err := strconv.ParseFloat(strings.TrimSpace(components[1]), 32)
	if err != nil {
		return vector2.Zero[float64](), fmt.Errorf("unable to parse tex x: %w", err)
	}

	parsedY, err := strconv.ParseFloat(strings.TrimSpace(components[2]), 32)
	if err != nil {
		return vector2.Zero[float64](), fmt.Errorf("unable to parse tex y: %w", err)
	}

	return vector2.New(parsedX, parsedY), nil
}

func parseMtllibLine(components []string) ([]string, error) {
	if len(components) == 1 {
		return nil, fmt.Errorf("mtllib line is empty")
	}

	files := make([]string, len(components)-1)
	for i := 1; i < len(components); i++ {
		files[i-1] = strings.TrimSpace(components[i])
	}

	return files, nil
}

func parseUsemtlLine(components []string) (string, error) {
	if len(components) == 1 {
		return "", fmt.Errorf("usemtl line is empty")
	}

	return strings.Join(components[1:], " "), nil
}

func parseObjFaceComponent(component string) (v int, vt int, vn int, err error) {
	v = -1
	vt = -1
	vn = -1

	if !strings.Contains(component, "/") {
		v, err = strconv.Atoi(component)
		v -= 1
		return
	}

	if strings.Contains(component, "//") {
		components := strings.Split(component, "//")
		v, err = strconv.Atoi(components[0])
		v -= 1
		if err != nil {
			return
		}
		vn, err = strconv.Atoi(components[1])
		vn -= 1
		return
	}

	components := strings.Split(component, "/")
	v, err = strconv.Atoi(components[0])
	v -= 1
	if err != nil {
		return
	}

	vt, err = strconv.Atoi(components[1])
	vt -= 1
	if err != nil {
		return
	}

	if len(components) == 3 {
		vn, err = strconv.Atoi(components[2])
		vn -= 1
	}
	return
}

func ReadMesh(in io.Reader) (*modeling.Mesh, []string, error) {
	scanner := bufio.NewScanner(in)

	tris := make([]int, 0)
	readVerts := make([]vector3.Float64, 0)
	readNormals := make([]vector3.Float64, 0)
	readUVs := make([]vector2.Float64, 0)
	readMaterialFiles := make([]string, 0)

	pointHash := make(map[string]int)
	verts := make([]vector3.Float64, 0)
	normals := make([]vector3.Float64, 0)
	uvs := make([]vector2.Float64, 0)
	meshMaterials := make([]modeling.MeshMaterial, 0)
	meshNameToMaterial := make(map[string]*modeling.Material)

	trisSenseLastMat := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		components := strings.Fields(line)
		switch components[0] {
		case "mtllib":
			materialFiles, err := parseMtllibLine(components)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to parse mtllib line '%s': %w", line, err)
			}
			readMaterialFiles = append(readMaterialFiles, materialFiles...)

		case "usemtl":
			matToUse, err := parseUsemtlLine(components)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to parse usemtl line '%s': %w", line, err)
			}

			if trisSenseLastMat > 0 {
				if len(meshMaterials) == 0 {
					meshMaterials = append(meshMaterials, modeling.MeshMaterial{
						PrimitiveCount: trisSenseLastMat,
						Material: &modeling.Material{
							Name: "Default",
						},
					})
				} else {
					meshMaterials[len(meshMaterials)-1].PrimitiveCount = trisSenseLastMat
				}
			}

			trisSenseLastMat = 0

			var meshMat *modeling.Material = nil

			if mat, ok := meshNameToMaterial[matToUse]; ok {
				meshMat = mat
			} else {
				meshMat = &modeling.Material{
					Name: matToUse,
				}
				meshNameToMaterial[matToUse] = meshMat
			}

			meshMaterials = append(meshMaterials, modeling.MeshMaterial{
				PrimitiveCount: 0,
				Material:       meshMat,
			})

		case "v":
			v, err := parseObjVectorLine(components)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to parse vertex line '%s': %w", line, err)
			}
			readVerts = append(readVerts, v)

		case "vn":
			vn, err := parseObjVectorLine(components)
			if err != nil {
				return nil, nil, err
			}
			readNormals = append(readNormals, vn)

		case "vt":
			vt, err := parseObjTextureLine(components)
			if err != nil {
				return nil, nil, err
			}
			readUVs = append(readUVs, vt)

		case "f":

			trisSenseLastMat++

			var p1 int
			if val, ok := pointHash[components[1]]; ok {
				p1 = val
			} else {
				v, vt, vn, err := parseObjFaceComponent(components[1])
				if err != nil {
					return nil, nil, err
				}
				p1 = len(pointHash)
				pointHash[components[1]] = p1

				verts = append(verts, readVerts[v])

				if vn != -1 {
					normals = append(normals, readNormals[vn])
				}

				if vt != -1 {
					uvs = append(uvs, readUVs[vt])
				}
			}

			var p2 int
			if val, ok := pointHash[components[2]]; ok {
				p2 = val
			} else {
				v, vt, vn, err := parseObjFaceComponent(components[2])
				if err != nil {
					return nil, nil, err
				}
				p2 = len(pointHash)
				pointHash[components[2]] = p2

				verts = append(verts, readVerts[v])

				if vn != -1 {
					normals = append(normals, readNormals[vn])
				}

				if vt != -1 {
					uvs = append(uvs, readUVs[vt])
				}
			}

			var p3 int
			if val, ok := pointHash[components[3]]; ok {
				p3 = val
			} else {
				v, vt, vn, err := parseObjFaceComponent(components[3])
				if err != nil {
					return nil, nil, err
				}
				p3 = len(pointHash)
				pointHash[components[3]] = p3

				verts = append(verts, readVerts[v])

				if vn != -1 {
					normals = append(normals, readNormals[vn])
				}

				if vt != -1 {
					uvs = append(uvs, readUVs[vt])
				}
			}

			tris = append(tris, p1, p2, p3)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	if trisSenseLastMat > 0 {
		if len(meshMaterials) > 0 {
			meshMaterials[len(meshMaterials)-1].PrimitiveCount = trisSenseLastMat
		}
	}

	mesh := modeling.NewMesh(
		tris,
		map[string][]vector3.Float64{
			modeling.PositionAttribute: verts,
			modeling.NormalAttribute:   normals,
		},
		map[string][]vector2.Float64{
			modeling.TexCoordAttribute: uvs,
		},
		nil,
		meshMaterials,
	)

	return &mesh, readMaterialFiles, nil
}
