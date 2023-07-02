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

func WriteMaterial(mat modeling.Material, out io.Writer) error {
	_, err := fmt.Fprintf(out, "newmtl %s\n", strings.Replace(mat.Name, " ", "", -1))
	if err != nil {
		return err
	}

	if mat.DiffuseColor != nil {
		_, err = fmt.Fprintf(out, "Kd %s\n", ColorString(mat.DiffuseColor))
		if err != nil {
			return err
		}
	}

	if mat.AmbientColor != nil {
		_, err = fmt.Fprintf(out, "Ka %s\n", ColorString(mat.AmbientColor))
		if err != nil {
			return err
		}
	}

	if mat.SpecularColor != nil {
		_, err = fmt.Fprintf(out, "Ks %s\n", ColorString(mat.SpecularColor))
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(out, "Ns %f\n", mat.SpecularHighlight)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "Ni %f\n", mat.OpticalDensity)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "d %f\n", 1-mat.Transparency)
	if err != nil {
		return err
	}

	if mat.ColorTextureURI != nil {
		_, err = fmt.Fprintf(out, "map_Kd %s\n", *mat.ColorTextureURI)
		if err != nil {
			return err
		}
	}

	if mat.NormalTextureURI != nil {
		_, err = fmt.Fprintf(out, "map_Bump %s\n", *mat.NormalTextureURI)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(out, "norm %s\n", *mat.NormalTextureURI)
		if err != nil {
			return err
		}
	}

	if mat.SpecularTextureURI != nil {
		_, err = fmt.Fprintf(out, "map_Ks %s\n", *mat.SpecularTextureURI)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(out, "")
	return err
}

func WriteMaterials(m modeling.Mesh, out io.Writer) error {
	fmt.Fprintln(out, "# Created with github.com/EliCDavis/polyform")

	defaultWritten := false

	written := make(map[*modeling.Material]bool)

	for _, mat := range m.Materials() {

		if mat.Material == nil {
			if !defaultWritten {
				err := WriteMaterial(modeling.DefaultMaterial(), out)
				if err != nil {
					return err
				}
				defaultWritten = true
			}
			continue
		}

		_, ok := written[mat.Material]
		if ok {
			continue
		}
		err := WriteMaterial(*mat.Material, out)
		if err != nil {
			return err
		}
		written[mat.Material] = true
	}
	return nil
}

func writeUsingMaterial(mat *modeling.Material, out io.Writer) {
	if mat == nil {
		fmt.Fprint(out, "usemtl DefaultDiffuse\n")
	} else {
		fmt.Fprintf(out, "usemtl %s\n", strings.Replace(mat.Name, " ", "", -1))
	}
}

func writeFaceVerts(tris iter.ArrayIterator[int], out io.Writer, start, end, offset int) error {
	for triIndex := start; triIndex < end; triIndex += 3 {
		p1 := tris.At(triIndex) + 1 + offset
		p2 := tris.At(triIndex+1) + 1 + offset
		p3 := tris.At(triIndex+2) + 1 + offset
		_, err := fmt.Fprintf(out, "f %d %d %d\n", p1, p2, p3)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFaceVertsAndUvs(tris iter.ArrayIterator[int], out io.Writer, start, end, offset int) error {
	for triIndex := start; triIndex < end; triIndex += 3 {
		p1 := tris.At(triIndex) + 1 + offset
		p2 := tris.At(triIndex+1) + 1 + offset
		p3 := tris.At(triIndex+2) + 1 + offset
		_, err := fmt.Fprintf(out, "f %d/%d %d/%d %d/%d\n", p1, p1, p2, p2, p3, p3)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFaceVertsAndNormals(tris iter.ArrayIterator[int], out io.Writer, start, end, offset int) error {
	for triIndex := start; triIndex < end; triIndex += 3 {
		p1 := tris.At(triIndex) + 1 + offset
		p2 := tris.At(triIndex+1) + 1 + offset
		p3 := tris.At(triIndex+2) + 1 + offset
		_, err := fmt.Fprintf(out, "f %d//%d %d//%d %d//%d\n", p1, p1, p2, p2, p3, p3)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFaceVertAndUvsAndNormals(tris iter.ArrayIterator[int], out io.Writer, start, end, offset int) error {
	for triIndex := start; triIndex < end; triIndex += 3 {
		p1 := tris.At(triIndex) + 1 + offset
		p2 := tris.At(triIndex+1) + 1 + offset
		p3 := tris.At(triIndex+2) + 1 + offset
		_, err := fmt.Fprintf(out, "f %d/%d/%d %d/%d/%d %d/%d/%d\n", p1, p1, p1, p2, p2, p2, p3, p3, p3)
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteMesh(m modeling.Mesh, materialFile string, out io.Writer) error {
	return WriteMeshes([]ObjMesh{{Mesh: m}}, materialFile, out)
}

func WriteMeshes(meshes []ObjMesh, materialFile string, out io.Writer) error {
	fmt.Fprintln(out, "# Created with github.com/EliCDavis/polyform")
	if materialFile != "" {
		fmt.Fprintf(out, "mtllib %s\no mesh\n", materialFile)
	}

	for _, objMesh := range meshes {
		m := objMesh.Mesh
		if m.HasFloat3Attribute(modeling.PositionAttribute) {
			posData := m.Float3Attribute(modeling.PositionAttribute)
			for i := 0; i < posData.Len(); i++ {
				v := posData.At(i)
				_, err := fmt.Fprintf(out, "v %f %f %f\n", v.X(), v.Y(), v.Z())
				if err != nil {
					return err
				}
			}
		}

		if m.HasFloat2Attribute(modeling.TexCoordAttribute) {
			uvData := m.Float2Attribute(modeling.TexCoordAttribute)
			for i := 0; i < uvData.Len(); i++ {
				uv := uvData.At(i)
				_, err := fmt.Fprintf(out, "vt %f %f\n", uv.X(), uv.Y())
				if err != nil {
					return err
				}
			}
		}

		if m.HasFloat3Attribute(modeling.NormalAttribute) {
			normalData := m.Float3Attribute(modeling.NormalAttribute)
			for i := 0; i < normalData.Len(); i++ {
				n := normalData.At(i)
				_, err := fmt.Fprintf(out, "vn %f %f %f\n", n.X(), n.Y(), n.Z())
				if err != nil {
					return err
				}
			}
		}
	}

	var faceWriter func(tris iter.ArrayIterator[int], out io.Writer, start, end, offset int) error

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
			err := faceWriter(indices, out, 0, indices.Len(), indexOffset)
			if err != nil {
				return err
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
