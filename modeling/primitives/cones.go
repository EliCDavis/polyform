package primitives

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Cone struct {
	Height float64
	Radius float64
	Sides  int
}

func (c Cone) ToMesh() modeling.Mesh {
	if c.Sides < 3 {
		panic("can not make cone with less that 3 sides")
	}

	verts := repeat.CirclePoints(c.Sides, c.Radius)
	lastVert := len(verts)
	verts = append(verts, vector3.New(0., c.Height, 0.))
	uvs := make([]vector2.Float64, len(verts))
	uvs[len(uvs)-1] = vector2.One[float64]()

	tris := make([]int, 0, c.Sides*3)
	for i := 0; i < c.Sides; i++ {
		tris = append(tris, i, lastVert, i+1)
	}
	tris[len(tris)-1] = 0

	return modeling.NewMesh(modeling.TriangleTopology, tris).
		SetFloat3Attribute(modeling.PositionAttribute, verts).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)

}

type ConeNode = nodes.Struct[ConeNodeData]

type ConeNodeData struct {
	Height nodes.Output[float64]
	Radius nodes.Output[float64]
	Sides  nodes.Output[int]
}

func (r ConeNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	cone := Cone{
		Height: 1,
		Radius: 0.5,
		Sides:  3,
	}

	if r.Sides != nil {
		cone.Sides = max(cone.Sides, r.Sides.Value())
	}

	if r.Radius != nil {
		cone.Radius = r.Radius.Value()
	}

	if r.Height != nil {
		cone.Height = r.Height.Value()
	}

	return nodes.NewStructOutput(cone.ToMesh())
}
