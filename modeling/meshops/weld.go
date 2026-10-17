package meshops

import (
	"encoding/binary"
	"fmt"
	"math"
	"slices"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type WeldTransformer struct {
	Attribute       string
	Tolerance       float64
	MatchAttributes bool
}

func (wt WeldTransformer) attribute() string {
	return wt.Attribute
}

func (wt WeldTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(wt, modeling.PositionAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	if wt.MatchAttributes {
		return WeldIdentical(m, attribute, wt.Tolerance), nil
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

// WeldIdentical merges vertices closer together than tolerance only when
// every other attribute they carry is equal, so no data is lost.
func WeldIdentical(m modeling.Mesh, attribute string, tolerance float64) modeling.Mesh {
	positions := m.Float3Attribute(attribute)
	weld := geometry.NewPointWelder3D(tolerance)
	components := vertexComponentsOf(m, attribute)

	first := make(map[string]int)
	representative := make([]int, positions.Len())
	var values []float64
	var key []byte
	for i := range representative {
		representative[i] = i
		key = binary.LittleEndian.AppendUint64(key[:0], uint64(weld.Index(positions.At(i))))

		values = values[:0]
		for _, read := range components {
			values = read(values, i)
		}
		if slices.ContainsFunc(values, math.IsNaN) {
			continue
		}
		for _, v := range values {
			// +0 turns -0 into 0, so equal values share a key.
			key = binary.LittleEndian.AppendUint64(key, math.Float64bits(v+0))
		}

		if seen, ok := first[string(key)]; ok {
			representative[i] = seen
		} else {
			first[string(key)] = i
		}
	}

	return rebuild(m, representative)
}

func vertexComponentsOf(m modeling.Mesh, skip string) []func(dst []float64, i int) []float64 {
	var components []func(dst []float64, i int) []float64
	for _, attr := range m.Float1Attributes() {
		data := m.Float1Attribute(attr)
		components = append(components, func(dst []float64, i int) []float64 {
			return append(dst, data.At(i))
		})
	}
	for _, attr := range m.Float2Attributes() {
		data := m.Float2Attribute(attr)
		components = append(components, func(dst []float64, i int) []float64 {
			v := data.At(i)
			return append(dst, v.X(), v.Y())
		})
	}
	for _, attr := range m.Float3Attributes() {
		if attr == skip {
			continue
		}
		data := m.Float3Attribute(attr)
		components = append(components, func(dst []float64, i int) []float64 {
			v := data.At(i)
			return append(dst, v.X(), v.Y(), v.Z())
		})
	}
	for _, attr := range m.Float4Attributes() {
		data := m.Float4Attribute(attr)
		components = append(components, func(dst []float64, i int) []float64 {
			v := data.At(i)
			return append(dst, v.X(), v.Y(), v.Z(), v.W())
		})
	}
	return components
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
	Mesh            nodes.Output[modeling.Mesh]
	Attribute       nodes.Output[string]  `description:"Attribute whose values decide which vertices are the same point (default: Position)."`
	Tolerance       nodes.Output[float64] `description:"How far apart two vertices can be and still merge. Zero merges only exact matches."`
	MatchAttributes nodes.Output[bool]    `description:"When true, vertices merge only if every other attribute is also identical, so no vertex data is lost. Defaults to false, which keeps the first vertex's data."`
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

	if nodes.TryGetOutputValue(out, n.MatchAttributes, false) {
		out.Set(WeldIdentical(mesh, attribute, tolerance))
		return
	}
	out.Set(Weld(mesh, attribute, tolerance))
}
