package obj

import (
	"fmt"
	"io"

	"github.com/EliCDavis/mesh"
)

func Write(m *mesh.Mesh, out io.Writer) error {
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
