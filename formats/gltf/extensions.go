package gltf

import "image/color"

type MaterialExtension interface {
	ExtensionID() string
	ToExtensionData(w *Writer) map[string]any
}

// https://kcoley.github.io/glTF/extensions/2.0/Khronos/KHR_materials_pbrSpecularGlossiness/
type PolyformPbrSpecularGlossiness struct {
	// The RGBA components of the reflected diffuse color of the material.
	// Metals have a diffuse value of [0.0, 0.0, 0.0]. The fourth component (A)
	// is the opacity of the material. The values are linear.
	DiffuseFactor color.Color

	// The diffuse texture. This texture contains RGB(A) components of the
	// reflected diffuse color of the material in sRGB color space. If the
	// fourth component (A) is present, it represents the alpha coverage of the
	// material. Otherwise, an alpha of 1.0 is assumed. The alphaMode property
	// specifies how alpha is interpreted. The stored texels must not be
	// premultiplied.
	DiffuseTexture *PolyformTexture

	// The specular RGB color of the material. This value is linear.
	SpecularFactor color.Color

	// The glossiness or smoothness of the material. A value of 1.0 means the
	// material has full glossiness or is perfectly smooth. A value of 0.0
	// means the material has no glossiness or is perfectly rough. This value
	// is linear.
	// Default value is 1.0
	GlossinessFactor float64

	// The specular-glossiness texture is a RGBA texture, containing the
	// specular color (RGB) in sRGB space and the glossiness value (A) in
	// linear space.
	SpecularGlossinessTexture *PolyformTexture
}

func (ppsg PolyformPbrSpecularGlossiness) ExtensionID() string {
	return "KHR_materials_pbrSpecularGlossiness"
}

func (sg PolyformPbrSpecularGlossiness) ToExtensionData(w *Writer) map[string]any {
	metadata := make(map[string]any)
	if sg.DiffuseFactor != nil {
		metadata["diffuseFactor"] = rgbaToFloatArr(sg.DiffuseFactor)
	}

	if sg.DiffuseTexture != nil {
		metadata["diffuseTexture"] = w.AddTexture(*sg.DiffuseTexture)
	}

	if sg.SpecularFactor != nil {
		metadata["specularFactor"] = rgbToFloatArr(sg.SpecularFactor)
	}

	metadata["glossinessFactor"] = sg.GlossinessFactor

	if sg.SpecularGlossinessTexture != nil {
		metadata["specularGlossinessTexture"] = w.AddTexture(*sg.SpecularGlossinessTexture)
	}
	return metadata
}

// https://github.com/KhronosGroup/glTF/blob/main/extensions/2.0/Khronos/KHR_materials_transmission/README.md
type PolyformTransmission struct {

	// The base percentage of light that is transmitted through the surface.
	// Default: 0.0
	TransmissionFactor float64

	// A texture that defines the transmission percentage of the surface,
	// stored in the R channel. This will be multiplied by transmissionFactor.
	TransmissionTexture *PolyformTexture
}

func (tr PolyformTransmission) ExtensionID() string {
	return "KHR_materials_transmission"
}

func (tr PolyformTransmission) ToExtensionData(w *Writer) map[string]any {
	metadata := make(map[string]any)

	metadata["transmissionFactor"] = tr.TransmissionFactor

	if tr.TransmissionTexture != nil {
		metadata["transmissionTexture"] = w.AddTexture(*tr.TransmissionTexture)
	}

	return metadata
}

// https://github.com/KhronosGroup/glTF/blob/main/extensions/2.0/Khronos/KHR_materials_volume/README.md
type PolyformVolume struct {
	// The thickness of the volume beneath the surface. The value is given in
	// the coordinate space of the mesh. If the value is 0 the material is
	// thin-walled. Otherwise the material is a volume boundary. The
	// doubleSided property has no effect on volume boundaries. Range is
	// [0, +inf).
	// Default: 0
	ThicknessFactor float64

	// A texture that defines the thickness, stored in the G channel. This will
	// be multiplied by thicknessFactor. Range is [0, 1].
	ThicknessTexture *PolyformTexture

	// Density of the medium given as the average distance that light travels
	// in the medium before interacting with a particle. The value is given in
	// world space. Range is (0, +inf).
	// Default: +Infinity
	AttenuationDistance *float64

	// The color that white light turns into due to absorption when reaching
	// the attenuation distance.
	AttenuationColor color.Color
}

func (v PolyformVolume) ExtensionID() string {
	return "KHR_materials_volume"
}

func (v PolyformVolume) ToExtensionData(w *Writer) map[string]any {
	metadata := make(map[string]any)

	metadata["thicknessFactor"] = v.ThicknessFactor

	if v.ThicknessTexture != nil {
		metadata["thicknessTexture"] = w.AddTexture(*v.ThicknessTexture)
	}

	if v.AttenuationDistance != nil {
		metadata["attenuationDistance"] = *v.AttenuationDistance
	}

	if v.AttenuationColor != nil {
		metadata["attenuationColor"] = rgbToFloatArr(v.AttenuationColor)
	}

	return metadata
}

// https://github.com/KhronosGroup/glTF/blob/main/extensions/2.0/Khronos/KHR_materials_ior/README.md
type PolyformIndexOfRefraction struct {
	// The index of refraction
	// Air	           1.0
	// Water     	   1.33
	// Eyes	           1.38
	// Window Glass	   1.52
	// Sapphire	       1.76
	// Diamond	       2.42
	IOR float64
}

func (sg PolyformIndexOfRefraction) ExtensionID() string {
	return "KHR_materials_ior"
}

func (sg PolyformIndexOfRefraction) ToExtensionData(w *Writer) map[string]any {
	return map[string]any{
		"ior": sg.IOR,
	}
}

// https://github.com/KhronosGroup/glTF/blob/main/extensions/2.0/Khronos/KHR_materials_specular/README.md
type PolyformSpecular struct {
	// The strength of the specular reflection.
	// Default: 1.0
	SpecularFactor float64

	// A texture that defines the strength of the specular reflection, stored
	// in the alpha (A) channel. This will be multiplied by specularFactor.
	SpecularTexture *PolyformTexture

	// The F0 color of the specular reflection (linear RGB).
	SpecularColorFactor color.Color

	// 	A texture that defines the F0 color of the specular reflection, stored
	// in the RGB channels and encoded in sRGB. This texture will be multiplied
	// by specularColorFactor.
	SpecularColorTexture *PolyformTexture
}

func (ps PolyformSpecular) ExtensionID() string {
	return "KHR_materials_specular"
}

func (ps PolyformSpecular) ToExtensionData(w *Writer) map[string]any {
	metadata := make(map[string]any)

	metadata["specularFactor"] = ps.SpecularFactor

	if ps.SpecularTexture != nil {
		metadata["specularTexture"] = w.AddTexture(*ps.SpecularTexture)
	}

	if ps.SpecularColorFactor != nil {
		metadata["specularColorFactor"] = rgbToFloatArr(ps.SpecularColorFactor)
	}

	if ps.SpecularColorTexture != nil {
		metadata["specularColorTexture"] = w.AddTexture(*ps.SpecularColorTexture)
	}

	return metadata
}

type PolyformMaterialsUnlit struct {
}

func (ps PolyformMaterialsUnlit) ExtensionID() string {
	return "KHR_materials_unlit"
}

func (ps PolyformMaterialsUnlit) ToExtensionData(w *Writer) map[string]any {
	return make(map[string]any)
}
