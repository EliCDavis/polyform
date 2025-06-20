package gltf_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLightFixtureModelLoading(t *testing.T) {
	// Path to the light fixture model
	lightFixturePath := filepath.Join("..", "..", "test-models", "light_fixture", "model.gltf")

	// Load the GLTF file
	doc, buffers, err := gltf.ExperimentalLoad(lightFixturePath)
	require.NoError(t, err, "Failed to load light fixture GLTF")
	require.NotNil(t, doc, "GLTF document should not be nil")
	require.Len(t, buffers, 1, "Should have one buffer")

	// Validate basic structure
	assert.Equal(t, "2.0", doc.Asset.Version, "Should be GLTF version 2.0")
	assert.Len(t, doc.Nodes, 5, "Should have 5 nodes")
	assert.Len(t, doc.Meshes, 1, "Should have 1 mesh")
	assert.Len(t, doc.Materials, 1, "Should have 1 material")
	assert.Len(t, doc.Textures, 4, "Should have 4 textures")
	assert.Len(t, doc.Images, 4, "Should have 4 images")

	// Validate buffer size
	expectedBufferSize := 4456
	assert.Equal(t, expectedBufferSize, len(buffers[0]), "Buffer size should match expected")

	// Validate images (texture files)
	expectedImages := []string{
		"Scene_-_Root_baseColor.jpeg",
		"Scene_-_Root_metallicRoughness.png",
		"Scene_-_Root_normal.png",
		"Scene_-_Root_emissive.jpeg",
	}

	require.Len(t, doc.Images, len(expectedImages), "Should have correct number of images")
	for i, expectedImage := range expectedImages {
		assert.Equal(t, expectedImage, doc.Images[i].URI, "Image %d should have correct URI", i)
	}

	// Test model decoding with materials and textures
	gltfDir := filepath.Dir(lightFixturePath)
	models, err := gltf.ExperimentalDecodeModels(doc, buffers, gltfDir)
	require.NoError(t, err, "Failed to decode models")
	require.Len(t, models, 1, "Should have decoded 1 model")

	model := models[0]

	// Validate mesh properties
	assert.NotNil(t, model.Mesh, "Model should have a mesh")
	assert.Equal(t, 234, model.Mesh.Indices().Len(), "Should have correct number of indices")

	// Check that all expected attributes are present
	assert.True(t, model.Mesh.HasFloat3Attribute("Position"), "Should have position attribute")
	assert.True(t, model.Mesh.HasFloat3Attribute("Normal"), "Should have normal attribute")
	assert.True(t, model.Mesh.HasFloat2Attribute("TexCoord"), "Should have UV attribute")

	// Validate material
	require.NotNil(t, model.Material, "Model should have a material")
	assert.Equal(t, "Scene_-_Root", model.Material.Name, "Material should have correct name")

	// Validate PBR material properties
	require.NotNil(t, model.Material.PbrMetallicRoughness, "Should have PBR properties")

	// Base color texture should be loaded
	assert.NotNil(t, model.Material.PbrMetallicRoughness.BaseColorTexture, "Should have base color texture")
	assert.NotNil(t, model.Material.PbrMetallicRoughness.BaseColorTexture.Image, "Base color texture should have image data")
	assert.Equal(t, "Scene_-_Root_baseColor.jpeg", model.Material.PbrMetallicRoughness.BaseColorTexture.URI, "Should have correct base color texture URI")

	// Metallic-roughness texture should be loaded
	assert.NotNil(t, model.Material.PbrMetallicRoughness.MetallicRoughnessTexture, "Should have metallic-roughness texture")
	assert.NotNil(t, model.Material.PbrMetallicRoughness.MetallicRoughnessTexture.Image, "Metallic-roughness texture should have image data")
	assert.Equal(t, "Scene_-_Root_metallicRoughness.png", model.Material.PbrMetallicRoughness.MetallicRoughnessTexture.URI, "Should have correct metallic-roughness texture URI")

	// Normal texture should be loaded
	assert.NotNil(t, model.Material.NormalTexture, "Should have normal texture")
	assert.NotNil(t, model.Material.NormalTexture.Image, "Normal texture should have image data")
	assert.Equal(t, "Scene_-_Root_normal.png", model.Material.NormalTexture.URI, "Should have correct normal texture URI")

	// Emissive texture should be loaded
	assert.NotNil(t, model.Material.EmissiveTexture, "Should have emissive texture")
	assert.NotNil(t, model.Material.EmissiveTexture.Image, "Emissive texture should have image data")
	assert.Equal(t, "Scene_-_Root_emissive.jpeg", model.Material.EmissiveTexture.URI, "Should have correct emissive texture URI")

	// Validate transformation matrix handling
	assert.NotNil(t, model.TRS, "Model should have transformation")
	// The model uses matrix transformations, so TRS should be populated from matrix decomposition
}

func TestLightFixtureImageLoading(t *testing.T) {
	// Test individual image loading
	lightFixtureDir := filepath.Join("..", "..", "test-models", "light_fixture")

	testCases := []struct {
		filename string
		format   string
	}{
		{"Scene_-_Root_baseColor.jpeg", "JPEG"},
		{"Scene_-_Root_emissive.jpeg", "JPEG"},
		{"Scene_-_Root_metallicRoughness.png", "PNG"},
		{"Scene_-_Root_normal.png", "PNG"},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			// This would use the loadImage function if it were exported
			// For now, we'll test via the full model loading which exercises the same code path
			lightFixturePath := filepath.Join(lightFixtureDir, "model.gltf")
			doc, buffers, err := gltf.ExperimentalLoad(lightFixturePath)
			require.NoError(t, err)

			models, err := gltf.ExperimentalDecodeModels(doc, buffers, lightFixtureDir)
			require.NoError(t, err)
			require.Len(t, models, 1)
			require.NotNil(t, models[0].Material)

			// Verify that at least one texture was loaded successfully
			material := models[0].Material
			textureLoaded := false

			if material.PbrMetallicRoughness != nil {
				if material.PbrMetallicRoughness.BaseColorTexture != nil && material.PbrMetallicRoughness.BaseColorTexture.Image != nil {
					textureLoaded = true
				}
				if material.PbrMetallicRoughness.MetallicRoughnessTexture != nil && material.PbrMetallicRoughness.MetallicRoughnessTexture.Image != nil {
					textureLoaded = true
				}
			}
			if material.NormalTexture != nil && material.NormalTexture.Image != nil {
				textureLoaded = true
			}

			assert.True(t, textureLoaded, "At least one texture should be loaded successfully")
		})
	}
}

func TestLightFixtureAccessorDecoding(t *testing.T) {
	lightFixturePath := filepath.Join("..", "..", "test-models", "light_fixture", "model.gltf")

	doc, _, err := gltf.ExperimentalLoad(lightFixturePath)
	require.NoError(t, err)

	// Test that all accessor types in the model are handled correctly
	require.Len(t, doc.Accessors, 4, "Should have 4 accessors")

	// Accessor 0: indices (SCALAR, UNSIGNED_INT)
	assert.Equal(t, gltf.AccessorType_SCALAR, doc.Accessors[0].Type)
	assert.Equal(t, gltf.AccessorComponentType_UNSIGNED_INT, doc.Accessors[0].ComponentType)
	assert.Equal(t, 234, doc.Accessors[0].Count)

	// Accessor 1: positions (VEC3, FLOAT)
	assert.Equal(t, gltf.AccessorType_VEC3, doc.Accessors[1].Type)
	assert.Equal(t, gltf.AccessorComponentType_FLOAT, doc.Accessors[1].ComponentType)
	assert.Equal(t, 110, doc.Accessors[1].Count)

	// Accessor 2: normals (VEC3, FLOAT)
	assert.Equal(t, gltf.AccessorType_VEC3, doc.Accessors[2].Type)
	assert.Equal(t, gltf.AccessorComponentType_FLOAT, doc.Accessors[2].ComponentType)
	assert.Equal(t, 110, doc.Accessors[2].Count)

	// Accessor 3: texture coordinates (VEC2, FLOAT)
	assert.Equal(t, gltf.AccessorType_VEC2, doc.Accessors[3].Type)
	assert.Equal(t, gltf.AccessorComponentType_FLOAT, doc.Accessors[3].ComponentType)
	assert.Equal(t, 110, doc.Accessors[3].Count)
}

func TestLightFixtureSceneHierarchy(t *testing.T) {
	lightFixturePath := filepath.Join("..", "..", "test-models", "light_fixture", "model.gltf")

	doc, buffers, err := gltf.ExperimentalLoad(lightFixturePath)
	require.NoError(t, err)

	// Test scene hierarchy reconstruction
	scene, err := gltf.ExperimentalDecodeScene(doc, buffers, filepath.Dir(lightFixturePath))
	require.NoError(t, err)
	require.NotNil(t, scene)

	// Should have 1 model (the cylinder mesh at the leaf node)
	require.Len(t, scene.Models, 1, "Should have exactly 1 model in the scene")

	model := scene.Models[0]

	// Verify the model has the correct transformations applied from the hierarchy
	assert.NotNil(t, model.TRS, "Model should have transformation")

	// The light fixture has a complex hierarchy:
	// Root -> Sketchfab_model -> 3594433b8e38456a8c6daedc6a52cb57.fbx -> RootNode -> Cylinder -> mesh
	// Each level has matrix transformations that should be accumulated

	// Verify material and mesh are still loaded correctly
	assert.NotNil(t, model.Material, "Model should have material")
	assert.NotNil(t, model.Mesh, "Model should have mesh")
	assert.Equal(t, "Scene_-_Root", model.Material.Name, "Should have correct material name")
}

func TestLightFixtureRoundTrip(t *testing.T) {
	// Load the original light fixture model
	lightFixturePath := filepath.Join("..", "..", "test-models", "light_fixture", "model.gltf")

	doc, buffers, err := gltf.ExperimentalLoad(lightFixturePath)
	require.NoError(t, err)

	// Decode the scene with full hierarchy
	scene, err := gltf.ExperimentalDecodeScene(doc, buffers, filepath.Dir(lightFixturePath))
	require.NoError(t, err)
	require.NotNil(t, scene)
	require.Len(t, scene.Models, 1, "Should have exactly 1 model")

	// Create output directory in the same location as test inputs
	testModelsDir := filepath.Join("..", "..", "test-models", "light_fixture")
	outputDir := filepath.Join(testModelsDir, "exported")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err, "Should be able to create output directory")
	outputPath := filepath.Join(outputDir, "light_fixture_exported.gltf")

	// Copy texture files to output directory so external references work
	textureFiles := []string{
		"Scene_-_Root_baseColor.jpeg",
		"Scene_-_Root_metallicRoughness.png",
		"Scene_-_Root_normal.png",
		"Scene_-_Root_emissive.jpeg",
	}

	sourceDir := filepath.Dir(lightFixturePath)
	for _, textureFile := range textureFiles {
		srcPath := filepath.Join(sourceDir, textureFile)
		dstPath := filepath.Join(outputDir, textureFile)

		srcData, err := os.ReadFile(srcPath)
		require.NoError(t, err, "Should be able to read source texture file %s", textureFile)

		err = os.WriteFile(dstPath, srcData, 0644)
		require.NoError(t, err, "Should be able to write texture file %s to output dir", textureFile)
	}

	// Export the scene as a text-based GLTF file
	err = gltf.SaveText(outputPath, *scene)
	require.NoError(t, err, "Failed to export light fixture as text GLTF")

	// Verify the exported file exists and can be read
	_, err = os.Stat(outputPath)
	require.NoError(t, err, "Exported GLTF file should exist")

	// Read the exported file to verify it's valid JSON
	exportedContent, err := os.ReadFile(outputPath)
	require.NoError(t, err, "Should be able to read exported file")

	// Parse the exported GLTF to ensure it's valid
	var exportedDoc gltf.Gltf
	err = json.Unmarshal(exportedContent, &exportedDoc)
	require.NoError(t, err, "Exported file should be valid JSON GLTF")

	// Validate the exported structure
	assert.Equal(t, "2.0", exportedDoc.Asset.Version, "Should maintain GLTF 2.0 version")
	assert.Len(t, exportedDoc.Meshes, 1, "Should have 1 mesh")
	assert.Len(t, exportedDoc.Materials, 1, "Should have 1 material")

	// Verify that all 4 images are exported (URIs may be external files or data URIs)
	require.Len(t, exportedDoc.Images, 4, "Should have 4 images")
	imageNames := make([]string, len(exportedDoc.Images))
	for i, image := range exportedDoc.Images {
		imageNames[i] = image.URI
		assert.NotEmpty(t, image.URI, "Image %d should have a URI", i)
	}
	t.Logf("Exported images: %v", imageNames)

	// Verify that buffers are present (may be embedded or external)
	require.Len(t, exportedDoc.Buffers, 1, "Should have 1 buffer")
	assert.NotEmpty(t, exportedDoc.Buffers[0].URI, "Buffer should have a URI")
	bufferURI := exportedDoc.Buffers[0].URI
	if len(bufferURI) > 50 {
		bufferURI = bufferURI[:50] + "..."
	}
	t.Logf("Exported buffer URI: %s", bufferURI)

	// Test that we can load the exported file back (round trip validation)
	reloadedDoc, reloadedBuffers, err := gltf.ExperimentalLoad(outputPath)
	require.NoError(t, err, "Should be able to load the exported file")

	// Decode the reloaded scene
	reloadedScene, err := gltf.ExperimentalDecodeScene(reloadedDoc, reloadedBuffers, filepath.Dir(outputPath))
	require.NoError(t, err, "Should be able to decode the reloaded scene")
	require.Len(t, reloadedScene.Models, 1, "Reloaded scene should have 1 model")

	// Compare key properties between original and reloaded
	originalModel := scene.Models[0]
	reloadedModel := reloadedScene.Models[0]

	// Verify mesh properties are preserved
	assert.Equal(t, originalModel.Mesh.Indices().Len(), reloadedModel.Mesh.Indices().Len(), "Index count should be preserved")
	assert.True(t, reloadedModel.Mesh.HasFloat3Attribute("Position"), "Position attribute should be preserved")
	assert.True(t, reloadedModel.Mesh.HasFloat3Attribute("Normal"), "Normal attribute should be preserved")
	assert.True(t, reloadedModel.Mesh.HasFloat2Attribute("TexCoord"), "TexCoord attribute should be preserved")

	// Verify material properties are preserved
	require.NotNil(t, reloadedModel.Material, "Material should be preserved")
	assert.Equal(t, originalModel.Material.Name, reloadedModel.Material.Name, "Material name should be preserved")

	// Verify textures are preserved
	require.NotNil(t, reloadedModel.Material.PbrMetallicRoughness, "PBR properties should be preserved")
	assert.NotNil(t, reloadedModel.Material.PbrMetallicRoughness.BaseColorTexture, "Base color texture should be preserved")
	assert.NotNil(t, reloadedModel.Material.PbrMetallicRoughness.MetallicRoughnessTexture, "Metallic-roughness texture should be preserved")
	assert.NotNil(t, reloadedModel.Material.NormalTexture, "Normal texture should be preserved")
	assert.NotNil(t, reloadedModel.Material.EmissiveTexture, "Emissive texture should be preserved")

	// Verify texture images were loaded successfully from embedded data
	assert.NotNil(t, reloadedModel.Material.PbrMetallicRoughness.BaseColorTexture.Image, "Base color texture image should be loaded")
	assert.NotNil(t, reloadedModel.Material.PbrMetallicRoughness.MetallicRoughnessTexture.Image, "Metallic-roughness texture image should be loaded")
	assert.NotNil(t, reloadedModel.Material.NormalTexture.Image, "Normal texture image should be loaded")
	assert.NotNil(t, reloadedModel.Material.EmissiveTexture.Image, "Emissive texture image should be loaded")

	t.Logf("Successfully completed round-trip test: %s -> %s -> validated", lightFixturePath, outputPath)
}
