package marching

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[MarchNode]](factory)
	refutil.RegisterType[nodes.Struct[ApplyColorFieldNode]](factory)

	registry.Register(factory)
}

type MarchNode struct {
	Field      nodes.Output[sample.Vec3ToFloat] `description:"The SDF to tesselate"`
	Resolution nodes.Output[float64]            `description:"Marching cube voxels per unit of space. Voxel size is 1/Resolution, so higher means finer detail. Must be large enough that 1/Resolution is smaller than the Domain, or nothing is produced." default:"1"`
	Surface    nodes.Output[float64]            `description:"Value of the SDF that represents the surface. Positive grows the shape outward by that distance, negative shrinks it inward."`
	Domain     nodes.Output[geometry.AABB]      `description:"The region in which the marching cubes algorithm runs" default:"0,0,0;10,10,10"`
}

func (cn MarchNode) Description() string {
	return "Turns a distance field into a mesh with marching cubes. Domain is the box that gets scanned; anything outside it is cut off. Higher Resolution gives finer detail: voxel size is 1/Resolution, so a Resolution smaller than 1/Domain size makes voxels bigger than the whole domain and yields an empty mesh."
}

func (cn MarchNode) Mesh(out *nodes.StructOutput[modeling.Mesh]) {
	if cn.Field == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	resolution := nodes.GetOutputValue(out, cn.Resolution)
	if resolution <= 0 {
		out.CaptureError(nodes.InvalidInputError{
			Input:   cn.Resolution,
			Message: fmt.Sprintf("value must be greater than 0 (recieved %f)", resolution),
		})
		return
	}

	domain := nodes.GetOutputValue(out, cn.Domain)
	voxelSize := 1 / resolution

	mesh := March(
		nodes.GetOutputValue(out, cn.Field),
		domain,
		voxelSize,
		nodes.GetOutputValue(out, cn.Surface),
	)
	out.Set(mesh)

	if mesh.PrimitiveCount() > 0 {
		return
	}

	size := domain.Size()
	smallestSide := math.Min(size.X(), math.Min(size.Y(), size.Z()))
	if voxelSize >= smallestSide {
		out.CaptureError(fmt.Errorf(
			"produced an empty mesh: Resolution %v makes voxels %v across, which is bigger than the domain's smallest side (%v). Raise Resolution above %v",
			resolution, voxelSize, smallestSide, 1/smallestSide,
		))
		return
	}

	out.CaptureError(fmt.Errorf(
		"produced an empty mesh: the field never crosses the Surface value inside the domain. Check the domain %v actually contains the shape, and sample_field a point that should be inside it",
		domain,
	))
}
