package obj

import (
	"fmt"
	"image/color"
	"io"
	"strings"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/formats/txt"
	"github.com/EliCDavis/polyform/modeling"
)

func writeMaterialColor(colorType string, color color.Color, writer *txt.Writer) error {
	if color == nil {
		return nil
	}
	writer.StartEntry()
	writer.String(colorType)
	writer.Space()

	r, g, b, _ := color.RGBA()
	writer.Float64MaxFigs(float64(r)/0xffff, 3)
	writer.Space()
	writer.Float64MaxFigs(float64(g)/0xffff, 3)
	writer.Space()
	writer.Float64MaxFigs(float64(b)/0xffff, 3)

	writer.NewLine()
	if _, err := writer.FinishEntry(); err != nil {
		return fmt.Errorf("failed to write %s: %w", colorType, err)
	}
	return nil
}

func writeMaterialFloat(floatType string, f float64, writer *txt.Writer) error {
	writer.StartEntry()
	writer.String(floatType)
	writer.Space()
	writer.Float64(f)
	writer.NewLine()

	if _, err := writer.FinishEntry(); err != nil {
		return fmt.Errorf("failed to write %s: %w", floatType, err)
	}
	return nil
}

func writeMaterialTexture(texType string, tex *string, writer *txt.Writer) error {
	if tex == nil {
		return nil
	}

	writer.StartEntry()

	writer.String(texType)
	writer.Space()
	writer.String(*tex)
	writer.NewLine()

	if _, err := writer.FinishEntry(); err != nil {
		return fmt.Errorf("failed to write %s: %w", texType, err)
	}
	return nil
}

func WriteMaterials(scene Scene, out io.Writer) error {
	_, _ = fmt.Fprintln(out, "# Created with github.com/EliCDavis/polyform")

	defaultWritten := false
	written := make(map[*Material]struct{})

	for _, o := range scene.Objects {
		for _, e := range o.Entries {
			if e.Material == nil {
				if !defaultWritten {
					if err := DefaultMaterial().write(out); err != nil {
						return fmt.Errorf("failed to write default material: %w", err)
					}
					defaultWritten = true
				}
				continue
			}

			if _, ok := written[e.Material]; ok {
				continue
			}

			if err := e.Material.write(out); err != nil {
				return fmt.Errorf("failed to write material %q on object %q: %w", e.Material.Name, o.Name, err)
			}

			written[e.Material] = struct{}{}
		}
	}

	return nil
}

func writeUsingMaterial(mat *Material, out *txt.Writer) {
	if mat == nil {
		out.StartEntry()
		out.String("usemtl DefaultDiffuse\n")
		out.FinishEntry()
	} else {
		out.StartEntry()
		out.String("usemtl ")
		out.String(strings.Replace(mat.Name, " ", "", -1))
		out.NewLine()
		out.FinishEntry()
	}
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
	return WriteScene(
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

func WriteScene(scene Scene, materialFile string, out io.Writer) error {
	if _, err := fmt.Fprintln(out, "# Created with github.com/EliCDavis/polyform"); err != nil {
		return fmt.Errorf("failed to write attribution comment: %w", err)
	}

	if materialFile != "" {
		if _, err := fmt.Fprintf(out, "mtllib %s\no mesh\n", materialFile); err != nil {
			return fmt.Errorf("failed to write matfile 'mesh': %w", err)
		}
	}

	writer := txt.NewWriter(out)

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

			if entry.Material != nil {
				mat := entry.Material
				writeUsingMaterial(mat, writer)
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
