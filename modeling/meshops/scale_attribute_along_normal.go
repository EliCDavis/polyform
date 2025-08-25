package meshops

import (
	"fmt"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

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

func ScaleAttributeAlongNormalWithTexture(
	m modeling.Mesh,
	attributeToScale, normalAttribute, uvAttribute string,
	amount float64,
	texture texturing.Texture[float64],
) modeling.Mesh {
	if err := RequireV3Attribute(m, attributeToScale); err != nil {
		panic(err)
	}

	if err := RequireV3Attribute(m, normalAttribute); err != nil {
		panic(err)
	}

	if err := RequireV2Attribute(m, uvAttribute); err != nil {
		panic(err)
	}

	positionData := m.Float3Attribute(attributeToScale)
	normalData := m.Float3Attribute(normalAttribute)
	uvData := m.Float2Attribute(uvAttribute)

	scaledData := make([]vector3.Float64, positionData.Len())
	for i := 0; i < positionData.Len(); i++ {
		uv := uvData.At(i)
		scaledData[i] = positionData.At(i).Add(normalData.At(i).Scale(amount * texture.UV(uv.X(), uv.Y())))
	}

	return m.SetFloat3Attribute(attributeToScale, scaledData)
}

type ScaleAttributeAlongNormalNode struct {
	Mesh             nodes.Output[modeling.Mesh]
	Amount           nodes.Output[float64]
	AttributeToScale nodes.Output[string]
	NormalAttribute  nodes.Output[string]

	Texture     nodes.Output[texturing.Texture[float64]]
	UvAttribute nodes.Output[string]
}

func (sa3dn ScaleAttributeAlongNormalNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if sa3dn.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	mesh := nodes.GetOutputValue(out, sa3dn.Mesh)

	attrToScale := nodes.TryGetOutputValue(out, sa3dn.AttributeToScale, modeling.PositionAttribute)
	if !mesh.HasFloat3Attribute(attrToScale) {
		out.CaptureError(fmt.Errorf("mesh does not contain %s to scale", attrToScale))
		out.Set(mesh)
		return
	}

	attrNormal := nodes.TryGetOutputValue(out, sa3dn.NormalAttribute, modeling.NormalAttribute)
	if !mesh.HasFloat3Attribute(attrNormal) {
		out.CaptureError(fmt.Errorf("mesh does not contain %s to scale along", attrNormal))
		out.Set(mesh)
		return
	}

	if sa3dn.Texture == nil {
		out.Set(ScaleAttributeAlongNormal(
			mesh,
			attrToScale,
			attrNormal,
			nodes.TryGetOutputValue(out, sa3dn.Amount, 0.),
		))
		return
	}

	attrUV := nodes.TryGetOutputValue(out, sa3dn.UvAttribute, modeling.TexCoordAttribute)
	if !mesh.HasFloat2Attribute(attrUV) {
		out.CaptureError(fmt.Errorf("mesh does not contain %s to scale along", attrUV))
		out.Set(mesh)
		return
	}

	out.Set(ScaleAttributeAlongNormalWithTexture(
		mesh,
		attrToScale,
		attrNormal,
		attrUV,
		nodes.TryGetOutputValue(out, sa3dn.Amount, 1.),
		nodes.GetOutputValue(out, sa3dn.Texture),
	))

}
