package gausops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type ScaleTransformer struct {
	Attribute string
	Scale     vector3.Float64
}

func (st ScaleTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	scaleAttr := getAttribute(st.Attribute, modeling.ScaleAttribute)

	if err = meshops.RequireV3Attribute(m, scaleAttr); err != nil {
		return
	}

	return Scale(m, scaleAttr, st.Scale), nil
}

func Scale(m modeling.Mesh, scaleAttr string, amount vector3.Float64) modeling.Mesh {
	if err := meshops.RequireV3Attribute(m, scaleAttr); err != nil {
		panic(err)
	}

	oldData := m.Float3Attribute(modeling.ScaleAttribute)
	scaledData := make([]vector3.Float64, oldData.Len())
	for i := 0; i < len(scaledData); i++ {
		scaledData[i] = oldData.At(i).Exp().MultByVector(amount).Log()
	}
	return m.SetFloat3Attribute(modeling.ScaleAttribute, scaledData)
}

type ScaleNode = nodes.Struct[modeling.Mesh, ScaleNodeData]

type ScaleNodeData struct {
	Mesh      nodes.NodeOutput[modeling.Mesh]
	Attribute nodes.NodeOutput[string]
	Amount    nodes.NodeOutput[vector3.Float64]
}

func (sa3dn ScaleNodeData) Process() (modeling.Mesh, error) {
	attr := modeling.ScaleAttribute
	if sa3dn.Attribute != nil {
		attr = sa3dn.Attribute.Value()
	}

	return Scale(sa3dn.Mesh.Value(), attr, sa3dn.Amount.Value()), nil
}
