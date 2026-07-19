package modeling

import (
	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

func iterToArr[T any](it *iter.ArrayIterator[T]) []T {
	data := make([]T, it.Len())
	for i := range it.Len() {
		data[i] = it.At(i)
	}
	return data
}

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[SelectFromMeshNode]](factory)
	refutil.RegisterType[nodes.Struct[MapEntryNode[[]float64]]](factory)
	refutil.RegisterType[nodes.Struct[MapEntryNode[[]vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[MapEntryNode[[]vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[MapEntryNode[[]vector4.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[NewMeshNode]](factory)
	refutil.RegisterType[nodes.Struct[SetAttribute1DNode]](factory)
	refutil.RegisterType[nodes.Struct[SetAttribute2DNode]](factory)
	refutil.RegisterType[nodes.Struct[SetAttribute3DNode]](factory)
	refutil.RegisterType[nodes.Struct[SetAttribute4DNode]](factory)
	refutil.RegisterType[TopologyNode](factory)
	refutil.RegisterType[AttributeNode](factory)
	generator.RegisterTypes(factory)
}

type SelectFromMeshNode struct {
	Mesh nodes.Output[Mesh]
}

func (n SelectFromMeshNode) Float3(recorder nodes.ExecutionRecorder, attr string) []vector3.Float64 {
	if n.Mesh == nil {
		return nil
	}

	mesh := nodes.GetOutputValue(recorder, n.Mesh)
	if !mesh.HasFloat3Attribute(attr) {
		return nil
	}

	return iterToArr(mesh.Float3Attribute(attr))
}

func (n SelectFromMeshNode) Float2(recorder nodes.ExecutionRecorder, attr string) []vector2.Float64 {
	if n.Mesh == nil {
		return nil
	}

	mesh := nodes.GetOutputValue(recorder, n.Mesh)
	if !mesh.HasFloat2Attribute(attr) {
		return nil
	}

	return iterToArr(mesh.Float2Attribute(attr))
}

func (n SelectFromMeshNode) Indices(out *nodes.StructOutput[[]int]) {
	if n.Mesh != nil {
		out.Set(iterToArr(nodes.GetOutputValue(out, n.Mesh).Indices()))
	}
}

func (n SelectFromMeshNode) Topology(out *nodes.StructOutput[Topology]) {
	if n.Mesh == nil {
		// TODO: EEEHHHHHHHHHHHHHHHH
		out.Set(TriangleTopology)
		return
	}
	out.Set(nodes.GetOutputValue(out, n.Mesh).topology)
}

func (n SelectFromMeshNode) Position(out *nodes.StructOutput[[]vector3.Float64]) {
	out.Set(n.Float3(out, PositionAttribute))
}

func (n SelectFromMeshNode) Normal(out *nodes.StructOutput[[]vector3.Float64]) {
	out.Set(n.Float3(out, NormalAttribute))
}

func (n SelectFromMeshNode) Color3(out *nodes.StructOutput[[]vector3.Float64]) {
	out.Set(n.Float3(out, ColorAttribute))
}

func (n SelectFromMeshNode) TexCoord(out *nodes.StructOutput[[]vector2.Float64]) {
	out.Set(n.Float2(out, TexCoordAttribute))
}

type MapEntry[T any] struct {
	Name string
	Data T
}

type MapEntryNode[T any] struct {
	Name nodes.Output[string]
	Data nodes.Output[T]
}

func (men MapEntryNode[T]) Out(out *nodes.StructOutput[MapEntry[T]]) {
	var val T
	out.Set(MapEntry[T]{
		Name: nodes.TryGetOutputValue(out, men.Name, ""),
		Data: nodes.TryGetOutputValue(out, men.Data, val),
	})
}

type NewMeshNode struct {
	Topology   nodes.Output[Topology]
	Indices    nodes.Output[[]int]
	Float1Data []nodes.Output[MapEntry[[]float64]]
	Float2Data []nodes.Output[MapEntry[[]vector2.Float64]]
	Float3Data []nodes.Output[MapEntry[[]vector3.Float64]]
	Float4Data []nodes.Output[MapEntry[[]vector4.Float64]]
}

func collapseMapEntries[T any](recorder nodes.ExecutionRecorder, entries []nodes.Output[MapEntry[T]]) map[string]T {
	result := make(map[string]T)
	resolvedEntries := nodes.GetOutputValues(recorder, entries)
	for _, val := range resolvedEntries {
		result[val.Name] = val.Data
	}
	return result
}

func (nmn NewMeshNode) Mesh(out *nodes.StructOutput[Mesh]) {
	mesh := NewMesh(
		nodes.TryGetOutputValue(out, nmn.Topology, TriangleTopology),
		nodes.TryGetOutputValue(out, nmn.Indices, nil),
	).
		SetFloat1Data(collapseMapEntries(out, nmn.Float1Data)).
		SetFloat2Data(collapseMapEntries(out, nmn.Float2Data)).
		SetFloat3Data(collapseMapEntries(out, nmn.Float3Data)).
		SetFloat4Data(collapseMapEntries(out, nmn.Float4Data))

	out.Set(mesh)
}

// ============================================================================

type TopologyNode struct{}

func (TopologyNode) Name() string {
	return "Topology"
}

func (TopologyNode) Inputs() map[string]nodes.InputPort {
	return nil
}

func (p *TopologyNode) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		"Triangle": nodes.ConstOutput[Topology]{
			Ref:      p,
			Val:      TriangleTopology,
			PortName: "Triangle",
		},

		"Point": nodes.ConstOutput[Topology]{
			Ref:      p,
			Val:      PointTopology,
			PortName: "Point",
		},

		"Quad": nodes.ConstOutput[Topology]{
			Ref:      p,
			Val:      QuadTopology,
			PortName: "Quad",
		},

		"Line": nodes.ConstOutput[Topology]{
			Ref:      p,
			Val:      LineTopology,
			PortName: "Line",
		},

		"Line Strip": nodes.ConstOutput[Topology]{
			Ref:      p,
			Val:      LineStripTopology,
			PortName: "Line Strip",
		},

		"Line Loop": nodes.ConstOutput[Topology]{
			Ref:      p,
			Val:      LineLoopTopology,
			PortName: "Line Loop",
		},
	}
}

// ============================================================================

type AttributeNode struct{}

func (AttributeNode) Name() string {
	return "Attribute"
}

func (AttributeNode) Inputs() map[string]nodes.InputPort {
	return nil
}

func (p *AttributeNode) stringConstOut(s string) nodes.ConstOutput[string] {
	return nodes.ConstOutput[string]{
		Ref:      p,
		Val:      s,
		PortName: s,
	}
}

func (p *AttributeNode) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		PositionAttribute:  p.stringConstOut(PositionAttribute),
		NormalAttribute:    p.stringConstOut(NormalAttribute),
		ColorAttribute:     p.stringConstOut(ColorAttribute),
		TexCoordAttribute:  p.stringConstOut(TexCoordAttribute),
		ClassAttribute:     p.stringConstOut(ClassAttribute),
		IntensityAttribute: p.stringConstOut(IntensityAttribute),
		JointAttribute:     p.stringConstOut(JointAttribute),
		WeightAttribute:    p.stringConstOut(WeightAttribute),
		ScaleAttribute:     p.stringConstOut(ScaleAttribute),
		RotationAttribute:  p.stringConstOut(RotationAttribute),
		OpacityAttribute:   p.stringConstOut(OpacityAttribute),
		FDCAttribute:       p.stringConstOut(FDCAttribute),
	}
}

// ============================================================================

func setAttribute[T any](
	out *nodes.StructOutput[Mesh],
	mesh nodes.Output[Mesh],
	attribute nodes.Output[string],
	data nodes.Output[[]T],
	setAttr func(Mesh, string, []T) Mesh,
) {
	if attribute == nil || data == nil {
		out.Set(nodes.TryGetOutputValue(out, mesh, EmptyMesh(PointTopology)))
		return
	}

	attr := nodes.GetOutputValue(out, attribute)
	vals := nodes.GetOutputValue(out, data)
	out.Set(setAttr(nodes.TryGetOutputValue(out, mesh, EmptyPointcloud()), attr, vals))
}

type SetAttribute1DNode struct {
	Mesh      nodes.Output[Mesh]
	Attribute nodes.Output[string]
	Data      nodes.Output[[]float64]
}

func (n SetAttribute1DNode) Out(out *nodes.StructOutput[Mesh]) {
	setAttribute(out, n.Mesh, n.Attribute, n.Data, Mesh.SetFloat1Attribute)
}

type SetAttribute2DNode struct {
	Mesh      nodes.Output[Mesh]
	Attribute nodes.Output[string]
	Data      nodes.Output[[]vector2.Float64]
}

func (n SetAttribute2DNode) Out(out *nodes.StructOutput[Mesh]) {
	setAttribute(out, n.Mesh, n.Attribute, n.Data, Mesh.SetFloat2Attribute)
}

type SetAttribute3DNode struct {
	Mesh      nodes.Output[Mesh]
	Attribute nodes.Output[string]
	Data      nodes.Output[[]vector3.Float64]
}

func (n SetAttribute3DNode) Out(out *nodes.StructOutput[Mesh]) {
	setAttribute(out, n.Mesh, n.Attribute, n.Data, Mesh.SetFloat3Attribute)
}

type SetAttribute4DNode struct {
	Mesh      nodes.Output[Mesh]
	Attribute nodes.Output[string]
	Data      nodes.Output[[]vector4.Float64]
}

func (n SetAttribute4DNode) Out(out *nodes.StructOutput[Mesh]) {
	setAttribute(out, n.Mesh, n.Attribute, n.Data, Mesh.SetFloat4Attribute)
}
