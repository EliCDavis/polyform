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
	Meshes []nodes.Output[modeling.Mesh] `description:"Solids to merge. Each must be closed."`
}

func (node UnionNode) Description() string {
	return "Merges solids into the volume enclosed by any of them, removing the geometry buried inside."
}

func (node UnionNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	fold(out, connected(out, "Meshes", node.Meshes), Union)
}

type IntersectionNode struct {
	Meshes []nodes.Output[modeling.Mesh] `description:"Solids to overlap. Each must be closed."`
}

func (node IntersectionNode) Description() string {
	return "Keeps only the volume every solid shares."
}

func (node IntersectionNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	fold(out, connected(out, "Meshes", node.Meshes), Intersect)
}

type SubtractNode struct {
	Base   nodes.Output[modeling.Mesh]   `description:"Solid to carve into."`
	Remove []nodes.Output[modeling.Mesh] `description:"Solids to carve away, applied in order."`
}

func (node SubtractNode) Description() string {
	return "Carves solids out of another, leaving the volume that only the first one covered."
}

func (node SubtractNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if node.Base == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	base := input{mesh: nodes.GetOutputValue(out, node.Base), port: "Base"}
	fold(out, append([]input{base}, connected(out, "Remove", node.Remove)...), Subtract)
}

type input struct {
	mesh modeling.Mesh
	port string
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

func fold(
	out *nodes.StructOutput[modeling.Mesh],
	inputs []input,
	op func(a, b modeling.Mesh) (modeling.Mesh, error),
) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	if len(inputs) == 0 {
		out.Set(empty)
		return
	}

	// Checked here as well as inside the operation so the port that actually
	// carries the bad mesh gets named. Folding pairwise otherwise reports the
	// third input of five as "the second mesh".
	broken := false
	for _, in := range inputs {
		if err := CheckClosed(in.mesh); err != nil {
			out.CaptureError(fmt.Errorf("%s: %w", in.port, err))
			broken = true
		}
	}
	if broken {
		out.Set(empty)
		return
	}

	result := inputs[0].mesh
	for _, next := range inputs[1:] {
		combined, err := op(result, next.mesh)
		if err != nil {
			out.CaptureError(fmt.Errorf("combining %s: %w", next.port, err))
			out.Set(result)
			return
		}
		result = combined
	}
	out.Set(result)
}
