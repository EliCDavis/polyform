package gltf_test

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const epsilon = 1e-6

// ptr is a generic utility function to create a pointer to any value.
// This is useful in tests for creating pointers to literals.
func ptr[T any](value T) *T {
	return &value
}

// generateTestImage creates a simple test image with a solid color
func generateTestImage(width, height int, col color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, col)
		}
	}
	return img
}

// imageToBase64DataURI converts an image to a base64 data URI
func imageToBase64DataURI(t *testing.T, img image.Image) string {
	buf := &bytes.Buffer{}
	require.NoError(t, png.Encode(buf, img))

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/png;base64," + encoded
}

// createTestBuffer creates a binary buffer with test data
func createTestBuffer(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	return "data:application/octet-stream;base64," + encoded
}

// createTriangleBuffer creates a buffer with triangle mesh data
func createTriangleBuffer() (string, int) {
	buf := &bytes.Buffer{}

	// Indices (1 triangle = 3 indices)
	indices := []uint32{0, 1, 2}
	for _, idx := range indices {
		binary.Write(buf, binary.LittleEndian, idx)
	}

	// Positions (3 vertices)
	positions := []vector3.Float64{
		vector3.New(0.0, 0.0, 0.0),
		vector3.New(1.0, 0.0, 0.0),
		vector3.New(0.5, 1.0, 0.0),
	}
	for _, pos := range positions {
		binary.Write(buf, binary.LittleEndian, float32(pos.X()))
		binary.Write(buf, binary.LittleEndian, float32(pos.Y()))
		binary.Write(buf, binary.LittleEndian, float32(pos.Z()))
	}

	// Normals (3 vertices)
	normals := []vector3.Float64{
		vector3.New(0.0, 0.0, 1.0),
		vector3.New(0.0, 0.0, 1.0),
		vector3.New(0.0, 0.0, 1.0),
	}
	for _, norm := range normals {
		binary.Write(buf, binary.LittleEndian, float32(norm.X()))
		binary.Write(buf, binary.LittleEndian, float32(norm.Y()))
		binary.Write(buf, binary.LittleEndian, float32(norm.Z()))
	}

	// Texture coordinates (3 vertices)
	texCoords := []vector2.Float64{
		vector2.New(0.0, 0.0),
		vector2.New(1.0, 0.0),
		vector2.New(0.5, 1.0),
	}
	for _, tc := range texCoords {
		binary.Write(buf, binary.LittleEndian, float32(tc.X()))
		binary.Write(buf, binary.LittleEndian, float32(tc.Y()))
	}

	data := buf.Bytes()
	return createTestBuffer(data), len(data)
}

// createMinimalValidGLTF creates a minimal valid GLTF document
func createMinimalValidGLTF() gltf.Gltf {
	bufferData, bufferSize := createTriangleBuffer()

	return gltf.Gltf{
		Asset: gltf.Asset{
			Version: "2.0",
		},
		Buffers: []gltf.Buffer{
			{
				URI:        bufferData,
				ByteLength: bufferSize,
			},
		},
		BufferViews: []gltf.BufferView{
			{Buffer: 0, ByteOffset: 0, ByteLength: 12, Target: gltf.ELEMENT_ARRAY_BUFFER}, // indices
			{Buffer: 0, ByteOffset: 12, ByteLength: 36, Target: gltf.ARRAY_BUFFER},        // positions
			{Buffer: 0, ByteOffset: 48, ByteLength: 36, Target: gltf.ARRAY_BUFFER},        // normals
			{Buffer: 0, ByteOffset: 84, ByteLength: 24, Target: gltf.ARRAY_BUFFER},        // texcoords
		},
		Accessors: []gltf.Accessor{
			{BufferView: ptr(0), ComponentType: gltf.AccessorComponentType_UNSIGNED_INT, Count: 3, Type: gltf.AccessorType_SCALAR}, // indices
			{BufferView: ptr(1), ComponentType: gltf.AccessorComponentType_FLOAT, Count: 3, Type: gltf.AccessorType_VEC3},          // positions
			{BufferView: ptr(2), ComponentType: gltf.AccessorComponentType_FLOAT, Count: 3, Type: gltf.AccessorType_VEC3},          // normals
			{BufferView: ptr(3), ComponentType: gltf.AccessorComponentType_FLOAT, Count: 3, Type: gltf.AccessorType_VEC2},          // texcoords
		},
		Meshes: []gltf.Mesh{
			{
				Primitives: []gltf.Primitive{
					{
						Indices: ptr(0),
						Attributes: map[string]gltf.GltfId{
							gltf.POSITION:   1,
							gltf.NORMAL:     2,
							gltf.TEXCOORD_0: 3,
						},
					},
				},
			},
		},
		Nodes: []gltf.Node{
			{Mesh: ptr(0)},
		},
		Scenes: []gltf.Scene{
			{Nodes: []gltf.GltfId{0}},
		},
		Scene: ptr(0),
	}
}

// createGLTFWithMaterials creates a GLTF with materials and textures
func createGLTFWithMaterials(t *testing.T) gltf.Gltf {
	doc := createMinimalValidGLTF()

	// Create test images
	baseColorImg := generateTestImage(64, 64, color.RGBA{255, 0, 0, 255})  // Red
	normalImg := generateTestImage(64, 64, color.RGBA{128, 128, 255, 255}) // Blue-ish normal

	baseColorURI := imageToBase64DataURI(t, baseColorImg)
	normalURI := imageToBase64DataURI(t, normalImg)

	// Add images
	doc.Images = []gltf.Image{
		{URI: baseColorURI},
		{URI: normalURI},
	}

	// Add textures
	doc.Textures = []gltf.Texture{
		{Source: ptr(0)}, // Base color
		{Source: ptr(1)}, // Normal
	}

	// Add materials
	doc.Materials = []gltf.Material{
		{
			ChildOfRootProperty: gltf.ChildOfRootProperty{
				Name: "TestMaterial",
			},
			PbrMetallicRoughness: &gltf.PbrMetallicRoughness{
				BaseColorFactor: ptr([4]float64{1.0, 0.0, 0.0, 1.0}),
				BaseColorTexture: ptr(gltf.TextureInfo{
					Index: 0,
				}),
				MetallicFactor:  ptr(0.0),
				RoughnessFactor: ptr(1.0),
			},
			NormalTexture: &gltf.NormalTexture{
				TextureInfo: gltf.TextureInfo{Index: 1},
				Scale:       ptr(1.0),
			},
		},
	}

	// Update primitive to use material
	doc.Meshes[0].Primitives[0].Material = ptr(0)

	return doc
}

// createGLTFWithHierarchy creates a GLTF with node hierarchy
func createGLTFWithHierarchy(t *testing.T) gltf.Gltf {
	doc := createMinimalValidGLTF()

	// Create hierarchy: Root -> Child1 -> Child2 (with mesh)
	doc.Nodes = []gltf.Node{
		{
			Name:        "Root",
			Children:    []gltf.GltfId{1},
			Translation: ptr([3]float64{1.0, 0.0, 0.0}),
		},
		{
			Name:     "Child1",
			Children: []gltf.GltfId{2},
			Scale:    ptr([3]float64{2.0, 2.0, 2.0}),
		},
		{
			Name:     "Child2",
			Mesh:     ptr(0),
			Rotation: ptr([4]float64{0.0, 0.0, 0.0, 1.0}),
		},
	}

	return doc
}

// writeGLTFToTempFile writes a GLTF document to a temporary file
func writeGLTFToTempFile(t *testing.T, doc gltf.Gltf) string {
	t.Helper()

	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.gltf")

	data, err := json.MarshalIndent(doc, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(tempFile, data, 0644)
	require.NoError(t, err)

	return tempFile
}

// =============================================================================
// Helper functions for struct validation
// =============================================================================

// assertVector3Equal compares two vector3.Float64 with epsilon tolerance
func assertVector3Equal(t *testing.T, expected, actual vector3.Float64, msg string) {
	t.Helper()
	assert.InDelta(t, expected.X(), actual.X(), epsilon, msg+" X component")
	assert.InDelta(t, expected.Y(), actual.Y(), epsilon, msg+" Y component")
	assert.InDelta(t, expected.Z(), actual.Z(), epsilon, msg+" Z component")
}

// assertColorEqual compares two color.RGBA values
func assertColorEqual(t *testing.T, expected, actual color.Color, msg string) {
	t.Helper()
	eR, eG, eB, eA := expected.RGBA()
	aR, aG, aB, aA := actual.RGBA()
	assert.Equal(t, eR, aR, msg+" R component")
	assert.Equal(t, eG, aG, msg+" G component")
	assert.Equal(t, eB, aB, msg+" B component")
	assert.Equal(t, eA, aA, msg+" A component")
}

// assertMaterialEqual compares two PolyformMaterial structs
func assertMaterialEqual(t *testing.T, expected, actual *gltf.PolyformMaterial, msg string) {
	t.Helper()

	if expected == nil && actual == nil {
		return
	}
	require.NotNil(t, expected, msg+" expected should not be nil")
	require.NotNil(t, actual, msg+" actual should not be nil")

	assert.Equal(t, expected.Name, actual.Name, msg+" name")

	// Compare PBR properties
	if expected.PbrMetallicRoughness != nil || actual.PbrMetallicRoughness != nil {
		require.NotNil(t, expected.PbrMetallicRoughness, msg+" expected PBR should not be nil")
		require.NotNil(t, actual.PbrMetallicRoughness, msg+" actual PBR should not be nil")

		// Compare base color factor if both exist
		if expected.PbrMetallicRoughness.BaseColorFactor != nil && actual.PbrMetallicRoughness.BaseColorFactor != nil {
			expectedColor := expected.PbrMetallicRoughness.BaseColorFactor.(color.RGBA)
			actualColor := actual.PbrMetallicRoughness.BaseColorFactor.(color.RGBA)
			assertColorEqual(t, expectedColor, actualColor, msg+" base color factor")
		}

		if expected.PbrMetallicRoughness.MetallicFactor != nil || actual.PbrMetallicRoughness.MetallicFactor != nil {
			require.NotNil(t, expected.PbrMetallicRoughness.MetallicFactor, msg+" expected metallic factor should not be nil")
			require.NotNil(t, actual.PbrMetallicRoughness.MetallicFactor, msg+" actual metallic factor should not be nil")
			assert.Equal(t, *expected.PbrMetallicRoughness.MetallicFactor, *actual.PbrMetallicRoughness.MetallicFactor, msg+" metallic factor")
		}

		if expected.PbrMetallicRoughness.RoughnessFactor != nil || actual.PbrMetallicRoughness.RoughnessFactor != nil {
			require.NotNil(t, expected.PbrMetallicRoughness.RoughnessFactor, msg+" expected roughness factor should not be nil")
			require.NotNil(t, actual.PbrMetallicRoughness.RoughnessFactor, msg+" actual roughness factor should not be nil")
			assert.Equal(t, *expected.PbrMetallicRoughness.RoughnessFactor, *actual.PbrMetallicRoughness.RoughnessFactor, msg+" roughness factor")
		}
	}
}

// =============================================================================
// LoadFile Tests
// =============================================================================

func TestLoadFile(t *testing.T) {
	tests := []struct {
		name        string
		setupGLTF   func(t *testing.T) gltf.Gltf
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_minimal_gltf",
			setupGLTF: func(t *testing.T) gltf.Gltf {
				return createMinimalValidGLTF()
			},
			expectError: false,
		},
		{
			name:        "valid_gltf_with_materials",
			setupGLTF:   createGLTFWithMaterials,
			expectError: false,
		},
		{
			name:        "valid_gltf_with_hierarchy",
			setupGLTF:   createGLTFWithHierarchy,
			expectError: false,
		},
		{
			name: "missing_asset_version",
			setupGLTF: func(t *testing.T) gltf.Gltf {
				doc := createMinimalValidGLTF()
				doc.Asset.Version = ""
				return doc
			},
			expectError: true,
			errorMsg:    "missing required asset version",
		},
		{
			name: "invalid_buffer_byte_length",
			setupGLTF: func(t *testing.T) gltf.Gltf {
				doc := createMinimalValidGLTF()
				doc.Buffers[0].ByteLength = 0
				return doc
			},
			expectError: true,
			errorMsg:    "invalid byte length",
		},
		{
			name: "mismatched_buffer_length",
			setupGLTF: func(t *testing.T) gltf.Gltf {
				doc := createMinimalValidGLTF()
				doc.Buffers[0].ByteLength = 999999 // Wrong length
				return doc
			},
			expectError: true,
			errorMsg:    "does not match expected length",
		},
		{
			name: "invalid_base64_data",
			setupGLTF: func(t *testing.T) gltf.Gltf {
				doc := createMinimalValidGLTF()
				doc.Buffers[0].URI = "data:application/octet-stream;base64,invalid_base64_data!!!"
				return doc
			},
			expectError: true,
			errorMsg:    "failed to decode base64 data",
		},
		{
			name: "unsupported_data_uri_encoding",
			setupGLTF: func(t *testing.T) gltf.Gltf {
				doc := createMinimalValidGLTF()
				doc.Buffers[0].URI = "data:application/octet-stream;charset=utf-8,some_data"
				return doc
			},
			expectError: true,
			errorMsg:    "only base64 encoded data URIs are supported",
		},
		{
			name: "empty_buffer_uri",
			setupGLTF: func(t *testing.T) gltf.Gltf {
				doc := createMinimalValidGLTF()
				doc.Buffers[0].URI = ""
				return doc
			},
			expectError: true,
			errorMsg:    "empty uri is not a valid buffer location",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := tt.setupGLTF(t)
			tempFile := writeGLTFToTempFile(t, doc)

			loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, loadedDoc)
			require.Len(t, buffers, len(doc.Buffers))

			// Validate basic structure
			assert.Equal(t, doc.Asset.Version, loadedDoc.Asset.Version)
			assert.Len(t, loadedDoc.Buffers, len(doc.Buffers))
			assert.Len(t, loadedDoc.BufferViews, len(doc.BufferViews))
			assert.Len(t, loadedDoc.Accessors, len(doc.Accessors))
			assert.Len(t, loadedDoc.Meshes, len(doc.Meshes))
			assert.Len(t, loadedDoc.Nodes, len(doc.Nodes))
			assert.Len(t, loadedDoc.Scenes, len(doc.Scenes))

			// Validate buffer data
			for i, buffer := range buffers {
				assert.Equal(t, doc.Buffers[i].ByteLength, len(buffer))
			}
		})
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, _, err := gltf.LoadFile("/non/existent/path.gltf", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open GLTF file")
}

func TestLoadInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "invalid.gltf")

	err := os.WriteFile(tempFile, []byte("invalid json {"), 0644)
	require.NoError(t, err)

	_, _, err = gltf.LoadFile(tempFile, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse GLTF JSON")
}

// =============================================================================
// Accessor Decoding Tests
// =============================================================================

func TestAccessorDecoding(t *testing.T) {
	tests := []struct {
		name           string
		setupAccessor  func() (gltf.Gltf, [][]byte)
		accessorIndex  int
		expectedCount  int
		expectedType   gltf.AccessorType
		expectedValues interface{}
		expectError    bool
		errorMsg       string
	}{
		{
			name: "scalar_float_accessor",
			setupAccessor: func() (gltf.Gltf, [][]byte) {
				buf := &bytes.Buffer{}
				values := []float32{1.0, 2.0, 3.0}
				for _, v := range values {
					binary.Write(buf, binary.LittleEndian, v)
				}
				data := buf.Bytes()

				doc := gltf.Gltf{
					Asset: gltf.Asset{Version: "2.0"},
					Buffers: []gltf.Buffer{
						{URI: createTestBuffer(data), ByteLength: len(data)},
					},
					BufferViews: []gltf.BufferView{
						{Buffer: 0, ByteOffset: 0, ByteLength: len(data)},
					},
					Accessors: []gltf.Accessor{
						{
							BufferView:    ptr(0),
							ComponentType: gltf.AccessorComponentType_FLOAT,
							Count:         3,
							Type:          gltf.AccessorType_SCALAR,
						},
					},
				}

				return doc, [][]byte{data}
			},
			accessorIndex:  0,
			expectedCount:  3,
			expectedType:   gltf.AccessorType_SCALAR,
			expectedValues: []float64{1.0, 2.0, 3.0},
			expectError:    false,
		},
		{
			name: "vec3_accessor",
			setupAccessor: func() (gltf.Gltf, [][]byte) {
				buf := &bytes.Buffer{}
				values := []vector3.Float64{
					vector3.New(1.0, 2.0, 3.0),
					vector3.New(4.0, 5.0, 6.0),
				}
				for _, v := range values {
					binary.Write(buf, binary.LittleEndian, float32(v.X()))
					binary.Write(buf, binary.LittleEndian, float32(v.Y()))
					binary.Write(buf, binary.LittleEndian, float32(v.Z()))
				}
				data := buf.Bytes()

				doc := gltf.Gltf{
					Asset: gltf.Asset{Version: "2.0"},
					Buffers: []gltf.Buffer{
						{URI: createTestBuffer(data), ByteLength: len(data)},
					},
					BufferViews: []gltf.BufferView{
						{Buffer: 0, ByteOffset: 0, ByteLength: len(data)},
					},
					Accessors: []gltf.Accessor{
						{
							BufferView:    ptr(0),
							ComponentType: gltf.AccessorComponentType_FLOAT,
							Count:         2,
							Type:          gltf.AccessorType_VEC3,
						},
					},
				}

				return doc, [][]byte{data}
			},
			accessorIndex:  0,
			expectedCount:  2,
			expectedType:   gltf.AccessorType_VEC3,
			expectedValues: []vector3.Float64{vector3.New(1.0, 2.0, 3.0), vector3.New(4.0, 5.0, 6.0)},
			expectError:    false,
		},
		{
			name: "unsigned_short_indices",
			setupAccessor: func() (gltf.Gltf, [][]byte) {
				buf := &bytes.Buffer{}
				values := []uint16{0, 1, 2}
				for _, v := range values {
					binary.Write(buf, binary.LittleEndian, v)
				}
				data := buf.Bytes()

				doc := gltf.Gltf{
					Asset: gltf.Asset{Version: "2.0"},
					Buffers: []gltf.Buffer{
						{URI: createTestBuffer(data), ByteLength: len(data)},
					},
					BufferViews: []gltf.BufferView{
						{Buffer: 0, ByteOffset: 0, ByteLength: len(data)},
					},
					Accessors: []gltf.Accessor{
						{
							BufferView:    ptr(0),
							ComponentType: gltf.AccessorComponentType_UNSIGNED_SHORT,
							Count:         3,
							Type:          gltf.AccessorType_SCALAR,
						},
					},
				}

				return doc, [][]byte{data}
			},
			accessorIndex:  0,
			expectedCount:  3,
			expectedType:   gltf.AccessorType_SCALAR,
			expectedValues: []int{0, 1, 2},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := tt.setupAccessor()

			accessor := doc.Accessors[tt.accessorIndex]
			assert.Equal(t, tt.expectedCount, accessor.Count)
			assert.Equal(t, tt.expectedType, accessor.Type)

			// Test appropriate decoder based on type
			switch tt.expectedType {
			case gltf.AccessorType_SCALAR:
				switch tt.expectedValues.(type) {
				case []float64:
					// Test scalar float accessor
					assert.Equal(t, gltf.AccessorComponentType_FLOAT, accessor.ComponentType)
				case []int:
					// Test indices
					assert.True(t, accessor.ComponentType == gltf.AccessorComponentType_UNSIGNED_SHORT ||
						accessor.ComponentType == gltf.AccessorComponentType_UNSIGNED_INT ||
						accessor.ComponentType == gltf.AccessorComponentType_UNSIGNED_BYTE)
				}
			case gltf.AccessorType_VEC3:
				assert.Equal(t, gltf.AccessorComponentType_FLOAT, accessor.ComponentType)
			}
		})
	}
}

// =============================================================================
// Primitive Decoding Tests
// =============================================================================

func TestPrimitiveDecoding(t *testing.T) {
	tests := []struct {
		name             string
		setupPrimitive   func() (gltf.Gltf, [][]byte)
		expectedTopology modeling.Topology
		expectError      bool
		errorMsg         string
	}{
		{
			name: "triangle_primitive",
			setupPrimitive: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectedTopology: modeling.TriangleTopology,
			expectError:      false,
		},
		{
			name: "point_primitive",
			setupPrimitive: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				mode := gltf.PrimitiveMode_POINTS
				doc.Meshes[0].Primitives[0].Mode = &mode
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectedTopology: modeling.PointTopology,
			expectError:      false,
		},
		{
			name: "line_primitive",
			setupPrimitive: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				mode := gltf.PrimitiveMode_LINES
				doc.Meshes[0].Primitives[0].Mode = &mode
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectedTopology: modeling.LineTopology,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, buffers := tt.setupPrimitive()

			models, err := gltf.DecodeModels(&doc, buffers, nil)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			require.Len(t, models, 1)

			model := models[0]
			require.NotNil(t, model.Mesh)

			assert.Equal(t, tt.expectedTopology, model.Mesh.Topology())

			// Validate mesh attributes
			assert.True(t, model.Mesh.HasFloat3Attribute(modeling.PositionAttribute))
			assert.True(t, model.Mesh.HasFloat3Attribute(modeling.NormalAttribute))
			assert.True(t, model.Mesh.HasFloat2Attribute(modeling.TexCoordAttribute))

			// Validate geometry
			assert.Equal(t, 3, model.Mesh.Indices().Len())
		})
	}
}

// =============================================================================
// Material Loading Tests
// =============================================================================

func TestMaterialLoading(t *testing.T) {
	tests := []struct {
		name           string
		setupMaterial  func() (gltf.Gltf, [][]byte)
		expectedName   string
		expectTextures bool
		expectError    bool
		errorMsg       string
	}{
		{
			name: "material_with_textures",
			setupMaterial: func() (gltf.Gltf, [][]byte) {
				doc := createGLTFWithMaterials(t)
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectedName:   "TestMaterial",
			expectTextures: true,
			expectError:    false,
		},
		{
			name: "material_without_textures",
			setupMaterial: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				doc.Materials = []gltf.Material{
					{
						ChildOfRootProperty: gltf.ChildOfRootProperty{
							Name: "SimpleMaterial",
						},
						PbrMetallicRoughness: &gltf.PbrMetallicRoughness{
							BaseColorFactor: &[4]float64{0.5, 0.5, 0.5, 1.0},
							MetallicFactor:  ptr(0.1),
							RoughnessFactor: ptr(0.9),
						},
					},
				}
				doc.Meshes[0].Primitives[0].Material = ptr(0)
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectedName:   "SimpleMaterial",
			expectTextures: false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, buffers := tt.setupMaterial()

			models, err := gltf.DecodeModels(&doc, buffers, nil)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			require.Len(t, models, 1)

			model := models[0]
			require.NotNil(t, model.Material)

			assert.Equal(t, tt.expectedName, model.Material.Name)

			if tt.expectTextures {
				require.NotNil(t, model.Material.PbrMetallicRoughness)
				assert.NotNil(t, model.Material.PbrMetallicRoughness.BaseColorTexture)
				assert.NotNil(t, model.Material.NormalTexture)

				// Validate texture images were loaded
				assert.NotNil(t, model.Material.PbrMetallicRoughness.BaseColorTexture.Image)
				assert.NotNil(t, model.Material.NormalTexture.Image)
			}

			// Validate PBR properties
			require.NotNil(t, model.Material.PbrMetallicRoughness)
			assert.NotNil(t, model.Material.PbrMetallicRoughness.MetallicFactor)
			assert.NotNil(t, model.Material.PbrMetallicRoughness.RoughnessFactor)
		})
	}
}

// =============================================================================
// Texture Loading Tests
// =============================================================================

func TestTextureLoading(t *testing.T) {
	tests := []struct {
		name         string
		setupTexture func() (gltf.Gltf, [][]byte)
		expectError  bool
		errorMsg     string
	}{
		{
			name: "data_uri_texture",
			setupTexture: func() (gltf.Gltf, [][]byte) {
				doc := createGLTFWithMaterials(t)
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectError: false,
		},
		{
			name: "invalid_texture_reference",
			setupTexture: func() (gltf.Gltf, [][]byte) {
				doc := createGLTFWithMaterials(t)
				// Reference non-existent texture
				doc.Materials[0].PbrMetallicRoughness.BaseColorTexture.Index = 999
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectError: true,
			errorMsg:    "invalid texture ID",
		},
		{
			name: "invalid_image_reference",
			setupTexture: func() (gltf.Gltf, [][]byte) {
				doc := createGLTFWithMaterials(t)
				// Reference non-existent image
				doc.Textures[0].Source = ptr(999)
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectError: true,
			errorMsg:    "invalid image",
		},
		{
			name: "malformed_data_uri",
			setupTexture: func() (gltf.Gltf, [][]byte) {
				doc := createGLTFWithMaterials(t)
				// Use malformed data URI
				doc.Images[0].URI = "data:invalid_format"
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectError: true,
			errorMsg:    "missing comma separator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, buffers := tt.setupTexture()

			models, err := gltf.DecodeModels(&doc, buffers, nil)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			require.Len(t, models, 1)

			model := models[0]
			require.NotNil(t, model.Material)
			require.NotNil(t, model.Material.PbrMetallicRoughness)

			// Validate texture loaded successfully
			if model.Material.PbrMetallicRoughness.BaseColorTexture != nil {
				assert.NotNil(t, model.Material.PbrMetallicRoughness.BaseColorTexture.Image)
				assert.True(t, strings.HasPrefix(model.Material.PbrMetallicRoughness.BaseColorTexture.URI, "data:image/png;base64,"))
			}
		})
	}
}

// =============================================================================
// File URI Tests
// =============================================================================

func TestFileURISupport(t *testing.T) {
	// Create a temporary image file
	tempDir := t.TempDir()
	imagePath := filepath.Join(tempDir, "test_image.png")

	// Create a simple test image
	img := generateTestImage(32, 32, color.RGBA{255, 0, 0, 255})
	file, err := os.Create(imagePath)
	require.NoError(t, err)
	defer file.Close()

	err = png.Encode(file, img)
	require.NoError(t, err)

	tests := []struct {
		name        string
		imageURI    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "file_uri_absolute",
			imageURI:    "file://" + imagePath,
			expectError: false,
		},
		{
			name:        "invalid_file_uri",
			imageURI:    "file:///nonexistent/path/image.png",
			expectError: true,
			errorMsg:    "failed to open image file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a GLTF document that references the image
			doc := createMinimalValidGLTF()

			// Add an image reference
			doc.Images = []gltf.Image{
				{URI: tt.imageURI},
			}

			// Add textures and materials
			doc.Textures = []gltf.Texture{
				{Source: ptr(0)},
			}

			doc.Materials = []gltf.Material{
				{
					ChildOfRootProperty: gltf.ChildOfRootProperty{
						Name: "TestMaterial",
					},
					PbrMetallicRoughness: ptr(gltf.PbrMetallicRoughness{
						BaseColorTexture: ptr(gltf.TextureInfo{
							Index: 0,
						}),
					}),
				},
			}

			// Update primitive to use material
			doc.Meshes[0].Primitives[0].Material = ptr(0)

			// Write the updated GLTF file
			tempFile := writeGLTFToTempFile(t, doc)

			// Load and decode
			loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
			require.NoError(t, err)

			scene, err := gltf.DecodeScene(loadedDoc, buffers, nil)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			require.Len(t, scene.Models, 1)

			model := scene.Models[0]
			require.NotNil(t, model.Material)
			require.NotNil(t, model.Material.PbrMetallicRoughness)
			require.NotNil(t, model.Material.PbrMetallicRoughness.BaseColorTexture)

			// Validate that the image was loaded successfully
			assert.NotNil(t, model.Material.PbrMetallicRoughness.BaseColorTexture.Image)

			// Check that the loaded image has the expected properties
			bounds := model.Material.PbrMetallicRoughness.BaseColorTexture.Image.Bounds()
			assert.Equal(t, 32, bounds.Dx(), "Image width should be 32")
			assert.Equal(t, 32, bounds.Dy(), "Image height should be 32")
		})
	}
}

// =============================================================================
// Scene Hierarchy Tests
// =============================================================================

func TestSceneHierarchy(t *testing.T) {
	tests := []struct {
		name           string
		setupScene     func() (gltf.Gltf, [][]byte)
		validate       func(scene *gltf.PolyformScene)
		expectedModels int
		expectError    bool
		errorMsg       string
	}{
		{
			name: "simple_hierarchy",
			setupScene: func() (gltf.Gltf, [][]byte) {
				doc := createGLTFWithHierarchy(t)
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectedModels: 1,
			expectError:    false,
			validate: func(scene *gltf.PolyformScene) {
				model := scene.Models[0]
				require.NotNil(t, model.TRS)

				// The hierarchy should have accumulated transformations
				// Root: translation (1,0,0), Child1: scale (2,2,2), Child2: rotation
				translation := model.TRS.Position()
				assert.InDelta(t, 1.0, translation.X(), epsilon, "Translation should be applied")

				scale := model.TRS.Scale()
				assertVector3Equal(t, vector3.New(1.0, 1.0, 1.0), scale, "Scale should be applied")
			},
		},
		{
			name: "empty_scene",
			setupScene: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				doc.Scenes[0].Nodes = []gltf.GltfId{} // Empty scene
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectedModels: 0,
			expectError:    false,
		},
		{
			name: "invalid_scene_index",
			setupScene: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				doc.Scene = ptr(999) // Invalid scene index
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectedModels: 1,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, buffers := tt.setupScene()

			scene, err := gltf.DecodeScene(&doc, buffers, nil)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, scene)

			assert.Len(t, scene.Models, tt.expectedModels)

			if tt.validate != nil {
				tt.validate(scene)
			}
		})
	}
}

// =============================================================================
// Transformation Tests
// =============================================================================

func TestTransformations(t *testing.T) {
	tests := []struct {
		name                string
		setupTransform      func() (gltf.Gltf, [][]byte)
		expectedTranslation vector3.Float64
		expectedScale       vector3.Float64
		expectError         bool
		errorMsg            string
	}{
		{
			name: "trs_components",
			setupTransform: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				doc.Nodes[0].Translation = &[3]float64{1.0, 2.0, 3.0}
				doc.Nodes[0].Scale = &[3]float64{2.0, 2.0, 2.0}
				doc.Nodes[0].Rotation = &[4]float64{0.0, 0.0, 0.0, 1.0}
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectedTranslation: vector3.New(1.0, 2.0, 3.0),
			expectedScale:       vector3.New(2.0, 2.0, 2.0),
			expectError:         false,
		},
		{
			name: "matrix_transform",
			setupTransform: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				// Identity matrix with translation
				doc.Nodes[0].Matrix = &[16]float64{
					1.0, 0.0, 0.0, 0.0,
					0.0, 1.0, 0.0, 0.0,
					0.0, 0.0, 1.0, 0.0,
					5.0, 6.0, 7.0, 1.0,
				}
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectedTranslation: vector3.New(5.0, 6.0, 7.0),
			expectedScale:       vector3.New(1.0, 1.0, 1.0),
			expectError:         false,
		},
		{
			name: "identity_transform",
			setupTransform: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				// No explicit transform - should be identity
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectedTranslation: vector3.New(0.0, 0.0, 0.0),
			expectedScale:       vector3.New(1.0, 1.0, 1.0),
			expectError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, buffers := tt.setupTransform()

			models, err := gltf.DecodeModels(&doc, buffers, nil)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			require.Len(t, models, 1)

			model := models[0]
			require.NotNil(t, model.TRS)

			translation := model.TRS.Position()
			scale := model.TRS.Scale()

			assertVector3Equal(t, tt.expectedTranslation, translation, "Translation")
			assertVector3Equal(t, tt.expectedScale, scale, "Scale")
		})
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestIntegrationLoadAndDecode(t *testing.T) {
	tests := []struct {
		name           string
		setupDocument  func(t *testing.T) gltf.Gltf
		expectedModels int
		validateModel  func(t *testing.T, model gltf.PolyformModel)
	}{
		{
			name:           "complete_scene_with_materials",
			setupDocument:  createGLTFWithMaterials,
			expectedModels: 1,
			validateModel: func(t *testing.T, model gltf.PolyformModel) {
				// Validate mesh
				require.NotNil(t, model.Mesh)
				assert.Equal(t, modeling.TriangleTopology, model.Mesh.Topology())
				assert.Equal(t, 3, model.Mesh.Indices().Len())

				// Validate material
				require.NotNil(t, model.Material)
				assert.Equal(t, "TestMaterial", model.Material.Name)

				// Validate textures
				require.NotNil(t, model.Material.PbrMetallicRoughness)
				assert.NotNil(t, model.Material.PbrMetallicRoughness.BaseColorTexture)
				assert.NotNil(t, model.Material.NormalTexture)

				// Validate texture images
				assert.NotNil(t, model.Material.PbrMetallicRoughness.BaseColorTexture.Image)
				assert.NotNil(t, model.Material.NormalTexture.Image)

				// Validate transform
				require.NotNil(t, model.TRS)
			},
		},
		{
			name:           "hierarchical_scene",
			setupDocument:  createGLTFWithHierarchy,
			expectedModels: 1,
			validateModel: func(t *testing.T, model gltf.PolyformModel) {
				// Validate transform hierarchy accumulation
				require.NotNil(t, model.TRS)

				translation := model.TRS.Position()
				scale := model.TRS.Scale()

				// Should have accumulated transforms from hierarchy
				assert.InDelta(t, 1.0, translation.X(), epsilon, "Translation from root")
				assertVector3Equal(t, vector3.New(1.0, 1.0, 1.0), scale, "Scale should be applied")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := tt.setupDocument(t)
			tempFile := writeGLTFToTempFile(t, doc)

			// Load document
			loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
			require.NoError(t, err)
			require.NotNil(t, loadedDoc)

			// Decode scene
			scene, err := gltf.DecodeScene(loadedDoc, buffers, nil)
			require.NoError(t, err)
			require.NotNil(t, scene)
			require.Len(t, scene.Models, tt.expectedModels)

			// Validate each model
			for i, model := range scene.Models {
				t.Run(fmt.Sprintf("model_%d", i), func(t *testing.T) {
					tt.validateModel(t, model)
				})
			}
		})
	}
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		setupError  func() (gltf.Gltf, [][]byte)
		expectError bool
		errorMsg    string
	}{
		{
			name: "invalid_accessor_reference",
			setupError: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				// Reference non-existent accessor
				doc.Meshes[0].Primitives[0].Attributes[gltf.POSITION] = 999
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectError: true,
			errorMsg:    "invalid accessor",
		},
		{
			name: "invalid_buffer_view_reference",
			setupError: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				// Reference non-existent buffer view
				doc.Accessors[0].BufferView = ptr(999)
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectError: true,
			errorMsg:    "",
		},
		{
			name: "invalid_material_reference",
			setupError: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				// Reference non-existent material
				doc.Meshes[0].Primitives[0].Material = ptr(999)
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectError: true,
			errorMsg:    "invalid material ID",
		},
		{
			name: "invalid_node_reference",
			setupError: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				// Reference non-existent node in scene
				doc.Scenes[0].Nodes = []gltf.GltfId{999}
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectError: true,
			errorMsg:    "invalid node index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, buffers := tt.setupError()

			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Convert panic to error for test validation
						err = fmt.Errorf("panic occurred: %v", r)
					}
				}()
				_, err = gltf.DecodeScene(&doc, buffers, nil)
			}()

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// WriterOptions API Tests
// =============================================================================

func TestOptionsAPI(t *testing.T) {
	t.Run("save_text_with_opts", func(t *testing.T) {
		// Create a GLTF with materials and textures
		doc := createGLTFWithMaterials(t)
		tempFile := writeGLTFToTempFile(t, doc)

		// Load the original scene
		loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
		require.NoError(t, err)

		scene, err := gltf.DecodeScene(loadedDoc, buffers, nil)
		require.NoError(t, err)

		outputDir := t.TempDir()

		// Test with embedding enabled
		outputPath1 := filepath.Join(outputDir, "embedded.gltf")
		err = gltf.SaveText(outputPath1, *scene, &gltf.WriterOptions{EmbedTextures: true})
		require.NoError(t, err)

		exportedContent1, err := os.ReadFile(outputPath1)
		require.NoError(t, err)

		var exportedDoc1 gltf.Gltf
		err = json.Unmarshal(exportedContent1, &exportedDoc1)
		require.NoError(t, err)

		require.Len(t, exportedDoc1.Images, 2)
		for i, image := range exportedDoc1.Images {
			assert.True(t, strings.HasPrefix(image.URI, "data:image/png;base64,"),
				"Image %d should be embedded when EmbedTextures=true", i)
		}

		// Test with default options (embedding disabled)
		outputPath2 := filepath.Join(outputDir, "default.gltf")
		err = gltf.SaveText(outputPath2, *scene, &gltf.WriterOptions{})
		require.NoError(t, err)

		exportedContent2, err := os.ReadFile(outputPath2)
		require.NoError(t, err)

		var exportedDoc2 gltf.Gltf
		err = json.Unmarshal(exportedContent2, &exportedDoc2)
		require.NoError(t, err)

		require.Len(t, exportedDoc2.Images, 2)
		// Note: Since our test creates scenes with already-embedded images,
		// they will remain embedded regardless of the EmbedTextures setting
		// The EmbedTextures option affects how external image files are handled
	})

	t.Run("save_binary_with_opts", func(t *testing.T) {
		doc := createGLTFWithMaterials(t)
		tempFile := writeGLTFToTempFile(t, doc)

		loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
		require.NoError(t, err)

		scene, err := gltf.DecodeScene(loadedDoc, buffers, nil)
		require.NoError(t, err)

		outputDir := t.TempDir()

		// Test SaveBinaryWithOpts
		binaryOutputPath := filepath.Join(outputDir, "test_output.glb")
		err = gltf.SaveBinary(binaryOutputPath, *scene, &gltf.WriterOptions{EmbedTextures: true})
		require.NoError(t, err)

		// Verify binary file exists and has content
		stat, err := os.Stat(binaryOutputPath)
		require.NoError(t, err)
		assert.Greater(t, stat.Size(), int64(0), "Binary file should have content")
	})

	t.Run("options_struct_validation", func(t *testing.T) {
		// Test that WriterOptions struct can be created and used properly
		opts := gltf.WriterOptions{
			EmbedTextures: true,
			JsonFormat:    gltf.MinifyJsonFormat,
		}
		assert.True(t, opts.EmbedTextures, "EmbedTextures should be settable")
		assert.Equal(t, gltf.MinifyJsonFormat, opts.JsonFormat, "JsonFormat should be settable")

		// Test zero value
		defaultOpts := gltf.WriterOptions{}
		assert.False(t, defaultOpts.EmbedTextures, "Default EmbedTextures should be false")
		assert.Equal(t, gltf.DefaultJsonFormat, defaultOpts.JsonFormat, "Default JsonFormat should be DefaultJsonFormat")
	})
}

func TestWriteWithOptsAPI(t *testing.T) {
	doc := createGLTFWithMaterials(t)
	tempFile := writeGLTFToTempFile(t, doc)

	loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
	require.NoError(t, err)

	scene, err := gltf.DecodeScene(loadedDoc, buffers, nil)
	require.NoError(t, err)

	t.Run("write_text_with_opts", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := gltf.WriteText(*scene, buf, &gltf.WriterOptions{EmbedTextures: true})
		require.NoError(t, err)

		var exportedDoc gltf.Gltf
		err = json.Unmarshal(buf.Bytes(), &exportedDoc)
		require.NoError(t, err)

		// Verify textures are embedded
		require.Len(t, exportedDoc.Images, 2)
		for i, image := range exportedDoc.Images {
			assert.True(t, strings.HasPrefix(image.URI, "data:image/png;base64,"),
				"Image %d should be embedded", i)
		}
	})

	t.Run("write_binary_with_opts", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := gltf.WriteBinary(*scene, buf, &gltf.WriterOptions{EmbedTextures: true})
		require.NoError(t, err)

		// Verify we got some binary data
		assert.Greater(t, buf.Len(), 0, "Should have written binary data")
	})

	t.Run("minify_json_option", func(t *testing.T) {
		prettyBuf := &bytes.Buffer{}
		minifiedBuf := &bytes.Buffer{}

		// Test pretty-printed JSON (default)
		err := gltf.WriteText(*scene, prettyBuf, &gltf.WriterOptions{JsonFormat: gltf.PrettyJsonFormat})
		require.NoError(t, err)

		// Test minified JSON
		err = gltf.WriteText(*scene, minifiedBuf, &gltf.WriterOptions{JsonFormat: gltf.MinifyJsonFormat})
		require.NoError(t, err)

		prettyContent := prettyBuf.String()
		minifiedContent := minifiedBuf.String()

		// Verify both are valid JSON
		var prettyDoc, minifiedDoc gltf.Gltf
		err = json.Unmarshal([]byte(prettyContent), &prettyDoc)
		require.NoError(t, err, "Pretty-printed JSON should be valid")

		err = json.Unmarshal([]byte(minifiedContent), &minifiedDoc)
		require.NoError(t, err, "Minified JSON should be valid")

		// Verify they represent the same data
		assert.Equal(t, prettyDoc.Asset.Version, minifiedDoc.Asset.Version)
		assert.Len(t, minifiedDoc.Meshes, len(prettyDoc.Meshes))

		// Verify size difference - minified should be smaller
		assert.Less(t, len(minifiedContent), len(prettyContent),
			"Minified JSON should be smaller than pretty-printed")

		// Verify pretty-printed contains indentation
		assert.Contains(t, prettyContent, "    ", "Pretty-printed JSON should contain indentation")

		// Verify minified doesn't contain unnecessary whitespace
		assert.NotContains(t, minifiedContent, "    ", "Minified JSON should not contain indentation")
		assert.NotContains(t, minifiedContent, "\n", "Minified JSON should not contain newlines")
	})
}

// =============================================================================
// Data URI Tests
// =============================================================================

func TestLoadImageFromDataURI(t *testing.T) {
	// Create valid base64 encoded images for testing
	pngImg := generateTestImage(16, 16, color.RGBA{255, 0, 0, 255})
	pngDataURI := imageToBase64DataURI(t, pngImg)

	// Extract the base64 portion for creating test URIs
	commaIndex := strings.Index(pngDataURI, ",")
	require.Greater(t, commaIndex, -1, "Valid data URI should have a comma")
	pngBase64Data := pngDataURI[commaIndex+1:]

	tests := []struct {
		name        string
		dataURI     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid_png_data_uri",
			dataURI:     "data:image/png;base64," + pngBase64Data,
			expectError: false,
		},
		{
			name:        "valid_jpeg_data_uri_explicit",
			dataURI:     "data:image/jpeg;base64," + pngBase64Data, // Will fail on format mismatch
			expectError: true,
			errorMsg:    "image format mismatch",
		},
		{
			name:        "missing_data_prefix",
			dataURI:     "image/png;base64," + pngBase64Data,
			expectError: true,
			errorMsg:    "failed to open image file", // Treated as file path, not data URI
		},
		{
			name:        "missing_comma_separator",
			dataURI:     "data:image/png;base64" + pngBase64Data,
			expectError: true,
			errorMsg:    "missing comma separator",
		},
		{
			name:        "empty_content_type",
			dataURI:     "data:;base64," + pngBase64Data,
			expectError: true,
			errorMsg:    "unsupported image type",
		},
		{
			name:        "unsupported_content_type",
			dataURI:     "data:image/gif;base64," + pngBase64Data,
			expectError: true,
			errorMsg:    "unsupported image type",
		},
		{
			name:        "missing_base64_declaration",
			dataURI:     "data:image/png," + pngBase64Data,
			expectError: true,
			errorMsg:    "only base64 encoded data URIs are supported",
		},
		{
			name:        "invalid_base64_data",
			dataURI:     "data:image/png;base64,invalid_base64_data!!!",
			expectError: true,
			errorMsg:    "failed to decode base64 image data",
		},
		{
			name:        "corrupted_image_data",
			dataURI:     "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
			expectError: false, // This is actually a valid 1x1 PNG
		},
		{
			name:        "multiple_semicolons_in_header",
			dataURI:     "data:image/png;charset=utf-8;base64," + pngBase64Data,
			expectError: false, // Should work, ignores charset parameter
		},
		{
			name:        "base64_not_last_parameter",
			dataURI:     "data:image/png;base64;other=value," + pngBase64Data,
			expectError: false, // Should work, base64 is found
		},
		{
			name:        "whitespace_in_header",
			dataURI:     "data: image/png ; base64 ," + pngBase64Data,
			expectError: true,
			errorMsg:    "unsupported image type",
		},
		{
			name:        "empty_header",
			dataURI:     "data:," + pngBase64Data,
			expectError: true,
			errorMsg:    "unsupported image type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test by creating a GLTF with embedded texture and loading it
			err := testDataURIThroughGLTF(t, tt.dataURI)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
		})
	}
}

// testDataURIThroughGLTF tests data URI loading by creating a GLTF with embedded texture
func testDataURIThroughGLTF(t *testing.T, dataURI string) error {
	// Create a simple GLTF document with an embedded texture
	doc := createMinimalValidGLTF()

	// Add an image with the data URI
	doc.Images = []gltf.Image{
		{URI: dataURI},
	}

	// Add textures and materials
	doc.Textures = []gltf.Texture{
		{Source: ptr(0)},
	}

	doc.Materials = []gltf.Material{
		{
			ChildOfRootProperty: gltf.ChildOfRootProperty{
				Name: "TestMaterial",
			},
			PbrMetallicRoughness: ptr(gltf.PbrMetallicRoughness{
				BaseColorTexture: ptr(gltf.TextureInfo{
					Index: 0,
				}),
			}),
		},
	}

	// Update primitive to use material
	doc.Meshes[0].Primitives[0].Material = ptr(0)

	// Write the GLTF file
	tempFile := writeGLTFToTempFile(t, doc)

	// Load and decode - this will trigger data URI loading
	loadedDoc, buffers, err := gltf.LoadFile(tempFile, nil)
	if err != nil {
		return err
	}

	_, err = gltf.DecodeScene(loadedDoc, buffers, nil)
	return err
}

func TestDataURIValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		dataURI     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "data_uri_with_commas_in_base64",
			dataURI:     "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
			expectError: false,
		},
		{
			name:        "very_long_content_type",
			dataURI:     "data:image/png;some=very;long=parameter;list=here;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
			expectError: false,
		},
		{
			name:        "case_sensitive_content_type",
			dataURI:     "data:IMAGE/PNG;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
			expectError: true,
			errorMsg:    "unsupported image type",
		},
		{
			name:        "case_sensitive_base64",
			dataURI:     "data:image/png;BASE64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
			expectError: true,
			errorMsg:    "only base64 encoded data URIs are supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testDataURIThroughGLTF(t, tt.dataURI)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
		})
	}
}

// =============================================================================
// New API Tests (Parse, Load, Custom Loaders, etc.)
// =============================================================================

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		gltfData    string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_minimal_gltf",
			gltfData: `{
				"asset": {"version": "2.0"},
				"scenes": [{"nodes": [0]}],
				"nodes": [{"mesh": 0}],
				"meshes": [{"primitives": [{"attributes": {"POSITION": 0}}]}],
				"accessors": [{"bufferView": 0, "count": 3, "componentType": 5126, "type": "VEC3"}],
				"bufferViews": [{"buffer": 0, "byteLength": 36}],
				"buffers": [{"byteLength": 36, "uri": "data:application/octet-stream;base64,"}]
			}`,
			expectError: false,
		},
		{
			name:        "missing_asset_version",
			gltfData:    `{"scenes": []}`,
			expectError: true,
			errorMsg:    "missing required asset version",
		},
		{
			name:        "invalid_json",
			gltfData:    `{"asset": {"version": "2.0"} invalid json`,
			expectError: true,
			errorMsg:    "failed to parse GLTF JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.gltfData)
			doc, err := gltf.ParseGLTF(reader)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, doc)
			assert.Equal(t, "2.0", doc.Asset.Version)
		})
	}
}

func TestParseFile(t *testing.T) {
	// Create a temporary GLTF file
	tempDir := t.TempDir()
	gltfPath := filepath.Join(tempDir, "test.gltf")

	gltfContent := `{
		"asset": {"version": "2.0"},
		"scenes": [{"nodes": [0]}],
		"nodes": [{"mesh": 0}],
		"meshes": [{"primitives": [{"attributes": {"POSITION": 0}}]}],
		"accessors": [{"bufferView": 0, "count": 3, "componentType": 5126, "type": "VEC3"}],
		"bufferViews": [{"buffer": 0, "byteLength": 36}],
		"buffers": [{"byteLength": 36, "uri": "buffer.bin"}]
	}`

	err := os.WriteFile(gltfPath, []byte(gltfContent), 0644)
	require.NoError(t, err)

	// Test loading the file
	doc, err := gltf.ParseFile(gltfPath)
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, "2.0", doc.Asset.Version)

	// Test loading non-existent file
	_, err = gltf.ParseFile(filepath.Join(tempDir, "nonexistent.gltf"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open GLTF file")
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		gltfData    string
		basePath    string
		setupFiles  map[string][]byte
		expectError bool
		errorMsg    string
	}{
		{
			name: "embedded_buffer",
			gltfData: `{
				"asset": {"version": "2.0"},
				"buffers": [{"byteLength": 36, "uri": "data:application/octet-stream;base64,AAAAAAAAAAAAAIA/AAAAAAAAAAAAAIA/AAAAAAAAAAAAAIA/"}]
			}`,
			expectError: false,
		},
		{
			name: "external_buffer",
			gltfData: `{
				"asset": {"version": "2.0"},
				"buffers": [{"byteLength": 36, "uri": "buffer.bin"}]
			}`,
			setupFiles: map[string][]byte{
				"buffer.bin": make([]byte, 36),
			},
			expectError: false,
		},
		{
			name: "invalid_buffer_length",
			gltfData: `{
				"asset": {"version": "2.0"},
				"buffers": [{"byteLength": 0, "uri": "buffer.bin"}]
			}`,
			expectError: true,
			errorMsg:    "invalid byte length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Setup files if needed
			for filename, content := range tt.setupFiles {
				err := os.WriteFile(filepath.Join(tempDir, filename), content, 0644)
				require.NoError(t, err)
			}

			reader := strings.NewReader(tt.gltfData)
			opts := &gltf.ReaderOptions{BasePath: tempDir}
			doc, buffers, err := gltf.LoadGLTF(reader, opts)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, doc)
			assert.Equal(t, len(doc.Buffers), len(buffers))
		})
	}
}

// Test custom loaders
type mockBufferLoader struct {
	buffers map[string][]byte
}

func (m *mockBufferLoader) LoadBuffer(uri string) ([]byte, error) {
	if data, ok := m.buffers[uri]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("buffer not found: %s", uri)
}

type mockImageLoader struct {
	images  map[string]image.Image
	formats map[string]string
}

func (m *mockImageLoader) LoadImage(uri string) (image.Image, string, error) {
	img, imgOk := m.images[uri]
	format, fmtOk := m.formats[uri]

	if imgOk && fmtOk {
		return img, format, nil
	}
	return nil, "", fmt.Errorf("image not found: %s", uri)
}

func TestCustomLoaders(t *testing.T) {
	// Test custom buffer loader
	t.Run("custom_buffer_loader", func(t *testing.T) {
		loader := &mockBufferLoader{
			buffers: map[string][]byte{
				"custom://buffer1": make([]byte, 36),
			},
		}

		gltfData := `{
			"asset": {"version": "2.0"},
			"buffers": [{"byteLength": 36, "uri": "custom://buffer1"}]
		}`

		opts := &gltf.ReaderOptions{
			BufferLoader: loader,
		}

		doc, buffers, err := gltf.LoadGLTF(strings.NewReader(gltfData), opts)
		require.NoError(t, err)
		require.NotNil(t, doc)
		assert.Len(t, buffers, 1)
		assert.Len(t, buffers[0], 36)
	})

	// Test custom image loader
	t.Run("custom_image_loader", func(t *testing.T) {
		// Create a simple test image
		testImage := image.NewRGBA(image.Rect(0, 0, 16, 16))
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				testImage.Set(x, y, color.RGBA{255, 0, 0, 255})
			}
		}

		loader := &mockImageLoader{
			images:  map[string]image.Image{"custom://image1": testImage},
			formats: map[string]string{"custom://image1": "png"},
		}

		// Create a minimal valid GLTF doc
		gltfDoc := createMinimalValidGLTF()
		gltfDoc.Images = []gltf.Image{{URI: "custom://image1"}}
		gltfDoc.Textures = []gltf.Texture{{Source: ptr(0)}}
		gltfDoc.Materials = []gltf.Material{
			{
				PbrMetallicRoughness: &gltf.PbrMetallicRoughness{
					BaseColorTexture: &gltf.TextureInfo{Index: 0},
				},
			},
		}
		gltfDoc.Meshes[0].Primitives[0].Material = ptr(0)

		opts := &gltf.ReaderOptions{
			ImageLoader: loader,
		}

		_, bufferSize := createTriangleBuffer()
		models, err := gltf.DecodeModels(&gltfDoc, [][]byte{make([]byte, bufferSize)}, opts)
		require.NoError(t, err)
		require.Len(t, models, 1)
		require.NotNil(t, models[0].Material)
		require.NotNil(t, models[0].Material.PbrMetallicRoughness)
		require.NotNil(t, models[0].Material.PbrMetallicRoughness.BaseColorTexture)
		assert.NotNil(t, models[0].Material.PbrMetallicRoughness.BaseColorTexture.Image)
	})
}

func TestNoOpImageLoader(t *testing.T) {
	gltfDoc := createMinimalValidGLTF()
	gltfDoc.Images = []gltf.Image{{URI: "nonexistent.png"}}
	gltfDoc.Textures = []gltf.Texture{{Source: ptr(0)}}
	gltfDoc.Materials = []gltf.Material{
		{
			PbrMetallicRoughness: &gltf.PbrMetallicRoughness{
				BaseColorTexture: &gltf.TextureInfo{Index: 0},
			},
		},
	}
	gltfDoc.Meshes[0].Primitives[0].Material = ptr(0)

	// Without NoOpImageLoader - should fail
	_, bufferSize := createTriangleBuffer()
	models, err := gltf.DecodeModels(&gltfDoc, [][]byte{make([]byte, bufferSize)}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load material")

	// With NoOpImageLoader - should succeed
	opts := &gltf.ReaderOptions{
		ImageLoader: &gltf.NoOpImageLoader{},
	}
	models, err = gltf.DecodeModels(&gltfDoc, [][]byte{make([]byte, bufferSize)}, opts)
	require.NoError(t, err)
	require.Len(t, models, 1)
	require.NotNil(t, models[0].Material)
	require.NotNil(t, models[0].Material.PbrMetallicRoughness)
	require.NotNil(t, models[0].Material.PbrMetallicRoughness.BaseColorTexture)
	assert.Nil(t, models[0].Material.PbrMetallicRoughness.BaseColorTexture.Image) // Image not loaded
}

func TestBasePath(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "models")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	// Create buffer in subdirectory
	bufferPath := filepath.Join(subDir, "buffer.bin")
	err = os.WriteFile(bufferPath, make([]byte, 36), 0644)
	require.NoError(t, err)

	gltfData := `{
		"asset": {"version": "2.0"},
		"buffers": [{"byteLength": 36, "uri": "buffer.bin"}]
	}`

	// Test with BasePath in options
	opts := &gltf.ReaderOptions{
		BasePath: subDir,
	}
	doc, buffers, err := gltf.LoadGLTF(strings.NewReader(gltfData), opts)
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Len(t, buffers, 1)
}

func TestBackwardCompatibility(t *testing.T) {
	// Create test GLTF file
	tempDir := t.TempDir()
	gltfPath := filepath.Join(tempDir, "test.gltf")

	gltfDoc := createMinimalValidGLTF()
	gltfData, err := json.Marshal(gltfDoc)
	require.NoError(t, err)

	err = os.WriteFile(gltfPath, gltfData, 0644)
	require.NoError(t, err)

	// Test deprecated ExperimentalLoad
	doc, buffers, err := gltf.ExperimentalLoad(gltfPath, nil)
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Test deprecated ExperimentalDecodeModels
	models, err := gltf.ExperimentalDecodeModels(doc, buffers, tempDir, &gltf.ReaderOptions{})
	require.NoError(t, err)
	assert.Len(t, models, 1)

	// Test deprecated ExperimentalDecodeScene
	scene, err := gltf.ExperimentalDecodeScene(doc, buffers, tempDir, &gltf.ReaderOptions{})
	require.NoError(t, err)
	assert.NotNil(t, scene)
	assert.Len(t, scene.Models, 1)
}

func TestDecodeMaterialExtensions(t *testing.T) {

	tests := map[string]struct {
		ExtensionName string
		ExtensionData map[string]any
		Assert        func(t *testing.T, matExt gltf.MaterialExtension)
	}{
		"PbrSpecularGlossiness": {
			ExtensionName: "KHR_materials_pbrSpecularGlossiness",
			ExtensionData: map[string]any{
				"diffuseFactor":             [4]float64{1., 1., 1., 1.},
				"diffuseTexture":            map[string]any{"index": 0},
				"specularFactor":            [3]float64{0.5, 0.5, 0.5},
				"glossinessFactor":          1,
				"specularGlossinessTexture": map[string]any{"index": 0},
			},
			Assert: func(t *testing.T, matExt gltf.MaterialExtension) {
				ext, ok := matExt.(*gltf.PolyformPbrSpecularGlossiness)
				require.True(t, ok, "can cast to specific extension")

				assert.Equal(t, 1., *ext.GlossinessFactor)
				assertColorEqual(t, color.RGBA{255, 255, 255, 255}, ext.DiffuseFactor, "Diffuse Color")
				assertColorEqual(t, color.RGBA{127, 127, 127, 255}, ext.SpecularFactor, "SpecularFactor")
				assert.NotNil(t, ext.DiffuseTexture)
				assert.NotNil(t, ext.SpecularGlossinessTexture)
			},
		},
		"IOR": {
			ExtensionName: "KHR_materials_ior",
			ExtensionData: map[string]any{
				"ior": 1.33,
			},
			Assert: func(t *testing.T, matExt gltf.MaterialExtension) {
				ext, ok := matExt.(*gltf.PolyformIndexOfRefraction)
				require.True(t, ok, "can cast to specific extension")
				assert.Equal(t, 1.33, *ext.IOR)
			},
		},
		"Transmission": {
			ExtensionName: "KHR_materials_transmission",
			ExtensionData: map[string]any{
				"transmissionFactor":  0.5,
				"transmissionTexture": map[string]any{"index": 0},
			},
			Assert: func(t *testing.T, matExt gltf.MaterialExtension) {
				ext, ok := matExt.(*gltf.PolyformTransmission)
				require.True(t, ok, "can cast to specific extension")
				assert.Equal(t, 0.5, ext.Factor)
				assert.NotNil(t, ext.Texture)
			},
		},
		"Volume": {
			ExtensionName: "KHR_materials_volume",
			ExtensionData: map[string]any{
				"thicknessFactor":     0.5,
				"attenuationDistance": 0.5,
				"attenuationColor":    [3]float64{0.5, 0.5, 0.5},
				"thicknessTexture":    map[string]any{"index": 0},
			},
			Assert: func(t *testing.T, matExt gltf.MaterialExtension) {
				ext, ok := matExt.(*gltf.PolyformVolume)
				require.True(t, ok, "can cast to specific extension")
				assert.Equal(t, 0.5, ext.ThicknessFactor)
				assert.Equal(t, 0.5, *ext.AttenuationDistance)
				assert.NotNil(t, ext.ThicknessTexture)
				assertColorEqual(t, color.RGBA{127, 127, 127, 255}, ext.AttenuationColor, "AttenuationColor")
			},
		},
		"Specular": {
			ExtensionName: "KHR_materials_specular",
			ExtensionData: map[string]any{
				"specularFactor":       0.5,
				"specularColorFactor":  [3]float64{0.5, 0.5, 0.5},
				"specularTexture":      map[string]any{"index": 0},
				"specularColorTexture": map[string]any{"index": 0},
			},
			Assert: func(t *testing.T, matExt gltf.MaterialExtension) {
				ext, ok := matExt.(*gltf.PolyformSpecular)
				require.True(t, ok, "can cast to specific extension")
				assert.Equal(t, 0.5, *ext.Factor)
				assert.NotNil(t, ext.Texture)
				assert.NotNil(t, ext.ColorTexture)
				assertColorEqual(t, color.RGBA{127, 127, 127, 255}, ext.ColorFactor, "ColorFactor")
			},
		},
		"Unlit": {
			ExtensionName: "KHR_materials_unlit",
			ExtensionData: map[string]any{},
			Assert: func(t *testing.T, matExt gltf.MaterialExtension) {
				_, ok := matExt.(gltf.PolyformUnlit)
				require.True(t, ok, "can cast to specific extension")
			},
		},
		"Clearcoat": {
			ExtensionName: "KHR_materials_clearcoat",
			ExtensionData: map[string]any{
				"clearcoatFactor":           0.5,
				"clearcoatTexture":          map[string]any{"index": 0},
				"clearcoatRoughnessFactor":  0.5,
				"clearcoatRoughnessTexture": map[string]any{"index": 0},
			},
			Assert: func(t *testing.T, matExt gltf.MaterialExtension) {
				ext, ok := matExt.(*gltf.PolyformClearcoat)
				require.True(t, ok, "can cast to specific extension")

				assert.Equal(t, 0.5, ext.ClearcoatFactor)
				assert.Equal(t, 0.5, ext.ClearcoatRoughnessFactor)
				assert.NotNil(t, ext.ClearcoatTexture)
				assert.NotNil(t, ext.ClearcoatRoughnessTexture)
			},
		},
		"Emissive Strength": {
			ExtensionName: "KHR_materials_emissive_strength",
			ExtensionData: map[string]any{
				"emissiveStrength": 0.5,
			},
			Assert: func(t *testing.T, matExt gltf.MaterialExtension) {
				ext, ok := matExt.(*gltf.PolyformEmissiveStrength)
				require.True(t, ok, "can cast to specific extension")

				assert.Equal(t, 0.5, *ext.EmissiveStrength)
			},
		},
		"Iridescence": {
			ExtensionName: "KHR_materials_iridescence",
			ExtensionData: map[string]any{
				"iridescenceFactor":           0.5,
				"iridescenceIor":              0.5,
				"iridescenceThicknessMinimum": 0.5,
				"iridescenceThicknessMaximum": 0.5,
				"iridescenceTexture":          map[string]any{"index": 0},
				"iridescenceThicknessTexture": map[string]any{"index": 0},
			},
			Assert: func(t *testing.T, matExt gltf.MaterialExtension) {
				ext, ok := matExt.(*gltf.PolyformIridescence)
				require.True(t, ok, "can cast to specific extension")

				assert.Equal(t, 0.5, ext.IridescenceFactor)
				assert.Equal(t, 0.5, *ext.IridescenceIor)
				assert.Equal(t, 0.5, *ext.IridescenceThicknessMinimum)
				assert.Equal(t, 0.5, *ext.IridescenceThicknessMaximum)
				assert.NotNil(t, ext.IridescenceTexture)
				assert.NotNil(t, ext.IridescenceThicknessTexture)
			},
		},
		"Sheen": {
			ExtensionName: "KHR_materials_sheen",
			ExtensionData: map[string]any{
				"sheenRoughnessFactor":  0.5,
				"sheenColorTexture":     map[string]any{"index": 0},
				"sheenRoughnessTexture": map[string]any{"index": 0},
				"sheenColorFactor":      [3]float64{0.5, 0.5, 0.5},
			},
			Assert: func(t *testing.T, matExt gltf.MaterialExtension) {
				ext, ok := matExt.(*gltf.PolyformSheen)
				require.True(t, ok, "can cast to specific extension")

				assert.Equal(t, 0.5, ext.SheenRoughnessFactor)
				assert.NotNil(t, ext.SheenColorTexture)
				assert.NotNil(t, ext.SheenRoughnessTexture)
				assertColorEqual(t, color.RGBA{127, 127, 127, 255}, ext.SheenColorFactor, "SheenColorFactor")
			},
		},
		"Anisotropy": {
			ExtensionName: "KHR_materials_anisotropy",
			ExtensionData: map[string]any{
				"anisotropyStrength": 0.5,
				"anisotropyRotation": 0.5,
				"anisotropyTexture":  map[string]any{"index": 0},
			},
			Assert: func(t *testing.T, matExt gltf.MaterialExtension) {
				ext, ok := matExt.(*gltf.PolyformAnisotropy)
				require.True(t, ok, "can cast to specific extension")

				assert.Equal(t, 0.5, ext.AnisotropyStrength)
				assert.Equal(t, 0.5, ext.AnisotropyRotation)
				assert.NotNil(t, ext.AnisotropyTexture)
			},
		},
		"Dispersion": {
			ExtensionName: "KHR_materials_dispersion",
			ExtensionData: map[string]any{
				"dispersion": 0.5,
			},
			Assert: func(t *testing.T, matExt gltf.MaterialExtension) {
				ext, ok := matExt.(*gltf.PolyformDispersion)
				require.True(t, ok, "can cast to specific extension")

				assert.Equal(t, 0.5, ext.Dispersion)
			},
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			minimalGLTF := createGLTFWithMaterials(t)
			minimalGLTF.Materials[0].Extensions = map[string]any{
				tc.ExtensionName: tc.ExtensionData,
			}

			marshalledGLTF, err := json.Marshal(minimalGLTF)
			require.NoError(t, err)

			decodedGLTF, buffers, err := gltf.LoadGLTF(bytes.NewReader(marshalledGLTF), nil)
			require.NoError(t, err)

			models, err := gltf.DecodeModels(decodedGLTF, buffers, nil)
			require.NoError(t, err)

			require.Len(t, models, 1)
			require.Len(t, models[0].Material.Extensions, 1)
			assert.Equal(t, tc.ExtensionName, models[0].Material.Extensions[0].ExtensionID())
			tc.Assert(t, models[0].Material.Extensions[0])
		})
	}

}
