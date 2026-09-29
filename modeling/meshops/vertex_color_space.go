package meshops

import (
	"fmt"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type VertexColorSpaceTransformation int

const (
	VertexColorSpaceSRGBToLinear VertexColorSpaceTransformation = iota
	VertexColorSpaceLinearToSRGB
)

type VertexColorSpaceTransformer struct {
	Attribute              string
	SkipOnMissingAttribute bool
	Transformation         VertexColorSpaceTransformation
}

func (vcst VertexColorSpaceTransformer) attribute() string {
	return vcst.Attribute
}

func (vcst VertexColorSpaceTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(vcst, modeling.ColorAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		if vcst.SkipOnMissingAttribute {
			return m, nil
		}
		return
	}

	return VertexColorSpace(m, attribute, vcst.Transformation), nil
}

func VertexColorSpace(m modeling.Mesh, attribute string, transformation VertexColorSpaceTransformation) modeling.Mesh {
	if err := RequireV3Attribute(m, attribute); err != nil {
		panic(err)
	}

	oldData := m.Float3Attribute(attribute)
	transferredData := make([]vector3.Float64, oldData.Len())

	switch transformation {
	case VertexColorSpaceLinearToSRGB:
		for i := 0; i < oldData.Len(); i++ {
			entry := oldData.At(i)
			transferredData[i] = vector3.New(
				coloring.LinearToSRGB(entry.X()),
				coloring.LinearToSRGB(entry.Y()),
				coloring.LinearToSRGB(entry.Z()),
			)
		}

	case VertexColorSpaceSRGBToLinear:
		for i := 0; i < oldData.Len(); i++ {
			entry := oldData.At(i)
			transferredData[i] = vector3.New(
				coloring.SRGBToLinear(entry.X()),
				coloring.SRGBToLinear(entry.Y()),
				coloring.SRGBToLinear(entry.Z()),
			)
		}
	}

	return m.SetFloat3Attribute(attribute, transferredData)
}

type SrgbToLinearNode struct {
	Attribute nodes.Output[string]
	Mesh      nodes.Output[modeling.Mesh]
}

func (n SrgbToLinearNode) Description() string {
	return "Converts a color attribute from sRGB to linear space."
}

func (n SrgbToLinearNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	convertVertexColorSpace(out, n.Mesh, n.Attribute, VertexColorSpaceSRGBToLinear)
}

func convertVertexColorSpace(out *nodes.StructOutput[modeling.Mesh], meshIn nodes.Output[modeling.Mesh], attrIn nodes.Output[string], transformation VertexColorSpaceTransformation) {
	if meshIn == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	mesh := nodes.GetOutputValue(out, meshIn)
	attr := nodes.TryGetOutputValue(out, attrIn, modeling.ColorAttribute)
	if mesh.AttributeLength() == 0 {
		out.CaptureError(fmt.Errorf("mesh has no %s vertex data to convert", attr))
		out.Set(mesh)
		return
	}

	out.Set(VertexColorSpace(mesh, attr, transformation))
}

type LinearToSRGBNode struct {
	Attribute nodes.Output[string]
	Mesh      nodes.Output[modeling.Mesh]
}

func (n LinearToSRGBNode) Description() string {
	return "Converts a color attribute from linear space to sRGB."
}

func (n LinearToSRGBNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	convertVertexColorSpace(out, n.Mesh, n.Attribute, VertexColorSpaceLinearToSRGB)
}
