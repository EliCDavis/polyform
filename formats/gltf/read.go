// Package gltf provides functionality to read and write GLTF 2.0 files.
//
// The reading API is designed to be flexible and support various use cases:
//
//  1. Simple file loading:
//     doc, buffers, err := gltf.LoadFile("model.gltf", nil)
//     models, err := gltf.DecodeModels(doc, buffers, "", nil)
//
//  2. Loading from memory/streams:
//     doc, buffers, err := gltf.Load(reader, nil)
//
//  3. Custom resource loading (e.g., from database):
//     opts := &gltf.ReaderOptions{
//     BufferLoader: myBufferLoader,
//     ImageLoader: myImageLoader,
//     }
//     doc, buffers, err := gltf.Load(reader, opts)
//
//  4. Parsing without loading resources:
//     doc, err := gltf.Parse(reader)
//     // Process doc structure, then load resources as needed
package gltf

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/EliCDavis/bitlib"
	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

// BufferLoader allows custom buffer resolution when loading GLTF files.
// This can be used to load buffers from a database, CDN, or other custom source.
type BufferLoader interface {
	// LoadBuffer loads binary data for the given URI.
	// The URI may be a relative path, absolute path, or custom scheme.
	LoadBuffer(uri string) ([]byte, error)
}

// ImageLoader allows custom image resolution when loading GLTF files.
// This can be used to load images from a database, CDN, or other custom source.
type ImageLoader interface {
	// LoadImage loads an image for the given URI.
	// The URI may be a relative path, absolute path, or custom scheme.
	LoadImage(uri string) (image.Image, string, error)
}

// ReaderOptions configures GLTF import behavior
type ReaderOptions struct {
	// BufferLoader provides custom buffer resolution. If nil, uses default file loading.
	BufferLoader BufferLoader

	// ImageLoader provides custom image resolution. If nil, uses default file loading.
	ImageLoader ImageLoader

	// BasePath overrides the base path for resolving relative URIs.
	// If empty, uses the directory of the GLTF file (for file-based loading) or current directory.
	BasePath string
}

func decodeTopology(mode *PrimitiveMode) modeling.Topology {
	if mode == nil {
		return modeling.TriangleTopology
	}

	switch *mode {
	case PrimitiveMode_POINTS:
		return modeling.PointTopology

	case PrimitiveMode_LINES:
		return modeling.LineTopology

	case PrimitiveMode_LINE_LOOP:
		return modeling.LineLoopTopology

	case PrimitiveMode_LINE_STRIP:
		return modeling.LineStripTopology

	default:
		return modeling.TriangleTopology
	}
}

type imgReaderCache map[int]image.Image

func (irc imgReaderCache) Load(doc *Gltf, bufferView int, buffers [][]byte) (image.Image, error) {

	if i, ok := irc[bufferView]; ok {
		return i, nil
	}

	bufView := resolveBufferview(doc.BufferViews[bufferView], buffers)
	img, _, err := image.Decode(bytes.NewReader(bufView))

	if err != nil {
		return nil, err
	}
	irc[bufferView] = img

	return img, nil
}

func resolveBufferview(bufferView BufferView, buffers [][]byte) []byte {
	start := bufferView.ByteOffset
	end := bufferView.ByteOffset + bufferView.ByteLength
	return buffers[bufferView.Buffer][start:end]
}

func decodeIndices(doc *Gltf, id *GltfId, buffers [][]byte) ([]int, error) {
	accessor := doc.Accessors[*id]
	if accessor.Type != AccessorType_SCALAR {
		return nil, fmt.Errorf("unexpected accessor type for indices: %s", accessor.Type)
	}

	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := resolveBufferview(bufferView, buffers)[accessor.ByteOffset:]

	indices := make([]int, accessor.Count)
	switch accessor.ComponentType {
	case AccessorComponentType_UNSIGNED_INT:
		for i := range indices {
			indices[i] = int(binary.LittleEndian.Uint32(buffer[i*4:]))
		}

	case AccessorComponentType_UNSIGNED_SHORT:
		for i := range indices {
			indices[i] = int(binary.LittleEndian.Uint16(buffer[i*2:]))
		}

	case AccessorComponentType_UNSIGNED_BYTE:
		for i := range indices {
			indices[i] = int(buffer[i])
		}

	default:
		return nil, fmt.Errorf("unexpected accessor component type for indices: %d", accessor.ComponentType)

	}

	return indices, nil
}

func decodePrimitiveAttributeName(name string) string {
	const texcoord = "TEXCOORD_"
	if strings.HasPrefix(name, texcoord) {
		i, err := strconv.Atoi(name[len(texcoord):])
		if err == nil {
			if i == 0 {
				return modeling.TexCoordAttribute
			}
			return fmt.Sprintf("%s%d", modeling.TexCoordAttribute, i+1)
		}
	}

	switch name {
	case POSITION:
		return modeling.PositionAttribute

	case COLOR_0:
		return modeling.ColorAttribute

	case JOINTS_0:
		return modeling.JointAttribute

	case WEIGHTS_0:
		return modeling.WeightAttribute

	case TEXCOORD_0:
		return modeling.TexCoordAttribute

	case NORMAL:
		return modeling.NormalAttribute

	default:
		return name
	}
}

func decodeScalarAccessor(doc *Gltf, id GltfId, buffers [][]byte) ([]float64, error) {
	accessor := doc.Accessors[id]
	if accessor.Type != AccessorType_SCALAR {
		return nil, fmt.Errorf("unexpected accessor type for scalar: %s", accessor.Type)
	}

	if accessor.BufferView == nil {
		return nil, fmt.Errorf("scalar accessor %d missing buffer view", id)
	}

	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := resolveBufferview(bufferView, buffers)[accessor.ByteOffset:]

	stride := accessor.ComponentType.Size()
	if bufferView.ByteStride != nil {
		stride = *bufferView.ByteStride
	}

	values := make([]float64, accessor.Count)

	switch accessor.ComponentType {
	case AccessorComponentType_FLOAT:
		for i := range accessor.Count {
			offset := i * stride
			values[i] = float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset:])))
		}

	case AccessorComponentType_UNSIGNED_BYTE:
		div := 1.0
		if accessor.Normalized {
			div = 255.0
		}
		for i := range accessor.Count {
			offset := i * stride
			values[i] = float64(buffer[offset]) / div
		}

	case AccessorComponentType_BYTE:
		div := 1.0
		if accessor.Normalized {
			div = 127.0
		}
		for i := range accessor.Count {
			offset := i * stride
			values[i] = float64(int8(buffer[offset])) / div
		}

	case AccessorComponentType_UNSIGNED_SHORT:
		div := 1.0
		if accessor.Normalized {
			div = 65535.0
		}
		for i := range accessor.Count {
			offset := i * stride
			values[i] = float64(binary.LittleEndian.Uint16(buffer[offset:])) / div
		}

	case AccessorComponentType_SHORT:
		div := 1.0
		if accessor.Normalized {
			div = 32767.0
		}
		for i := range accessor.Count {
			offset := i * stride
			values[i] = float64(int16(binary.LittleEndian.Uint16(buffer[offset:]))) / div
		}

	case AccessorComponentType_UNSIGNED_INT:
		for i := range accessor.Count {
			offset := i * stride
			values[i] = float64(binary.LittleEndian.Uint32(buffer[offset:]))
		}

	default:
		return nil, fmt.Errorf("unsupported accessor component type for scalar: %d", accessor.ComponentType)
	}

	return values, nil
}

func decodeVector2Accessor(doc *Gltf, id GltfId, buffers [][]byte) ([]vector2.Float64, error) {
	accessor := doc.Accessors[id]
	if accessor.Type != AccessorType_VEC2 {
		return nil, fmt.Errorf("unexpected accessor type for vec2: %s", accessor.Type)
	}

	if accessor.BufferView == nil {
		return nil, fmt.Errorf("vec2 accessor %d missing buffer view", id)
	}

	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := resolveBufferview(bufferView, buffers)[accessor.ByteOffset:]

	stride := accessor.ComponentType.Size() * 2
	if bufferView.ByteStride != nil {
		stride = *bufferView.ByteStride
	}

	vectors := make([]vector2.Float64, accessor.Count)

	switch accessor.ComponentType {

	case AccessorComponentType_FLOAT:
		for i := range accessor.Count {
			offset := i * stride
			x := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset:])))
			y := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset+4:])))
			vectors[i] = vector2.New(x, y)
		}

	case AccessorComponentType_UNSIGNED_BYTE:
		div := 1.
		if accessor.Normalized {
			div = 255.
		}

		for i := range accessor.Count {
			offset := i * stride
			x := float64(buffer[offset]) / div
			y := float64(buffer[offset+1]) / div
			vectors[i] = vector2.New(x, y)
		}

	case AccessorComponentType_BYTE:
		div := 1.0
		if accessor.Normalized {
			div = 127.0
		}
		for i := range accessor.Count {
			offset := i * stride
			x := float64(int8(buffer[offset])) / div
			y := float64(int8(buffer[offset+1])) / div
			vectors[i] = vector2.New(x, y)
		}

	case AccessorComponentType_UNSIGNED_SHORT:
		div := 1.0
		if accessor.Normalized {
			div = 65535.0
		}
		for i := range accessor.Count {
			offset := i * stride
			x := float64(binary.LittleEndian.Uint16(buffer[offset:])) / div
			y := float64(binary.LittleEndian.Uint16(buffer[offset+2:])) / div
			vectors[i] = vector2.New(x, y)
		}

	case AccessorComponentType_SHORT:
		div := 1.0
		if accessor.Normalized {
			div = 32767.0
		}
		for i := range accessor.Count {
			offset := i * stride
			x := float64(int16(binary.LittleEndian.Uint16(buffer[offset:]))) / div
			y := float64(int16(binary.LittleEndian.Uint16(buffer[offset+2:]))) / div
			vectors[i] = vector2.New(x, y)
		}

	default:
		return nil, fmt.Errorf("unsupported accessor component type for vec2: %d", accessor.ComponentType)
	}

	return vectors, nil
}

func decodeVector3Accessor(doc *Gltf, id GltfId, buffers [][]byte) ([]vector3.Float64, error) {
	accessor := doc.Accessors[id]
	if accessor.Type != AccessorType_VEC3 {
		return nil, fmt.Errorf("unexpected accessor type for vec3: %s", accessor.Type)
	}

	if accessor.BufferView == nil {
		return nil, fmt.Errorf("vec3 accessor %d missing buffer view", id)
	}

	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := resolveBufferview(bufferView, buffers)[accessor.ByteOffset:]

	stride := accessor.ComponentType.Size() * 3
	if bufferView.ByteStride != nil {
		stride = *bufferView.ByteStride
	}

	vectors := make([]vector3.Float64, accessor.Count)

	switch accessor.ComponentType {

	case AccessorComponentType_FLOAT:
		for i := range accessor.Count {
			offset := i * stride
			x := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset:])))
			y := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset+4:])))
			z := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset+8:])))
			vectors[i] = vector3.New(x, y, z)
		}

	case AccessorComponentType_UNSIGNED_BYTE:
		div := 1.
		if accessor.Normalized {
			div = 255.
		}

		for i := range accessor.Count {
			offset := i * stride
			x := float64(buffer[offset]) / div
			y := float64(buffer[offset+1]) / div
			z := float64(buffer[offset+2]) / div
			vectors[i] = vector3.New(x, y, z)
		}

	case AccessorComponentType_BYTE:
		div := 1.0
		if accessor.Normalized {
			div = 127.0
		}
		for i := range accessor.Count {
			offset := i * stride
			x := float64(int8(buffer[offset])) / div
			y := float64(int8(buffer[offset+1])) / div
			z := float64(int8(buffer[offset+2])) / div
			vectors[i] = vector3.New(x, y, z)
		}

	case AccessorComponentType_UNSIGNED_SHORT:
		div := 1.0
		if accessor.Normalized {
			div = 65535.0
		}
		for i := range accessor.Count {
			offset := i * stride
			x := float64(binary.LittleEndian.Uint16(buffer[offset:])) / div
			y := float64(binary.LittleEndian.Uint16(buffer[offset+2:])) / div
			z := float64(binary.LittleEndian.Uint16(buffer[offset+4:])) / div
			vectors[i] = vector3.New(x, y, z)
		}

	case AccessorComponentType_SHORT:
		div := 1.0
		if accessor.Normalized {
			div = 32767.0
		}
		for i := range accessor.Count {
			offset := i * stride
			x := float64(int16(binary.LittleEndian.Uint16(buffer[offset:]))) / div
			y := float64(int16(binary.LittleEndian.Uint16(buffer[offset+2:]))) / div
			z := float64(int16(binary.LittleEndian.Uint16(buffer[offset+4:]))) / div
			vectors[i] = vector3.New(x, y, z)
		}

	default:
		return nil, fmt.Errorf("unsupported accessor component type for vec3: %d", accessor.ComponentType)
	}

	return vectors, nil
}

func decodeVector4Accessor(doc *Gltf, id GltfId, buffers [][]byte) ([]vector4.Float64, error) {
	accessor := doc.Accessors[id]
	if accessor.Type != AccessorType_VEC4 {
		return nil, fmt.Errorf("unexpected accessor type for vec4: %s", accessor.Type)
	}

	if accessor.BufferView == nil {
		return nil, fmt.Errorf("vec4 accessor %d missing buffer view", id)
	}

	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := resolveBufferview(bufferView, buffers)[accessor.ByteOffset:]

	stride := accessor.ComponentType.Size() * 4
	if bufferView.ByteStride != nil {
		stride = *bufferView.ByteStride
	}

	vectors := make([]vector4.Float64, accessor.Count)
	switch accessor.ComponentType {
	case AccessorComponentType_FLOAT:
		for i := range vectors {
			offset := i * stride
			x := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset:])))
			y := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset+4:])))
			z := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset+8:])))
			w := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset+12:])))
			vectors[i] = vector4.New(x, y, z, w)
		}

	case AccessorComponentType_UNSIGNED_BYTE:
		div := 1.
		if accessor.Normalized {
			div = 255.
		}

		for i := range vectors {
			offset := i * stride
			x := float64(buffer[offset]) / div
			y := float64(buffer[offset+1]) / div
			z := float64(buffer[offset+2]) / div
			w := float64(buffer[offset+3]) / div
			vectors[i] = vector4.New(x, y, z, w)
		}

	case AccessorComponentType_BYTE:
		div := 1.0
		if accessor.Normalized {
			div = 127.0
		}
		for i := range vectors {
			offset := i * stride
			x := float64(int8(buffer[offset])) / div
			y := float64(int8(buffer[offset+1])) / div
			z := float64(int8(buffer[offset+2])) / div
			w := float64(int8(buffer[offset+3])) / div
			vectors[i] = vector4.New(x, y, z, w)
		}

	case AccessorComponentType_UNSIGNED_SHORT:
		div := 1.0
		if accessor.Normalized {
			div = 65535.0
		}
		for i := range vectors {
			offset := i * stride
			x := float64(binary.LittleEndian.Uint16(buffer[offset:])) / div
			y := float64(binary.LittleEndian.Uint16(buffer[offset+2:])) / div
			z := float64(binary.LittleEndian.Uint16(buffer[offset+4:])) / div
			w := float64(binary.LittleEndian.Uint16(buffer[offset+6:])) / div
			vectors[i] = vector4.New(x, y, z, w)
		}

	case AccessorComponentType_SHORT:
		div := 1.0
		if accessor.Normalized {
			div = 32767.0
		}
		for i := range vectors {
			offset := i * stride
			x := float64(int16(binary.LittleEndian.Uint16(buffer[offset:]))) / div
			y := float64(int16(binary.LittleEndian.Uint16(buffer[offset+2:]))) / div
			z := float64(int16(binary.LittleEndian.Uint16(buffer[offset+4:]))) / div
			w := float64(int16(binary.LittleEndian.Uint16(buffer[offset+6:]))) / div
			vectors[i] = vector4.New(x, y, z, w)
		}

	default:
		return nil, fmt.Errorf("unsupported accessor component type for vec4: %d", accessor.ComponentType)
	}

	return vectors, nil
}

// loadTexture loads a texture from the GLTF document
func loadTexture(doc *Gltf, textureId GltfId, opts ReaderOptions, buffers [][]byte, imgCache imgReaderCache) (*PolyformTexture, error) {
	if textureId >= len(doc.Textures) || textureId < 0 {
		return nil, fmt.Errorf("invalid texture ID: %d", textureId)
	}

	texture := doc.Textures[textureId]
	polyformTexture := &PolyformTexture{}

	// Load image if present
	if texture.Source != nil {
		if *texture.Source >= len(doc.Images) {
			return nil, fmt.Errorf("texture %d references invalid image %d", textureId, *texture.Source)
		}

		imageRef := doc.Images[*texture.Source]
		if imageRef.URI != "" {
			// Use the configured image loader
			img, _, err := opts.ImageLoader.LoadImage(imageRef.URI)
			if err != nil {
				return nil, fmt.Errorf("failed to load image for texture %d: %w", textureId, err)
			}
			polyformTexture.Image = img
			polyformTexture.URI = imageRef.URI
		} else if imageRef.BufferView != nil {
			if *imageRef.BufferView >= len(doc.BufferViews) {
				return nil, fmt.Errorf("texture %d image %d references invalid buffer view %d", textureId, *texture.Source, *imageRef.BufferView)
			}

			img, err := imgCache.Load(doc, *imageRef.BufferView, buffers)
			if err != nil {
				return nil, fmt.Errorf("unable to interpret texture %d image %d: %w", textureId, *texture.Source, err)
			}
			polyformTexture.Image = img
		} else {
			return nil, fmt.Errorf("texture %d image %d does not reference a URI or Buffer View", textureId, *texture.Source)
		}
	}

	// Load sampler if present
	if texture.Sampler != nil {
		if *texture.Sampler >= len(doc.Samplers) || *texture.Sampler < 0 {
			return nil, fmt.Errorf("texture %d references invalid sampler %d", textureId, *texture.Sampler)
		}
		sampler := doc.Samplers[*texture.Sampler]
		polyformTexture.Sampler = &sampler
	}

	return polyformTexture, nil
}

// loadMaterial loads a material from the GLTF document
func loadMaterial(doc *Gltf, materialId GltfId, opts ReaderOptions, buffers [][]byte, imgCache imgReaderCache) (*PolyformMaterial, error) {
	if materialId >= len(doc.Materials) || materialId < 0 {
		return nil, fmt.Errorf("invalid material ID: %d", materialId)
	}

	gltfMaterial := doc.Materials[materialId]
	material := &PolyformMaterial{
		Name: gltfMaterial.Name,
	}

	// Load PBR metallic-roughness properties
	if gltfMaterial.PbrMetallicRoughness != nil {
		pbr := &PolyformPbrMetallicRoughness{}

		// Base color factor
		if gltfMaterial.PbrMetallicRoughness.BaseColorFactor != nil {
			factor := *gltfMaterial.PbrMetallicRoughness.BaseColorFactor
			pbr.BaseColorFactor = color.RGBA{
				R: uint8(factor[0] * 255),
				G: uint8(factor[1] * 255),
				B: uint8(factor[2] * 255),
				A: uint8(factor[3] * 255),
			}
		}

		// Base color texture
		if gltfMaterial.PbrMetallicRoughness.BaseColorTexture != nil {
			texture, err := loadTexture(doc, gltfMaterial.PbrMetallicRoughness.BaseColorTexture.Index, opts, buffers, imgCache)
			if err != nil {
				return nil, fmt.Errorf("failed to load base color texture: %w", err)
			}
			pbr.BaseColorTexture = texture
		}

		// Metallic and roughness factors
		if gltfMaterial.PbrMetallicRoughness.MetallicFactor != nil {
			pbr.MetallicFactor = gltfMaterial.PbrMetallicRoughness.MetallicFactor
		}
		if gltfMaterial.PbrMetallicRoughness.RoughnessFactor != nil {
			pbr.RoughnessFactor = gltfMaterial.PbrMetallicRoughness.RoughnessFactor
		}

		// Metallic-roughness texture
		if gltfMaterial.PbrMetallicRoughness.MetallicRoughnessTexture != nil {
			texture, err := loadTexture(doc, gltfMaterial.PbrMetallicRoughness.MetallicRoughnessTexture.Index, opts, buffers, imgCache)
			if err != nil {
				return nil, fmt.Errorf("failed to load metallic-roughness texture: %w", err)
			}
			pbr.MetallicRoughnessTexture = texture
		}

		material.PbrMetallicRoughness = pbr
	}

	// Load normal texture
	if gltfMaterial.NormalTexture != nil {
		texture, err := loadTexture(doc, gltfMaterial.NormalTexture.Index, opts, buffers, imgCache)
		if err != nil {
			return nil, fmt.Errorf("failed to load normal texture: %w", err)
		}
		material.NormalTexture = &PolyformNormal{
			PolyformTexture: texture,
			Scale:           gltfMaterial.NormalTexture.Scale,
		}
	}

	// Load emissive texture
	if gltfMaterial.EmissiveTexture != nil {
		texture, err := loadTexture(doc, gltfMaterial.EmissiveTexture.Index, opts, buffers, imgCache)
		if err != nil {
			return nil, fmt.Errorf("failed to load emissive texture: %w", err)
		}
		material.EmissiveTexture = texture
	}

	// Load emissive factor
	if gltfMaterial.EmissiveFactor != nil {
		factor := *gltfMaterial.EmissiveFactor
		material.EmissiveFactor = color.RGBA{
			R: uint8(factor[0] * 255),
			G: uint8(factor[1] * 255),
			B: uint8(factor[2] * 255),
			A: 255,
		}
	}

	material.AlphaMode = gltfMaterial.AlphaMode
	material.AlphaCutoff = gltfMaterial.AlphaCutoff

	return material, nil
}

func decodePrimitive(doc *Gltf, buffers [][]byte, n Node, m Mesh, p Primitive, opts ReaderOptions, imgCache imgReaderCache) (*PolyformModel, error) {
	// Handle indices - they might be nil for non-indexed geometry
	var indices []int
	var err error
	if p.Indices != nil {
		indices, err = decodeIndices(doc, p.Indices, buffers)
		if err != nil {
			return nil, fmt.Errorf("failed to decode indices: %w", err)
		}
	}

	mesh := modeling.NewMesh(decodeTopology(p.Mode), indices)

	// Process all attributes
	for attr, gltfId := range p.Attributes {
		if gltfId >= len(doc.Accessors) {
			return nil, fmt.Errorf("attribute %s references invalid accessor %d", attr, gltfId)
		}

		accessor := doc.Accessors[gltfId]
		attributeName := decodePrimitiveAttributeName(attr)

		switch accessor.Type {
		case AccessorType_SCALAR:
			values, err := decodeScalarAccessor(doc, gltfId, buffers)
			if err != nil {
				return nil, fmt.Errorf("failed to decode scalar attribute %s: %w", attr, err)
			}
			mesh = mesh.SetFloat1Attribute(attributeName, values)

		case AccessorType_VEC2:
			v2, err := decodeVector2Accessor(doc, gltfId, buffers)
			if err != nil {
				return nil, fmt.Errorf("failed to decode vec2 attribute %s: %w", attr, err)
			}
			mesh = mesh.SetFloat2Attribute(attributeName, v2)

		case AccessorType_VEC3:
			v3, err := decodeVector3Accessor(doc, gltfId, buffers)
			if err != nil {
				return nil, fmt.Errorf("failed to decode vec3 attribute %s: %w", attr, err)
			}
			mesh = mesh.SetFloat3Attribute(attributeName, v3)

		case AccessorType_VEC4:
			v4, err := decodeVector4Accessor(doc, gltfId, buffers)
			if err != nil {
				return nil, fmt.Errorf("failed to decode vec4 attribute %s: %w", attr, err)
			}
			mesh = mesh.SetFloat4Attribute(attributeName, v4)

		default:
			return nil, fmt.Errorf("unsupported accessor type %s for attribute %s", accessor.Type, attr)
		}
	}

	transform := trs.Identity()

	// Handle matrix transformation if present
	if n.Matrix != nil {
		transform = trs.FromMatrix(mat.FromColArray(*n.Matrix))
	} else {
		// Handle TRS components
		if n.Translation != nil {
			data := *n.Translation
			p := vector3.New(data[0], data[1], data[2])
			transform = transform.SetTranslation(p)
		}

		if n.Scale != nil {
			data := *n.Scale
			p := vector3.New(data[0], data[1], data[2])
			transform = transform.SetScale(p)
		}

		if n.Rotation != nil {
			data := *n.Rotation
			p := quaternion.New(vector3.New(data[0], data[1], data[2]), data[3])
			transform = transform.SetRotation(p)
		}
	}

	// Load material if present
	var material *PolyformMaterial
	if p.Material != nil {
		m, err := loadMaterial(doc, *p.Material, opts, buffers, imgCache)
		if err != nil {
			return nil, fmt.Errorf("failed to load material: %w", err)
		}
		material = m
	}

	return &PolyformModel{
		Name:     n.Name,
		Mesh:     &mesh,
		Material: material,
		TRS:      &transform,
	}, nil
}

// processNodeHierarchy recursively processes a node and its children, accumulating transformations
func processNodeHierarchy(doc *Gltf, buffers [][]byte, nodeIndex int, parentTransform trs.TRS, scene *PolyformScene, opts ReaderOptions, imgCache imgReaderCache) error {
	if nodeIndex >= len(doc.Nodes) {
		return fmt.Errorf("invalid node index: %d", nodeIndex)
	}

	node := doc.Nodes[nodeIndex]

	// Calculate node transformation
	nodeTransform := trs.Identity()
	if node.Matrix != nil {
		nodeTransform = trs.FromMatrix(mat.FromColArray(*node.Matrix))
	} else {
		if node.Translation != nil {
			data := *node.Translation
			nodeTransform = nodeTransform.SetTranslation(vector3.New(data[0], data[1], data[2]))
		}
		if node.Scale != nil {
			data := *node.Scale
			nodeTransform = nodeTransform.SetScale(vector3.New(data[0], data[1], data[2]))
		}
		if node.Rotation != nil {
			data := *node.Rotation
			nodeTransform = nodeTransform.SetRotation(quaternion.New(vector3.New(data[0], data[1], data[2]), data[3]))
		}
	}

	// Combine with parent transformation
	worldTransform := parentTransform.Multiply(nodeTransform)

	// Process mesh if present
	if node.Mesh != nil {
		mesh := doc.Meshes[*node.Mesh]
		for primitiveIndex, p := range mesh.Primitives {
			model, err := decodePrimitive(doc, buffers, node, mesh, p, opts, imgCache)
			if err != nil {
				return fmt.Errorf("Node %d Meshes[%d].primitives[%d]: %w", nodeIndex, *node.Mesh, primitiveIndex, err)
			}

			// Apply world transformation
			model.TRS = &worldTransform
			scene.Models = append(scene.Models, *model)
		}
	}

	// Process children recursively
	for _, childIndex := range node.Children {
		err := processNodeHierarchy(doc, buffers, childIndex, worldTransform, scene, opts, imgCache)
		if err != nil {
			return fmt.Errorf("failed to process child node %d of node %d: %w", childIndex, nodeIndex, err)
		}
	}

	return nil
}

// ParseGLTF reads and parses GLTF JSON from an io.Reader.
// This only parses the JSON structure without loading any external resources like buffers or images.
// Use Load or LoadFile if you need to load external resources.
//
// Example:
//
//	jsonData := strings.NewReader(gltfJSON)
//	doc, err := gltf.ParseGLTF(jsonData)
//	if err != nil {
//	    log.Fatal(err)
//	}
func ParseGLTF(r io.Reader) (*Gltf, error) {
	decoder := json.NewDecoder(r)
	g := &Gltf{}

	if err := decoder.Decode(g); err != nil {
		return nil, fmt.Errorf("failed to parse GLTF JSON: %w", err)
	}

	// Validate basic GLTF structure
	if g.Asset.Version == "" {
		return nil, fmt.Errorf("missing required asset version in GLTF file")
	}

	return g, nil
}

// ParseFile reads and parses a GLTF JSON file from disk.
// This only parses the JSON structure without loading any external resources like buffers or images.
// Use LoadFile if you need to load external resources.
//
// Example:
//
//	doc, err := gltf.ParseFile("model.gltf")
//	if err != nil {
//	    log.Fatal(err)
//	}
func ParseFile(gltfPath string) (*Gltf, error) {
	file, err := os.Open(gltfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open GLTF file %q: %w", gltfPath, err)
	}
	defer file.Close()

	return ParseGLTF(file)
}

// LoadGLTF reads GLTF JSON from an io.Reader and loads all referenced buffers.
//
// Example with embedded buffer:
//
//	jsonData := strings.NewReader(gltfJSON) // GLTF with data URI buffers
//	doc, buffers, err := gltf.LoadGLTF(jsonData, nil)
//
// Example with custom loader:
//
//	opts := &gltf.ReaderOptions{
//	    BufferLoader: myCustomLoader,
//	}
//	doc, buffers, err := gltf.LoadGLTF(jsonData, opts)
//
// If a custom BufferLoader is provided in options, it will be used to load buffers.
// Otherwise, a StandardLoader is used with options.BasePath (or current directory).
func LoadGLTF(r io.Reader, options *ReaderOptions) (*Gltf, [][]byte, error) {
	// Parse the GLTF JSON
	g, err := ParseGLTF(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse GLTF data: %w", err)
	}

	// Set up options with defaults
	var opts ReaderOptions
	if options != nil {
		opts = *options
	}

	// Use standard buffer loader if none provided
	if opts.BufferLoader == nil {
		opts.BufferLoader = &StandardLoader{
			BasePath: opts.BasePath,
		}
	}

	// Load buffers
	allBuffers := make([][]byte, 0, len(g.Buffers))
	for bufIndex, buf := range g.Buffers {
		if buf.ByteLength <= 0 {
			return nil, nil, fmt.Errorf("buffer %d has invalid byte length: %d", bufIndex, buf.ByteLength)
		}

		bufferData, err := opts.BufferLoader.LoadBuffer(buf.URI)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load buffer %d: %w", bufIndex, err)
		}

		if len(bufferData) != buf.ByteLength {
			return nil, nil, fmt.Errorf("buffer %d: loaded size %d does not match expected length %d", bufIndex, len(bufferData), buf.ByteLength)
		}

		allBuffers = append(allBuffers, bufferData)
	}

	return g, allBuffers, nil
}

func LoadGLB(r io.Reader, options *ReaderOptions) (*Gltf, [][]byte, error) {
	reader := bitlib.NewReader(r, binary.LittleEndian)

	readMagic := reader.UInt32()
	if readMagic != magicNumber {
		return nil, nil, fmt.Errorf("unexpected read magic number %d", readMagic)
	}

	readVersion := reader.UInt32()
	if readVersion != version {
		return nil, nil, fmt.Errorf("unexpected version %d", readVersion)
	}

	documentLength := reader.UInt32()
	jsonLength := reader.UInt32()
	jsonIdentifier := reader.UInt32()
	if jsonIdentifier != jsonChunkIdentifier {
		return nil, nil, fmt.Errorf("unrecognized json chunk identifier: %d", jsonIdentifier)
	}

	jsonPayload := make([]byte, jsonLength)
	reader.Read(jsonPayload)
	doc := &Gltf{}
	err := json.Unmarshal(jsonPayload, doc)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse GLB embedded JSON doc: %w", err)
	}

	remainingLength := documentLength - jsonLength - 8 - 12
	if remainingLength == 0 {
		return doc, nil, nil
	}

	bufLen := reader.UInt32()
	bufIdentifier := reader.UInt32()
	if bufIdentifier != binChunkIdentifier {
		return nil, nil, fmt.Errorf("unrecognized bin chunk identifier: %d", bufIdentifier)
	}

	binPayload := make([]byte, bufLen)
	reader.Read(binPayload)

	// Load buffers
	allBuffers := make([][]byte, 0, len(doc.Buffers))
	for bufIndex, buf := range doc.Buffers {

		if buf.ByteLength <= 0 {
			return nil, nil, fmt.Errorf("buffer %d has invalid byte length: %d", bufIndex, buf.ByteLength)
		}

		if buf.URI == "" {
			allBuffers = append(allBuffers, binPayload)
			continue
		}

		bufferData, err := options.BufferLoader.LoadBuffer(buf.URI)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load buffer %d: %w", bufIndex, err)
		}

		if len(bufferData) != buf.ByteLength {
			return nil, nil, fmt.Errorf("buffer %d: loaded size %d does not match expected length %d", bufIndex, len(bufferData), buf.ByteLength)
		}

		allBuffers = append(allBuffers, bufferData)
	}

	return doc, allBuffers, reader.Error()
}

// LoadFile reads a GLTF file from disk and loads all referenced buffers.
// The directory containing the GLTF file is used as the base path for resolving relative URIs.
//
// Example:
//
//	doc, buffers, err := gltf.LoadFile("model.gltf", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Example with custom image loader:
//
//	opts := &gltf.ReaderOptions{
//	    ImageLoader: &gltf.NoOpImageLoader{},  // Skip loading texture files
//	}
//	doc, buffers, err := gltf.LoadFile("model.gltf", opts)
func LoadFile(gltfPath string, options *ReaderOptions) (*Gltf, [][]byte, error) {
	file, err := os.Open(gltfPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open GLTF file %q: %w", gltfPath, err)
	}
	defer file.Close()

	// Set up options with base path
	opts := &ReaderOptions{}
	if options != nil {
		*opts = *options
	}

	// Use the GLTF file's directory as base path if not specified
	if opts.BasePath == "" {
		opts.BasePath = filepath.Dir(gltfPath)
	}

	ext := strings.ToLower(filepath.Ext(gltfPath))
	switch ext {
	case ".gltf":
		return LoadGLTF(file, opts)

	case ".glb":
		return LoadGLB(file, opts)
	}

	return nil, nil, fmt.Errorf("unrecognized flie: %s", ext)
}

// DecodeModels converts a GLTF document into a flat list of Polyform models.
// The options.BasePath field is used to resolve relative image URIs.
// If a custom ImageLoader is provided in options, it will be used to load images.
//
// Example:
//
//	doc, buffers, _ := gltf.LoadFile("model.gltf", nil)
//	models, err := gltf.DecodeModels(doc, buffers, "", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, model := range models {
//	    fmt.Printf("Model: %s, Vertices: %d\n", model.Name, model.Mesh.AttributeLength())
//	}
func DecodeModels(doc *Gltf, buffers [][]byte, options *ReaderOptions) ([]PolyformModel, error) {
	// Set up options with defaults
	var opts ReaderOptions
	if options != nil {
		opts = *options
	}

	// Use standard image loader if none provided
	if opts.ImageLoader == nil {
		opts.ImageLoader = &StandardLoader{
			BasePath: opts.BasePath,
		}
	}
	imgCache := make(imgReaderCache)

	models := make([]PolyformModel, 0)

	for nodeIndex, node := range doc.Nodes {
		if node.Mesh == nil {
			continue
		}

		mesh := doc.Meshes[*node.Mesh]

		for primitiveIndex, p := range mesh.Primitives {
			model, err := decodePrimitive(doc, buffers, node, mesh, p, opts, imgCache)
			if err != nil {
				return nil, fmt.Errorf("Node %d Meshes[%d].primitives[%d]: %w", nodeIndex, *node.Mesh, primitiveIndex, err)
			}
			models = append(models, *model)
		}
	}

	return models, nil
}

// DecodeScene reconstructs the complete scene hierarchy with proper parent-child relationships.
// The options.BasePath field is used to resolve relative image URIs.
// If a custom ImageLoader is provided in options, it will be used to load images.
//
// Example:
//
//	doc, buffers, _ := gltf.LoadFile("scene.gltf", nil)
//	scene, err := gltf.DecodeScene(doc, buffers, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Scene has %d models and %d lights\n", len(scene.Models), len(scene.Lights))
func DecodeScene(doc *Gltf, buffers [][]byte, options *ReaderOptions) (*PolyformScene, error) {
	// Set up options with defaults
	var opts ReaderOptions
	if options != nil {
		opts = *options
	}

	// Use standard image loader if none provided
	if opts.ImageLoader == nil {
		opts.ImageLoader = &StandardLoader{
			BasePath: opts.BasePath,
		}
	}

	// Build scene hierarchy
	scene := &PolyformScene{
		Models: make([]PolyformModel, 0),
		Lights: make([]KHR_LightsPunctual, 0),
	}

	imgCache := make(imgReaderCache)

	// Get the main scene or use the first one
	var sceneIndex int
	if doc.Scene != nil && *doc.Scene >= 0 && *doc.Scene < len(doc.Scenes) {
		sceneIndex = *doc.Scene
	}

	if len(doc.Scenes) == 0 {
		return scene, nil
	}

	if sceneIndex >= len(doc.Scenes) {
		return nil, fmt.Errorf("invalid scene index %d: only %d scenes available", sceneIndex, len(doc.Scenes))
	}

	gltfScene := doc.Scenes[sceneIndex]

	// Process root nodes and their children recursively
	for _, rootNodeIndex := range gltfScene.Nodes {
		err := processNodeHierarchy(doc, buffers, rootNodeIndex, trs.Identity(), scene, opts, imgCache)
		if err != nil {
			return nil, fmt.Errorf("failed to process root node %d: %w", rootNodeIndex, err)
		}
	}

	return scene, nil
}

// Deprecated: Use DecodeModels instead.
// ExperimentalDecodeModels converts a GLTF document into a flat list of Polyform models.
func ExperimentalDecodeModels(doc *Gltf, buffers [][]byte, gltfDir string, options *ReaderOptions) ([]PolyformModel, error) {
	opts := &ReaderOptions{}
	if options != nil {
		*opts = *options
	}
	if opts.BasePath == "" && gltfDir != "" {
		opts.BasePath = gltfDir
	}
	return DecodeModels(doc, buffers, options)
}

// Deprecated: Use DecodeScene instead.
// ExperimentalDecodeScene reconstructs the complete scene hierarchy with proper parent-child relationships.
func ExperimentalDecodeScene(doc *Gltf, buffers [][]byte, gltfDir string, options *ReaderOptions) (*PolyformScene, error) {
	opts := &ReaderOptions{}
	if options != nil {
		*opts = *options
	}
	if opts.BasePath == "" && gltfDir != "" {
		opts.BasePath = gltfDir
	}

	return DecodeScene(doc, buffers, options)
}

// Deprecated: Use LoadFile instead.
// ExperimentalLoad reads a GLTF file from disk and loads all referenced buffers.
func ExperimentalLoad(gltfPath string, options *ReaderOptions) (*Gltf, [][]byte, error) {
	return LoadFile(gltfPath, options)
}
