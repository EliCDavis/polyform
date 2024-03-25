package meshops

import (
	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
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

	// Right now we only support point topology
	check(RequireTopology(m, modeling.PointTopology))

	oldV4 := make(map[string]*iter.ArrayIterator[vector4.Float64])
	v4 := make(map[string][]vector4.Float64)
	for _, attr := range m.Float4Attributes() {
		oldV4[attr] = m.Float4Attribute(attr)
		v4[attr] = make([]vector4.Vector[float64], 0)
	}

	oldV3 := make(map[string]*iter.ArrayIterator[vector3.Float64])
	v3 := make(map[string][]vector3.Float64)
	for _, attr := range m.Float3Attributes() {
		oldV3[attr] = m.Float3Attribute(attr)
		v3[attr] = make([]vector3.Vector[float64], 0)
	}

	oldV2 := make(map[string]*iter.ArrayIterator[vector2.Float64])
	v2 := make(map[string][]vector2.Float64)
	for _, attr := range m.Float2Attributes() {
		oldV2[attr] = m.Float2Attribute(attr)
		v2[attr] = make([]vector2.Vector[float64], 0)
	}

	oldV1 := make(map[string]*iter.ArrayIterator[float64])
	v1 := make(map[string][]float64)
	for _, attr := range m.Float1Attributes() {
		oldV1[attr] = m.Float1Attribute(attr)
		v1[attr] = make([]float64, 0)
	}

	decidingAttribute := m.Float3Attribute(attr)
	for i := 0; i < decidingAttribute.Len(); i++ {
		if !boundingBox.Contains(decidingAttribute.At(i)) {
			continue
		}

		for _, attr := range m.Float4Attributes() {
			v4[attr] = append(v4[attr], oldV4[attr].At(i))
		}

		for _, attr := range m.Float3Attributes() {
			v3[attr] = append(v3[attr], oldV3[attr].At(i))
		}

		for _, attr := range m.Float2Attributes() {
			v2[attr] = append(v2[attr], oldV2[attr].At(i))
		}

		for _, attr := range m.Float1Attributes() {
			v1[attr] = append(v1[attr], oldV1[attr].At(i))
		}
	}

	return modeling.NewPointCloud(v4, v3, v2, v1, m.Materials())
}

type CropAttribute3DNode struct {
	nodes.StructData[modeling.Mesh]

	Attribute nodes.NodeOutput[string]
	Mesh      nodes.NodeOutput[modeling.Mesh]
	AABB      nodes.NodeOutput[geometry.AABB]
}

func (ca3dn *CropAttribute3DNode) Out() nodes.NodeOutput[modeling.Mesh] {
	return &nodes.StructNodeOutput[modeling.Mesh]{Definition: ca3dn}
}

func (ca3dn CropAttribute3DNode) Process() (modeling.Mesh, error) {
	attr := modeling.PositionAttribute

	if ca3dn.Attribute != nil {
		attr = ca3dn.Attribute.Data()
	}

	return CropFloat3Attribute(ca3dn.Mesh.Data(), attr, ca3dn.AABB.Data()), nil
}
