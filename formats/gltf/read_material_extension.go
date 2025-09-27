package gltf

import (
	"encoding/json"
	"fmt"
	"image/color"
)

var defaultMaterialExtensionLoaders = map[string]MaterialExtensionLoader{
	khr_materials_pbrSpecularGlossiness: pbrSpecularGlossinessExtensionLoader{},
	khr_materials_transmission:          transmissionExtensionLoader{},
	khr_materials_ior:                   iorExtensionLoader{},
	khr_materials_unlit:                 unlitExtensionLoader{},
	khr_materials_emissive_strength:     emissiveStrengthExtensionLoader{},
	khr_materials_dispersion:            dispersionExtensionLoader{},
	khr_materials_volume:                volumeExtensionLoader{},
	khr_materials_iridescence:           iridescenceExtensionLoader{},
	khr_materials_specular:              specularExtensionLoader{},
	khr_materials_clearcoat:             clearcoatExtensionLoader{},
	khr_materials_sheen:                 sheenExtensionLoader{},
	khr_materials_anisotropy:            anisotropyExtensionLoader{},
}

type MaterialExtensionLoaderContext struct {
	doc      *Gltf
	opts     ReaderOptions
	buffers  [][]byte
	imgCache imgReaderCache
}

func (ctx MaterialExtensionLoaderContext) LoadTexture(textureInfo TextureInfo) (*PolyformTexture, error) {
	return loadTexture(ctx.doc, textureInfo, ctx.opts, ctx.buffers, ctx.imgCache)
}

type MaterialExtensionLoader interface {
	LoadMaterialExtension(ctx *MaterialExtensionLoaderContext, extensionData any) (MaterialExtension, error)
}

func decodeRGBA(rgba []float64) color.RGBA {
	return color.RGBA{
		R: uint8(rgba[0] * 255),
		G: uint8(rgba[1] * 255),
		B: uint8(rgba[2] * 255),
		A: uint8(rgba[3] * 255),
	}
}

func decodeRGB(rgb []float64) color.RGBA {
	return color.RGBA{
		R: uint8(rgb[0] * 255),
		G: uint8(rgb[1] * 255),
		B: uint8(rgb[2] * 255),
		A: 255,
	}
}

func tryGetMapData[T any](m map[string]any, key string) (T, bool) {
	var v T
	val, ok := m[key]
	if !ok {
		return v, false
	}

	v, ok = val.(T)
	return v, ok
}

func tryGetMapArrayData[T any](m map[string]any, key string) ([]T, bool) {
	var v []T
	val, ok := m[key]
	if !ok {
		return v, false
	}

	arr, ok := val.([]any)
	if !ok {
		return nil, false
	}

	v = make([]T, len(arr))
	for i, ele := range arr {
		v[i], ok = ele.(T)

	}

	return v, ok
}

func tryReinterpretMapData[T any](m map[string]any, key string) (T, bool) {
	var v T
	val, ok := m[key]
	if !ok {
		return v, false
	}

	marshallData, err := json.Marshal(val)
	if err != nil {
		return v, false
	}

	err = json.Unmarshal(marshallData, &v)
	if err != nil {
		return v, false
	}
	return v, true
}

type pbrSpecularGlossinessExtensionLoader struct{}

func (pbrSpecularGlossinessExtensionLoader) LoadMaterialExtension(ctx *MaterialExtensionLoaderContext, extensionData any) (MaterialExtension, error) {
	pbrSpecularData, ok := extensionData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unable to interpret extension data")
	}

	ext := &PolyformPbrSpecularGlossiness{}

	if diffuseFactor, ok := tryGetMapArrayData[float64](pbrSpecularData, "diffuseFactor"); ok {
		ext.DiffuseFactor = decodeRGBA(diffuseFactor)
	}

	if specularFactor, ok := tryGetMapArrayData[float64](pbrSpecularData, "specularFactor"); ok {
		ext.SpecularFactor = decodeRGB(specularFactor)
	}

	if glossinessFactor, ok := tryGetMapData[float64](pbrSpecularData, "glossinessFactor"); ok {
		ext.GlossinessFactor = &glossinessFactor
	}

	if diffuseTexture, ok := tryReinterpretMapData[TextureInfo](pbrSpecularData, "diffuseTexture"); ok {
		loadedTex, err := ctx.LoadTexture(diffuseTexture)
		if err != nil {
			return nil, fmt.Errorf("unable to interpret diffuseTexture texture: %w", err)
		}
		ext.DiffuseTexture = loadedTex
	}

	if specularGlossinessTexture, ok := tryReinterpretMapData[TextureInfo](pbrSpecularData, "specularGlossinessTexture"); ok {
		loadedTex, err := ctx.LoadTexture(specularGlossinessTexture)
		if err != nil {
			return nil, fmt.Errorf("unable to interpret specularGlossinessTexture texture: %w", err)
		}
		ext.SpecularGlossinessTexture = loadedTex
	}

	return ext, nil
}

type unlitExtensionLoader struct{}

func (unlitExtensionLoader) LoadMaterialExtension(ctx *MaterialExtensionLoaderContext, extensionData any) (MaterialExtension, error) {
	return PolyformUnlit{}, nil
}

type transmissionExtensionLoader struct{}

func (transmissionExtensionLoader) LoadMaterialExtension(ctx *MaterialExtensionLoaderContext, extensionData any) (MaterialExtension, error) {
	pbrSpecularData, ok := extensionData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unable to interpret extension data")
	}

	ext := &PolyformTransmission{}

	if transmissionFactor, ok := tryGetMapData[float64](pbrSpecularData, "transmissionFactor"); ok {
		ext.Factor = transmissionFactor
	}

	if transmissionTexture, ok := tryReinterpretMapData[TextureInfo](pbrSpecularData, "transmissionTexture"); ok {
		loadedTex, err := ctx.LoadTexture(transmissionTexture)
		if err != nil {
			return nil, fmt.Errorf("unable to interpret transmission texture: %w", err)
		}
		ext.Texture = loadedTex
	}

	return ext, nil
}

type iorExtensionLoader struct{}

func (iorExtensionLoader) LoadMaterialExtension(ctx *MaterialExtensionLoaderContext, extensionData any) (MaterialExtension, error) {
	iorData, ok := extensionData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unable to interpret extension data")
	}

	ext := &PolyformIndexOfRefraction{}

	if ior, ok := tryGetMapData[float64](iorData, "ior"); ok {
		ext.IOR = &ior
	}

	return ext, nil
}

type emissiveStrengthExtensionLoader struct{}

func (emissiveStrengthExtensionLoader) LoadMaterialExtension(ctx *MaterialExtensionLoaderContext, extensionData any) (MaterialExtension, error) {
	emissiveData, ok := extensionData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unable to interpret extension data")
	}

	ext := &PolyformEmissiveStrength{}

	if emissiveStrength, ok := tryGetMapData[float64](emissiveData, "emissiveStrength"); ok {
		ext.EmissiveStrength = &emissiveStrength
	}

	return ext, nil
}

type dispersionExtensionLoader struct{}

func (dispersionExtensionLoader) LoadMaterialExtension(ctx *MaterialExtensionLoaderContext, extensionData any) (MaterialExtension, error) {
	emissiveData, ok := extensionData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unable to interpret extension data")
	}

	ext := &PolyformDispersion{}

	if dispersion, ok := tryGetMapData[float64](emissiveData, "dispersion"); ok {
		ext.Dispersion = dispersion
	}

	return ext, nil
}

type volumeExtensionLoader struct{}

func (volumeExtensionLoader) LoadMaterialExtension(ctx *MaterialExtensionLoaderContext, extensionData any) (MaterialExtension, error) {
	volumeData, ok := extensionData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unable to interpret extension data")
	}

	ext := &PolyformVolume{}

	if thicknessFactor, ok := tryGetMapData[float64](volumeData, "thicknessFactor"); ok {
		ext.ThicknessFactor = thicknessFactor
	}

	if attenuationDistance, ok := tryGetMapData[float64](volumeData, "attenuationDistance"); ok {
		ext.AttenuationDistance = &attenuationDistance
	}

	if thicknessTexture, ok := tryReinterpretMapData[TextureInfo](volumeData, "thicknessTexture"); ok {
		loadedTex, err := ctx.LoadTexture(thicknessTexture)
		if err != nil {
			return nil, fmt.Errorf("unable to interpret thicknessTexture texture: %w", err)
		}
		ext.ThicknessTexture = loadedTex
	}

	if attenuationColor, ok := tryGetMapArrayData[float64](volumeData, "attenuationColor"); ok {
		ext.AttenuationColor = decodeRGB(attenuationColor)
	}

	return ext, nil
}

type iridescenceExtensionLoader struct{}

func (iridescenceExtensionLoader) LoadMaterialExtension(ctx *MaterialExtensionLoaderContext, extensionData any) (MaterialExtension, error) {
	iridescenceData, ok := extensionData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unable to interpret extension data")
	}

	ext := &PolyformIridescence{}

	if iridescenceFactor, ok := tryGetMapData[float64](iridescenceData, "iridescenceFactor"); ok {
		ext.IridescenceFactor = iridescenceFactor
	}

	if iridescenceIor, ok := tryGetMapData[float64](iridescenceData, "iridescenceIor"); ok {
		ext.IridescenceIor = &iridescenceIor
	}

	if iridescenceThicknessMinimum, ok := tryGetMapData[float64](iridescenceData, "iridescenceThicknessMinimum"); ok {
		ext.IridescenceThicknessMinimum = &iridescenceThicknessMinimum
	}

	if iridescenceThicknessMaximum, ok := tryGetMapData[float64](iridescenceData, "iridescenceThicknessMaximum"); ok {
		ext.IridescenceThicknessMaximum = &iridescenceThicknessMaximum
	}

	if iridescenceTexture, ok := tryReinterpretMapData[TextureInfo](iridescenceData, "iridescenceTexture"); ok {
		loadedTex, err := ctx.LoadTexture(iridescenceTexture)
		if err != nil {
			return nil, fmt.Errorf("unable to interpret iridescenceTexture texture: %w", err)
		}
		ext.IridescenceTexture = loadedTex
	}

	if iridescenceThicknessTexture, ok := tryReinterpretMapData[TextureInfo](iridescenceData, "iridescenceThicknessTexture"); ok {
		loadedTex, err := ctx.LoadTexture(iridescenceThicknessTexture)
		if err != nil {
			return nil, fmt.Errorf("unable to interpret iridescenceThicknessTexture texture: %w", err)
		}
		ext.IridescenceThicknessTexture = loadedTex
	}

	return ext, nil
}

type specularExtensionLoader struct{}

func (specularExtensionLoader) LoadMaterialExtension(ctx *MaterialExtensionLoaderContext, extensionData any) (MaterialExtension, error) {
	specularData, ok := extensionData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unable to interpret extension data")
	}

	ext := &PolyformSpecular{}

	if specularFactor, ok := tryGetMapData[float64](specularData, "specularFactor"); ok {
		ext.Factor = &specularFactor
	}

	if specularColorFactor, ok := tryGetMapArrayData[float64](specularData, "specularColorFactor"); ok {
		ext.ColorFactor = decodeRGB(specularColorFactor)
	}

	if specularTexture, ok := tryReinterpretMapData[TextureInfo](specularData, "specularTexture"); ok {
		loadedTex, err := ctx.LoadTexture(specularTexture)
		if err != nil {
			return nil, fmt.Errorf("unable to interpret specularTexture texture: %w", err)
		}
		ext.Texture = loadedTex
	}

	if specularColorTexture, ok := tryReinterpretMapData[TextureInfo](specularData, "specularColorTexture"); ok {
		loadedTex, err := ctx.LoadTexture(specularColorTexture)
		if err != nil {
			return nil, fmt.Errorf("unable to interpret specularColorTexture texture: %w", err)
		}
		ext.ColorTexture = loadedTex
	}

	return ext, nil
}

type clearcoatExtensionLoader struct{}

func (clearcoatExtensionLoader) LoadMaterialExtension(ctx *MaterialExtensionLoaderContext, extensionData any) (MaterialExtension, error) {
	specularData, ok := extensionData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unable to interpret extension data")
	}

	ext := &PolyformClearcoat{}

	if clearcoatFactor, ok := tryGetMapData[float64](specularData, "clearcoatFactor"); ok {
		ext.ClearcoatFactor = clearcoatFactor
	}

	if clearcoatRoughnessFactor, ok := tryGetMapData[float64](specularData, "clearcoatRoughnessFactor"); ok {
		ext.ClearcoatRoughnessFactor = clearcoatRoughnessFactor
	}

	if clearcoatTexture, ok := tryReinterpretMapData[TextureInfo](specularData, "clearcoatTexture"); ok {
		loadedTex, err := ctx.LoadTexture(clearcoatTexture)
		if err != nil {
			return nil, fmt.Errorf("unable to interpret clearcoatTexture texture: %w", err)
		}
		ext.ClearcoatTexture = loadedTex
	}

	if clearcoatRoughnessTexture, ok := tryReinterpretMapData[TextureInfo](specularData, "clearcoatRoughnessTexture"); ok {
		loadedTex, err := ctx.LoadTexture(clearcoatRoughnessTexture)
		if err != nil {
			return nil, fmt.Errorf("unable to interpret clearcoatRoughnessTexture texture: %w", err)
		}
		ext.ClearcoatRoughnessTexture = loadedTex
	}

	return ext, nil
}

type sheenExtensionLoader struct{}

func (sheenExtensionLoader) LoadMaterialExtension(ctx *MaterialExtensionLoaderContext, extensionData any) (MaterialExtension, error) {
	specularData, ok := extensionData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unable to interpret extension data")
	}

	ext := &PolyformSheen{}

	if sheenColorFactor, ok := tryGetMapArrayData[float64](specularData, "sheenColorFactor"); ok {
		ext.SheenColorFactor = decodeRGB(sheenColorFactor)
	}

	if sheenRoughnessFactor, ok := tryGetMapData[float64](specularData, "sheenRoughnessFactor"); ok {
		ext.SheenRoughnessFactor = sheenRoughnessFactor
	}

	if sheenColorTexture, ok := tryReinterpretMapData[TextureInfo](specularData, "sheenColorTexture"); ok {
		loadedTex, err := ctx.LoadTexture(sheenColorTexture)
		if err != nil {
			return nil, fmt.Errorf("unable to interpret sheenColorTexture texture: %w", err)
		}
		ext.SheenColorTexture = loadedTex
	}

	if sheenRoughnessTexture, ok := tryReinterpretMapData[TextureInfo](specularData, "sheenRoughnessTexture"); ok {
		loadedTex, err := ctx.LoadTexture(sheenRoughnessTexture)
		if err != nil {
			return nil, fmt.Errorf("unable to interpret sheenRoughnessTexture texture: %w", err)
		}
		ext.SheenRoughnessTexture = loadedTex
	}

	return ext, nil
}

type anisotropyExtensionLoader struct{}

func (anisotropyExtensionLoader) LoadMaterialExtension(ctx *MaterialExtensionLoaderContext, extensionData any) (MaterialExtension, error) {
	specularData, ok := extensionData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unable to interpret extension data")
	}

	ext := &PolyformAnisotropy{}

	if anisotropyStrength, ok := tryGetMapData[float64](specularData, "anisotropyStrength"); ok {
		ext.AnisotropyStrength = anisotropyStrength
	}

	if anisotropyRotation, ok := tryGetMapData[float64](specularData, "anisotropyRotation"); ok {
		ext.AnisotropyRotation = anisotropyRotation
	}

	if anisotropyTexture, ok := tryReinterpretMapData[TextureInfo](specularData, "anisotropyTexture"); ok {
		loadedTex, err := ctx.LoadTexture(anisotropyTexture)
		if err != nil {
			return nil, fmt.Errorf("unable to interpret anisotropyTexture texture: %w", err)
		}
		ext.AnisotropyTexture = loadedTex
	}

	return ext, nil
}
