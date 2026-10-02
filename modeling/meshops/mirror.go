package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

// Reflection reverses orientation, so the winding is reversed to match.
func Mirror(m modeling.Mesh, origin, normal vector3.Float64) modeling.Mesh {
	if normal.Length() == 0 {
		return m
	}
	n := normal.Normalized()

	mirrored := m
	if m.HasFloat3Attribute(modeling.PositionAttribute) {
		positions := m.Float3Attribute(modeling.PositionAttribute)
		reflected := make([]vector3.Float64, positions.Len())
		for i := range reflected {
			p := positions.At(i)
			reflected[i] = p.Sub(n.Scale(2 * p.Sub(origin).Dot(n)))
		}
		mirrored = mirrored.SetFloat3Attribute(modeling.PositionAttribute, reflected)
	}

	// A direction reflects about the plane itself, without the origin.
	if m.HasFloat3Attribute(modeling.NormalAttribute) {
		normals := m.Float3Attribute(modeling.NormalAttribute)
		reflected := make([]vector3.Float64, normals.Len())
		for i := range reflected {
			v := normals.At(i)
			reflected[i] = v.Sub(n.Scale(2 * v.Dot(n)))
		}
		mirrored = mirrored.SetFloat3Attribute(modeling.NormalAttribute, reflected)
	}

	return ReverseWinding(mirrored)
}

// Vertex order only: normal attributes are left as they are.
func ReverseWinding(m modeling.Mesh) modeling.Mesh {
	if m.Topology() != modeling.TriangleTopology {
		return m
	}

	tris := m.Indices()
	rewound := make([]int, tris.Len())
	for i := 0; i < tris.Len(); i += 3 {
		rewound[i] = tris.At(i + 1)
		rewound[i+1] = tris.At(i)
		rewound[i+2] = tris.At(i + 2)
	}
	return m.SetIndices(rewound)
}

type MirrorNode struct {
	Mesh   nodes.Output[modeling.Mesh]
	Origin nodes.Output[vector3.Float64] `description:"A point on the mirror plane. Defaults to the world origin."`
	Normal nodes.Output[vector3.Float64] `description:"The plane's normal. Defaults to +X."`
	Union  nodes.Output[bool]            `description:"When true (default), appends the reflection to the original so both sides are present. When false, this is a pure reflection and the original is dropped."`
}

func (n MirrorNode) Description() string {
	return "Reflects a mesh across a plane, reversing the winding so it still faces outward."
}

func (n MirrorNode) Keywords() []string {
	return []string{"reflect", "symmetry"}
}

func (n MirrorNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
	if n.Mesh == nil {
		return
	}

	mesh := nodes.GetOutputValue(out, n.Mesh)
	mirrored := Mirror(
		mesh,
		nodes.TryGetOutputValue(out, n.Origin, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, n.Normal, vector3.Right[float64]()),
	)

	if nodes.TryGetOutputValue(out, n.Union, true) {
		mirrored = mesh.Append(mirrored)
	}
	out.Set(mirrored)
}
