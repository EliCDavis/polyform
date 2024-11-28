package stl

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

func Read(in io.Reader) (*Binary, error) {

	header := new(Header)
	if err := binary.Read(in, binary.LittleEndian, header); err != nil {
		return nil, fmt.Errorf("unable to read header %w", err)
	}

	var triCount uint32
	if err := binary.Read(in, binary.LittleEndian, &triCount); err != nil {
		return nil, fmt.Errorf("unable to read tri count: %w", err)
	}

	tris := make([]Triangle, triCount)
	if err := binary.Read(in, binary.LittleEndian, &tris); err != nil {
		return nil, fmt.Errorf("unable to read tris: %w", err)
	}

	return &Binary{
		Header:    *header,
		Triangles: tris,
	}, nil
}

func ReadMesh(in io.Reader) (*modeling.Mesh, error) {
	bin, err := Read(in)
	if err != nil {
		return nil, err
	}

	if len(bin.Triangles) == 0 {
		empty := modeling.EmptyMesh(modeling.TriangleTopology)
		return &empty, nil
	}

	indices := make([]int, len(bin.Triangles)*3)
	position := make([]vector3.Float64, len(bin.Triangles)*3)
	normals := make([]vector3.Float64, len(bin.Triangles)*3)
	normalExists := false

	for i, tri := range bin.Triangles {
		start := i * 3
		indices[start] = start
		indices[start+1] = start + 1
		indices[start+2] = start + 2

		position[start] = tri.Vertex1.Float64()
		position[start+1] = tri.Vertex2.Float64()
		position[start+2] = tri.Vertex3.Float64()

		var normal vector3.Float64
		if tri.Normal.Zero() {
			// Calculate a flat normal
			normal = tri.Vertex2.Float64().Sub(tri.Vertex1.Float64()).
				Cross(tri.Vertex3.Float64().Sub(tri.Vertex1.Float64())).
				Normalized()
		} else {
			normalExists = true
			normal = tri.Normal.Float64()
		}

		normals[start] = normal
		normals[start+1] = normal
		normals[start+2] = normal
	}

	mesh := modeling.NewTriangleMesh(indices).
		SetFloat3Attribute(modeling.PositionAttribute, position)

	if normalExists {
		mesh = mesh.SetFloat3Attribute(modeling.NormalAttribute, normals)
	}

	return &mesh, nil
}
