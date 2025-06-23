package gltf

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg" // Import for side effects to register JPEG decoder
	_ "image/png"  // Import for side effects to register PNG decoder
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

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

func decodeIndices(doc *Gltf, id *GltfId, buffers [][]byte) ([]int, error) {
	accessor := doc.Accessors[*id]
	if accessor.Type != AccessorType_SCALAR {
		return nil, fmt.Errorf("unexpected accessor type for indices: %s", accessor.Type)
	}

	bufferView := doc.BufferViews[*accessor.BufferView]
	start := bufferView.ByteOffset + accessor.ByteOffset
	end := bufferView.ByteOffset + bufferView.ByteLength
	buffer := buffers[bufferView.Buffer][start:end]

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
	start := bufferView.ByteOffset + accessor.ByteOffset
	end := bufferView.ByteOffset + bufferView.ByteLength
	buffer := buffers[bufferView.Buffer][start:end]

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
	start := bufferView.ByteOffset + accessor.ByteOffset
	end := bufferView.ByteOffset + bufferView.ByteLength
	buffer := buffers[bufferView.Buffer][start:end]

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
	start := bufferView.ByteOffset + accessor.ByteOffset
	end := bufferView.ByteOffset + bufferView.ByteLength
	buffer := buffers[bufferView.Buffer][start:end]

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
	start := bufferView.ByteOffset + accessor.ByteOffset
	bufferEnd := bufferView.ByteOffset + bufferView.ByteLength
	buffer := buffers[bufferView.Buffer][start:bufferEnd]

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

// loadImage loads an image from a file path or file:// URI and returns it
func loadImage(imagePath string) (image.Image, error) {
	// Convert file:// URI to local path if needed
	actualPath, err := resolveImagePath(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve image path %q: %w", imagePath, err)
	}

	file, err := os.Open(actualPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file %q: %w", actualPath, err)
	}
	defer file.Close()

	// Use image.Decode to automatically detect format from file content
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image %q: %w", actualPath, err)
	}

	// Log the detected format for debugging purposes
	_ = format // format contains the detected image format (e.g., "jpeg", "png")

	return img, nil
}

// resolveImagePath converts file:// URIs to local file paths, or returns the path unchanged
func resolveImagePath(imagePath string) (string, error) {
	// Check if it's a file:// URI
	if strings.HasPrefix(imagePath, "file://") {
		parsedURL, err := url.Parse(imagePath)
		if err != nil {
			return "", fmt.Errorf("invalid file URI: %w", err)
		}

		if parsedURL.Scheme != "file" {
			return "", fmt.Errorf("expected file:// scheme, got %s://", parsedURL.Scheme)
		}

		// Convert URL path to local file path
		// parsedURL.Path already handles URL decoding
		return parsedURL.Path, nil
	}

	// Return path unchanged if it's not a file:// URI
	return imagePath, nil
}

// loadImageFromDataURI loads an image from a data URI
// Data URI format: data:[<mediatype>][;base64],<data>
func loadImageFromDataURI(dataURI string) (image.Image, error) {
	// Check if it's a valid data URI
	if !strings.HasPrefix(dataURI, "data:") {
		return nil, fmt.Errorf("invalid data URI: must start with 'data:'")
	}

	// Find the first comma that separates header from data
	commaIndex := strings.Index(dataURI, ",")
	if commaIndex == -1 {
		return nil, fmt.Errorf("invalid data URI: missing comma separator")
	}

	header := dataURI[5:commaIndex] // Skip "data:" prefix
	data := dataURI[commaIndex+1:]

	// Parse the header to extract mediatype and parameters
	// Format: [<mediatype>][;base64]
	var mediaType string
	var isBase64 bool

	// Split header by semicolons to get mediatype and parameters
	headerParts := strings.Split(header, ";")

	if len(headerParts) == 0 {
		return nil, fmt.Errorf("invalid data URI: empty header")
	}

	// First part is the media type
	mediaType = strings.TrimSpace(headerParts[0])

	// Check remaining parts for base64 parameter
	for i := 1; i < len(headerParts); i++ {
		param := strings.TrimSpace(headerParts[i])
		if param == "base64" {
			isBase64 = true
			break
		}
	}

	// Validate content type is present and supported
	if mediaType == "" {
		return nil, fmt.Errorf("invalid data URI: missing content type (required by GLTF specification)")
	}

	// Check if the content type is supported
	switch mediaType {
	case "image/jpeg", "image/png":
		// Supported image formats
	default:
		return nil, fmt.Errorf("unsupported content type %q: only image/jpeg and image/png are supported", mediaType)
	}

	// Check for base64 encoding declaration
	if !isBase64 {
		return nil, fmt.Errorf("invalid data URI: base64 encoding declaration is required")
	}

	// Decode base64 data
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// Create reader from decoded data
	reader := bytes.NewReader(decoded)

	// Use image.Decode to automatically detect format from data content
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image from data URI: %w", err)
	}

	// Verify the detected format matches the declared content type
	switch format {
	case "jpeg":
		if mediaType != "image/jpeg" {
			return nil, fmt.Errorf("content type mismatch: declared %q but detected JPEG", mediaType)
		}
	case "png":
		if mediaType != "image/png" {
			return nil, fmt.Errorf("content type mismatch: declared %q but detected PNG", mediaType)
		}
	default:
		return nil, fmt.Errorf("unsupported image format detected: %q", format)
	}

	return img, nil
}

// loadTexture loads a texture from the GLTF document
func loadTexture(doc *Gltf, textureId GltfId, gltfDir string) (*PolyformTexture, error) {
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
			if strings.HasPrefix(imageRef.URI, "data:") {
				// Handle data URI (embedded image)
				img, err := loadImageFromDataURI(imageRef.URI)
				if err != nil {
					return nil, fmt.Errorf("failed to load embedded image for texture %d: %w", textureId, err)
				}
				polyformTexture.Image = img
				polyformTexture.URI = imageRef.URI
			} else if strings.HasPrefix(imageRef.URI, "file://") {
				// Handle file:// URI (absolute path)
				img, err := loadImage(imageRef.URI)
				if err != nil {
					return nil, fmt.Errorf("failed to load image for texture %d: %w", textureId, err)
				}
				polyformTexture.Image = img
				polyformTexture.URI = imageRef.URI
			} else {
				// Load external image file (relative path)
				imagePath := filepath.Join(gltfDir, imageRef.URI)
				img, err := loadImage(imagePath)
				if err != nil {
					return nil, fmt.Errorf("failed to load image for texture %d: %w", textureId, err)
				}
				polyformTexture.Image = img
				polyformTexture.URI = imageRef.URI
			}
		}
		// TODO: Handle embedded images via buffer views
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
func loadMaterial(doc *Gltf, materialId GltfId, gltfDir string) (*PolyformMaterial, error) {
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
			texture, err := loadTexture(doc, gltfMaterial.PbrMetallicRoughness.BaseColorTexture.Index, gltfDir)
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
			texture, err := loadTexture(doc, gltfMaterial.PbrMetallicRoughness.MetallicRoughnessTexture.Index, gltfDir)
			if err != nil {
				return nil, fmt.Errorf("failed to load metallic-roughness texture: %w", err)
			}
			pbr.MetallicRoughnessTexture = texture
		}

		material.PbrMetallicRoughness = pbr
	}

	// Load normal texture
	if gltfMaterial.NormalTexture != nil {
		texture, err := loadTexture(doc, gltfMaterial.NormalTexture.Index, gltfDir)
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
		texture, err := loadTexture(doc, gltfMaterial.EmissiveTexture.Index, gltfDir)
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

	// Set alpha mode and cutoff
	if gltfMaterial.AlphaMode != nil {
		material.AlphaMode = gltfMaterial.AlphaMode
	}
	if gltfMaterial.AlphaCutoff != nil {
		material.AlphaCutoff = gltfMaterial.AlphaCutoff
	}

	return material, nil
}

func decodePrimitive(doc *Gltf, buffers [][]byte, n Node, m Mesh, p Primitive, gltfDir string) (*PolyformModel, error) {
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
		mat, err := loadMaterial(doc, *p.Material, gltfDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load material: %w", err)
		}
		material = mat
	}

	return &PolyformModel{
		Name:     n.Name,
		Mesh:     &mesh,
		Material: material,
		TRS:      &transform,
	}, nil
}

func ExperimentalDecodeModels(doc *Gltf, buffers [][]byte, gltfDir string) ([]PolyformModel, error) {
	models := make([]PolyformModel, 0)

	for nodeIndex, node := range doc.Nodes {
		if node.Mesh == nil {
			continue
		}

		mesh := doc.Meshes[*node.Mesh]

		for primitiveIndex, p := range mesh.Primitives {
			model, err := decodePrimitive(doc, buffers, node, mesh, p, gltfDir)
			if err != nil {
				return nil, fmt.Errorf("Node %d Meshes[%d].primitives[%d]: %w", nodeIndex, *node.Mesh, primitiveIndex, err)
			}
			models = append(models, *model)
		}

	}

	return models, nil
}

// ExperimentalDecodeScene reconstructs the complete scene hierarchy with proper parent-child relationships
func ExperimentalDecodeScene(doc *Gltf, buffers [][]byte, gltfDir string) (*PolyformScene, error) {
	// Build scene hierarchy
	scene := &PolyformScene{
		Models: make([]PolyformModel, 0),
	}

	// Process each scene (typically there's just one)
	if len(doc.Scenes) == 0 {
		return scene, nil
	}

	// Use the default scene or the first one
	sceneIndex := 0
	if doc.Scene != 0 {
		sceneIndex = doc.Scene
	}
	if sceneIndex >= len(doc.Scenes) {
		return nil, fmt.Errorf("invalid scene index: %d", sceneIndex)
	}

	gltfScene := doc.Scenes[sceneIndex]

	// Process root nodes and their children recursively
	for _, rootNodeIndex := range gltfScene.Nodes {
		err := processNodeHierarchy(doc, buffers, gltfDir, rootNodeIndex, trs.Identity(), scene)
		if err != nil {
			return nil, fmt.Errorf("failed to process root node %d: %w", rootNodeIndex, err)
		}
	}

	return scene, nil
}

// processNodeHierarchy recursively processes a node and its children, accumulating transformations
func processNodeHierarchy(doc *Gltf, buffers [][]byte, gltfDir string, nodeIndex int, parentTransform trs.TRS, scene *PolyformScene) error {
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
			model, err := decodePrimitive(doc, buffers, node, mesh, p, gltfDir)
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
		err := processNodeHierarchy(doc, buffers, gltfDir, childIndex, worldTransform, scene)
		if err != nil {
			return fmt.Errorf("failed to process child node %d of node %d: %w", childIndex, nodeIndex, err)
		}
	}

	return nil
}

func ExperimentalLoad(gltfPath string) (*Gltf, [][]byte, error) {
	gltfContents, err := os.ReadFile(gltfPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read GLTF file %q: %w", gltfPath, err)
	}

	g := &Gltf{}
	err = json.Unmarshal(gltfContents, g)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse GLTF JSON: %w", err)
	}

	// Validate basic GLTF structure
	if g.Asset.Version == "" {
		return nil, nil, fmt.Errorf("missing required asset version in GLTF file")
	}

	allBuffers := make([][]byte, 0, len(g.Buffers))
	gltfDir := filepath.Dir(gltfPath)

	for bufIndex, buf := range g.Buffers {
		if buf.ByteLength <= 0 {
			return nil, nil, fmt.Errorf("buffer %d has invalid byte length: %d", bufIndex, buf.ByteLength)
		}

		if strings.HasPrefix(buf.URI, "data:") {
			// Handle embedded data URI
			stringBuf := buf.URI[5:]

			base64Str := "application/octet-stream;base64,"
			if strings.HasPrefix(stringBuf, base64Str) {
				buf64, err := base64.StdEncoding.DecodeString(stringBuf[len(base64Str):])
				if err != nil {
					return nil, nil, fmt.Errorf("failed to decode base64 data in buffer %d: %w", bufIndex, err)
				}

				if len(buf64) != buf.ByteLength {
					return nil, nil, fmt.Errorf("buffer %d: decoded data length %d does not match expected length %d", bufIndex, len(buf64), buf.ByteLength)
				}

				allBuffers = append(allBuffers, buf64)
			} else {
				return nil, nil, fmt.Errorf("unsupported data URI encoding in buffer %d: %s", bufIndex, stringBuf[:min(50, len(stringBuf))])
			}
		} else {
			// Handle external file reference
			if buf.URI == "" {
				return nil, nil, fmt.Errorf("buffer %d has empty URI", bufIndex)
			}

			bufferPath := filepath.Join(gltfDir, buf.URI)
			bufferData, err := os.ReadFile(bufferPath)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read buffer file %q for buffer %d: %w", bufferPath, bufIndex, err)
			}

			if len(bufferData) != buf.ByteLength {
				return nil, nil, fmt.Errorf("buffer %d: file %q size %d does not match expected length %d", bufIndex, buf.URI, len(bufferData), buf.ByteLength)
			}

			allBuffers = append(allBuffers, bufferData)
		}
	}

	return g, allBuffers, nil
}
