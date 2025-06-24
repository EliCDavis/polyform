package gltf

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// New API Tests
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
			doc, err := Parse(reader)

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
	doc, err := ParseFile(gltfPath)
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, "2.0", doc.Asset.Version)

	// Test loading non-existent file
	_, err = ParseFile(filepath.Join(tempDir, "nonexistent.gltf"))
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
			opts := &ReaderOptions{BasePath: tempDir}
			doc, buffers, err := Load(reader, opts)

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

func TestLoadFile(t *testing.T) {
	tempDir := t.TempDir()
	gltfPath := filepath.Join(tempDir, "test.gltf")
	bufferPath := filepath.Join(tempDir, "buffer.bin")
	
	// Create GLTF file
	gltfContent := `{
		"asset": {"version": "2.0"},
		"buffers": [{"byteLength": 36, "uri": "buffer.bin"}]
	}`
	err := os.WriteFile(gltfPath, []byte(gltfContent), 0644)
	require.NoError(t, err)
	
	// Create buffer file
	err = os.WriteFile(bufferPath, make([]byte, 36), 0644)
	require.NoError(t, err)

	// Test loading
	doc, buffers, err := LoadFile(gltfPath, nil)
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Len(t, buffers, 1)
	assert.Len(t, buffers[0], 36)
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
	images map[string]image.Image
}

func (m *mockImageLoader) LoadImage(uri string) (image.Image, error) {
	if img, ok := m.images[uri]; ok {
		return img, nil
	}
	return nil, fmt.Errorf("image not found: %s", uri)
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

		opts := &ReaderOptions{
			BufferLoader: loader,
		}

		doc, buffers, err := Load(strings.NewReader(gltfData), opts)
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
			images: map[string]image.Image{
				"custom://image1": testImage,
			},
		}

		// Create a minimal valid GLTF doc
		gltfDoc := createMinimalValidGLTF()
		gltfDoc.Images = []Image{{URI: "custom://image1"}}
		gltfDoc.Textures = []Texture{{Source: ptr(0)}}
		gltfDoc.Materials = []Material{
			{
				PbrMetallicRoughness: &PbrMetallicRoughness{
					BaseColorTexture: &TextureInfo{Index: 0},
				},
			},
		}
		gltfDoc.Meshes[0].Primitives[0].Material = ptr(0)

		opts := &ReaderOptions{
			ImageLoader: loader,
		}

		models, err := DecodeModels(gltfDoc, [][]byte{make([]byte, 102)}, opts)
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
	gltfDoc.Images = []Image{{URI: "nonexistent.png"}}
	gltfDoc.Textures = []Texture{{Source: ptr(0)}}
	gltfDoc.Materials = []Material{
		{
			PbrMetallicRoughness: &PbrMetallicRoughness{
				BaseColorTexture: &TextureInfo{Index: 0},
			},
		},
	}
	gltfDoc.Meshes[0].Primitives[0].Material = ptr(0)

	// Without NoOpImageLoader - should fail
	models, err := DecodeModels(gltfDoc, [][]byte{make([]byte, 102)}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load material")

	// With NoOpImageLoader - should succeed
	opts := &ReaderOptions{
		ImageLoader: &NoOpImageLoader{},
	}
	models, err = DecodeModels(gltfDoc, [][]byte{make([]byte, 102)}, opts)
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
	opts := &ReaderOptions{
		BasePath: subDir,
	}
	doc, buffers, err := Load(strings.NewReader(gltfData), opts)
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
	doc, buffers, err := ExperimentalLoad(gltfPath, nil)
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Test deprecated ExperimentalDecodeModels
	models, err := ExperimentalDecodeModels(doc, buffers, tempDir, &ReaderOptions{})
	require.NoError(t, err)
	assert.Len(t, models, 1)

	// Test deprecated ExperimentalDecodeScene
	scene, err := ExperimentalDecodeScene(doc, buffers, tempDir, &ReaderOptions{})
	require.NoError(t, err)
	assert.NotNil(t, scene)
	assert.Len(t, scene.Models, 1)
}

// Helper functions

func ptr[T any](v T) *T {
	return &v
}

func createMinimalValidGLTF() *Gltf {
	// Create a minimal triangle
	// Positions: 3 vertices * 3 floats * 4 bytes = 36 bytes
	positions := []byte{
		0, 0, 0, 0,       // vertex 0 x: 0.0
		0, 0, 0, 0,       // vertex 0 y: 0.0
		0, 0, 0, 0,       // vertex 0 z: 0.0
		0, 0, 128, 63,    // vertex 1 x: 1.0
		0, 0, 0, 0,       // vertex 1 y: 0.0
		0, 0, 0, 0,       // vertex 1 z: 0.0
		0, 0, 0, 0,       // vertex 2 x: 0.0
		0, 0, 128, 63,    // vertex 2 y: 1.0
		0, 0, 0, 0,       // vertex 2 z: 0.0
	}
	
	// Normals: 3 vertices * 3 floats * 4 bytes = 36 bytes
	normals := []byte{
		0, 0, 0, 0,       // normal 0: (0, 0, 1)
		0, 0, 0, 0,       //
		0, 0, 128, 63,    //
		0, 0, 0, 0,       // normal 1: (0, 0, 1)
		0, 0, 0, 0,       //
		0, 0, 128, 63,    //
		0, 0, 0, 0,       // normal 2: (0, 0, 1)
		0, 0, 0, 0,       //
		0, 0, 128, 63,    //
	}
	
	// UVs: 3 vertices * 2 floats * 4 bytes = 24 bytes
	uvs := []byte{
		0, 0, 0, 0,       // uv 0: (0, 0)
		0, 0, 0, 0,       //
		0, 0, 128, 63,    // uv 1: (1, 0)
		0, 0, 0, 0,       //
		0, 0, 128, 63,    // uv 2: (1, 1)
		0, 0, 128, 63,    //
	}
	
	// Indices: 3 indices * 2 bytes = 6 bytes
	indices := []byte{0, 0, 1, 0, 2, 0} // unsigned short indices
	
	// Combine all data
	allData := append(positions, normals...)
	allData = append(allData, uvs...)
	allData = append(allData, indices...)
	
	// Convert to base64
	bufferData := bytes.NewBuffer(allData).Bytes()
	
	// Debug: print actual sizes
	// fmt.Printf("Positions: %d bytes\n", len(positions))
	// fmt.Printf("Normals: %d bytes\n", len(normals))  
	// fmt.Printf("UVs: %d bytes\n", len(uvs))
	// fmt.Printf("Indices: %d bytes\n", len(indices))
	// fmt.Printf("Total buffer: %d bytes\n", len(bufferData))
	
	base64Data := "data:application/octet-stream;base64," + bytesToBase64(bufferData)

	return &Gltf{
		Asset: Asset{Version: "2.0"},
		Buffers: []Buffer{
			{URI: base64Data, ByteLength: len(bufferData)},
		},
		BufferViews: []BufferView{
			{Buffer: 0, ByteOffset: 0, ByteLength: 36, Target: ARRAY_BUFFER},          // positions
			{Buffer: 0, ByteOffset: 36, ByteLength: 36, Target: ARRAY_BUFFER},         // normals
			{Buffer: 0, ByteOffset: 72, ByteLength: 24, Target: ARRAY_BUFFER},         // uvs
			{Buffer: 0, ByteOffset: 96, ByteLength: 6, Target: ELEMENT_ARRAY_BUFFER},  // indices
		},
		Accessors: []Accessor{
			{BufferView: ptr(GltfId(0)), ByteOffset: 0, ComponentType: AccessorComponentType_FLOAT, Count: 3, Type: AccessorType_VEC3}, // positions
			{BufferView: ptr(GltfId(1)), ByteOffset: 0, ComponentType: AccessorComponentType_FLOAT, Count: 3, Type: AccessorType_VEC3}, // normals
			{BufferView: ptr(GltfId(2)), ByteOffset: 0, ComponentType: AccessorComponentType_FLOAT, Count: 3, Type: AccessorType_VEC2}, // uvs
			{BufferView: ptr(GltfId(3)), ByteOffset: 0, ComponentType: AccessorComponentType_UNSIGNED_SHORT, Count: 3, Type: AccessorType_SCALAR}, // indices
		},
		Meshes: []Mesh{
			{
				Primitives: []Primitive{
					{
						Attributes: map[string]GltfId{
							POSITION:   0,
							NORMAL:     1,
							TEXCOORD_0: 2,
						},
						Indices: ptr(GltfId(3)),
					},
				},
			},
		},
		Nodes: []Node{
			{Mesh: ptr(GltfId(0))},
		},
		Scenes: []Scene{
			{Nodes: []GltfId{0}},
		},
		Scene: 0,
	}
}

func bytesToBase64(data []byte) string {
	buf := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	encoder.Write(data)
	encoder.Close()
	return buf.String()
}