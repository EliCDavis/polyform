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

func WriteMaterial(mat modeling.Material, out io.Writer) (err error) {

	writer := txt.NewWriter(out)

	writer.StartEntry()
	writer.String("newmtl ")
	writer.String(strings.Replace(mat.Name, " ", "", -1))
	writer.NewLine()
	if _, err = writer.FinishEntry(); err != nil {
		return fmt.Errorf("failed to write newmtl: %w", err)
	}

	if err = writeMaterialColor("Kd", mat.DiffuseColor, writer); err != nil {
		return err
	}

	if err = writeMaterialColor("Ka", mat.AmbientColor, writer); err != nil {
		return err
	}

	if err = writeMaterialColor("Ks", mat.SpecularColor, writer); err != nil {
		return err
	}

	if err = writeMaterialFloat("Ns", mat.SpecularHighlight, writer); err != nil {
		return err
	}

	if err = writeMaterialFloat("Ni", mat.OpticalDensity, writer); err != nil {
		return err
	}

	if err = writeMaterialFloat("d", 1-mat.Transparency, writer); err != nil {
		return err
	}

	if err = writeMaterialTexture("map_Kd", mat.ColorTextureURI, writer); err != nil {
		return err
	}

	if err = writeMaterialTexture("map_Ks", mat.SpecularTextureURI, writer); err != nil {
		return err
	}

	if err = writeMaterialTexture("map_Bump", mat.NormalTextureURI, writer); err != nil {
		return err
	}

	if err = writeMaterialTexture("norm", mat.NormalTextureURI, writer); err != nil {
		return err
	}

	writer.StartEntry()
	writer.NewLine()
	writer.FinishEntry()

	if err = writer.Error(); err != nil {
		return fmt.Errorf("failed to write out: %w", err)
	}
	return nil
}

func WriteMaterialsFromMesh(m modeling.Mesh, out io.Writer) error {
	return WriteMaterials(m.Materials(), out)
}

func WriteMaterials(ms []modeling.MeshMaterial, out io.Writer) error {
	_, _ = fmt.Fprintln(out, "# Created with github.com/EliCDavis/polyform")

	defaultWritten := false

	written := make(map[*modeling.Material]bool)

	for _, mat := range ms {
		if mat.Material == nil {
			if !defaultWritten {
				if err := WriteMaterial(modeling.DefaultMaterial(), out); err != nil {
					return fmt.Errorf("failed to write default material: %w", err)
				}
				defaultWritten = true
			}
			continue
		}

		if _, ok := written[mat.Material]; ok {
			continue
		}
		if err := WriteMaterial(*mat.Material, out); err != nil {
			return fmt.Errorf("failed to write material %s: %w", mat.Material.Name, err)
		}
		written[mat.Material] = true
	}
	return nil
}

func writeUsingMaterial(mat *modeling.Material, out *txt.Writer) {
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

func writeFaceVerts(tris *iter.ArrayIterator[int], out *txt.Writer, start, end, offset int) {
	shift := 1 + offset
	for triIndex := start; triIndex < end; triIndex += 3 {
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

func writeFaceVertsAndUvs(tris *iter.ArrayIterator[int], out *txt.Writer, start, end, offset int) {
	shift := 1 + offset
	for triIndex := start; triIndex < end; triIndex += 3 {
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

func writeFaceVertsAndNormals(tris *iter.ArrayIterator[int], out *txt.Writer, start, end, offset int) {
	shift := 1 + offset
	for triIndex := start; triIndex < end; triIndex += 3 {
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

func writeFaceVertAndUvsAndNormals(tris *iter.ArrayIterator[int], out *txt.Writer, start, end, offset int) {
	shift := 1 + offset
	for triIndex := start; triIndex < end; triIndex += 3 {
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
	return WriteMeshes([]ObjMesh{{Mesh: m}}, materialFile, out)
}

func WriteMeshes(meshes []ObjMesh, materialFile string, out io.Writer) error {
	if _, err := fmt.Fprintln(out, "# Created with github.com/EliCDavis/polyform"); err != nil {
		return fmt.Errorf("failed to write attribution comment: %w", err)
	}

	if materialFile != "" {
		if _, err := fmt.Fprintf(out, "mtllib %s\no mesh\n", materialFile); err != nil {
			return fmt.Errorf("failed to write matfile 'mesh': %w", err)
		}
	}

	writer := txt.NewWriter(out)

	for _, objMesh := range meshes {
		m := objMesh.Mesh
		if m.HasFloat3Attribute(modeling.PositionAttribute) {
			posData := m.Float3Attribute(modeling.PositionAttribute)
			vtxt := []byte("v ")
			for i := 0; i < posData.Len(); i++ {
				v := posData.At(i)
				writer.StartEntry()
				writer.Append(vtxt)
				writer.Float64(v.X())
				writer.Space()
				writer.Float64(v.Y())
				writer.Space()
				writer.Float64(v.Z())
				writer.NewLine()
				writer.FinishEntry()
			}

			if err := writer.Error(); err != nil {
				return fmt.Errorf("failed to write position attr: %w", err)
			}
		}

		if m.HasFloat2Attribute(modeling.TexCoordAttribute) {
			uvData := m.Float2Attribute(modeling.TexCoordAttribute)
			vt := []byte("vt ")
			for i := 0; i < uvData.Len(); i++ {
				uv := uvData.At(i)
				writer.StartEntry()
				writer.Append(vt)
				writer.Float64(uv.X())
				writer.Space()
				writer.Float64(uv.Y())
				writer.NewLine()
				writer.FinishEntry()
			}
			if err := writer.Error(); err != nil {
				return fmt.Errorf("failed to write UV attr: %w", err)
			}
		}

		if m.HasFloat3Attribute(modeling.NormalAttribute) {
			normalData := m.Float3Attribute(modeling.NormalAttribute)
			vn := []byte("vn ")
			for i := 0; i < normalData.Len(); i++ {
				n := normalData.At(i)
				writer.StartEntry()
				writer.Append(vn)
				writer.Float64(n.X())
				writer.Space()
				writer.Float64(n.Y())
				writer.Space()
				writer.Float64(n.Z())
				writer.NewLine()
				writer.FinishEntry()
			}

			if err := writer.Error(); err != nil {
				return fmt.Errorf("failed to write UV normal attr: %w", err)
			}
		}
	}

	var faceWriter func(tris *iter.ArrayIterator[int], out *txt.Writer, start, end, offset int)

	indexOffset := 0
	for _, objMesh := range meshes {
		if len(meshes) > 1 || objMesh.Name != "" {
			fmt.Fprintf(out, "g %s\n", objMesh.Name)
		}

		m := objMesh.Mesh
		if m.HasVertexAttribute(modeling.NormalAttribute) && m.HasVertexAttribute(modeling.TexCoordAttribute) {
			faceWriter = writeFaceVertAndUvsAndNormals
		} else if m.HasVertexAttribute(modeling.NormalAttribute) {
			faceWriter = writeFaceVertsAndNormals
		} else if m.HasVertexAttribute(modeling.TexCoordAttribute) {
			faceWriter = writeFaceVertsAndUvs
		} else {
			faceWriter = writeFaceVerts
		}

		mats := m.Materials()
		indices := m.Indices()
		if len(mats) == 0 {
			faceWriter(indices, writer, 0, indices.Len(), indexOffset)
			if err := writer.Error(); err != nil {
				return fmt.Errorf("failed to write faces: %w", err)
			}
		} else {
			offset := 0
			for _, mat := range mats {
				writeUsingMaterial(mat.Material, writer)
				if err := writer.Error(); err != nil {
					return fmt.Errorf("failed to write materials: %w", err)
				}

				nextOffset := offset + (mat.PrimitiveCount * 3)
				faceWriter(indices, writer, offset, nextOffset, indexOffset)
				if err := writer.Error(); err != nil {
					return fmt.Errorf("failed to write faces: %w", err)
				}

				offset = nextOffset
			}
		}
		indexOffset += m.AttributeLength()
	}

	return nil
}
