package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
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

type TranslateAttribute3DNode = nodes.Struct[modeling.Mesh, TranslateAttribute3DNodeData]

type TranslateAttribute3DNodeData struct {
	Attribute nodes.NodeOutput[string]
	Mesh      nodes.NodeOutput[modeling.Mesh]
	Amount    nodes.NodeOutput[vector3.Float64]
}

func (ta3dn TranslateAttribute3DNodeData) Process() (modeling.Mesh, error) {
	attr := modeling.PositionAttribute

	if ta3dn.Attribute != nil {
		attr = ta3dn.Attribute.Value()
	}

	return TranslateAttribute3D(ta3dn.Mesh.Value(), attr, ta3dn.Amount.Value()), nil
}
