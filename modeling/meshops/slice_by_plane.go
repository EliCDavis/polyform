package meshops

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type SliceByPlaneTransformerSide int

const (
	AbovePlane SliceByPlaneTransformerSide = iota
	BelowPlane
)

type SliceByPlaneTransformer struct {
	Attribute   string
	SliceToKeep SliceByPlaneTransformerSide
	Plane       geometry.Plane
}

func (sbpt SliceByPlaneTransformer) attribute() string {
	return sbpt.Attribute
}

func (sbpt SliceByPlaneTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(sbpt, modeling.PositionAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	above, below := SliceByPlaneWithAttribute(m, sbpt.Plane, attribute)
	if sbpt.SliceToKeep == AbovePlane {
		return above, nil
	}
	return below, nil
}

func SliceByPlaneWithAttribute(m modeling.Mesh, plane geometry.Plane, attribute string) (modeling.Mesh, modeling.Mesh) {
	return sliceTrianglesByPlaneWithAttribute(m, plane, attribute)
}

type SliceAttributeByPlaneNode struct {
	Mesh      nodes.Output[modeling.Mesh]
	Plane     nodes.Output[geometry.Plane]
	Attribute nodes.Output[string] `description:"The vector3 attribute measured against the plane. Defaults to Position."`
}

func (n SliceAttributeByPlaneNode) Description() string {
	return "Cuts the mesh with an infinite plane. Above Plane is the side the plane's normal points to."
}

func (n SliceAttributeByPlaneNode) slice(out *nodes.StructOutput[modeling.Mesh], above bool) {
	if n.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	mesh := nodes.GetOutputValue(out, n.Mesh)
	if n.Plane == nil {
		out.Set(mesh)
		return
	}
	plane := nodes.GetOutputValue(out, n.Plane)

	aboveM, belowM := SliceByPlaneWithAttribute(
		mesh,
		plane,
		nodes.TryGetOutputValue(out, n.Attribute, modeling.PositionAttribute),
	)
	if above {
		out.Set(aboveM)
	} else {
		out.Set(belowM)
	}
}

func (n SliceAttributeByPlaneNode) AbovePlane(out *nodes.StructOutput[modeling.Mesh]) {
	n.slice(out, true)
}

func (n SliceAttributeByPlaneNode) BelowPlane(out *nodes.StructOutput[modeling.Mesh]) {
	n.slice(out, false)
}

type planeCut struct {
	from, to int
	t        float64
}

type planeCuts struct {
	distances []float64
	byEdge    map[[2]int]int
	cuts      []planeCut
}

func (c *planeCuts) vertex(a, b int) int {
	from, to := min(a, b), max(a, b)
	key := [2]int{from, to}
	if index, ok := c.byEdge[key]; ok {
		return index
	}
	index := len(c.distances) + len(c.cuts)
	c.byEdge[key] = index
	c.cuts = append(c.cuts, planeCut{
		from: from,
		to:   to,
		t:    c.distances[from] / (c.distances[from] - c.distances[to]),
	})
	return index
}

func (c *planeCuts) clip(indices []int, tri [3]int, side float64) []int {
	polygon := make([]int, 0, 4)
	for k := range 3 {
		a, b := tri[k], tri[(k+1)%3]
		da, db := side*c.distances[a], side*c.distances[b]
		if da >= 0 {
			polygon = append(polygon, a)
		}
		if (da > 0 && db < 0) || (da < 0 && db > 0) {
			polygon = append(polygon, c.vertex(a, b))
		}
	}
	for j := 1; j+1 < len(polygon); j++ {
		indices = append(indices, polygon[0], polygon[j], polygon[j+1])
	}
	return indices
}

func appendCuts[T any](data []T, cuts []planeCut, lerp func(a, b T, t float64) T) []T {
	for _, cut := range cuts {
		data = append(data, lerp(data[cut.from], data[cut.to], cut.t))
	}
	return data
}

func (c *planeCuts) extend(m modeling.Mesh) modeling.Mesh {
	for attr, data := range readAllFloat4Data(m) {
		m = m.SetFloat4Attribute(attr, appendCuts(data, c.cuts, func(a, b vector4.Float64, t float64) vector4.Float64 {
			return a.Scale(1 - t).Add(b.Scale(t))
		}))
	}
	for attr, data := range readAllFloat3Data(m) {
		data = appendCuts(data, c.cuts, func(a, b vector3.Float64, t float64) vector3.Float64 {
			return a.Scale(1 - t).Add(b.Scale(t))
		})
		if attr == modeling.NormalAttribute {
			for i := len(data) - len(c.cuts); i < len(data); i++ {
				if length := data[i].Length(); length > 0 {
					data[i] = data[i].Scale(1 / length)
				}
			}
		}
		m = m.SetFloat3Attribute(attr, data)
	}
	for attr, data := range readAllFloat2Data(m) {
		m = m.SetFloat2Attribute(attr, appendCuts(data, c.cuts, func(a, b vector2.Float64, t float64) vector2.Float64 {
			return a.Scale(1 - t).Add(b.Scale(t))
		}))
	}
	for attr, data := range readAllFloat1Data(m) {
		m = m.SetFloat1Attribute(attr, appendCuts(data, c.cuts, func(a, b float64, t float64) float64 {
			return a*(1-t) + b*t
		}))
	}
	return m
}

func sliceTrianglesByPlaneWithAttribute(m modeling.Mesh, plane geometry.Plane, attribute string) (modeling.Mesh, modeling.Mesh) {
	if err := RequireTopology(m, modeling.TriangleTopology); err != nil {
		panic(err)
	}
	if m.PrimitiveCount() == 0 {
		return m, m
	}
	if err := RequireV3Attribute(m, attribute); err != nil {
		panic(err)
	}

	positions := m.Float3Attribute(attribute)
	cuts := planeCuts{
		distances: make([]float64, positions.Len()),
		byEdge:    make(map[[2]int]int),
	}
	for i := range cuts.distances {
		cuts.distances[i] = plane.Normal().Dot(positions.At(i).Sub(plane.Origin()))
	}

	indices := m.Indices()
	above := make([]int, 0, indices.Len())
	below := make([]int, 0, indices.Len())
	for i := 0; i < indices.Len(); i += 3 {
		tri := [3]int{indices.At(i), indices.At(i + 1), indices.At(i + 2)}
		// A face on the plane belongs to the solid behind it, the side its normal points away from.
		if cuts.distances[tri[0]] == 0 && cuts.distances[tri[1]] == 0 && cuts.distances[tri[2]] == 0 {
			a, b, c := positions.At(tri[0]), positions.At(tri[1]), positions.At(tri[2])
			if b.Sub(a).Cross(c.Sub(a)).Dot(plane.Normal()) > 0 {
				below = append(below, tri[:]...)
			} else {
				above = append(above, tri[:]...)
			}
			continue
		}
		above = cuts.clip(above, tri, 1)
		below = cuts.clip(below, tri, -1)
	}

	m = cuts.extend(m)
	return RemovedUnreferencedVertices(m.SetIndices(above)), RemovedUnreferencedVertices(m.SetIndices(below))
}
