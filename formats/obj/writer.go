package obj

import (
	"fmt"
	"image/color"
	"io"
	"strings"

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

func writeFaceVerts(tris []int, out io.Writer, start, end int) error {
	for triIndex := start; triIndex < end; triIndex += 3 {
		p1 := tris[triIndex] + 1
		p2 := tris[triIndex+1] + 1
		p3 := tris[triIndex+2] + 1
		_, err := fmt.Fprintf(out, "f %d %d %d\n", p1, p2, p3)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFaceVertsAndUvs(tris []int, out io.Writer, start, end int) error {
	for triIndex := start; triIndex < end; triIndex += 3 {
		p1 := tris[triIndex] + 1
		p2 := tris[triIndex+1] + 1
		p3 := tris[triIndex+2] + 1
		_, err := fmt.Fprintf(out, "f %d/%d %d/%d %d/%d\n", p1, p1, p2, p2, p3, p3)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFaceVertsAndNormals(tris []int, out io.Writer, start, end int) error {
	for triIndex := start; triIndex < end; triIndex += 3 {
		p1 := tris[triIndex] + 1
		p2 := tris[triIndex+1] + 1
		p3 := tris[triIndex+2] + 1
		_, err := fmt.Fprintf(out, "f %d//%d %d//%d %d//%d\n", p1, p1, p2, p2, p3, p3)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFaceVertAndUvsAndNormals(tris []int, out io.Writer, start, end int) error {
	for triIndex := start; triIndex < end; triIndex += 3 {
		p1 := tris[triIndex] + 1
		p2 := tris[triIndex+1] + 1
		p3 := tris[triIndex+2] + 1
		_, err := fmt.Fprintf(out, "f %d/%d/%d %d/%d/%d %d/%d/%d\n", p1, p1, p1, p2, p2, p2, p3, p3, p3)
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteMesh(m modeling.Mesh, materialFile string, out io.Writer) error {
	fmt.Fprintln(out, "# Created with github.com/EliCDavis/polyform")
	if materialFile != "" {
		fmt.Fprintf(out, "mtllib %s\no mesh\n", materialFile)
	}

	view := m.View()
	if vals, ok := view.Float3Data[modeling.PositionAttribute]; ok {
		for _, v := range vals {
			_, err := fmt.Fprintf(out, "v %f %f %f\n", v.X(), v.Y(), v.Z())
			if err != nil {
				return err
			}
		}
	}

	if vals, ok := view.Float2Data[modeling.TexCoordAttribute]; ok {
		for _, uv := range vals {
			_, err := fmt.Fprintf(out, "vt %f %f\n", uv.X(), uv.Y())
			if err != nil {
				return err
			}
		}
	}

	if vals, ok := view.Float3Data[modeling.NormalAttribute]; ok {
		for _, n := range vals {
			_, err := fmt.Fprintf(out, "vn %f %f %f\n", n.X(), n.Y(), n.Z())
			if err != nil {
				return err
			}
		}
	}

	var faceWriter func(tris []int, out io.Writer, start, end int) error

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
	if len(mats) == 0 {
		err := faceWriter(view.Indices, out, 0, len(view.Indices))
		if err != nil {
			return err
		}
	} else {
		offset := 0
		for _, mat := range mats {
			writeUsingMaterial(mat.Material, out)
			nextOffset := offset + (mat.PrimitiveCount * 3)
			faceWriter(view.Indices, out, offset, nextOffset)
			offset = nextOffset
		}
	}
	return nil
}
