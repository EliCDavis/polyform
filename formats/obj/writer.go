package obj

import (
	"fmt"
	"io"
	"strings"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/formats/txt"
	"github.com/EliCDavis/polyform/modeling"
)

func calculateMaterialNames(scene Scene) map[*Material]string {
	written := make(map[*Material]string)
	writtenNames := make(map[string]struct{})

	for _, o := range scene.Objects {
		for _, e := range o.Entries {
			if _, ok := written[e.Material]; ok {
				continue
			}

			nameToUse := "Default Diffuse"
			if e.Material != nil {
				nameToUse = e.Material.Name
			}

			if strings.TrimSpace(nameToUse) == "" {
				nameToUse = "unamed"
			}

			if _, ok := writtenNames[nameToUse]; ok {
				duplicateNameCount := 2
				for {
					attempt := fmt.Sprintf("%s%d", nameToUse, duplicateNameCount)
					if _, ok := writtenNames[attempt]; !ok {
						nameToUse = attempt
						break
					}
					duplicateNameCount++
				}
			}

			written[e.Material] = nameToUse
			writtenNames[nameToUse] = struct{}{}
		}
	}

	return written
}

func WriteMaterials(scene Scene, out io.Writer) error {
	_, _ = fmt.Fprintln(out, "# Created with github.com/EliCDavis/polyform")

	written := make(map[*Material]struct{})
	names := calculateMaterialNames(scene)
	dm := DefaultMaterial()

	for _, o := range scene.Objects {
		for _, e := range o.Entries {
			if _, ok := written[e.Material]; ok {
				continue
			}

			mat := e.Material
			if mat == nil {
				mat = &dm
			}

			if err := mat.write(names[e.Material], out); err != nil {
				return fmt.Errorf("failed to write material %q on object %q: %w", mat.Name, o.Name, err)
			}

			written[e.Material] = struct{}{}
		}
	}

	return nil
}

func writeUsingMaterial(matName string, out *txt.Writer) {
	out.StartEntry()
	out.String("usemtl ")
	out.String(strings.Replace(matName, " ", "", -1))
	out.NewLine()
	out.FinishEntry()
}

func writeFaceVerts(tris *iter.ArrayIterator[int], out *txt.Writer, offset int) {
	shift := 1 + offset
	for triIndex := 0; triIndex < tris.Len(); triIndex += 3 {
		out.StartEntry()
		out.String("f ")
		out.Int(tris.At(triIndex) + shift)
		out.Space()
		out.Int(tris.At(triIndex+1) + shift)
		out.Space()
		out.Int(tris.At(triIndex+2) + shift)
		out.NewLine()
		out.FinishEntry()
	}
}

func writeFaceVertsAndUvs(tris *iter.ArrayIterator[int], out *txt.Writer, offset int) {
	shift := 1 + offset
	for triIndex := 0; triIndex < tris.Len(); triIndex += 3 {
		p1 := tris.At(triIndex) + shift
		p2 := tris.At(triIndex+1) + shift
		p3 := tris.At(triIndex+2) + shift

		out.StartEntry()
		out.String("f ")

		out.Int(p1)
		out.String("/")
		out.Int(p1)
		out.Space()

		out.Int(p2)
		out.String("/")
		out.Int(p2)
		out.Space()

		out.Int(p3)
		out.String("/")
		out.Int(p3)
		out.NewLine()
		out.FinishEntry()
	}
}

func writeFaceVertsAndNormals(tris *iter.ArrayIterator[int], out *txt.Writer, offset int) {
	shift := 1 + offset
	for triIndex := 0; triIndex < tris.Len(); triIndex += 3 {
		p1 := tris.At(triIndex) + shift
		p2 := tris.At(triIndex+1) + shift
		p3 := tris.At(triIndex+2) + shift

		out.StartEntry()
		out.String("f ")

		out.Int(p1)
		out.String("//")
		out.Int(p1)
		out.Space()

		out.Int(p2)
		out.String("//")
		out.Int(p2)
		out.Space()

		out.Int(p3)
		out.String("//")
		out.Int(p3)
		out.NewLine()
		out.FinishEntry()
	}
}

func writeFaceVertAndUvsAndNormals(tris *iter.ArrayIterator[int], out *txt.Writer, offset int) {
	shift := 1 + offset
	for triIndex := 0; triIndex < tris.Len(); triIndex += 3 {
		p1 := tris.At(triIndex) + shift
		p2 := tris.At(triIndex+1) + shift
		p3 := tris.At(triIndex+2) + shift

		out.StartEntry()
		out.String("f ")

		out.Int(p1)
		out.String("/")
		out.Int(p1)
		out.String("/")
		out.Int(p1)
		out.Space()

		out.Int(p2)
		out.String("/")
		out.Int(p2)
		out.String("/")
		out.Int(p2)
		out.Space()

		out.Int(p3)
		out.String("/")
		out.Int(p3)
		out.String("/")
		out.Int(p3)
		out.NewLine()
		out.FinishEntry()
	}
}

func WriteMesh(m modeling.Mesh, materialFile string, out io.Writer) error {
	return Write(
		Scene{
			Objects: []Object{
				{
					Entries: []Entry{{Mesh: m}},
				},
			},
		},
		materialFile,
		out,
	)
}

func Write(scene Scene, materialFile string, out io.Writer) error {
	if _, err := fmt.Fprintln(out, "# Created with github.com/EliCDavis/polyform"); err != nil {
		return fmt.Errorf("failed to write attribution comment: %w", err)
	}

	if materialFile != "" {
		if _, err := fmt.Fprintf(out, "mtllib %s\no mesh\n", materialFile); err != nil {
			return fmt.Errorf("failed to write matfile 'mesh': %w", err)
		}
	}

	writer := txt.NewWriter(out)
	var matNames map[*Material]string = nil
	if materialFile != "" {
		matNames = calculateMaterialNames(scene)
	}

	if err := scene.writeVertexData(writer); err != nil {
		return err
	}

	var faceWriter func(tris *iter.ArrayIterator[int], out *txt.Writer, offset int)

	indexOffset := 0
	for _, obj := range scene.Objects {
		name := strings.TrimSpace(obj.Name)
		if name == "" && len(scene.Objects) > 1 {
			name = "Unamed"
		}

		if strings.Contains(name, "\n") {
			return fmt.Errorf("object name contains linebreaks: %q", obj.Name)
		}

		if name != "" {
			fmt.Fprintf(out, "o %s\n", name)
		}

		for _, entry := range obj.Entries {
			m := entry.Mesh
			if m.HasVertexAttribute(modeling.NormalAttribute) && m.HasVertexAttribute(modeling.TexCoordAttribute) {
				faceWriter = writeFaceVertAndUvsAndNormals
			} else if m.HasVertexAttribute(modeling.NormalAttribute) {
				faceWriter = writeFaceVertsAndNormals
			} else if m.HasVertexAttribute(modeling.TexCoordAttribute) {
				faceWriter = writeFaceVertsAndUvs
			} else {
				faceWriter = writeFaceVerts
			}

			if matNames != nil {
				writeUsingMaterial(matNames[entry.Material], writer)
				if err := writer.Error(); err != nil {
					return fmt.Errorf("failed to write materials: %w", err)
				}
			}

			indices := m.Indices()
			faceWriter(indices, writer, indexOffset)
			if err := writer.Error(); err != nil {
				return fmt.Errorf("failed to write faces: %w", err)
			}
			indexOffset += m.AttributeLength()
		}
	}

	return nil
}
