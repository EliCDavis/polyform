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

type TranslateAttribute3DNode = nodes.Struct[TranslateAttribute3DNodeData]

type TranslateAttribute3DNodeData struct {
	Attribute nodes.Output[string]
	Mesh      nodes.Output[modeling.Mesh]
	Amount    nodes.Output[vector3.Float64]
}

func (ta3dn TranslateAttribute3DNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	if ta3dn.Mesh == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
	}

	return nodes.NewStructOutput(TranslateAttribute3D(
		ta3dn.Mesh.Value(),
		nodes.TryGetOutputValue(ta3dn.Attribute, modeling.PositionAttribute),
		nodes.TryGetOutputValue(ta3dn.Amount, vector3.Zero[float64]()),
	))
}
