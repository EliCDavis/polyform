package meshops

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func ProjectUV(m modeling.Mesh, normal vector3.Float64, scale, offset vector2.Float64, rotation float64) modeling.Mesh {
	if !m.HasFloat3Attribute(modeling.PositionAttribute) {
		return m
	}
	n := normal.Normalized()
	up := vector3.Up[float64]()
	if math.Abs(n.Dot(up)) > 0.999 {
		up = vector3.Backwards[float64]()
	}
	uAxis := up.Cross(n).Normalized()
	vAxis := n.Cross(uAxis)

	if rotation != 0 {
		cos, sin := math.Cos(rotation), math.Sin(rotation)
		uAxis, vAxis = uAxis.Scale(cos).Add(vAxis.Scale(sin)), vAxis.Scale(cos).Sub(uAxis.Scale(sin))
	}

	su, sv := scale.X(), scale.Y()
	if su == 0 {
		su = 1
	}
	if sv == 0 {
		sv = 1
	}

	positions := m.Float3Attribute(modeling.PositionAttribute)
	uvs := make([]vector2.Float64, positions.Len())
	for i := range uvs {
		p := positions.At(i)
		uvs[i] = vector2.New(p.Dot(uAxis)/su+offset.X(), p.Dot(vAxis)/sv+offset.Y())
	}
	return m.SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
}

type ProjectUVNode struct {
	Mesh   nodes.Output[modeling.Mesh]
	Normal nodes.Output[vector3.Float64] `description:"Direction to project along. Along +Z, u follows world X and v follows world Y." default:"0,0,1"`
	Scale  nodes.Output[vector2.Float64] `description:"World size of one texture repeat along u and v, in meters." default:"1,1"`
	Offset nodes.Output[vector2.Float64] `description:"Added to every uv after scaling."`

	Rotation nodes.Output[float64] `description:"Turns u and v within the projection plane, in radians, before scaling."`
}

func (ProjectUVNode) Description() string {
	return "Gives any mesh planar UVs from its positions."
}

func (ProjectUVNode) Keywords() []string {
	return []string{"unwrap"}
}

func (n ProjectUVNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
	if n.Mesh == nil {
		return
	}
	out.Set(ProjectUV(
		nodes.GetOutputValue(out, n.Mesh),
		nodes.GetOutputValue(out, n.Normal),
		nodes.GetOutputValue(out, n.Scale),
		nodes.GetOutputValue(out, n.Offset),
		nodes.GetOutputValue(out, n.Rotation),
	))
}
