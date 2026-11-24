package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type TranslateAttribute3DTransformer struct {
	Attribute string
	Amount    vector3.Float64
}

func (tat TranslateAttribute3DTransformer) attribute() string {
	return tat.Attribute
}

func (tat TranslateAttribute3DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(tat, modeling.PositionAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	return TranslateAttribute3D(m, attribute, tat.Amount), nil
}

func TranslateAttribute3D(m modeling.Mesh, attribute string, amount vector3.Float64) modeling.Mesh {
	if err := RequireV3Attribute(m, attribute); err != nil {
		panic(err)
	}

	oldData := m.Float3Attribute(attribute)
	scaledData := make([]vector3.Float64, oldData.Len())
	for i := 0; i < oldData.Len(); i++ {
		scaledData[i] = oldData.At(i).Add(amount)
	}

	return m.SetFloat3Attribute(attribute, scaledData)
}

type TranslateAttribute3DNode struct {
	Attribute nodes.Output[string] `default:"Position"`
	Mesh      nodes.Output[modeling.Mesh]
	Amount    nodes.Output[vector3.Float64]
}

func (ta3dn TranslateAttribute3DNode) Description() string {
	return "Moves one named vector attribute of a mesh by an offset."
}

func (ta3dn TranslateAttribute3DNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if ta3dn.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	mesh := nodes.GetOutputValue(out, ta3dn.Mesh)
	out.Set(TranslateAttribute3D(
		mesh,
		nodes.GetOutputValue(out, ta3dn.Attribute),
		nodes.GetOutputValue(out, ta3dn.Amount),
	))
}

func zipAttribute[T any](out *nodes.Lifted[modeling.Mesh], attribute string, mesh nodes.LiftedPort[modeling.Mesh], amount nodes.LiftedPort[T], apply func(modeling.Mesh, string, T) modeling.Mesh) {
	nodes.Zip2(
		out,
		nodes.LiftedOr(mesh, modeling.EmptyMesh(modeling.TriangleTopology)),
		amount,
		func(m modeling.Mesh, amount T) modeling.Mesh {
			if m.PrimitiveCount() == 0 {
				return m
			}
			return apply(m, attribute, amount)
		},
	)
}

func TranslateAttribute1D(m modeling.Mesh, attribute string, amount float64) modeling.Mesh {
	return m.ModifyFloat1Attribute(attribute, func(_ int, v float64) float64 { return v + amount })
}

func TranslateAttribute2D(m modeling.Mesh, attribute string, amount vector2.Float64) modeling.Mesh {
	return m.ModifyFloat2Attribute(attribute, func(_ int, v vector2.Float64) vector2.Float64 { return v.Add(amount) })
}

func TranslateAttribute4D(m modeling.Mesh, attribute string, amount vector4.Float64) modeling.Mesh {
	return modifyFloat4(m, attribute, func(v vector4.Float64) vector4.Float64 { return v.Add(amount) })
}

func modifyFloat4(m modeling.Mesh, attribute string, f func(vector4.Float64) vector4.Float64) modeling.Mesh {
	if err := RequireV4Attribute(m, attribute); err != nil {
		panic(err)
	}
	data := m.Float4Attribute(attribute)
	modified := make([]vector4.Float64, data.Len())
	for i := range modified {
		modified[i] = f(data.At(i))
	}
	return m.SetFloat4Attribute(attribute, modified)
}

type TranslateAttribute1DNode struct {
	Attribute nodes.Output[string]
	Mesh      nodes.LiftedPort[modeling.Mesh]
	Amount    nodes.LiftedPort[float64]
}

func (n TranslateAttribute1DNode) Description() string {
	return "Offsets one named scalar attribute of a mesh."
}

func (n TranslateAttribute1DNode) Out(out *nodes.Lifted[modeling.Mesh]) {
	zipAttribute(out, nodes.GetOutputValue(out, n.Attribute), n.Mesh, n.Amount, TranslateAttribute1D)
}

type TranslateAttribute2DNode struct {
	Attribute nodes.Output[string] `default:"TexCoord"`
	Mesh      nodes.LiftedPort[modeling.Mesh]
	Amount    nodes.LiftedPort[vector2.Float64]
}

func (n TranslateAttribute2DNode) Description() string {
	return "Moves one named 2D attribute of a mesh by an offset."
}

func (n TranslateAttribute2DNode) Out(out *nodes.Lifted[modeling.Mesh]) {
	zipAttribute(out, nodes.GetOutputValue(out, n.Attribute), n.Mesh, n.Amount, TranslateAttribute2D)
}

type TranslateAttribute4DNode struct {
	Attribute nodes.Output[string]
	Mesh      nodes.LiftedPort[modeling.Mesh]
	Amount    nodes.LiftedPort[vector4.Float64]
}

func (n TranslateAttribute4DNode) Description() string {
	return "Moves one named 4D attribute of a mesh by an offset."
}

func (n TranslateAttribute4DNode) Out(out *nodes.Lifted[modeling.Mesh]) {
	zipAttribute(out, nodes.GetOutputValue(out, n.Attribute), n.Mesh, n.Amount, TranslateAttribute4D)
}
