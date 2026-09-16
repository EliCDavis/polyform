package meshops

import (
	"fmt"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/predicate"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector1"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type FillHolesTransformer struct{}

func (fht FillHolesTransformer) Transform(m modeling.Mesh) (modeling.Mesh, error) {
	return FillHoles(m)
}

// The triangulator may merge close points, leaving the patch short an outline edge.
func closesOutline(patch [][3]int, n int) bool {
	if len(patch) != n-2 {
		return false
	}

	indices := make([]int, 0, len(patch)*3)
	for _, t := range patch {
		indices = append(indices, t[0], t[1], t[2])
	}

	boundary := modeling.NewTriangleMesh(indices).BoundaryEdges()
	if len(boundary) != n {
		return false
	}
	for k, edge := range boundary {
		if edge != [2]int{k, (k + n - 1) % n} {
			return false
		}
	}
	return true
}

func triangulateOutline(flat []vector2.Float64) ([][3]int, bool) {
	if len(flat) < 3 {
		return nil, false
	}

	patch, err := triangulation.ConstrainedDelaunay(flat, []triangulation.Constraint{
		triangulation.NewConstraint(flat),
	})
	if err != nil {
		return nil, false
	}

	rimWinding := geometry.Shape(flat).SignedArea()
	indices := patch.Indices()
	out := make([][3]int, 0, indices.Len()/3)

	for i := 0; i+3 <= indices.Len(); i += 3 {
		a, b, c := indices.At(i), indices.At(i+1), indices.At(i+2)
		if a >= len(flat) || b >= len(flat) || c >= len(flat) {
			return nil, false
		}

		if (predicate.Orient2D(flat[a], flat[b], flat[c]) > 0) == (rimWinding > 0) {
			b, c = c, b
		}
		out = append(out, [3]int{a, b, c})
	}

	if !closesOutline(out, len(flat)) {
		return nil, false
	}
	return out, true
}

// FillHoles closes each rim BoundaryLoops finds with triangles across its own
// vertices, or a fan from a new middle vertex when its outline will not
// triangulate. A rim enclosing no area is left open.
func FillHoles(m modeling.Mesh) (modeling.Mesh, error) {
	loops := m.BoundaryLoops()
	if len(loops) == 0 {
		return m, nil
	}

	if err := RequireV3Attribute(m, modeling.PositionAttribute); err != nil {
		return m, err
	}

	indices := iter.ReadFull(m.Indices())
	positions := m.Float3Attribute(modeling.PositionAttribute)

	var v4 map[string][]vector4.Float64
	var v3 map[string][]vector3.Float64
	var v2 map[string][]vector2.Float64
	var v1 map[string][]float64

	for _, loop := range loops {
		rim := make([]vector3.Float64, len(loop))
		for i, v := range loop {
			rim[i] = positions.At(v)
		}

		plane := geometry.NewPlaneFromPolygon(rim)
		if plane.Normal().Length() == 0 {
			continue
		}

		if patch, ok := triangulateOutline(plane.Project(rim)); ok {
			for _, t := range patch {
				indices = append(indices, loop[t[0]], loop[t[1]], loop[t[2]])
			}
			continue
		}

		if v3 == nil {
			v4 = readAllFloat4Data(m)
			v3 = readAllFloat3Data(m)
			v2 = readAllFloat2Data(m)
			v1 = readAllFloat1Data(m)
		}
		middle := len(v3[modeling.PositionAttribute])

		for atr, data := range v4 {
			v4[atr] = append(data, average(data, loop, vector4.Space[float64]{}))
		}
		for atr, data := range v3 {
			value := average(data, loop, vector3.Space[float64]{})
			if atr == modeling.NormalAttribute {
				// The fan winds against the rim, so it faces the other way.
				value = plane.Normal().Scale(-1)
			}
			v3[atr] = append(data, value)
		}
		for atr, data := range v2 {
			v2[atr] = append(data, average(data, loop, vector2.Space[float64]{}))
		}
		for atr, data := range v1 {
			v1[atr] = append(data, average(data, loop, vector1.Space[float64]{}))
		}

		// Wound against the rim so their shared edges pair up.
		for i := range loop {
			indices = append(indices, middle, loop[(i+1)%len(loop)], loop[i])
		}
	}

	if v3 == nil {
		return m.SetIndices(indices), nil
	}

	return modeling.
		NewMesh(m.Topology(), indices).
		SetFloat4Data(v4).
		SetFloat3Data(v3).
		SetFloat2Data(v2).
		SetFloat1Data(v1), nil
}

type FillHolesNode struct {
	Mesh nodes.Output[modeling.Mesh]
}

func (n FillHolesNode) Description() string {
	return "Closes each hole with triangles across its own rim, or a fan from a new middle vertex where the rim will not triangulate. Holes meeting at a pinched vertex and rims enclosing no area are left open."
}

func (n FillHolesNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
	if n.Mesh == nil {
		return
	}

	mesh := nodes.GetOutputValue(out, n.Mesh)
	if mesh.Topology() != modeling.TriangleTopology {
		out.CaptureError(fmt.Errorf("can only fill holes in a triangle mesh"))
		out.Set(mesh)
		return
	}
	filled, err := FillHoles(mesh)
	out.Set(filled)
	if err != nil {
		out.CaptureError(err)
		return
	}

	// Edges, not rims: pinched holes have no rim to count.
	if remaining := len(filled.BoundaryEdges()); remaining > 0 {
		out.CaptureError(fmt.Errorf(
			"%d of %d open edges could not be closed: they belong to holes that meet at a vertex, to an edge shared by more than two faces, or to a rim enclosing no area",
			remaining, len(mesh.BoundaryEdges())))
	}
}
