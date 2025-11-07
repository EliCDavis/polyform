package primitives

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Quad struct {
	Width   float64
	Depth   float64
	Columns int
	Rows    int
	UVs     EuclideanUVSpace
}

func (q Quad) simpleQuad() modeling.Mesh {
	up := vector3.Up[float64]()

	v2Data := make(map[string][]vector2.Float64)
	if q.UVs != nil {
		v2Data[modeling.TexCoordAttribute] = q.UVs.AtXYs([]vector2.Float64{
			vector2.New(0., 0.),
			vector2.New(0., 1.),
			vector2.New(1., 1.),
			vector2.New(1., 0.),
		})
	}

	halfWidth := q.Width / 2
	halfHeight := q.Depth / 2

	return modeling.NewTriangleMesh([]int{0, 1, 2, 2, 3, 0}).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New(-halfWidth, 0., -halfHeight),
				vector3.New(-halfWidth, 0., halfHeight),
				vector3.New(halfWidth, 0., halfHeight),
				vector3.New(halfWidth, 0., -halfHeight),
			},
			modeling.NormalAttribute: {
				up, up, up, up,
			},
		}).
		SetFloat2Data(v2Data)
}

func (q Quad) ToMesh() modeling.Mesh {
	if q.Rows <= 1 && q.Columns <= 1 {
		return q.simpleQuad()
	}
	up := vector3.Up[float64]()

	// A 3x1 quad has 8 verts
	// . -- . -- . -- .
	// |    |    |    |
	// . -- . -- . -- .

	vertexCount := (q.Rows + 1) * (q.Columns + 1)
	verts := make([]vector3.Float64, 0, vertexCount)
	normals := make([]vector3.Float64, 0, vertexCount)
	uvs := make([]vector2.Float64, 0, vertexCount)

	uvStrip := q.UVs

	halfWidth := q.Width / 2.
	halfHeight := q.Depth / 2.
	rowInc := q.Width / float64(q.Rows)
	colInc := q.Depth / float64(q.Columns)

	// I'm lazy and don't want to think hard
	vertToIndex := make(map[vector2.Int]int)

	for x := range q.Rows + 1 {
		for y := range q.Columns + 1 {
			vertToIndex[vector2.New(x, y)] = len(verts)
			verts = append(verts, vector3.New(
				(rowInc*float64(x))-halfWidth,
				0,
				(colInc*float64(y))-halfHeight,
			))
			normals = append(normals, up)

			if uvStrip != nil {
				uvs = append(uvs, uvStrip.AtXY(
					vector2.New(
						float64(x)/float64(q.Rows),
						float64(y)/float64(q.Columns),
					),
				))
			}
		}
	}

	indices := make([]int, q.Rows*q.Columns*6)
	for x := range q.Rows {
		for y := range q.Columns {
			indices = append(
				indices,
				vertToIndex[vector2.New(x, y)],
				vertToIndex[vector2.New(x, y+1)],
				vertToIndex[vector2.New(x+1, y+1)],
				vertToIndex[vector2.New(x+1, y+1)],
				vertToIndex[vector2.New(x+1, y)],
				vertToIndex[vector2.New(x, y)],
			)
		}
	}

	return modeling.NewTriangleMesh(indices).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: verts,
			modeling.NormalAttribute:   normals,
		}).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)

}

type QuadNode struct {
	Width   nodes.Output[float64]
	Depth   nodes.Output[float64]
	Columns nodes.Output[int]
	Rows    nodes.Output[int]
	UVs     nodes.Output[StripUVs]
}

func (c QuadNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	quad := Quad{
		Width:   nodes.TryGetOutputValue(out, c.Width, 1.),
		Depth:   nodes.TryGetOutputValue(out, c.Depth, 1.),
		Rows:    max(nodes.TryGetOutputValue(out, c.Rows, 1), 1),
		Columns: max(nodes.TryGetOutputValue(out, c.Columns, 1), 1),
		UVs:     nodes.TryGetOutputReference(out, c.UVs, nil),
	}
	out.Set(quad.ToMesh())
}
