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

type ScaleAttribute3DNode = nodes.StructNode[modeling.Mesh, ScaleAttribute3DNodeData]

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
