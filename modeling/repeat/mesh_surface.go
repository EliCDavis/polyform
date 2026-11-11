package repeat

import (
	"fmt"
	"sort"

	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type MeshSurface struct {
	Mesh      modeling.Mesh
	Attribute string
	Samples   int
	Seed      int
}

func (ms MeshSurface) TRS() []trs.TRS {
	if ms.Samples <= 0 {
		return nil
	}

	if ms.Mesh.Topology() != modeling.TriangleTopology {
		panic(fmt.Errorf("can only sample triangle mesh"))
	}

	triCount := ms.Mesh.PrimitiveCount()
	if triCount == 0 {
		return nil
	}

	cumulative := make([]float64, triCount)
	total := 0.
	for i := range triCount {
		total += ms.Mesh.Tri(i).Area3D(ms.Attribute)
		cumulative[i] = total
	}
	if total <= 0 {
		return nil
	}

	transforms := make([]trs.TRS, ms.Samples)
	for i := range transforms {
		next := chance.Roll(ms.Seed, i)
		target := next() * total
		triIndex := min(sort.SearchFloat64s(cumulative, target), triCount-1)
		for triIndex < triCount-1 && cumulative[triIndex] <= target {
			triIndex++
		}
		tri := ms.Mesh.Tri(triIndex)
		p := tri.UniformSample(ms.Attribute, vector2.New(next(), next()))
		transforms[i] = trs.Position(p).LookAt(p.Add(tri.Normal(ms.Attribute)))
	}

	return transforms
}

type SampleMeshSurfaceNode struct {
	Mesh      nodes.Output[modeling.Mesh] `description:"The mesh to scatter points across. Must have triangle topology."`
	Attribute nodes.Output[string]        `description:"Which mesh attribute to sample area/position/normal from." default:"Position"`
	Samples   nodes.Output[int]           `description:"How many points to scatter, distributed so larger triangles get proportionally more samples (uniform per unit area, not per triangle)."`
	Seed      nodes.Output[int]           `description:"The same seed always produces the same scatter."`
}

func (rnd SampleMeshSurfaceNode) Description() string {
	return "Scatters Samples points across the surface of Mesh, weighted by triangle area, each rotated to face along the surface normal at that point."
}

func (rnd SampleMeshSurfaceNode) Out(out *nodes.StructOutput[[]trs.TRS]) {
	if rnd.Mesh == nil {
		return
	}

	mesh := nodes.GetOutputValue(out, rnd.Mesh)
	if mesh.Topology() != modeling.TriangleTopology {
		out.CaptureError(fmt.Errorf("mesh must have triangle topology to sample surface"))
		return
	}

	surface := MeshSurface{
		Mesh:      mesh,
		Attribute: nodes.GetOutputValue(out, rnd.Attribute),
		Samples:   nodes.GetOutputValue(out, rnd.Samples),
		Seed:      nodes.GetOutputValue(out, rnd.Seed),
	}
	out.Set(surface.TRS())
}
