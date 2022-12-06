package obj

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/vector"
)

func parseObjVectorLine(components []string) (vector.Vector3, error) {
	parsedX, err := strconv.ParseFloat(strings.TrimSpace(components[1]), 32)
	if err != nil {
		return vector.Vector3Zero(), fmt.Errorf("unable to parse x component '%s': %w", components[1], err)
	}

	parsedY, err := strconv.ParseFloat(strings.TrimSpace(components[2]), 32)
	if err != nil {
		return vector.Vector3Zero(), fmt.Errorf("unable to parse y component '%s': %w", components[2], err)
	}

	parsedZ, err := strconv.ParseFloat(strings.TrimSpace(components[3]), 32)
	if err != nil {
		return vector.Vector3Zero(), fmt.Errorf("unable to parse z component '%s': %w", components[3], err)
	}

	return vector.NewVector3(parsedX, parsedY, parsedZ), nil
}

func parseObjTextureLine(components []string) (vector.Vector2, error) {
	parsedX, err := strconv.ParseFloat(strings.TrimSpace(components[1]), 32)
	if err != nil {
		return vector.Vector2Zero(), fmt.Errorf("unable to parse tex x: %w", err)
	}

	parsedY, err := strconv.ParseFloat(strings.TrimSpace(components[2]), 32)
	if err != nil {
		return vector.Vector2Zero(), fmt.Errorf("unable to parse tex y: %w", err)
	}

	return vector.NewVector2(parsedX, parsedY), nil
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

func ReadMesh(in io.Reader) (*mesh.Mesh, []string, error) {
	scanner := bufio.NewScanner(in)

	tris := make([]int, 0)
	readVerts := make([]vector.Vector3, 0)
	readNormals := make([]vector.Vector3, 0)
	readUVs := make([]vector.Vector2, 0)
	readMaterialFiles := make([]string, 0)

	pointHash := make(map[string]int)
	verts := make([]vector.Vector3, 0)
	normals := make([]vector.Vector3, 0)
	uvs := make([]vector.Vector2, 0)
	materials := make([]mesh.MeshMaterial, 0)

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
				if len(materials) == 0 {
					materials = append(materials, mesh.MeshMaterial{
						NumOfTris: trisSenseLastMat,
						Material: &mesh.Material{
							Name: "Default",
						},
					})
				} else {
					materials[len(materials)-1].NumOfTris = trisSenseLastMat
				}
			}

			trisSenseLastMat = 0

			materials = append(materials, mesh.MeshMaterial{
				NumOfTris: 0,
				Material: &mesh.Material{
					Name: matToUse,
				},
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
		if len(materials) > 0 {
			materials[len(materials)-1].NumOfTris = trisSenseLastMat
		}
	}

	mesh := mesh.NewMeshWithMaterials(
		tris,
		verts,
		normals,
		[][]vector.Vector2{
			uvs,
		},
		materials,
	)

	return &mesh, readMaterialFiles, nil
}
