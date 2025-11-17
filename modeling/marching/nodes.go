package marching

import (
	"fmt"

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
	Field      nodes.Output[sample.Vec3ToFloat] `description:"The SDF to tesselate"`
	Resolution nodes.Output[float64]            `description:"Number of marching cube voxels contained in a single 'unit'"`
	Surface    nodes.Output[float64]            `description:"value of the SDF that represents the surface (default: 0)"`
	Domain     nodes.Output[geometry.AABB]      `description:"The region in which the marching cubes algorithm runs"`
}

func (cn MarchNode) Mesh(out *nodes.StructOutput[modeling.Mesh]) {
	if cn.Field == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	resolution := nodes.TryGetOutputValue(out, cn.Resolution, 1.)
	if resolution <= 0 {
		out.CaptureError(nodes.InvalidInputError{
			Input:   cn.Resolution,
			Message: fmt.Sprintf("value must be greater than 0 (recieved %f)", resolution),
		})
		return
	}

	out.Set(March(
		nodes.GetOutputValue(out, cn.Field),
		nodes.TryGetOutputValue(
			out,
			cn.Domain,
			geometry.NewAABB(vector3.Zero[float64](), vector3.Fill(10.)),
		),
		1/resolution,
		nodes.TryGetOutputValue(out, cn.Surface, 0.),
	))
}
