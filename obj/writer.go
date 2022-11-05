package obj

import (
	"fmt"
	"image/color"
	"io"
	"strings"

	"github.com/EliCDavis/mesh"
)

func ColorString(color color.Color) string {
	r, g, b, _ := color.RGBA()
	return fmt.Sprintf("%f %f %f", float64(r)/0xffff, float64(g)/0xffff, float64(b)/0xffff)
}

func WriteMaterial(mat mesh.Material, out io.Writer) {
	fmt.Fprintf(out, "newmtl %s\n", strings.Replace(mat.Name, " ", "", -1))

	fmt.Fprintf(out, "Kd %s\n", ColorString(mat.DiffuseColor))
	fmt.Fprintf(out, "Ka %s\n", ColorString(mat.AmbientColor))
	fmt.Fprintf(out, "Ks %s\n", ColorString(mat.SpecularColor))

	fmt.Fprintf(out, "Ns %f\n", mat.SpecularHighlight)
	fmt.Fprintf(out, "Ni %f\n", mat.OpticalDensity)
	fmt.Fprintf(out, "d %f\n", mat.Dissolve)

	if mat.ColorTextureURI != nil {
		fmt.Fprintf(out, "map_Kd %s\n", *mat.ColorTextureURI)
	}

	fmt.Fprintln(out, "")
}

func WriteMaterials(m *mesh.Mesh, out io.Writer) {
	defaultWritten := false

	written := make(map[mesh.Material]bool)

	for _, mat := range m.Materials() {

		if mat.Material == nil {
			if !defaultWritten {
				WriteMaterial(mesh.DefaultMaterial(), out)
				defaultWritten = true
			}
			continue
		}

		_, ok := written[*mat.Material]
		if ok {
			continue
		}
		WriteMaterial(*mat.Material, out)
		written[*mat.Material] = true
	}
}

func writeUsingMaterial(mat *mesh.Material, out io.Writer) {
	if mat == nil {
		fmt.Fprint(out, "usemtl DefaultDiffuse\n")
	} else {
		fmt.Fprintf(out, "usemtl %s\n", strings.Replace(mat.Name, " ", "", -1))
	}
}

func WriteMesh(m *mesh.Mesh, materialFile string, out io.Writer) error {
	if materialFile != "" {
		fmt.Fprintf(out, "mtllib %s\no mesh\n", materialFile)
	}

	view := m.View()
	for _, v := range view.Vertices {
		_, err := fmt.Fprintf(out, "v %f %f %f\n", v.X(), v.Y(), v.Z())
		if err != nil {
			return err
		}
	}

	for _, uvChannel := range view.UVs {
		for _, uv := range uvChannel {
			_, err := fmt.Fprintf(out, "vt %f %f\n", uv.X(), uv.Y())
			if err != nil {
				return err
			}
		}
	}

	for _, n := range view.Normals {
		_, err := fmt.Fprintf(out, "vn %f %f %f\n", n.X(), n.Y(), n.Z())
		if err != nil {
			return err
		}
	}

	mats := m.Materials()
	matIndex := 0
	matOffset := 0
	if len(mats) > 0 {
		writeUsingMaterial(mats[0].Material, out)
	}

	if len(view.Normals) > 0 && len(view.UVs) > 0 && len(view.UVs[0]) > 0 {
		for triIndex := 0; triIndex < len(view.Triangles); triIndex += 3 {
			p1 := view.Triangles[triIndex] + 1
			p2 := view.Triangles[triIndex+1] + 1
			p3 := view.Triangles[triIndex+2] + 1
			_, err := fmt.Fprintf(out, "f %d/%d/%d %d/%d/%d %d/%d/%d\n", p1, p1, p1, p2, p2, p2, p3, p3, p3)
			if err != nil {
				return err
			}
		}
	} else if len(view.Normals) > 0 {
		for triIndex := 0; triIndex < len(view.Triangles); triIndex += 3 {
			if len(mats) > 0 && triIndex/3 == mats[matIndex].NumOfTris+matOffset {
				matOffset = triIndex / 3
				matIndex++
				writeUsingMaterial(mats[matIndex].Material, out)
			}

			p1 := view.Triangles[triIndex] + 1
			p2 := view.Triangles[triIndex+1] + 1
			p3 := view.Triangles[triIndex+2] + 1
			_, err := fmt.Fprintf(out, "f %d//%d %d//%d %d//%d\n", p1, p1, p2, p2, p3, p3)
			if err != nil {
				return err
			}
		}
	} else if len(view.UVs) > 0 {
		for triIndex := 0; triIndex < len(view.Triangles); triIndex += 3 {
			p1 := view.Triangles[triIndex] + 1
			p2 := view.Triangles[triIndex+1] + 1
			p3 := view.Triangles[triIndex+2] + 1
			_, err := fmt.Fprintf(out, "f %d/%d %d/%d %d/%d\n", p1, p1, p2, p2, p3, p3)
			if err != nil {
				return err
			}
		}
	} else {
		for triIndex := 0; triIndex < len(view.Triangles); triIndex += 3 {
			p1 := view.Triangles[triIndex] + 1
			p2 := view.Triangles[triIndex+1] + 1
			p3 := view.Triangles[triIndex+2] + 1
			_, err := fmt.Fprintf(out, "f %d %d %d\n", p1, p2, p3)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
