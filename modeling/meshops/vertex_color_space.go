package meshops

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type VertexColorSpaceTransformation int

const (
	VertexColorSpaceSRGBToLinear VertexColorSpaceTransformation = iota
	VertexColorSpaceLinearToSRGB
)

type VertexColorSpaceTransformer struct {
	Attribute      string
	Transformation VertexColorSpaceTransformation
}

func (vcst VertexColorSpaceTransformer) attribute() string {
	return vcst.Attribute
}

func (vcst VertexColorSpaceTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(vcst, modeling.ColorAttribute)

	if err = requireV3Attribute(m, attribute); err != nil {
		return
	}

	return VertexColorSpace(m, attribute, vcst.Transformation), nil
}

// From Three.js
// https://github.com/mrdoob/three.js/blob/e6f7c4e677cb8869502739da2640791d020d8d2f/src/math/ColorManagement.js#L5
func sRGBToLinear(c float64) float64 {
	if c < 0.04045 {
		return c * 0.0773993808
	}
	return math.Pow(c*0.9478672986+0.0521327014, 2.4)
}

// From Three.js
// https://github.com/mrdoob/three.js/blob/e6f7c4e677cb8869502739da2640791d020d8d2f/src/math/ColorManagement.js#L5
func linearToSRGB(c float64) float64 {
	if c < 0.0031308 {
		return c * 12.92
	}
	return 1.055*(math.Pow(c, 0.41666)) - 0.055
}

func VertexColorSpace(m modeling.Mesh, attribute string, transformation VertexColorSpaceTransformation) modeling.Mesh {
	if err := requireV3Attribute(m, attribute); err != nil {
		panic(err)
	}

	oldData := m.Float3Attribute(attribute)
	transferredData := make([]vector3.Float64, oldData.Len())

	switch transformation {
	case VertexColorSpaceLinearToSRGB:
		for i := 0; i < oldData.Len(); i++ {
			entry := oldData.At(i)
			transferredData[i] = vector3.New(
				linearToSRGB(entry.X()),
				linearToSRGB(entry.Y()),
				linearToSRGB(entry.Z()),
			)
		}

	case VertexColorSpaceSRGBToLinear:
		for i := 0; i < oldData.Len(); i++ {
			entry := oldData.At(i)
			transferredData[i] = vector3.New(
				sRGBToLinear(entry.X()),
				sRGBToLinear(entry.Y()),
				sRGBToLinear(entry.Z()),
			)
		}
	}

	return m.SetFloat3Attribute(attribute, transferredData)
}
