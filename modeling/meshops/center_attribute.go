package meshops

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type CenterAttribute3DTransformer struct {
	Attribute string
}

func (cat CenterAttribute3DTransformer) attribute() string {
	return cat.Attribute
}

func (cat CenterAttribute3DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(cat, modeling.PositionAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	return CenterFloat3Attribute(m, attribute), nil
}

func CenterFloat3Attribute(m modeling.Mesh, attr string) modeling.Mesh {
	check(RequireV3Attribute(m, attr))
	oldData := m.Float3Attribute(attr)
	modified := make([]vector3.Float64, oldData.Len())

	min := vector3.New(math.Inf(1), math.Inf(1), math.Inf(1))
	max := vector3.New(math.Inf(-1), math.Inf(-1), math.Inf(-1))
	for i := 0; i < oldData.Len(); i++ {
		v := oldData.At(i)
		min = vector3.Min(min, v)
		max = vector3.Max(max, v)
	}

	center := min.Midpoint(max)
	for i := 0; i < oldData.Len(); i++ {
		modified[i] = oldData.At(i).Sub(center)
	}

	return m.SetFloat3Attribute(attr, modified)
}

type CenterAttribute3DNode struct {
	Attribute nodes.Output[string]
	Mesh      nodes.Output[modeling.Mesh]
}

func (ca3dn CenterAttribute3DNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if ca3dn.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	out.Set(CenterFloat3Attribute(
		nodes.GetOutputValue(out, ca3dn.Mesh),
		nodes.TryGetOutputValue(out, ca3dn.Attribute, modeling.PositionAttribute),
	))
}
