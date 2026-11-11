package meshops

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type CropAttribute3DTransformer struct {
	Attribute   string
	BoundingBox geometry.AABB
}

func (cat CropAttribute3DTransformer) attribute() string {
	return cat.Attribute
}

func (cat CropAttribute3DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(cat, modeling.PositionAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	return CropFloat3Attribute(m, attribute, cat.BoundingBox), nil
}

func CropFloat3Attribute(m modeling.Mesh, attr string, boundingBox geometry.AABB) modeling.Mesh {

	check(RequireTopology(m, modeling.PointTopology))

	positions := m.Float3Attribute(attr)
	kept := SelectIndices(positions.Len(), func(i int) bool {
		return boundingBox.Contains(positions.At(i))
	})
	return GatherPoints(m, kept)
}

type CropAttribute3DNode struct {
	Attribute nodes.Output[string] `default:"Position"`
	Mesh      nodes.Output[modeling.Mesh]
	AABB      nodes.Output[geometry.AABB]
}

func (ca3dn CropAttribute3DNode) Description() string {
	return "Deletes vertices whose attribute falls outside the given box."
}

func (ca3dn CropAttribute3DNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if ca3dn.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	mesh := nodes.GetOutputValue(out, ca3dn.Mesh)
	attr := nodes.GetOutputValue(out, ca3dn.Attribute)
	aabb := nodes.GetOutputValue(out, ca3dn.AABB)
	if aabb.Size() == vector3.Zero[float64]() {
		out.Set(mesh)
		return
	}
	out.Set(CropFloat3Attribute(mesh, attr, aabb))
}
