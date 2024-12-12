package gltf

import (
	"image/color"

	"github.com/EliCDavis/vector/vector2"
)

type MaterialExtension interface {
	ExtensionID() string
	ToMaterialExtensionData(w *Writer) map[string]any
}

type TextureExtension interface {
	ExtensionID() string
	ToTextureExtensionData(w *Writer) map[string]any
	IsRequired() bool

	// indicates if the extension should be applied to Texture or TextureInfo object
	IsInfo() bool
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
	GlossinessFactor *float64

	// The specular-glossiness texture is a RGBA texture, containing the
	// specular color (RGB) in sRGB space and the glossiness value (A) in
	// linear space.
	SpecularGlossinessTexture *PolyformTexture
}

func (ppsg PolyformPbrSpecularGlossiness) ExtensionID() string {
	return "KHR_materials_pbrSpecularGlossiness"
}

func (sg PolyformPbrSpecularGlossiness) ToMaterialExtensionData(w *Writer) map[string]any {
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

	if sg.GlossinessFactor != nil {
		metadata["glossinessFactor"] = *sg.GlossinessFactor
	}

	if sg.SpecularGlossinessTexture != nil {
		metadata["specularGlossinessTexture"] = w.AddTexture(*sg.SpecularGlossinessTexture)
	}
	return metadata
}

// https://github.com/KhronosGroup/glTF/blob/main/extensions/2.0/Khronos/KHR_materials_transmission/README.md
type PolyformTransmission struct {

	// The base percentage of light that is transmitted through the surface.
	// Default: 0.0
	Factor float64

	// A texture that defines the transmission percentage of the surface,
	// stored in the R channel. This will be multiplied by transmissionFactor.
	Texture *PolyformTexture
}

func (tr PolyformTransmission) ExtensionID() string {
	return "KHR_materials_transmission"
}

func (tr PolyformTransmission) ToMaterialExtensionData(w *Writer) map[string]any {
	metadata := make(map[string]any)

	metadata["transmissionFactor"] = tr.Factor

	if tr.Texture != nil {
		metadata["transmissionTexture"] = w.AddTexture(*tr.Texture)
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

func (v PolyformVolume) ToMaterialExtensionData(w *Writer) map[string]any {
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

// KHR_materials_ior ==========================================================

// https://github.com/KhronosGroup/glTF/blob/main/extensions/2.0/Khronos/KHR_materials_ior/README.md
type PolyformIndexOfRefraction struct {
	// The index of refraction
	// Air	           1.0
	// Water     	   1.33
	// Eyes	           1.38
	// Window Glass	   1.52
	// Sapphire	       1.76
	// Diamond	       2.42
	IOR *float64
}

func (sg PolyformIndexOfRefraction) ExtensionID() string {
	return "KHR_materials_ior"
}

func (sg PolyformIndexOfRefraction) ToMaterialExtensionData(w *Writer) map[string]any {
	if sg.IOR == nil {
		return map[string]any{}
	}
	return map[string]any{
		"ior": *sg.IOR,
	}
}

// KHR_materials_specular =====================================================

// https://github.com/KhronosGroup/glTF/blob/main/extensions/2.0/Khronos/KHR_materials_specular/README.md
type PolyformSpecular struct {
	// The strength of the specular reflection.
	// Default: 1.0
	Factor *float64

	// A texture that defines the strength of the specular reflection, stored
	// in the alpha (A) channel. This will be multiplied by specularFactor.
	Texture *PolyformTexture

	// The F0 color of the specular reflection (linear RGB).
	ColorFactor color.Color

	// 	A texture that defines the F0 color of the specular reflection, stored
	// in the RGB channels and encoded in sRGB. This texture will be multiplied
	// by specularColorFactor.
	ColorTexture *PolyformTexture
}

func (ps PolyformSpecular) ExtensionID() string {
	return "KHR_materials_specular"
}

func (ps PolyformSpecular) ToMaterialExtensionData(w *Writer) map[string]any {
	metadata := make(map[string]any)

	if ps.Factor != nil {
		metadata["specularFactor"] = *ps.Factor
	}

	if ps.Texture != nil {
		metadata["specularTexture"] = w.AddTexture(*ps.Texture)
	}

	if ps.ColorFactor != nil {
		metadata["specularColorFactor"] = rgbToFloatArr(ps.ColorFactor)
	}

	if ps.ColorTexture != nil {
		metadata["specularColorTexture"] = w.AddTexture(*ps.ColorTexture)
	}

	return metadata
}

type PolyformUnlit struct {
}

func (ps PolyformUnlit) ExtensionID() string {
	return "KHR_materials_unlit"
}

func (ps PolyformUnlit) ToMaterialExtensionData(w *Writer) map[string]any {
	return make(map[string]any)
}

// KHR_materials_clearcoat ====================================================

type PolyformClearcoat struct {
	ClearcoatFactor           float64
	ClearcoatTexture          *PolyformTexture
	ClearcoatRoughnessFactor  float64
	ClearcoatRoughnessTexture *PolyformTexture
	ClearcoatNormalTexture    *PolyformNormal
}

func (pmc PolyformClearcoat) ExtensionID() string {
	return "KHR_materials_clearcoat"
}

func (pmc PolyformClearcoat) ToMaterialExtensionData(w *Writer) map[string]any {
	metadata := make(map[string]any)

	metadata["clearcoatFactor"] = pmc.ClearcoatFactor
	if pmc.ClearcoatTexture != nil {
		metadata["clearcoatTexture"] = w.AddTexture(*pmc.ClearcoatTexture)
	}

	metadata["clearcoatRoughnessFactor"] = pmc.ClearcoatRoughnessFactor
	if pmc.ClearcoatRoughnessTexture != nil {
		metadata["clearcoatRoughnessTexture"] = w.AddTexture(*pmc.ClearcoatRoughnessTexture)
	}

	// if pmc.ClearcoatNormalTexture != nil {
	// 	metadata["clearcoatNormalTexture"] = w.AddTexture(*pmc.ClearcoatNormalTexture)
	// }

	return metadata
}

// KHR_materials_emissive_strength ============================================

// glTF extension that adjusts the strength of emissive material properties.
type PolyformEmissiveStrength struct {
	// The strength adjustment to be multiplied with the material's emissive value.
	EmissiveStrength *float64
}

func (pmes PolyformEmissiveStrength) ExtensionID() string {
	return "KHR_materials_emissive_strength"
}

func (pmes PolyformEmissiveStrength) ToMaterialExtensionData(w *Writer) map[string]any {
	metadata := make(map[string]any)

	if pmes.EmissiveStrength != nil {
		metadata["emissiveStrength"] = *pmes.EmissiveStrength
	}

	return metadata
}

// KHR_materials_iridescence ==================================================

// glTF extension that defines an iridescence effect
type PolyformIridescence struct {
	// The iridescence intensity factor
	IridescenceFactor float64

	// The iridescence intensity texture. The values are sampled from the R
	// channel. These values are linear. If a texture is not given, a value
	// of `1.0` **MUST** be assumed. If other channels are present (GBA), they
	// are ignored for iridescence intensity calculations
	IridescenceTexture *PolyformTexture

	// The index of refraction of the dielectric thin-film layer.
	IridescenceIor *float64

	// The minimum thickness of the thin-film layer given in nanometers. The
	// value **MUST** be less than or equal to the value of
	// `iridescenceThicknessMaximum`.
	IridescenceThicknessMinimum *float64

	// The maximum thickness of the thin-film layer given in nanometers. The
	// value **MUST** be greater than or equal to the value of
	// `iridescenceThicknessMinimum`.
	IridescenceThicknessMaximum *float64

	// The thickness texture of the thin-film layer to linearly interpolate
	// between the minimum and maximum thickness given by the corresponding
	// properties, where a sampled value of `0.0` represents the minimum
	// thickness and a sampled value of `1.0` represents the maximum thickness.
	// The values are sampled from the G channel. These values are linear. If a
	// texture is not given, the maximum thickness **MUST** be assumed. If
	// other channels are present (RBA), they are ignored for thickness
	// calculations.
	IridescenceThicknessTexture *PolyformTexture
}

func (pmi PolyformIridescence) ExtensionID() string {
	return "KHR_materials_iridescence"
}

func (pmi PolyformIridescence) ToMaterialExtensionData(w *Writer) map[string]any {
	metadata := make(map[string]any)

	metadata["iridescenceFactor"] = pmi.IridescenceFactor

	if pmi.IridescenceTexture != nil {
		metadata["iridescenceTexture"] = w.AddTexture(*pmi.IridescenceTexture)
	}

	if pmi.IridescenceIor != nil {
		metadata["iridescenceIor"] = *pmi.IridescenceIor
	}

	if pmi.IridescenceThicknessMinimum != nil {
		metadata["iridescenceThicknessMinimum"] = *pmi.IridescenceThicknessMinimum
	}

	if pmi.IridescenceThicknessMaximum != nil {
		metadata["iridescenceThicknessMaximum"] = *pmi.IridescenceThicknessMaximum
	}

	if pmi.IridescenceThicknessTexture != nil {
		metadata["iridescenceThicknessTexture"] = w.AddTexture(*pmi.IridescenceThicknessTexture)
	}

	return metadata
}

// KHR_materials_sheen ========================================================

// glTF extension that defines the sheen material model.
type PolyformSheen struct {
	// Color of the sheen layer (in linear space).
	SheenColorFactor color.Color

	// The sheen color (RGB) texture. Stored in channel RGB, the sheen color is
	// in sRGB transfer function.
	SheenColorTexture *PolyformTexture

	// The sheen layer roughness of the material.
	SheenRoughnessFactor float64

	// The sheen roughness (Alpha) texture. Stored in alpha channel, the
	// roughness value is in linear space.
	SheenRoughnessTexture *PolyformTexture
}

func (ps PolyformSheen) ExtensionID() string {
	return "KHR_materials_sheen"
}

func (ps PolyformSheen) ToMaterialExtensionData(w *Writer) map[string]any {
	metadata := make(map[string]any)

	if ps.SheenColorFactor != nil {
		metadata["sheenColorFactor"] = rgbToFloatArr(ps.SheenColorFactor)
	}

	if ps.SheenColorTexture != nil {
		metadata["sheenColorTexture"] = w.AddTexture(*ps.SheenColorTexture)
	}

	metadata["sheenRoughnessFactor"] = ps.SheenRoughnessFactor

	if ps.SheenRoughnessTexture != nil {
		metadata["sheenRoughnessTexture"] = w.AddTexture(*ps.SheenRoughnessTexture)
	}

	return metadata
}

// KHR_materials_anisotropy ===================================================

// glTF extension that defines anisotropy
type PolyformAnisotropy struct {
	// The anisotropy strength. When the anisotropy texture is present, this
	// value is multiplied by the texture's blue channel.
	AnisotropyStrength float64

	// The rotation of the anisotropy in tangent, bitangent space, measured in
	// radians counter-clockwise from the tangent. When the anisotropy texture
	// is present, this value provides additional rotation to the vectors in
	// the texture.
	AnisotropyRotation float64

	// The anisotropy texture. Red and green channels represent the anisotropy
	// direction in $[-1, 1]$ tangent, bitangent space, to be rotated by the
	// anisotropy rotation. The blue channel contains strength as $[0, 1]$ to
	// be multiplied by the anisotropy strength.
	AnisotropyTexture *PolyformTexture
}

func (pa PolyformAnisotropy) ExtensionID() string {
	return "KHR_materials_anisotropy"
}

func (pa PolyformAnisotropy) ToMaterialExtensionData(w *Writer) map[string]any {
	metadata := make(map[string]any)

	metadata["anisotropyStrength"] = pa.AnisotropyStrength
	metadata["anisotropyRotation"] = pa.AnisotropyRotation

	if pa.AnisotropyTexture != nil {
		metadata["anisotropyTexture"] = w.AddTexture(*pa.AnisotropyTexture)
	}

	return metadata
}

// KHR_materials_dispersion ===================================================

// glTF extension that defines the strength of dispersion.
type PolyformDispersion struct {
	// This parameter defines dispersion in terms of the 20/Abbe number
	// formulation.
	Dispersion float64
}

func (pd PolyformDispersion) ExtensionID() string {
	return "KHR_materials_dispersion"
}

func (pd PolyformDispersion) ToMaterialExtensionData(w *Writer) map[string]any {
	metadata := make(map[string]any)
	metadata["dispersion"] = pd.Dispersion
	return metadata
}

// KHR_texture_transform ===================================================
// https://github.com/KhronosGroup/glTF/tree/main/extensions/2.0/Khronos/KHR_texture_transform

var _ TextureExtension = PolyformTextureTransform{}

// PolyformTextureTransform is a glTF extension that defines texture transformations.
type PolyformTextureTransform struct {
	// Whether to make this extension required for the given model.
	// If set - extension will be added to `extensionsRequired` list at the GLTF file top level.
	Required bool

	Offset   *vector2.Float64 // The offset of the UV coordinate origin as a factor of the texture dimensions.
	Rotation *float64         // Rotate the UVs by this many radians counter-clockwise around the origin. This is equivalent to a similar rotation of the image clockwise.
	Scale    *vector2.Float64 // The scale factor applied to the components of the UV coordinates.
	TexCoord *int             // Overrides the textureInfo texCoord value if supplied, and if this extension is supported.
}

func (ptt PolyformTextureTransform) ExtensionID() string {
	return "KHR_texture_transform"
}

func (ptt PolyformTextureTransform) IsRequired() bool { return ptt.Required }
func (ptt PolyformTextureTransform) IsInfo() bool     { return true }

func (ptt PolyformTextureTransform) ToTextureExtensionData(w *Writer) map[string]any {
	metadata := make(map[string]any)
	if ptt.Offset != nil {
		metadata["offset"] = ptt.Offset.ToFixedArr()
	}
	if ptt.Rotation != nil {
		metadata["rotation"] = *ptt.Rotation
	}
	if ptt.Scale != nil {
		metadata["scale"] = ptt.Scale.ToFixedArr()
	}
	if ptt.TexCoord != nil {
		metadata["texCoord"] = *ptt.TexCoord
	}
	return metadata
}
