package marching

import (
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

// ApplyColorFieldNode samples a colored field's color at each mesh vertex
// and writes it to the mesh's Color attribute.
type ApplyColorFieldNode struct {
	Mesh  nodes.Output[modeling.Mesh]    `description:"The mesh to color. Must already have a Position attribute."`
	Field nodes.Output[sdf.ColoredField] `description:"The colored field to sample for color."`
}

func (n ApplyColorFieldNode) Description() string {
	return "Samples a colored field's color at each mesh vertex and writes it to the mesh's per-vertex Color attribute."
}

func (n ApplyColorFieldNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	mesh := nodes.TryGetOutputValue(out, n.Mesh, modeling.EmptyMesh(modeling.TriangleTopology))
	field := nodes.TryGetOutputValue(out, n.Field, sdf.ColoredField{})

	if !mesh.HasFloat3Attribute(modeling.PositionAttribute) || field.Color == nil {
		out.Set(mesh)
		return
	}

	positions := mesh.Float3Attribute(modeling.PositionAttribute)
	colors := make([]vector3.Float64, positions.Len())
	for i := 0; i < positions.Len(); i++ {
		c := field.Color(positions.At(i))
		colors[i] = vector3.New(c.R, c.G, c.B)
	}

	out.Set(mesh.SetFloat3Attribute(modeling.ColorAttribute, colors))
}
