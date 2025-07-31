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
	"github.com/EliCDavis/vector/vector3"
)

type MeshSurface struct {
	Mesh      modeling.Mesh
	Attribute string
	Samples   int
	Up        vector3.Float64
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
	Mesh      nodes.Output[modeling.Mesh]
	Attribute nodes.Output[string]
	Samples   nodes.Output[int]
	Up        nodes.Output[vector3.Float64]
}

func (rnd SampleMeshSurfaceNode) Out() nodes.StructOutput[[]trs.TRS] {
	out := nodes.StructOutput[[]trs.TRS]{}
	if rnd.Mesh == nil {
		return out
	}

	mesh := nodes.GetOutputValue(out, rnd.Mesh)
	if mesh.Topology() != modeling.TriangleTopology {
		out.CaptureError(fmt.Errorf("mesh must have triangle topology to sample surface"))
		return out
	}

	surface := MeshSurface{
		Mesh:      mesh,
		Attribute: nodes.TryGetOutputValue(&out, rnd.Attribute, modeling.PositionAttribute),
		Samples:   nodes.TryGetOutputValue(&out, rnd.Samples, 0),
		Up:        nodes.TryGetOutputValue(&out, rnd.Up, vector3.Up[float64]()),
	}
	out.Set(surface.TRS())
	return out
}
