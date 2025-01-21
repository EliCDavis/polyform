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

type ScaleAttributeAlongNormalNode = nodes.Struct[modeling.Mesh, ScaleAttributeAlongNormalNodeData]

type ScaleAttributeAlongNormalNodeData struct {
	Mesh             nodes.NodeOutput[modeling.Mesh]
	Amount           nodes.NodeOutput[float64]
	AttributeToScale nodes.NodeOutput[string]
	NormalAttribute  nodes.NodeOutput[string]
}

func (sa3dn ScaleAttributeAlongNormalNodeData) Process() (modeling.Mesh, error) {

	if sa3dn.Mesh == nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	mesh := sa3dn.Mesh.Value()

	attrToScale := modeling.PositionAttribute
	if sa3dn.AttributeToScale != nil {
		attrToScale = sa3dn.AttributeToScale.Value()
	}

	if !mesh.HasFloat3Attribute(attrToScale) {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	attrNormal := modeling.NormalAttribute
	if sa3dn.NormalAttribute != nil {
		attrNormal = sa3dn.NormalAttribute.Value()
	}

	if !mesh.HasFloat3Attribute(attrNormal) {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	ammount := 0.
	if sa3dn.Amount != nil {
		ammount = sa3dn.Amount.Value()
	}

	return ScaleAttributeAlongNormal(mesh, attrToScale, attrNormal, ammount), nil
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

type ScaleAttribute3DNode = nodes.Struct[modeling.Mesh, ScaleAttribute3DNodeData]

type ScaleAttribute3DNodeData struct {
	Attribute nodes.NodeOutput[string]
	Mesh      nodes.NodeOutput[modeling.Mesh]
	Amount    nodes.NodeOutput[vector3.Float64]
	Origin    nodes.NodeOutput[vector3.Float64]
}

func (sa3dn ScaleAttribute3DNodeData) Process() (modeling.Mesh, error) {
	attr := modeling.PositionAttribute
	if sa3dn.Attribute != nil {
		attr = sa3dn.Attribute.Value()
	}

	origin := vector3.Zero[float64]()
	if sa3dn.Origin != nil {
		origin = sa3dn.Origin.Value()
	}

	return ScaleAttribute3D(sa3dn.Mesh.Value(), attr, origin, sa3dn.Amount.Value()), nil
}
