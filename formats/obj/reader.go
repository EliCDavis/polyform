package obj

import (
	"bufio"
	"errors"
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
		return vector3.Zero[float64](), fmt.Errorf("unable to parse x component %q: %w", components[1], err)
	}

	parsedY, err := strconv.ParseFloat(strings.TrimSpace(components[2]), 32)
	if err != nil {
		return vector3.Zero[float64](), fmt.Errorf("unable to parse y component %q: %w", components[2], err)
	}

	parsedZ, err := strconv.ParseFloat(strings.TrimSpace(components[3]), 32)
	if err != nil {
		return vector3.Zero[float64](), fmt.Errorf("unable to parse z component %q: %w", components[3], err)
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
		return nil, errors.New("mtllib line is empty")
	}

	files := make([]string, len(components)-1)
	for i := 1; i < len(components); i++ {
		files[i-1] = strings.TrimSpace(components[i])
	}

	return files, nil
}

func parseUsemtlLine(components []string) (string, error) {
	if len(components) == 1 {
		return "", errors.New("usemtl line is empty")
	}

	return strings.Join(components[1:], " "), nil
}

func parseObjectLine(components []string) (string, error) {
	if len(components) == 1 {
		return "", errors.New("o line is empty")
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
		if err != nil {
			return v, vt, vn, fmt.Errorf("failed to convert component %q to int: %w", component, err)
		}
		return v, vt, vn, nil
	}

	if strings.Contains(component, "//") {
		components := strings.Split(component, "//")
		v, err = strconv.Atoi(components[0])
		v -= 1
		if err != nil {
			return v, vt, vn, fmt.Errorf("failed to convert //component[0] %q to int: %w", components[0], err)
		}

		if len(components) > 1 && strings.TrimSpace(components[1]) != "" {
			vn, err = strconv.Atoi(components[1])
			vn -= 1
			if err != nil {
				return v, vt, vn, fmt.Errorf("failed to convert //component[1] %q to int: %w", components[1], err)
			}
		}

		return v, vt, vn, nil
	}

	components := strings.Split(component, "/")
	v, err = strconv.Atoi(components[0])
	v -= 1
	if err != nil {
		return v, vt, vn, fmt.Errorf("failed to convert /component[0] %q to int: %w", components[0], err)
	}

	vt, err = strconv.Atoi(components[1])
	vt -= 1
	if err != nil {
		return v, vt, vn, fmt.Errorf("failed to convert /component[1] %q to int: %w", components[1], err)
	}

	if len(components) == 3 {
		vn, err = strconv.Atoi(components[2])
		vn -= 1
		if err != nil {
			return v, vt, vn, fmt.Errorf("failed to convert /component[2] %q to int: %w", components[2], err)
		}
	}
	return v, vt, vn, nil
}

type objMeshReading struct {
	pointHash map[string]int
	tris      []int
	verts     []vector3.Float64
	normals   []vector3.Float64
	uvs       []vector2.Float64
	material  *Material
}

func newObjMeshReading() objMeshReading {
	return objMeshReading{
		pointHash: make(map[string]int),
		tris:      make([]int, 0),
		verts:     make([]vector3.Float64, 0),
		normals:   make([]vector3.Float64, 0),
		uvs:       make([]vector2.Float64, 0),
	}
}

func (omr objMeshReading) empty() bool {
	return len(omr.tris) == 0
}

func (omr objMeshReading) toEntry() Entry {
	mesh := modeling.NewTriangleMesh(omr.tris).
		SetFloat3Attribute(modeling.PositionAttribute, omr.verts)

	if len(omr.normals) > 0 {
		mesh = mesh.SetFloat3Attribute(modeling.NormalAttribute, omr.normals)
	}

	if len(omr.uvs) > 0 {
		mesh = mesh.SetFloat2Attribute(modeling.TexCoordAttribute, omr.uvs)
	}
	return Entry{
		Mesh:     mesh,
		Material: omr.material,
	}
}

func ReadMesh(in io.Reader) (*Scene, []string, error) {
	scanner := bufio.NewScanner(in)

	readVerts := make([]vector3.Float64, 0)
	readNormals := make([]vector3.Float64, 0)
	readUVs := make([]vector2.Float64, 0)
	readMaterialFiles := make([]string, 0)

	meshNameToMaterial := make(map[string]*Material)

	scene := Scene{
		Objects: make([]Object, 0),
	}

	workingEntry := newObjMeshReading()
	workingObject := Object{
		Entries: make([]Entry, 0),
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLined := strings.TrimSpace(line)
		if trimmedLined == "" {
			continue
		}

		components := strings.Fields(trimmedLined)
		switch Keyword(components[0]) {

		case Group, SmoothingGroup, Comment:
			// we eat these

		case MaterialLibrary:
			materialFiles, err := parseMtllibLine(components)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse 'mtllib' line %q: %w", line, err)
			}
			readMaterialFiles = append(readMaterialFiles, materialFiles...)

		case MaterialUsage:
			matToUse, err := parseUsemtlLine(components)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse 'usemtl' line %q: %w", line, err)
			}

			var meshMat *Material

			if mat, ok := meshNameToMaterial[matToUse]; ok {
				meshMat = mat
			} else {
				meshMat = &Material{
					Name: matToUse,
				}
				meshNameToMaterial[matToUse] = meshMat
			}

			if !workingEntry.empty() {
				workingObject.Entries = append(workingObject.Entries, workingEntry.toEntry())
				workingEntry = newObjMeshReading()
			}

			workingEntry.material = meshMat

		case Vertex:
			v, err := parseObjVectorLine(components)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse %q line %q: %w", Vertex, line, err)
			}
			readVerts = append(readVerts, v)

		case Normal:
			vn, err := parseObjVectorLine(components)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse 'vn' line %q: %w", line, err)
			}
			readNormals = append(readNormals, vn)

		case TextureCoordinate:
			vt, err := parseObjTextureLine(components)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse 'vt' line %q: %w", line, err)
			}
			readUVs = append(readUVs, vt)

		case ObjectName:
			objectName, err := parseObjectLine(components)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse 'o' line %q: %w", line, err)
			}

			// Have we been working on an object's geometry already?
			if !workingEntry.empty() || len(workingObject.Entries) > 0 {

				// Finish off the current entry if we started one.
				if !workingEntry.empty() {
					workingObject.Entries = append(workingObject.Entries, workingEntry.toEntry())
					workingEntry = newObjMeshReading()
				}

				// Add the object to the scene.
				scene.Objects = append(scene.Objects, workingObject)
				workingObject = Object{Entries: make([]Entry, 0)}
			}

			workingObject.Name = objectName

		case Face:
			var p1 int
			if val, ok := workingEntry.pointHash[components[1]]; ok {
				p1 = val
			} else {
				v, vt, vn, err := parseObjFaceComponent(components[1])
				if err != nil {
					return nil, nil, fmt.Errorf("failed to parse 'f' line component[1] %q: %w", line, err)
				}
				p1 = len(workingEntry.pointHash)
				workingEntry.pointHash[components[1]] = p1

				workingEntry.verts = append(workingEntry.verts, readVerts[v])

				if vn != -1 {
					workingEntry.normals = append(workingEntry.normals, readNormals[vn])
				}

				if vt != -1 {
					workingEntry.uvs = append(workingEntry.uvs, readUVs[vt])
				}
			}

			var p2 int
			if val, ok := workingEntry.pointHash[components[2]]; ok {
				p2 = val
			} else {
				v, vt, vn, err := parseObjFaceComponent(components[2])
				if err != nil {
					return nil, nil, fmt.Errorf("failed to parse 'f' line component[2] %q: %w", line, err)
				}
				p2 = len(workingEntry.pointHash)
				workingEntry.pointHash[components[2]] = p2

				workingEntry.verts = append(workingEntry.verts, readVerts[v])

				if vn != -1 {
					workingEntry.normals = append(workingEntry.normals, readNormals[vn])
				}

				if vt != -1 {
					workingEntry.uvs = append(workingEntry.uvs, readUVs[vt])
				}
			}

			var p3 int
			if val, ok := workingEntry.pointHash[components[3]]; ok {
				p3 = val
			} else {
				v, vt, vn, err := parseObjFaceComponent(components[3])
				if err != nil {
					return nil, nil, fmt.Errorf("failed to parse 'f' line component[3] %q: %w", line, err)
				}
				p3 = len(workingEntry.pointHash)
				workingEntry.pointHash[components[3]] = p3

				workingEntry.verts = append(workingEntry.verts, readVerts[v])

				if vn != -1 {
					workingEntry.normals = append(workingEntry.normals, readNormals[vn])
				}

				if vt != -1 {
					workingEntry.uvs = append(workingEntry.uvs, readUVs[vt])
				}
			}

			workingEntry.tris = append(workingEntry.tris, p1, p2, p3)

		default:
			return nil, nil, fmt.Errorf("unexpected keyword %s on line %q", components[0], line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to run scanner: %w", err)
	}

	if !workingEntry.empty() || len(workingObject.Entries) > 0 {

		// Finish off the current entry if we started one.
		if !workingEntry.empty() {
			workingObject.Entries = append(workingObject.Entries, workingEntry.toEntry())
		}

		// Add the object to the scene.
		scene.Objects = append(scene.Objects, workingObject)
		workingObject = Object{Entries: make([]Entry, 0)}
	}

	return &scene, readMaterialFiles, nil
}
