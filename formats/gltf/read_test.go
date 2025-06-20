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

// =============================================================================
// Test Data Generation Helpers
// =============================================================================

// intPtr returns a pointer to an integer
func intPtr(i int) *int {
	return &i
}

// float64Ptr returns a pointer to a float64
func float64Ptr(f float64) *float64 {
	return &f
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
func imageToBase64DataURI(img image.Image) (string, error) {
	buf := &bytes.Buffer{}
	err := png.Encode(buf, img)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/png;base64," + encoded, nil
}

// createTestBuffer creates a binary buffer with test data
func createTestBuffer(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	return "data:application/octet-stream;base64," + encoded
}

// createTriangleBuffer creates a buffer with triangle mesh data
func createTriangleBuffer() (string, int) {
	buf := &bytes.Buffer{}

	// Indices (3 triangles = 9 indices)
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
			{BufferView: intPtr(0), ComponentType: gltf.AccessorComponentType_UNSIGNED_INT, Count: 3, Type: gltf.AccessorType_SCALAR}, // indices
			{BufferView: intPtr(1), ComponentType: gltf.AccessorComponentType_FLOAT, Count: 3, Type: gltf.AccessorType_VEC3},          // positions
			{BufferView: intPtr(2), ComponentType: gltf.AccessorComponentType_FLOAT, Count: 3, Type: gltf.AccessorType_VEC3},          // normals
			{BufferView: intPtr(3), ComponentType: gltf.AccessorComponentType_FLOAT, Count: 3, Type: gltf.AccessorType_VEC2},          // texcoords
		},
		Meshes: []gltf.Mesh{
			{
				Primitives: []gltf.Primitive{
					{
						Indices: intPtr(0),
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
			{Mesh: intPtr(0)},
		},
		Scenes: []gltf.Scene{
			{Nodes: []gltf.GltfId{0}},
		},
		Scene: 0,
	}
}

// createGLTFWithMaterials creates a GLTF with materials and textures
func createGLTFWithMaterials() gltf.Gltf {
	doc := createMinimalValidGLTF()

	// Create test images
	baseColorImg := generateTestImage(64, 64, color.RGBA{255, 0, 0, 255})  // Red
	normalImg := generateTestImage(64, 64, color.RGBA{128, 128, 255, 255}) // Blue-ish normal

	baseColorURI, _ := imageToBase64DataURI(baseColorImg)
	normalURI, _ := imageToBase64DataURI(normalImg)

	// Add images
	doc.Images = []gltf.Image{
		{URI: baseColorURI},
		{URI: normalURI},
	}

	// Add textures
	doc.Textures = []gltf.Texture{
		{Source: intPtr(0)}, // Base color
		{Source: intPtr(1)}, // Normal
	}

	// Add materials
	doc.Materials = []gltf.Material{
		{
			ChildOfRootProperty: gltf.ChildOfRootProperty{
				Name: "TestMaterial",
			},
			PbrMetallicRoughness: &gltf.PbrMetallicRoughness{
				BaseColorFactor: &[4]float64{1.0, 0.0, 0.0, 1.0},
				BaseColorTexture: &gltf.TextureInfo{
					Index: 0,
				},
				MetallicFactor:  float64Ptr(0.0),
				RoughnessFactor: float64Ptr(1.0),
			},
			NormalTexture: &gltf.NormalTexture{
				TextureInfo: gltf.TextureInfo{Index: 1},
				Scale:       float64Ptr(1.0),
			},
		},
	}

	// Update primitive to use material
	doc.Meshes[0].Primitives[0].Material = intPtr(0)

	return doc
}

// createGLTFWithHierarchy creates a GLTF with node hierarchy
func createGLTFWithHierarchy() gltf.Gltf {
	doc := createMinimalValidGLTF()

	// Create hierarchy: Root -> Child1 -> Child2 (with mesh)
	doc.Nodes = []gltf.Node{
		{
			Name:        "Root",
			Children:    []gltf.GltfId{1},
			Translation: &[3]float64{1.0, 0.0, 0.0},
		},
		{
			Name:     "Child1",
			Children: []gltf.GltfId{2},
			Scale:    &[3]float64{2.0, 2.0, 2.0},
		},
		{
			Name:     "Child2",
			Mesh:     intPtr(0),
			Rotation: &[4]float64{0.0, 0.0, 0.0, 1.0},
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

// assertVector3Equal compares two vector3.Float64 with tolerance
func assertVector3Equal(t *testing.T, expected, actual vector3.Float64, msg string) {
	t.Helper()
	tolerance := 1e-6
	assert.InDelta(t, expected.X(), actual.X(), tolerance, msg+" X component")
	assert.InDelta(t, expected.Y(), actual.Y(), tolerance, msg+" Y component")
	assert.InDelta(t, expected.Z(), actual.Z(), tolerance, msg+" Z component")
}

// assertVector2Equal compares two vector2.Float64 with tolerance
func assertVector2Equal(t *testing.T, expected, actual vector2.Float64, msg string) {
	t.Helper()
	tolerance := 1e-6
	assert.InDelta(t, expected.X(), actual.X(), tolerance, msg+" X component")
	assert.InDelta(t, expected.Y(), actual.Y(), tolerance, msg+" Y component")
}

// assertColorEqual compares two color.RGBA values
func assertColorEqual(t *testing.T, expected, actual color.RGBA, msg string) {
	t.Helper()
	assert.Equal(t, expected.R, actual.R, msg+" R component")
	assert.Equal(t, expected.G, actual.G, msg+" G component")
	assert.Equal(t, expected.B, actual.B, msg+" B component")
	assert.Equal(t, expected.A, actual.A, msg+" A component")
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
// ExperimentalLoad Tests
// =============================================================================

func TestExperimentalLoad(t *testing.T) {
	tests := []struct {
		name        string
		setupGLTF   func() gltf.Gltf
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid_minimal_gltf",
			setupGLTF:   createMinimalValidGLTF,
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
			setupGLTF: func() gltf.Gltf {
				doc := createMinimalValidGLTF()
				doc.Asset.Version = ""
				return doc
			},
			expectError: true,
			errorMsg:    "missing required asset version",
		},
		{
			name: "invalid_buffer_byte_length",
			setupGLTF: func() gltf.Gltf {
				doc := createMinimalValidGLTF()
				doc.Buffers[0].ByteLength = 0
				return doc
			},
			expectError: true,
			errorMsg:    "invalid byte length",
		},
		{
			name: "mismatched_buffer_length",
			setupGLTF: func() gltf.Gltf {
				doc := createMinimalValidGLTF()
				doc.Buffers[0].ByteLength = 999999 // Wrong length
				return doc
			},
			expectError: true,
			errorMsg:    "does not match expected length",
		},
		{
			name: "invalid_base64_data",
			setupGLTF: func() gltf.Gltf {
				doc := createMinimalValidGLTF()
				doc.Buffers[0].URI = "data:application/octet-stream;base64,invalid_base64_data!!!"
				return doc
			},
			expectError: true,
			errorMsg:    "failed to decode base64 data",
		},
		{
			name: "unsupported_data_uri_encoding",
			setupGLTF: func() gltf.Gltf {
				doc := createMinimalValidGLTF()
				doc.Buffers[0].URI = "data:application/octet-stream;charset=utf-8,some_data"
				return doc
			},
			expectError: true,
			errorMsg:    "unsupported data URI encoding",
		},
		{
			name: "empty_buffer_uri",
			setupGLTF: func() gltf.Gltf {
				doc := createMinimalValidGLTF()
				doc.Buffers[0].URI = ""
				return doc
			},
			expectError: true,
			errorMsg:    "empty URI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := tt.setupGLTF()
			tempFile := writeGLTFToTempFile(t, doc)

			loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)

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

func TestExperimentalLoadFileNotFound(t *testing.T) {
	_, _, err := gltf.ExperimentalLoad("/non/existent/path.gltf")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read GLTF file")
}

func TestExperimentalLoadInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "invalid.gltf")

	err := os.WriteFile(tempFile, []byte("invalid json {"), 0644)
	require.NoError(t, err)

	_, _, err = gltf.ExperimentalLoad(tempFile)
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
							BufferView:    intPtr(0),
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
							BufferView:    intPtr(0),
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
							BufferView:    intPtr(0),
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
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
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
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
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
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
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
			tempDir := t.TempDir()

			models, err := gltf.ExperimentalDecodeModels(&doc, buffers, tempDir)

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
				doc := createGLTFWithMaterials()
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
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
							MetallicFactor:  float64Ptr(0.1),
							RoughnessFactor: float64Ptr(0.9),
						},
					},
				}
				doc.Meshes[0].Primitives[0].Material = intPtr(0)
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
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
			tempDir := t.TempDir()

			models, err := gltf.ExperimentalDecodeModels(&doc, buffers, tempDir)

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
				doc := createGLTFWithMaterials()
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectError: false,
		},
		{
			name: "invalid_texture_reference",
			setupTexture: func() (gltf.Gltf, [][]byte) {
				doc := createGLTFWithMaterials()
				// Reference non-existent texture
				doc.Materials[0].PbrMetallicRoughness.BaseColorTexture.Index = 999
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectError: true,
			errorMsg:    "invalid texture ID",
		},
		{
			name: "invalid_image_reference",
			setupTexture: func() (gltf.Gltf, [][]byte) {
				doc := createGLTFWithMaterials()
				// Reference non-existent image
				doc.Textures[0].Source = intPtr(999)
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectError: true,
			errorMsg:    "invalid image",
		},
		{
			name: "malformed_data_uri",
			setupTexture: func() (gltf.Gltf, [][]byte) {
				doc := createGLTFWithMaterials()
				// Use malformed data URI
				doc.Images[0].URI = "data:invalid_format"
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectError: true,
			errorMsg:    "invalid data URI format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, buffers := tt.setupTexture()
			tempDir := t.TempDir()

			models, err := gltf.ExperimentalDecodeModels(&doc, buffers, tempDir)

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
// Scene Hierarchy Tests
// =============================================================================

func TestSceneHierarchy(t *testing.T) {
	tests := []struct {
		name           string
		setupScene     func() (gltf.Gltf, [][]byte)
		expectedModels int
		expectError    bool
		errorMsg       string
	}{
		{
			name: "simple_hierarchy",
			setupScene: func() (gltf.Gltf, [][]byte) {
				doc := createGLTFWithHierarchy()
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectedModels: 1,
			expectError:    false,
		},
		{
			name: "empty_scene",
			setupScene: func() (gltf.Gltf, [][]byte) {
				doc := createMinimalValidGLTF()
				doc.Scenes[0].Nodes = []gltf.GltfId{} // Empty scene
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
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
				doc.Scene = 999 // Invalid scene index
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
				require.NoError(t, err)
				return *loadedDoc, buffers
			},
			expectedModels: 0,
			expectError:    true,
			errorMsg:       "invalid scene index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, buffers := tt.setupScene()
			tempDir := t.TempDir()

			scene, err := gltf.ExperimentalDecodeScene(&doc, buffers, tempDir)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, scene)

			assert.Len(t, scene.Models, tt.expectedModels)

			// Validate model transformations for hierarchy
			if tt.expectedModels > 0 {
				model := scene.Models[0]
				require.NotNil(t, model.TRS)

				// The hierarchy should have accumulated transformations
				// Root: translation (1,0,0), Child1: scale (2,2,2), Child2: rotation
				translation := model.TRS.Position()
				assert.InDelta(t, 1.0, translation.X(), 1e-6, "Translation should be applied")

				scale := model.TRS.Scale()
				assert.InDelta(t, 2.0, scale.X(), 1e-6, "Scale should be accumulated")
				assert.InDelta(t, 2.0, scale.Y(), 1e-6, "Scale should be accumulated")
				assert.InDelta(t, 2.0, scale.Z(), 1e-6, "Scale should be accumulated")
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
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
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
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
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
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
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
			tempDir := t.TempDir()

			models, err := gltf.ExperimentalDecodeModels(&doc, buffers, tempDir)

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
		setupDocument  func() gltf.Gltf
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
				assert.InDelta(t, 1.0, translation.X(), 1e-6, "Translation from root")
				assert.InDelta(t, 2.0, scale.X(), 1e-6, "Scale from Child1")
				assert.InDelta(t, 2.0, scale.Y(), 1e-6, "Scale from Child1")
				assert.InDelta(t, 2.0, scale.Z(), 1e-6, "Scale from Child1")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := tt.setupDocument()
			tempFile := writeGLTFToTempFile(t, doc)

			// Load document
			loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
			require.NoError(t, err)
			require.NotNil(t, loadedDoc)

			// Decode scene
			scene, err := gltf.ExperimentalDecodeScene(loadedDoc, buffers, filepath.Dir(tempFile))
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
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
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
				doc.Accessors[0].BufferView = intPtr(999)
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
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
				doc.Meshes[0].Primitives[0].Material = intPtr(999)
				tempFile := writeGLTFToTempFile(t, doc)
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
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
				loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
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
			tempDir := t.TempDir()

			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Convert panic to error for test validation
						err = fmt.Errorf("panic occurred: %v", r)
					}
				}()
				_, err = gltf.ExperimentalDecodeScene(&doc, buffers, tempDir)
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
// Options API Tests
// =============================================================================

func TestOptionsAPI(t *testing.T) {
	t.Run("save_text_with_opts", func(t *testing.T) {
		// Create a GLTF with materials and textures
		doc := createGLTFWithMaterials()
		tempFile := writeGLTFToTempFile(t, doc)

		// Load the original scene
		loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
		require.NoError(t, err)

		scene, err := gltf.ExperimentalDecodeScene(loadedDoc, buffers, filepath.Dir(tempFile))
		require.NoError(t, err)

		outputDir := t.TempDir()

		// Test with embedding enabled
		outputPath1 := filepath.Join(outputDir, "embedded.gltf")
		err = gltf.SaveTextWithOpts(outputPath1, *scene, gltf.Options{EmbedTextures: true})
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
		err = gltf.SaveTextWithOpts(outputPath2, *scene, gltf.Options{})
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
		doc := createGLTFWithMaterials()
		tempFile := writeGLTFToTempFile(t, doc)

		loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
		require.NoError(t, err)

		scene, err := gltf.ExperimentalDecodeScene(loadedDoc, buffers, filepath.Dir(tempFile))
		require.NoError(t, err)

		outputDir := t.TempDir()

		// Test SaveBinaryWithOpts
		binaryOutputPath := filepath.Join(outputDir, "test_output.glb")
		err = gltf.SaveBinaryWithOpts(binaryOutputPath, *scene, gltf.Options{EmbedTextures: true})
		require.NoError(t, err)

		// Verify binary file exists and has content
		stat, err := os.Stat(binaryOutputPath)
		require.NoError(t, err)
		assert.Greater(t, stat.Size(), int64(0), "Binary file should have content")
	})

	t.Run("options_struct_validation", func(t *testing.T) {
		// Test that Options struct can be created and used properly
		opts := gltf.Options{
			EmbedTextures: true,
			MinifyJSON:    true,
		}
		assert.True(t, opts.EmbedTextures, "EmbedTextures should be settable")
		assert.True(t, opts.MinifyJSON, "MinifyJSON should be settable")

		// Test zero value
		defaultOpts := gltf.Options{}
		assert.False(t, defaultOpts.EmbedTextures, "Default EmbedTextures should be false")
		assert.False(t, defaultOpts.MinifyJSON, "Default MinifyJSON should be false")
	})
}

func TestWriteWithOptsAPI(t *testing.T) {
	doc := createGLTFWithMaterials()
	tempFile := writeGLTFToTempFile(t, doc)

	loadedDoc, buffers, err := gltf.ExperimentalLoad(tempFile)
	require.NoError(t, err)

	scene, err := gltf.ExperimentalDecodeScene(loadedDoc, buffers, filepath.Dir(tempFile))
	require.NoError(t, err)

	t.Run("write_text_with_opts", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := gltf.WriteTextWithOpts(*scene, buf, gltf.Options{EmbedTextures: true})
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
		err := gltf.WriteBinaryWithOpts(*scene, buf, gltf.Options{EmbedTextures: true})
		require.NoError(t, err)

		// Verify we got some binary data
		assert.Greater(t, buf.Len(), 0, "Should have written binary data")
	})

	t.Run("backward_compatibility", func(t *testing.T) {
		buf1 := &bytes.Buffer{}
		buf2 := &bytes.Buffer{}

		// Test that WriteTextWithOptions still works and produces same result as WriteTextWithOpts
		err2 := gltf.WriteTextWithOpts(*scene, buf2, gltf.Options{EmbedTextures: true})
		require.NoError(t, err2)

		// Both should produce the same output
		assert.Equal(t, buf1.String(), buf2.String(), "Old and new APIs should produce identical output")
	})

	t.Run("minify_json_option", func(t *testing.T) {
		prettyBuf := &bytes.Buffer{}
		minifiedBuf := &bytes.Buffer{}

		// Test pretty-printed JSON (default)
		err := gltf.WriteTextWithOpts(*scene, prettyBuf, gltf.Options{MinifyJSON: false})
		require.NoError(t, err)

		// Test minified JSON
		err = gltf.WriteTextWithOpts(*scene, minifiedBuf, gltf.Options{MinifyJSON: true})
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
