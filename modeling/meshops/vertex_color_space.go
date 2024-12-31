package meshops

import (
	"github.com/EliCDavis/polyform/math/colors"
	"github.com/EliCDavis/polyform/modeling"
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
				colors.LinearToSRGB(entry.X()),
				colors.LinearToSRGB(entry.Y()),
				colors.LinearToSRGB(entry.Z()),
			)
		}

	case VertexColorSpaceSRGBToLinear:
		for i := 0; i < oldData.Len(); i++ {
			entry := oldData.At(i)
			transferredData[i] = vector3.New(
				colors.SRGBToLinear(entry.X()),
				colors.SRGBToLinear(entry.Y()),
				colors.SRGBToLinear(entry.Z()),
			)
		}
	}

	return m.SetFloat3Attribute(attribute, transferredData)
}
