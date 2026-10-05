package csg

import (
	"fmt"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[UnionNode]](factory)
	refutil.RegisterType[nodes.Struct[IntersectionNode]](factory)
	refutil.RegisterType[nodes.Struct[SubtractNode]](factory)
	generator.RegisterTypes(factory)
}

type UnionNode struct {
	Meshes []nodes.Output[modeling.Mesh] `description:"Solids to merge. Empty meshes are skipped."`
}

func (node UnionNode) Description() string {
	return "Merges solids into the volume enclosed by any of them, removing the geometry buried inside."
}

func (node UnionNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	runCSG(out, nonEmpty(connected(out, "Meshes", node.Meshes)), (*Solid).Union)
}

type IntersectionNode struct {
	Meshes []nodes.Output[modeling.Mesh] `description:"Solids to overlap. An empty mesh makes the result empty."`
}

func (node IntersectionNode) Description() string {
	return "Keeps only the volume every solid shares."
}

func (node IntersectionNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	inputs := connected(out, "Meshes", node.Meshes)
	if len(nonEmpty(inputs)) < len(inputs) {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}
	runCSG(out, inputs, (*Solid).Intersect)
}

type SubtractNode struct {
	Base   nodes.Output[modeling.Mesh]   `description:"Solid to carve into."`
	Remove []nodes.Output[modeling.Mesh] `description:"Solids to carve away, applied in order. Empty meshes are skipped."`
}

func (node SubtractNode) Description() string {
	return "Carves solids out of another."
}

func (node SubtractNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if node.Base == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	base := input{mesh: nodes.GetOutputValue(out, node.Base), port: "Base"}
	if base.empty() {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}
	runCSG(out, append([]input{base}, nonEmpty(connected(out, "Remove", node.Remove))...), (*Solid).Subtract)
}

type input struct {
	mesh modeling.Mesh
	port string
}

func (in input) empty() bool {
	return in.mesh.Topology() == modeling.TriangleTopology && in.mesh.PrimitiveCount() == 0
}

func nonEmpty(inputs []input) []input {
	kept := make([]input, 0, len(inputs))
	for _, in := range inputs {
		if !in.empty() {
			kept = append(kept, in)
		}
	}
	return kept
}

// Named by the port each mesh arrived on, unconnected ports included in the
// count, so an error points at the port the user can see.
func connected(out *nodes.StructOutput[modeling.Mesh], name string, ports []nodes.Output[modeling.Mesh]) []input {
	inputs := make([]input, 0, len(ports))
	for i, port := range ports {
		if port == nil {
			continue
		}
		inputs = append(inputs, input{
			mesh: nodes.GetOutputValue(out, port),
			port: fmt.Sprintf("%s.%d", name, i),
		})
	}
	return inputs
}

// Applies op to the inputs left to right, naming the port of any that fails.
func runCSG(
	out *nodes.StructOutput[modeling.Mesh],
	inputs []input,
	op func(a, b *Solid) (*Solid, error),
) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	if len(inputs) == 0 {
		out.Set(empty)
		return
	}

	solids := make([]*Solid, 0, len(inputs))
	for _, in := range inputs {
		solid, err := NewSolid(in.mesh)
		if err != nil {
			out.CaptureError(fmt.Errorf("%s: %w", in.port, err))
			continue
		}
		solids = append(solids, solid)
	}
	if len(solids) < len(inputs) {
		out.Set(empty)
		return
	}

	result := solids[0]
	for i, next := range solids[1:] {
		combined, err := op(result, next)
		if err != nil {
			out.CaptureError(fmt.Errorf("combining %s: %w", inputs[i+1].port, err))
			out.Set(empty)
			return
		}
		if err := combined.prepare(); err != nil {
			out.CaptureError(fmt.Errorf("combining %s left an unclosed solid: %w", inputs[i+1].port, err))
			out.Set(combined.Mesh())
			return
		}
		result = combined
	}
	out.Set(result.Mesh())
}
