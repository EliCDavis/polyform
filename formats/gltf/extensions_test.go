package gltf_test

import (
	"image/color"
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/stretchr/testify/assert"
)

func Pointer[T any](v T) *T {
	return &v
}

func TestExtensionsID(t *testing.T) {
	tests := map[string]struct {
		extension gltf.MaterialExtension
		want      string
	}{
		"pbrSpecularGlossiness": {
			extension: gltf.PolyformPbrSpecularGlossiness{},
			want:      "KHR_materials_pbrSpecularGlossiness",
		},
		"transmission": {
			extension: gltf.PolyformTransmission{},
			want:      "KHR_materials_transmission",
		},
		"volume": {
			extension: gltf.PolyformVolume{},
			want:      "KHR_materials_volume",
		},
		"ior": {
			extension: gltf.PolyformIndexOfRefraction{},
			want:      "KHR_materials_ior",
		},
		"specular": {
			extension: gltf.PolyformSpecular{},
			want:      "KHR_materials_specular",
		},
		"unlit": {
			extension: gltf.PolyformUnlit{},
			want:      "KHR_materials_unlit",
		},
		"clearcoat": {
			extension: gltf.PolyformClearcoat{},
			want:      "KHR_materials_clearcoat",
		},
		"emissive": {
			extension: gltf.PolyformEmissiveStrength{},
			want:      "KHR_materials_emissive_strength",
		},
		"iridescence": {
			extension: gltf.PolyformIridescence{},
			want:      "KHR_materials_iridescence",
		},
		"sheen": {
			extension: gltf.PolyformSheen{},
			want:      "KHR_materials_sheen",
		},
		"anisotropy": {
			extension: gltf.PolyformAnisotropy{},
			want:      "KHR_materials_anisotropy",
		},
		"dispersion": {
			extension: gltf.PolyformDispersion{},
			want:      "KHR_materials_dispersion",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.extension.ExtensionID())
		})
	}
}

func TestMaterialExtension_ToExtensionData(t *testing.T) {
	tests := map[string]struct {
		extension gltf.MaterialExtension
		want      map[string]any
	}{
		"SpecularGlossiness/empty": {
			extension: gltf.PolyformPbrSpecularGlossiness{},
			want:      map[string]any{},
		},
		"SpecularGlossiness/diffuseFactor-white": {
			extension: gltf.PolyformPbrSpecularGlossiness{
				DiffuseFactor: color.White,
			},
			want: map[string]any{
				"diffuseFactor": [4]float64{1.0, 1.0, 1.0, 1.0},
			},
		},
		"SpecularGlossiness/diffuseFactor-black": {
			extension: gltf.PolyformPbrSpecularGlossiness{
				DiffuseFactor: color.Black,
			},
			want: map[string]any{
				"diffuseFactor": [4]float64{0.0, 0.0, 0.0, 1.0},
			},
		},
		"SpecularGlossiness/diffuseTexture": {
			extension: gltf.PolyformPbrSpecularGlossiness{
				DiffuseTexture: &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"diffuseTexture": &gltf.TextureInfo{},
			},
		},
		"SpecularGlossiness/specularFactor": {
			extension: gltf.PolyformPbrSpecularGlossiness{
				SpecularFactor: color.White,
			},
			want: map[string]any{
				"specularFactor": [3]float64{1.0, 1.0, 1.0},
			},
		},
		"SpecularGlossiness/glossinessFactor": {
			extension: gltf.PolyformPbrSpecularGlossiness{
				GlossinessFactor: Pointer(1.),
			},
			want: map[string]any{
				"glossinessFactor": 1.,
			},
		},
		"SpecularGlossiness/specularGlossinessTexture": {
			extension: gltf.PolyformPbrSpecularGlossiness{
				SpecularGlossinessTexture: &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"specularGlossinessTexture": &gltf.TextureInfo{},
			},
		},
		"SpecularGlossiness/everything": {
			extension: gltf.PolyformPbrSpecularGlossiness{
				DiffuseFactor:             color.Black,
				SpecularFactor:            color.White,
				DiffuseTexture:            &gltf.PolyformTexture{},
				GlossinessFactor:          Pointer(.5),
				SpecularGlossinessTexture: &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"diffuseFactor":    [4]float64{0., 0., 0., 1.},
				"diffuseTexture":   &gltf.TextureInfo{},
				"specularFactor":   [3]float64{1., 1., 1.},
				"glossinessFactor": 0.5,
				"specularGlossinessTexture": &gltf.TextureInfo{
					Index: 1,
				},
			},
		},
		"Transmission/empty": {
			extension: gltf.PolyformTransmission{},
			want: map[string]any{
				"transmissionFactor": 0.,
			},
		},
		"Transmission/everything": {
			extension: gltf.PolyformTransmission{
				Factor:  1,
				Texture: &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"transmissionFactor":  1.,
				"transmissionTexture": &gltf.TextureInfo{},
			},
		},
		"Iridescence/empty": {
			extension: gltf.PolyformIridescence{},
			want: map[string]any{
				"iridescenceFactor": 0.,
			},
		},
		"Iridescence/iridescenceTexture": {
			extension: gltf.PolyformIridescence{
				IridescenceTexture: &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"iridescenceFactor":  0.,
				"iridescenceTexture": &gltf.TextureInfo{},
			},
		},
		"Iridescence/iridescenceIor": {
			extension: gltf.PolyformIridescence{
				IridescenceIor: Pointer(1.),
			},
			want: map[string]any{
				"iridescenceFactor": 0.,
				"iridescenceIor":    1.,
			},
		},
		"Iridescence/iridescenceThicknessMinimum": {
			extension: gltf.PolyformIridescence{
				IridescenceThicknessMinimum: Pointer(1.),
			},
			want: map[string]any{
				"iridescenceFactor":           0.,
				"iridescenceThicknessMinimum": 1.,
			},
		},
		"Iridescence/iridescenceThicknessMaximum": {
			extension: gltf.PolyformIridescence{
				IridescenceThicknessMaximum: Pointer(1.),
			},
			want: map[string]any{
				"iridescenceFactor":           0.,
				"iridescenceThicknessMaximum": 1.,
			},
		},
		"Iridescence/iridescenceThicknessTexture": {
			extension: gltf.PolyformIridescence{
				IridescenceThicknessTexture: &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"iridescenceFactor":           0.,
				"iridescenceThicknessTexture": &gltf.TextureInfo{},
			},
		},
		"Iridescence/everything": {
			extension: gltf.PolyformIridescence{
				IridescenceFactor:           1,
				IridescenceTexture:          &gltf.PolyformTexture{},
				IridescenceIor:              Pointer(1.),
				IridescenceThicknessMinimum: Pointer(1.),
				IridescenceThicknessMaximum: Pointer(1.),
				IridescenceThicknessTexture: &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"iridescenceFactor":           1.,
				"iridescenceTexture":          &gltf.TextureInfo{},
				"iridescenceIor":              1.,
				"iridescenceThicknessMinimum": 1.,
				"iridescenceThicknessMaximum": 1.,
				"iridescenceThicknessTexture": &gltf.TextureInfo{Index: 1},
			},
		},
		"Sheen/empty": {
			extension: gltf.PolyformSheen{},
			want: map[string]any{
				"sheenRoughnessFactor": 0.,
			},
		},
		"Sheen/sheenColorTexture": {
			extension: gltf.PolyformSheen{
				SheenColorTexture: &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"sheenRoughnessFactor": 0.,
				"sheenColorTexture":    &gltf.TextureInfo{},
			},
		},
		"Sheen/sheenRoughnessTexture": {
			extension: gltf.PolyformSheen{
				SheenRoughnessTexture: &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"sheenRoughnessFactor":  0.,
				"sheenRoughnessTexture": &gltf.TextureInfo{},
			},
		},
		"Sheen/sheenColorFactor": {
			extension: gltf.PolyformSheen{
				SheenColorFactor: color.White,
			},
			want: map[string]any{
				"sheenRoughnessFactor": 0.,
				"sheenColorFactor":     [3]float64{1., 1., 1.},
			},
		},
		"Sheen/everything": {
			extension: gltf.PolyformSheen{
				SheenRoughnessFactor:  1,
				SheenRoughnessTexture: &gltf.PolyformTexture{},
				SheenColorTexture:     &gltf.PolyformTexture{},
				SheenColorFactor:      color.White,
			},
			want: map[string]any{
				"sheenColorFactor":      [3]float64{1., 1., 1.},
				"sheenRoughnessFactor":  1.,
				"sheenColorTexture":     &gltf.TextureInfo{},
				"sheenRoughnessTexture": &gltf.TextureInfo{Index: 1},
			},
		},
		"Anisotropy/empty": {
			extension: gltf.PolyformAnisotropy{},
			want: map[string]any{
				"anisotropyStrength": 0.,
				"anisotropyRotation": 0.,
			},
		},
		"Anisotropy/everything": {
			extension: gltf.PolyformAnisotropy{
				AnisotropyStrength: 0.5,
				AnisotropyRotation: 1,
				AnisotropyTexture:  &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"anisotropyStrength": 0.5,
				"anisotropyRotation": 1.,
				"anisotropyTexture":  &gltf.TextureInfo{},
			},
		},
		"Dispersion/empty": {
			extension: gltf.PolyformDispersion{},
			want: map[string]any{
				"dispersion": 0.,
			},
		},
		"Dispersion/everything": {
			extension: gltf.PolyformDispersion{
				Dispersion: 1,
			},
			want: map[string]any{
				"dispersion": 1.,
			},
		},
		"Unlit/empty": {
			extension: gltf.PolyformUnlit{},
			want:      map[string]any{},
		},
		"EmissiveStrength/empty": {
			extension: gltf.PolyformEmissiveStrength{},
			want:      map[string]any{},
		},
		"EmissiveStrength/everything": {
			extension: gltf.PolyformEmissiveStrength{
				EmissiveStrength: Pointer(1.0),
			},
			want: map[string]any{
				"emissiveStrength": 1.,
			},
		},
		"IndexOfRefraction/empty": {
			extension: gltf.PolyformIndexOfRefraction{},
			want:      map[string]any{},
		},
		"IndexOfRefraction/everything": {
			extension: gltf.PolyformIndexOfRefraction{
				IOR: Pointer(1.0),
			},
			want: map[string]any{
				"ior": 1.,
			},
		},
		"Volume/empty": {
			extension: gltf.PolyformVolume{},
			want: map[string]any{
				"thicknessFactor": 0.,
			},
		},
		"Volume/thicknessTexture": {
			extension: gltf.PolyformVolume{
				ThicknessTexture: &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"thicknessFactor":  0.,
				"thicknessTexture": &gltf.TextureInfo{},
			},
		},
		"Volume/attenuationDistance": {
			extension: gltf.PolyformVolume{
				AttenuationDistance: Pointer(1.),
			},
			want: map[string]any{
				"thicknessFactor":     0.,
				"attenuationDistance": 1.,
			},
		},
		"Volume/attenuationColor": {
			extension: gltf.PolyformVolume{
				AttenuationColor: color.White,
			},
			want: map[string]any{
				"thicknessFactor":  0.,
				"attenuationColor": [3]float64{1., 1., 1.},
			},
		},
		"Volume/everything": {
			extension: gltf.PolyformVolume{
				ThicknessFactor:     1.,
				ThicknessTexture:    &gltf.PolyformTexture{},
				AttenuationDistance: Pointer(1.),
				AttenuationColor:    color.White,
			},
			want: map[string]any{
				"thicknessFactor":     1.,
				"thicknessTexture":    &gltf.TextureInfo{},
				"attenuationDistance": 1.,
				"attenuationColor":    [3]float64{1., 1., 1.},
			},
		},
		"Specular/empty": {
			extension: gltf.PolyformSpecular{},
			want:      map[string]any{},
		},
		"Specular/specularFactor": {
			extension: gltf.PolyformSpecular{
				Factor: Pointer(1.),
			},
			want: map[string]any{
				"specularFactor": 1.0,
			},
		},
		"Specular/specularTexture": {
			extension: gltf.PolyformSpecular{
				Texture: &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"specularTexture": &gltf.TextureInfo{},
			},
		},
		"Specular/specularColorTexture": {
			extension: gltf.PolyformSpecular{
				ColorTexture: &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"specularColorTexture": &gltf.TextureInfo{},
			},
		},
		"Specular/specularColorFactor": {
			extension: gltf.PolyformSpecular{
				ColorFactor: color.White,
			},
			want: map[string]any{
				"specularColorFactor": [3]float64{1., 1., 1.},
			},
		},
		"Specular/everything": {
			extension: gltf.PolyformSpecular{
				Factor:       Pointer(1.),
				Texture:      &gltf.PolyformTexture{},
				ColorFactor:  color.White,
				ColorTexture: &gltf.PolyformTexture{},
			},
			want: map[string]any{
				"specularFactor":       1.0,
				"specularTexture":      &gltf.TextureInfo{},
				"specularColorFactor":  [3]float64{1., 1., 1.},
				"specularColorTexture": &gltf.TextureInfo{Index: 1},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			writer := gltf.NewWriter()
			data := tc.extension.ToExtensionData(writer)

			if !assert.Len(t, data, len(tc.want)) {
				return
			}

			for k, v := range tc.want {
				assert.Equal(t, v, data[k])
			}
		})
	}
}
