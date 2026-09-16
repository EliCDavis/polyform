package meshops

import (
	"fmt"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type WeldTransformer struct {
	Attribute string
	Tolerance float64
}

func (wt WeldTransformer) attribute() string {
	return wt.Attribute
}

func (wt WeldTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(wt, modeling.PositionAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	return Weld(m, attribute, wt.Tolerance), nil
}

// Merges vertices closer together than tolerance, keeping the data of the
// first one seen, and drops any primitive left with a repeated corner.
func Weld(m modeling.Mesh, attribute string, tolerance float64) modeling.Mesh {
	positions := m.Float3Attribute(attribute)
	weld := geometry.NewPointWelder3D(tolerance)

	firstVertex := make([]int, 0, positions.Len())
	representative := make([]int, positions.Len())
	for i := range representative {
		id := weld.Index(positions.At(i))
		if id == len(firstVertex) {
			firstVertex = append(firstVertex, i)
		}
		representative[i] = firstVertex[id]
	}

	return rebuild(m, representative)
}

// Keeps only the primitives that survive the merge and only the vertices
// those primitives still reach.
func rebuild(m modeling.Mesh, representative []int) modeling.Mesh {
	original := m.Indices()
	stride := m.Topology().IndexSize()

	kept := make([]int, 0, original.Len())
	for i := 0; i+stride <= original.Len(); i += stride {
		corners := make([]int, stride)
		degenerate := false
		for k := 0; k < stride; k++ {
			corners[k] = representative[original.At(i+k)]
			for j := 0; j < k; j++ {
				if corners[j] == corners[k] {
					degenerate = true
				}
			}
		}
		if degenerate {
			continue
		}
		kept = append(kept, corners...)
	}

	compact := make(map[int]int)
	order := make([]int, 0)
	for i, index := range kept {
		to, ok := compact[index]
		if !ok {
			to = len(order)
			compact[index] = to
			order = append(order, index)
		}
		kept[i] = to
	}

	return modeling.
		NewMesh(m.Topology(), kept).
		SetFloat4Data(gatherByIndex(order, m.Float4Attributes(),
			func(s string) *iter.ArrayIterator[vector4.Float64] { return m.Float4Attribute(s) })).
		SetFloat3Data(gatherByIndex(order, m.Float3Attributes(),
			func(s string) *iter.ArrayIterator[vector3.Float64] { return m.Float3Attribute(s) })).
		SetFloat2Data(gatherByIndex(order, m.Float2Attributes(),
			func(s string) *iter.ArrayIterator[vector2.Float64] { return m.Float2Attribute(s) })).
		SetFloat1Data(gatherByIndex(order, m.Float1Attributes(),
			func(s string) *iter.ArrayIterator[float64] { return m.Float1Attribute(s) }))
}

type WeldNode struct {
	Mesh      nodes.Output[modeling.Mesh]
	Attribute nodes.Output[string]  `description:"Attribute whose values decide which vertices are the same point (default: Position)."`
	Tolerance nodes.Output[float64] `description:"How far apart two vertices can be and still merge. Zero merges only exact matches."`
}

func (n WeldNode) Description() string {
	return "Merges vertices that sit within a tolerance of each other into one, so a mesh split along its seams becomes continuous. Primitives left with a repeated corner (degenerate) are dropped."
}

func (n WeldNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
	if n.Mesh == nil {
		return
	}

	mesh := nodes.GetOutputValue(out, n.Mesh)
	attribute := nodes.TryGetOutputValue(out, n.Attribute, modeling.PositionAttribute)
	if !mesh.HasFloat3Attribute(attribute) {
		out.CaptureError(fmt.Errorf("input is missing %s data", attribute))
		out.Set(mesh)
		return
	}

	tolerance := nodes.TryGetOutputValue(out, n.Tolerance, 0.)
	if tolerance < 0 {
		out.CaptureError(fmt.Errorf("Tolerance cannot be negative (received %f)", tolerance))
		out.Set(mesh)
		return
	}

	out.Set(Weld(mesh, attribute, tolerance))
}
