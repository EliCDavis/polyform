package sdf

import (
	"errors"
	"math"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector3"
)

type meshEdge struct{ a, b int }

func newMeshEdge(a, b int) meshEdge {
	if a > b {
		a, b = b, a
	}
	return meshEdge{a, b}
}

type meshField struct {
	tree          *trees.OctTree
	triangles     []geometry.Triangle
	corners       [][3]int
	faceNormals   []vector3.Float64
	vertexNormals []vector3.Float64
	edgeNormals   map[meshEdge]vector3.Float64
}

type weldCell struct{ x, y, z int64 }

func weldPositions(positions *iter.ArrayIterator[vector3.Float64], bounds geometry.AABB) ([]int, int) {
	tolerance := 1e-9 * math.Max(1, bounds.Size().Length())
	cells := make(map[weldCell][]int)
	representatives := make([]vector3.Float64, 0, positions.Len())
	ids := make([]int, positions.Len())

	for i := range ids {
		p := positions.At(i)
		cell := weldCell{
			int64(math.Floor(p.X() / tolerance)),
			int64(math.Floor(p.Y() / tolerance)),
			int64(math.Floor(p.Z() / tolerance)),
		}

		id := -1
		for dx := int64(-1); dx <= 1 && id < 0; dx++ {
			for dy := int64(-1); dy <= 1 && id < 0; dy++ {
				for dz := int64(-1); dz <= 1 && id < 0; dz++ {
					for _, candidate := range cells[weldCell{cell.x + dx, cell.y + dy, cell.z + dz}] {
						if representatives[candidate].Distance(p) <= tolerance {
							id = candidate
							break
						}
					}
				}
			}
		}

		if id < 0 {
			id = len(representatives)
			representatives = append(representatives, p)
			cells[cell] = append(cells[cell], id)
		}
		ids[i] = id
	}
	return ids, len(representatives)
}

func newMeshField(mesh modeling.Mesh) meshField {
	positions := mesh.Float3Attribute(modeling.PositionAttribute)
	indices := mesh.Indices()

	weldedIDs, weldedCount := weldPositions(positions, mesh.BoundingBox(modeling.PositionAttribute))

	count := indices.Len() / 3
	field := meshField{
		triangles:     make([]geometry.Triangle, count),
		corners:       make([][3]int, count),
		faceNormals:   make([]vector3.Float64, count),
		vertexNormals: make([]vector3.Float64, weldedCount),
		edgeNormals:   make(map[meshEdge]vector3.Float64),
	}
	elements := make([]trees.Element, count)

	for t := 0; t < count; t++ {
		tri := geometry.Triangle{
			positions.At(indices.At(t * 3)),
			positions.At(indices.At(t*3 + 1)),
			positions.At(indices.At(t*3 + 2)),
		}
		ids := [3]int{
			weldedIDs[indices.At(t*3)],
			weldedIDs[indices.At(t*3+1)],
			weldedIDs[indices.At(t*3+2)],
		}
		field.triangles[t] = tri
		field.corners[t] = ids
		elements[t] = tri

		cross := tri[1].Sub(tri[0]).Cross(tri[2].Sub(tri[0]))
		if cross.Length() == 0 {
			continue
		}
		normal := cross.Normalized()
		field.faceNormals[t] = normal

		for i := 0; i < 3; i++ {
			toNext := tri[(i+1)%3].Sub(tri[i])
			toPrev := tri[(i+2)%3].Sub(tri[i])
			angle := math.Acos(math.Max(-1, math.Min(1, toNext.Normalized().Dot(toPrev.Normalized()))))
			if !math.IsNaN(angle) {
				field.vertexNormals[ids[i]] = field.vertexNormals[ids[i]].Add(normal.Scale(angle))
			}

			edge := newMeshEdge(ids[i], ids[(i+1)%3])
			field.edgeNormals[edge] = field.edgeNormals[edge].Add(normal)
		}
	}

	field.tree = trees.NewOctree(elements, trees.MaxDepth(trees.OctreeDepthFromCount(count)))
	return field
}

func barycentric(tri geometry.Triangle, p vector3.Float64) vector3.Float64 {
	v0, v1, v2 := tri[1].Sub(tri[0]), tri[2].Sub(tri[0]), p.Sub(tri[0])
	d00, d01, d11 := v0.Dot(v0), v0.Dot(v1), v1.Dot(v1)
	d20, d21 := v2.Dot(v0), v2.Dot(v1)
	denom := d00*d11 - d01*d01
	if denom == 0 {
		return vector3.New(1., 0., 0.)
	}
	v := (d11*d20 - d01*d21) / denom
	w := (d00*d21 - d01*d20) / denom
	return vector3.New(1-v-w, v, w)
}

func (f meshField) pseudoNormal(t int, closest vector3.Float64) vector3.Float64 {
	const onFeature = 1e-7
	weights := barycentric(f.triangles[t], closest)
	ids := f.corners[t]
	zero := [3]bool{weights.X() < onFeature, weights.Y() < onFeature, weights.Z() < onFeature}

	switch {
	case zero[1] && zero[2]:
		return f.vertexNormals[ids[0]]
	case zero[0] && zero[2]:
		return f.vertexNormals[ids[1]]
	case zero[0] && zero[1]:
		return f.vertexNormals[ids[2]]
	case zero[0]:
		return f.edgeNormals[newMeshEdge(ids[1], ids[2])]
	case zero[1]:
		return f.edgeNormals[newMeshEdge(ids[0], ids[2])]
	case zero[2]:
		return f.edgeNormals[newMeshEdge(ids[0], ids[1])]
	}
	return f.faceNormals[t]
}

func (f meshField) distance(v vector3.Float64) float64 {
	t, closest := f.tree.ClosestPoint(v)
	offset := v.Sub(closest)
	dist := offset.Length()
	if offset.Dot(f.pseudoNormal(t, closest)) < 0 {
		return -dist
	}
	return dist
}

func Mesh(mesh modeling.Mesh) (sample.Vec3ToFloat, error) {
	if mesh.Topology() != modeling.TriangleTopology {
		return nil, errors.New("mesh must have triangle topology")
	}
	if !mesh.HasFloat3Attribute(modeling.PositionAttribute) || mesh.PrimitiveCount() == 0 {
		return nullField, nil
	}
	return newMeshField(mesh).distance, nil
}

type MeshNode struct {
	Mesh nodes.LiftedPort[modeling.Mesh]
}

func (cn MeshNode) Description() string {
	return "Signed distance to a triangle mesh, negative behind its faces."
}

func (cn MeshNode) Field(out *nodes.Lifted[sample.Vec3ToFloat]) {
	wired := cn.Mesh != nil
	nodes.Zip1(out, cn.Mesh, func(mesh modeling.Mesh) sample.Vec3ToFloat {
		if !wired {
			return nullField
		}
		field, err := Mesh(mesh)
		out.CaptureError(err)
		return field
	})
}
