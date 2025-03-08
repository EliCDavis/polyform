package primitives

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Quad struct {
	Width float64
	Depth float64
	UVs   *StripUVs
}

func (q Quad) ToMesh() modeling.Mesh {
	up := vector3.Up[float64]()

	v2Data := make(map[string][]vector2.Float64)
	if q.UVs != nil {
		v2Data[modeling.TexCoordAttribute] = []vector2.Vector[float64]{
			q.UVs.StartLeft(),
			q.UVs.EndLeft(),
			q.UVs.EndRight(),
			q.UVs.StartRight(),
		}
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

type QuadNode = nodes.Struct[QuadNodeData]

type QuadNodeData struct {
	Width nodes.Output[float64]
	Depth nodes.Output[float64]
	UVs   nodes.Output[StripUVs]
}

func (c QuadNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	quad := Quad{
		Width: 1,
		Depth: 1,
	}

	if c.Width != nil {
		quad.Width = c.Width.Value()
	}

	if c.Depth != nil {
		quad.Depth = c.Depth.Value()
	}

	if c.UVs != nil {
		uvs := c.UVs.Value()
		quad.UVs = &uvs
	}

	return nodes.NewStructOutput(quad.ToMesh())
}
