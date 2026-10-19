package mcp

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/stretchr/testify/assert"
)

func TestRepeatsRecognisesInstancesAndLiftedOutputs(t *testing.T) {
	model := schema.Node{Type: "nodes.Struct[formats/gltf.ModelNode]", AssignedInput: map[string]schema.PortReference{"Gpu Instances": {NodeId: "Node-1", PortName: "Out"}}}
	assert.True(t, repeats(model))
	assert.False(t, repeats(schema.Node{Type: "nodes.Struct[formats/gltf.ModelNode]"}))

	lifted := schema.Node{Type: "nodes.Struct[modeling/extrude.RegionNode]", Output: map[string]schema.NodeOutputPort{"Out": {Type: "[]modeling.Mesh"}}}
	assert.True(t, repeats(lifted))
	assert.False(t, repeats(schema.Node{Type: "nodes.Struct[modeling/extrude.RegionNode]"}))
}

func TestMatchLeafPrefersTypeThenThickness(t *testing.T) {
	leaves := []thinnestLeaf{
		{path: "dog/tail/tip", world: 0.015, primitive: "line"},
		{path: "dog/head/ear", world: 0.02, primitive: "sphere"},
		{path: "dog/torso/chest", world: 0.25, primitive: "sphere"},
	}
	got, ok := matchLeaf(leaves, "sphere", 0.24)
	assert.True(t, ok)
	assert.Equal(t, "dog/torso/chest", got.path, "a chest-sized sphere is the chest, not the ear")

	got, _ = matchLeaf(leaves, "sphere", 0)
	assert.Equal(t, "dog/head/ear", got.path, "with no size to go on, the thinnest of its type")

	got, _ = matchLeaf(leaves, "cone", 0.016)
	assert.Equal(t, "dog/tail/tip", got.path, "an unplanned primitive type falls back to the closest thickness anywhere")
}

func TestRootGeometryCountsMathOnlyWhenItFeedsGeometry(t *testing.T) {
	types := map[string]schema.NodeType{
		"sphere":   {Outputs: map[string]schema.NodeTypeOutput{"Field": {Type: "math/sample.Vec3ToFloat"}}},
		"multiply": {Outputs: map[string]schema.NodeTypeOutput{"Float": {Type: "float64"}}},
		"vec":      {Outputs: map[string]schema.NodeTypeOutput{"Out": {Type: "vector3.Float64"}}},
		"model":    {Outputs: map[string]schema.NodeTypeOutput{"Model": {Type: "gltf.PolyformModel"}}},
	}
	nodes := map[string]schema.Node{
		"sphere":   {Type: "sphere", AssignedInput: map[string]schema.PortReference{"Radius": {NodeId: "multiply", PortName: "Float"}}},
		"multiply": {Type: "multiply"},
		"vec":      {Type: "vec"},
		"model":    {Type: "nodes.Struct[formats/gltf.ModelNode]", AssignedInput: map[string]schema.PortReference{"Translation": {NodeId: "vec", PortName: "Out"}}},
		"dead":     {Type: "multiply"},
	}
	assert.Equal(t, []string{"dead", "multiply", "sphere"}, rootGeometry(nodes, types),
		"the sphere and the multiply that feeds it are geometry; the vector feeding a model is wiring; a dead multiply is clutter")
}
