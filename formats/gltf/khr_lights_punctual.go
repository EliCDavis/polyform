package gltf

import (
	"image/color"

	"github.com/EliCDavis/vector/vector3"
)

type KHR_LightsPunctualType string

const (
	KHR_LightsPunctualType_Directional KHR_LightsPunctualType = "directional"
	KHR_LightsPunctualType_Point       KHR_LightsPunctualType = "point"
	KHR_LightsPunctualType_Spot        KHR_LightsPunctualType = "spot"
)

type KHR_LightsPunctual struct {
	Type      KHR_LightsPunctualType
	Name      *string
	Color     color.Color
	Intensity *float64
	Range     *float64
	Position  vector3.Float64
}

func (khr_lp KHR_LightsPunctual) ToExtension() map[string]any {
	data := make(map[string]any)

	if khr_lp.Type == "" {
		data["type"] = KHR_LightsPunctualType_Point
	} else {
		data["type"] = khr_lp.Type
	}

	if khr_lp.Color != nil {
		data["color"] = rgbToFloatArr(khr_lp.Color)
	}

	if khr_lp.Range != nil {
		data["range"] = *khr_lp.Range
	}

	if khr_lp.Intensity != nil {
		data["intensity"] = *khr_lp.Intensity
	}

	return data
}
