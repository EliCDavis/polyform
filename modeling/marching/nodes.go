package marching

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[MarchNode]](factory)

	generator.RegisterTypes(factory)
}

type MarchNode struct {
	Field      nodes.Output[sample.Vec3ToFloat]
	Resolution nodes.Output[float64]
	Domain     nodes.Output[geometry.AABB]
}

func (cn MarchNode) Mesh(out *nodes.StructOutput[modeling.Mesh]) {
	if cn.Field == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	// canvas := NewMarchingCanvas(
	// 	nodes.TryGetOutputValue(out, cn.Resolution, 1.),
	// )

	// canvas.AddField(Field{
	// 	Domain: nodes.TryGetOutputValue(
	// 		out,
	// 		cn.Domain,
	// 		geometry.NewAABB(vector3.Zero[float64](), vector3.One[float64]()),
	// 	),
	// 	Float1Functions: map[string]sample.Vec3ToFloat{
	// 		modeling.PositionAttribute: nodes.GetOutputValue(out, cn.Field),
	// 	},
	// })

	// out.Set(canvas.March(0))

	out.Set(March(
		nodes.GetOutputValue(out, cn.Field),
		nodes.TryGetOutputValue(
			out,
			cn.Domain,
			geometry.NewAABB(vector3.Zero[float64](), vector3.One[float64]()),
		),
		1/nodes.TryGetOutputValue(out, cn.Resolution, 1.),
	))
}
