package stl

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/EliCDavis/polyform/modeling"
)

func Write(out io.Writer, bin Binary) error {
	if _, err := out.Write(bin.Header[:]); err != nil {
		return fmt.Errorf("unable to write header %w", err)
	}

	if err := binary.Write(out, binary.LittleEndian, uint32(len(bin.Triangles))); err != nil {
		return fmt.Errorf("unable to write tri count %w", err)
	}

	if err := binary.Write(out, binary.LittleEndian, bin.Triangles); err != nil {
		return fmt.Errorf("unable to write triangles: %w", err)
	}

	return nil
}

func WriteMesh(out io.Writer, m modeling.Mesh) error {
	if m.Topology() != modeling.TriangleTopology {
		panic(fmt.Errorf("stl format does not supoprt %s topology", m.Topology()))
	}

	if !m.HasFloat3Attribute(modeling.PositionAttribute) {
		return Write(out, Binary{
			Triangles: make([]Triangle, 0),
		})
	}

	count := m.PrimitiveCount()
	tris := make([]Triangle, count)
	for i := 0; i < count; i++ {
		tri := m.Tri(i)
		v1 := tri.P1Vec3Attr(modeling.PositionAttribute).ToFloat32()
		v2 := tri.P2Vec3Attr(modeling.PositionAttribute).ToFloat32()
		v3 := tri.P3Vec3Attr(modeling.PositionAttribute).ToFloat32()

		tris[i] = Triangle{
			Vertex1: Vec{
				X: v1.X(),
				Y: v1.Y(),
				Z: v1.Z(),
			},
			Vertex2: Vec{
				X: v2.X(),
				Y: v2.Y(),
				Z: v2.Z(),
			},
			Vertex3: Vec{
				X: v3.X(),
				Y: v3.Y(),
				Z: v3.Z(),
			},
		}
	}

	if m.HasFloat3Attribute(modeling.NormalAttribute) {
		for i := 0; i < count; i++ {
			tri := m.Tri(i)
			v1 := tri.P1Vec3Attr(modeling.NormalAttribute)
			v2 := tri.P2Vec3Attr(modeling.NormalAttribute)
			v3 := tri.P3Vec3Attr(modeling.NormalAttribute)

			n := v1.Add(v2).Add(v3).
				DivByConstant(3).
				Normalized().
				ToFloat32()

			t := tris[i]
			t.Normal = Vec{
				X: n.X(),
				Y: n.Y(),
				Z: n.Z(),
			}
			tris[i] = t
		}
	}

	return Write(out, Binary{
		Triangles: tris,
	})
}
