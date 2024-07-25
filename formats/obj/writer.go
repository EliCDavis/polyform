package obj

import (
	"fmt"
	"image/color"
	"io"
	"strings"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling"
)

func ColorString(color color.Color) string {
	r, g, b, _ := color.RGBA()
	return fmt.Sprintf("%f %f %f", float64(r)/0xffff, float64(g)/0xffff, float64(b)/0xffff)
}

func WriteMaterial(mat modeling.Material, out io.Writer) (err error) {
	if _, err = fmt.Fprintf(out, "newmtl %s\n", strings.Replace(mat.Name, " ", "", -1)); err != nil {
		return fmt.Errorf("failed to write newmtl: %w", err)
	}

	if mat.DiffuseColor != nil {
		if _, err = fmt.Fprintf(out, "Kd %s\n", ColorString(mat.DiffuseColor)); err != nil {
			return fmt.Errorf("failed to write Kd: %w", err)
		}
	}

	if mat.AmbientColor != nil {
		if _, err = fmt.Fprintf(out, "Ka %s\n", ColorString(mat.AmbientColor)); err != nil {
			return fmt.Errorf("failed to write Ka: %w", err)
		}
	}

	if mat.SpecularColor != nil {
		if _, err = fmt.Fprintf(out, "Ks %s\n", ColorString(mat.SpecularColor)); err != nil {
			return fmt.Errorf("failed to write Ks: %w", err)
		}
	}

	if _, err = fmt.Fprintf(out, "Ns %f\n", mat.SpecularHighlight); err != nil {
		return fmt.Errorf("failed to write Ns: %w", err)
	}

	if _, err = fmt.Fprintf(out, "Ni %f\n", mat.OpticalDensity); err != nil {
		return fmt.Errorf("failed to write Ni: %w", err)
	}

	if _, err = fmt.Fprintf(out, "d %f\n", 1-mat.Transparency); err != nil {
		return fmt.Errorf("failed to write d: %w", err)
	}

	if mat.ColorTextureURI != nil {
		if _, err = fmt.Fprintf(out, "map_Kd %s\n", *mat.ColorTextureURI); err != nil {
			return fmt.Errorf("failed to write map_Kd: %w", err)
		}
	}

	if mat.NormalTextureURI != nil {
		if _, err = fmt.Fprintf(out, "map_Bump %s\n", *mat.NormalTextureURI); err != nil {
			return fmt.Errorf("failed to write map_Bump: %w", err)
		}

		if _, err = fmt.Fprintf(out, "norm %s\n", *mat.NormalTextureURI); err != nil {
			return fmt.Errorf("failed to write norm: %w", err)
		}
	}

	if mat.SpecularTextureURI != nil {
		if _, err = fmt.Fprintf(out, "map_Ks %s\n", *mat.SpecularTextureURI); err != nil {
			return fmt.Errorf("failed to write map_Ks: %w", err)
		}
	}

	if _, err = fmt.Fprintln(out, ""); err != nil {
		return fmt.Errorf("failed to write out: %w", err)
	}
	return nil
}

func WriteMaterialsFromMesh(m modeling.Mesh, out io.Writer) error {
	fmt.Fprintln(out, "# Created with github.com/EliCDavis/polyform")

	defaultWritten := false

	written := make(map[*modeling.Material]bool)

	for _, mat := range m.Materials() {
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

func WriteMaterials(ms []modeling.MeshMaterial, out io.Writer) error {
	fmt.Fprintln(out, "# Created with github.com/EliCDavis/polyform")

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

func writeUsingMaterial(mat *modeling.Material, out io.Writer) {
	if mat == nil {
		_, _ = fmt.Fprint(out, "usemtl DefaultDiffuse\n")
	} else {
		_, _ = fmt.Fprintf(out, "usemtl %s\n", strings.Replace(mat.Name, " ", "", -1))
	}
}

func writeFaceVerts(tris *iter.ArrayIterator[int], out io.Writer, start, end, offset int) error {
	for triIndex := start; triIndex < end; triIndex += 3 {
		p1 := tris.At(triIndex) + 1 + offset
		p2 := tris.At(triIndex+1) + 1 + offset
		p3 := tris.At(triIndex+2) + 1 + offset

		if _, err := fmt.Fprintf(out, "f %d %d %d\n", p1, p2, p3); err != nil {
			return fmt.Errorf("failed to write face verts: %w", err)
		}
	}
	return nil
}

func writeFaceVertsAndUvs(tris *iter.ArrayIterator[int], out io.Writer, start, end, offset int) error {
	for triIndex := start; triIndex < end; triIndex += 3 {
		p1 := tris.At(triIndex) + 1 + offset
		p2 := tris.At(triIndex+1) + 1 + offset
		p3 := tris.At(triIndex+2) + 1 + offset
		if _, err := fmt.Fprintf(out, "f %d/%d %d/%d %d/%d\n", p1, p1, p2, p2, p3, p3); err != nil {
			return fmt.Errorf("failed to write face verts and UVs: %w", err)
		}
	}
	return nil
}

func writeFaceVertsAndNormals(tris *iter.ArrayIterator[int], out io.Writer, start, end, offset int) error {
	for triIndex := start; triIndex < end; triIndex += 3 {
		p1 := tris.At(triIndex) + 1 + offset
		p2 := tris.At(triIndex+1) + 1 + offset
		p3 := tris.At(triIndex+2) + 1 + offset
		if _, err := fmt.Fprintf(out, "f %d//%d %d//%d %d//%d\n", p1, p1, p2, p2, p3, p3); err != nil {
			return fmt.Errorf("failed to write face verts and normals: %w", err)
		}
	}
	return nil
}

func writeFaceVertAndUvsAndNormals(tris *iter.ArrayIterator[int], out io.Writer, start, end, offset int) error {
	for triIndex := start; triIndex < end; triIndex += 3 {
		p1 := tris.At(triIndex) + 1 + offset
		p2 := tris.At(triIndex+1) + 1 + offset
		p3 := tris.At(triIndex+2) + 1 + offset
		_, err := fmt.Fprintf(out, "f %d/%d/%d %d/%d/%d %d/%d/%d\n", p1, p1, p1, p2, p2, p2, p3, p3, p3)
		if err != nil {
			return fmt.Errorf("failed to write face verts, UVs and normals: %w", err)
		}
	}
	return nil
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

	for _, objMesh := range meshes {
		m := objMesh.Mesh
		if m.HasFloat3Attribute(modeling.PositionAttribute) {
			posData := m.Float3Attribute(modeling.PositionAttribute)
			for i := 0; i < posData.Len(); i++ {
				v := posData.At(i)
				if _, err := fmt.Fprintf(out, "v %f %f %f\n", v.X(), v.Y(), v.Z()); err != nil {
					return fmt.Errorf("failed to write position attr: %w", err)
				}
			}
		}

		if m.HasFloat2Attribute(modeling.TexCoordAttribute) {
			uvData := m.Float2Attribute(modeling.TexCoordAttribute)
			for i := 0; i < uvData.Len(); i++ {
				uv := uvData.At(i)
				if _, err := fmt.Fprintf(out, "vt %f %f\n", uv.X(), uv.Y()); err != nil {
					return fmt.Errorf("failed to write UV attr: %w", err)
				}
			}
		}

		if m.HasFloat3Attribute(modeling.NormalAttribute) {
			normalData := m.Float3Attribute(modeling.NormalAttribute)
			for i := 0; i < normalData.Len(); i++ {
				n := normalData.At(i)
				if _, err := fmt.Fprintf(out, "vn %f %f %f\n", n.X(), n.Y(), n.Z()); err != nil {
					return fmt.Errorf("failed to write UV normal attr: %w", err)
				}
			}
		}
	}

	var faceWriter func(tris *iter.ArrayIterator[int], out io.Writer, start, end, offset int) error

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
			if err := faceWriter(indices, out, 0, indices.Len(), indexOffset); err != nil {
				return fmt.Errorf("failed to call faceWriter: %w", err)
			}
		} else {
			offset := 0
			for _, mat := range mats {
				writeUsingMaterial(mat.Material, out)
				nextOffset := offset + (mat.PrimitiveCount * 3)
				faceWriter(indices, out, offset, nextOffset, indexOffset)
				offset = nextOffset
			}
		}
		indexOffset += m.AttributeLength()
	}

	return nil
}
