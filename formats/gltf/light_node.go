package gltf

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/nodes"
)

type LightType int

const (
	LightTypePoint LightType = iota
	LightTypeSpot
	LightTypeDirectional
)

var _ = parameter.DefineEnum(
	enum.Value(LightTypePoint, "Point", "Point lights emit light in all directions from their position in space."),
	enum.Value(LightTypeSpot, "Spot", "Spot lights emit light in a cone in the direction of the local -z axis."),
	enum.Value(LightTypeDirectional, "Directional", "Directional lights act as though they are infinitely far away and emit light in the direction of the local -z axis."),
)

func (t LightType) punctual() KHR_LightsPunctualType {
	switch t {
	case LightTypeSpot:
		return KHR_LightsPunctualType_Spot
	case LightTypeDirectional:
		return KHR_LightsPunctualType_Directional
	}
	return KHR_LightsPunctualType_Point
}

type LightNode struct {
	Type           nodes.Output[LightType]              `default:"Point"`
	Color          nodes.Output[coloring.Color]         `description:"Color as picked on screen (sRGB); written as linear." default:"#ffffff"`
	Intensity      nodes.Output[float64]                `description:"Candela for point and spot lights, lux for directional." default:"1"`
	Range          nodes.Output[float64]                `description:"Distance where intensity reaches zero; 0 is infinite."`
	InnerConeAngle nodes.Output[float64]                `description:"Angle in radians from centre of spotlight where falloff begins." advanced:"true"`
	OuterConeAngle nodes.Output[float64]                `description:"Angle in radians from centre of spotlight where falloff ends." default:"0.7853981633974483" advanced:"true"`
}

func (LightNode) Description() string {
	return "A KHR_lights_punctual light, emitted from the model it is attached to."
}

func (n LightNode) Out(out *nodes.StructOutput[KHR_LightsPunctual]) {
	lightType := nodes.GetOutputValue(out, n.Type).punctual()
	intensity := nodes.GetOutputValue(out, n.Intensity)
	light := KHR_LightsPunctual{
		Type:      lightType,
		Color:     nodes.GetOutputValue(out, n.Color).SRGBToLinear(),
		Intensity: &intensity,
	}

	if r := nodes.GetOutputValue(out, n.Range); r > 0 {
		light.Range = &r
	} else if r < 0 {
		out.CaptureError(fmt.Errorf("range must be greater than 0, or 0 for infinite: got %g", r))
	}

	if lightType == KHR_LightsPunctualType_Spot {
		inner := nodes.GetOutputValue(out, n.InnerConeAngle)
		outer := nodes.GetOutputValue(out, n.OuterConeAngle)
		if inner < 0 || inner >= outer || outer > math.Pi/2 {
			out.CaptureError(fmt.Errorf("cone angles need 0 <= inner < outer <= pi/2: got inner %g, outer %g", inner, outer))
		}
		light.InnerConeAngle = &inner
		light.OuterConeAngle = &outer
	}

	out.Set(light)
}
