package repeat

import (
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/EliCDavis/polyform/math/bias"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type MeshSurface struct {
	Mesh      modeling.Mesh
	Attribute string
	Samples   int
}

func phi(d int) float64 {
	x := 2.0000
	for range 10 {
		x = math.Pow(1+x, 1./(float64(d)+1))
	}
	return x
}

func (ms MeshSurface) TRS() []trs.TRS {
	if ms.Samples == 0 {
		return nil
	}

	if ms.Mesh.Topology() != modeling.TriangleTopology {
		panic(fmt.Errorf("can only sample triangle mesh"))
	}

	attr := ms.Attribute

	triCount := ms.Mesh.PrimitiveCount()
	if triCount == 0 {
		return nil
	}

	items := make([]bias.ListItem[int], triCount)
	for i := range triCount {
		items = append(items, bias.ListItem[int]{
			Item:   i,
			Weight: ms.Mesh.Tri(i).Area3D(attr),
		})
	}

	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano())))
	list := bias.NewList(items, bias.ListConfig{Seed: r})

	samplesPerTri := make([]int, triCount)
	for range ms.Samples {
		samplesPerTri[list.Next()]++
	}

	g := phi(2)
	a1 := 1. / g
	a2 := 1. / (g * g)
	alpha := vector2.New(a1, a2)
	seed := vector2.New(0.5, 0.5)

	transforms := make([]trs.TRS, 0, ms.Samples)
	for triIndex, sampleCount := range samplesPerTri {
		tri := ms.Mesh.Tri(triIndex)
		normal := tri.Normal(ms.Attribute)

		for range sampleCount {
			n := float64(len(transforms))
			// u := vector2.New(math.Mod(0.5+a1*n, 1.), math.Mod(0.5+a2*n, 1.))
			u := alpha.Scale(n).Add(seed).Mod(1)
			p := tri.UniformSample(ms.Attribute, u)
			transforms = append(transforms, trs.Position(p).LookAt(p.Add(normal)))
		}
	}

	return transforms
}

type SampleMeshSurfaceNode struct {
	Mesh      nodes.Output[modeling.Mesh] `description:"The mesh to scatter points across. Must have triangle topology."`
	Attribute nodes.Output[string]        `description:"Which mesh attribute to sample area/position/normal from. Defaults to modeling.PositionAttribute (\"Position\")."`
	Samples   nodes.Output[int]           `description:"How many points to scatter, distributed so larger triangles get proportionally more samples (uniform per unit area, not per triangle). Defaults to 0."`
}

func (rnd SampleMeshSurfaceNode) Description() string {
	return "Scatters Samples points across the surface of Mesh, weighted by triangle area, each rotated to face along the surface normal at that point. Non-deterministic — every call produces a different scatter."
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
		Attribute: nodes.TryGetOutputValue(out, rnd.Attribute, modeling.PositionAttribute),
		Samples:   nodes.TryGetOutputValue(out, rnd.Samples, 0),
	}
	out.Set(surface.TRS())
}
