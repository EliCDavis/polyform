package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type ScaleAttribute3DTransformer struct {
	Attribute string
	Origin    vector3.Float64
	Amount    vector3.Float64
}

func (st ScaleAttribute3DTransformer) attribute() string {
	return st.Attribute
}

func (st ScaleAttribute3DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(st, modeling.PositionAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	return ScaleAttribute3D(m, attribute, st.Origin, st.Amount), nil
}

func ScaleAttribute3D(m modeling.Mesh, attribute string, origin, amount vector3.Float64) modeling.Mesh {
	if err := RequireV3Attribute(m, attribute); err != nil {
		panic(err)
	}

	oldData := m.Float3Attribute(attribute)
	scaledData := make([]vector3.Float64, oldData.Len())
	for i := 0; i < oldData.Len(); i++ {
		scaledData[i] = origin.Add(oldData.At(i).Sub(origin).MultByVector(amount))
	}

	return m.SetFloat3Attribute(attribute, scaledData)
}

// ============================================================================

type ScaleAttributeAlongNormalTransformer struct {
	AttributeToScale string
	NormalAttribute  string
	Amount           float64
}

func (st ScaleAttributeAlongNormalTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := fallbackAttribute(st.AttributeToScale, modeling.PositionAttribute)
	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	normalAttribute := fallbackAttribute(st.NormalAttribute, modeling.NormalAttribute)
	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	return ScaleAttributeAlongNormal(m, attribute, normalAttribute, st.Amount), nil
}

func ScaleAttributeAlongNormal(m modeling.Mesh, attributeToScale, normalAttribute string, amount float64) modeling.Mesh {
	if err := RequireV3Attribute(m, attributeToScale); err != nil {
		panic(err)
	}

	if err := RequireV3Attribute(m, normalAttribute); err != nil {
		panic(err)
	}

	positionData := m.Float3Attribute(attributeToScale)
	normalData := m.Float3Attribute(normalAttribute)
	scaledData := make([]vector3.Float64, positionData.Len())
	for i := 0; i < positionData.Len(); i++ {
		scaledData[i] = positionData.At(i).Add(normalData.At(i).Scale(amount))
	}

	return m.SetFloat3Attribute(attributeToScale, scaledData)
}

type ScaleAttributeAlongNormalNode = nodes.Struct[ScaleAttributeAlongNormalNodeData]

type ScaleAttributeAlongNormalNodeData struct {
	Mesh             nodes.Output[modeling.Mesh]
	Amount           nodes.Output[float64]
	AttributeToScale nodes.Output[string]
	NormalAttribute  nodes.Output[string]
}

func (sa3dn ScaleAttributeAlongNormalNodeData) Out() nodes.StructOutput[modeling.Mesh] {

	if sa3dn.Mesh == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
	}

	mesh := sa3dn.Mesh.Value()

	attrToScale := modeling.PositionAttribute
	if sa3dn.AttributeToScale != nil {
		attrToScale = sa3dn.AttributeToScale.Value()
	}

	if !mesh.HasFloat3Attribute(attrToScale) {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
	}

	attrNormal := modeling.NormalAttribute
	if sa3dn.NormalAttribute != nil {
		attrNormal = sa3dn.NormalAttribute.Value()
	}

	if !mesh.HasFloat3Attribute(attrNormal) {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
	}

	ammount := 0.
	if sa3dn.Amount != nil {
		ammount = sa3dn.Amount.Value()
	}

	return nodes.NewStructOutput(ScaleAttributeAlongNormal(mesh, attrToScale, attrNormal, ammount))
}

// ============================================================================

type ScaleAttribute2DTransformer struct {
	Attribute string
	Origin    vector2.Float64
	Amount    vector2.Float64
}

func (st ScaleAttribute2DTransformer) attribute() string {
	return st.Attribute
}

func (st ScaleAttribute2DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(st, modeling.TexCoordAttribute)

	if err = RequireV2Attribute(m, attribute); err != nil {
		return
	}

	return ScaleAttribute2D(m, attribute, st.Origin, st.Amount), nil
}

func ScaleAttribute2D(m modeling.Mesh, attribute string, origin, amount vector2.Float64) modeling.Mesh {
	if err := RequireV2Attribute(m, attribute); err != nil {
		panic(err)
	}

	oldData := m.Float2Attribute(attribute)
	scaledData := make([]vector2.Float64, oldData.Len())
	for i := 0; i < oldData.Len(); i++ {
		scaledData[i] = origin.Add(oldData.At(i).Sub(origin).MultByVector(amount))
	}

	return m.SetFloat2Attribute(attribute, scaledData)
}

type ScaleAttribute3DNode = nodes.Struct[ScaleAttribute3DNodeData]

type ScaleAttribute3DNodeData struct {
	Attribute nodes.Output[string]
	Mesh      nodes.Output[modeling.Mesh]
	Amount    nodes.Output[vector3.Float64]
	Origin    nodes.Output[vector3.Float64]
}

func (sa3dn ScaleAttribute3DNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	attr := modeling.PositionAttribute
	if sa3dn.Attribute != nil {
		attr = sa3dn.Attribute.Value()
	}

	origin := vector3.Zero[float64]()
	if sa3dn.Origin != nil {
		origin = sa3dn.Origin.Value()
	}

	return nodes.NewStructOutput(ScaleAttribute3D(sa3dn.Mesh.Value(), attr, origin, sa3dn.Amount.Value()))
}
