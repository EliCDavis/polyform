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
	fold(out, nodes.GetOutputValues(out, node.Meshes), meshPort("Meshes"), Union)
}

type IntersectionNode struct {
	Meshes []nodes.Output[modeling.Mesh] `description:"Solids to overlap. Each must be closed."`
}

func (node IntersectionNode) Description() string {
	return "Keeps only the volume every solid shares."
}

func (node IntersectionNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	fold(out, nodes.GetOutputValues(out, node.Meshes), meshPort("Meshes"), Intersect)
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

	label := func(i int) string {
		if i == 0 {
			return "Base"
		}
		return fmt.Sprintf("Remove.%d", i-1)
	}

	fold(out, append(
		[]modeling.Mesh{nodes.GetOutputValue(out, node.Base)},
		nodes.GetOutputValues(out, node.Remove)...,
	), label, Subtract)
}

func meshPort(name string) func(int) string {
	return func(i int) string { return fmt.Sprintf("%s.%d", name, i) }
}

func fold(
	out *nodes.StructOutput[modeling.Mesh],
	meshes []modeling.Mesh,
	label func(int) string,
	op func(a, b modeling.Mesh) (modeling.Mesh, error),
) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	if len(meshes) == 0 {
		out.Set(empty)
		return
	}

	// Checked here as well as inside the operation so the port that actually
	// carries the bad mesh gets named. Folding pairwise otherwise reports the
	// third input of five as "the second mesh".
	broken := false
	for i, m := range meshes {
		if err := CheckClosed(m); err != nil {
			out.CaptureError(fmt.Errorf("%s: %w", label(i), err))
			broken = true
		}
	}
	if broken {
		out.Set(empty)
		return
	}

	result := meshes[0]
	for i, next := range meshes[1:] {
		combined, err := op(result, next)
		if err != nil {
			out.CaptureError(fmt.Errorf("combining %s: %w", label(i+1), err))
			out.Set(result)
			return
		}
		result = combined
	}
	out.Set(result)
}
