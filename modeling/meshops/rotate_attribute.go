package meshops

import (
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type RotateAttribute3DTransformer struct {
	Attribute string
	Amount    quaternion.Quaternion
}

func (rat RotateAttribute3DTransformer) attribute() string {
	return rat.Attribute
}

func (rat RotateAttribute3DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(rat, modeling.PositionAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	return RotateAttribute3D(m, attribute, rat.Amount), nil
}

func RotateAttribute3D(m modeling.Mesh, attribute string, q quaternion.Quaternion) modeling.Mesh {
	if err := RequireV3Attribute(m, attribute); err != nil {
		panic(err)
	}

	return m.ModifyFloat3AttributeParallel(attribute, func(_ int, v vector3.Float64) vector3.Float64 {
		return q.Rotate(v)
	})
}

type RotateAttribute3DNode struct {
	Attribute nodes.Output[string] `default:"Position"`
	Mesh      nodes.Output[modeling.Mesh]
	Amount    nodes.Output[quaternion.Quaternion]
}

func (ra3dn RotateAttribute3DNode) Description() string {
	return "Rotates one named vector attribute of a mesh. Use Transform to rotate a whole mesh including its normals."
}

func (ra3dn RotateAttribute3DNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if ra3dn.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	mesh := nodes.GetOutputValue(out, ra3dn.Mesh)
	attr := nodes.GetOutputValue(out, ra3dn.Attribute)
	ammount := nodes.GetOutputValue(out, ra3dn.Amount)
	out.Set(RotateAttribute3D(mesh, attr, ammount))
}
