package mcp_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	transmissionExtType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.MaterialTransmissionExtensionNode]"
	volumeExtType       = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.MaterialVolumeExtensionNode]"
	dispersionExtType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.MaterialDispersionExtensionNode]"
	anisotropyExtType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.MaterialAnisotropyExtensionNode]"
	uniformFramesType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/animation.UniformFramesNode[github.com/EliCDavis/vector/vector3.Vector[float64]]]"
	translationChannel  = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.TranslationAnimationChannelNode]"
	animationNodeType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.AnimationNode]"
)

func TestDescribeManifestListsMaterialsAndFileSize(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "sphere", "type": sphereNodeType},
		map[string]any{"alias": "gold", "type": materialNodeType, "inputs": map[string]any{
			"Name":            map[string]any{"value": `"Gold"`},
			"Color":           map[string]any{"value": `"#ffc040"`},
			"MetallicFactor":  map[string]any{"value": "1"},
			"RoughnessFactor": map[string]any{"value": "0.3"},
		}},
		map[string]any{"alias": "left", "type": gltfModelType, "inputs": map[string]any{
			"Mesh":     map[string]any{"nodeId": "sphere", "port": "Out"},
			"Material": map[string]any{"nodeId": "gold", "port": "Out"},
		}},
		map[string]any{"alias": "right", "type": gltfModelType, "inputs": map[string]any{
			"Mesh":        map[string]any{"nodeId": "sphere", "port": "Out"},
			"Material":    map[string]any{"nodeId": "gold", "port": "Out"},
			"Translation": map[string]any{"value": `{"x":2,"y":0,"z":0}`},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"elements": []any{
				map[string]any{"nodeId": "left", "port": "Out"},
				map[string]any{"nodeId": "right", "port": "Out"},
			}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var out polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": created.Nodes["manifest"]}, &out)

	require.Len(t, out.Materials, 1, "one material shared by both models")
	gold := out.Materials[0]
	assert.Equal(t, "Gold", gold.Name)
	assert.Equal(t, "#ffc040", gold.Color)
	require.NotNil(t, gold.Metallic)
	assert.Equal(t, 1.0, *gold.Metallic)
	assert.Equal(t, 2, gold.Models)
	assert.Empty(t, gold.Extensions)
	assert.Empty(t, out.ExtensionsUsed)
	for _, m := range out.Models {
		assert.Equal(t, "Gold", m.Material)
	}
	assert.Equal(t, out.Models[0].Triangles*2, out.TotalTriangles)
	assert.Equal(t, out.TotalTriangles, out.StoredTriangles, "each ModelNode gets its own copy of the sphere's mesh, so the file stores both")

	dir := t.TempDir()
	callTool(t, session, "set_producer", map[string]any{"nodeId": created.Nodes["manifest"], "port": "Out", "name": "gold.glb"}, &polyformmcp.SetProducerOutput{})
	callTool(t, session, "generate", map[string]any{"outputDir": dir}, &polyformmcp.GenerateOutput{})
	written, err := os.Stat(filepath.Join(dir, "gold.glb", "model.glb"))
	require.NoError(t, err)
	require.Len(t, out.Files, 1)
	assert.Equal(t, "model.glb", out.Files[0].Name)
	assert.Equal(t, written.Size(), out.Files[0].MeasuredBytes)
}

func TestDescribeManifestReportsExtensionsAttributesInstancesAndBounds(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "R", "type": "float64", "value": "1"}},
	}, &polyformmcp.CreateVariablesOutput{})

	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType, "inputs": map[string]any{
			"Width":  map[string]any{"variable": "R"},
			"Height": map[string]any{"variable": "R"},
			"Depth":  map[string]any{"variable": "R"},
		}},
		map[string]any{"alias": "transmission", "type": transmissionExtType, "inputs": map[string]any{
			"Factor": map[string]any{"value": "0.9"},
		}},
		map[string]any{"alias": "volume", "type": volumeExtType, "inputs": map[string]any{
			"ThicknessFactor": map[string]any{"value": "0.5"},
		}},
		map[string]any{"alias": "dispersion", "type": dispersionExtType, "inputs": map[string]any{
			"Dispersion": map[string]any{"value": "0.36"},
		}},
		map[string]any{"alias": "glass", "type": materialNodeType, "inputs": map[string]any{
			"Name":              map[string]any{"value": `"Glass"`},
			"Color":             map[string]any{"value": `"#ffffff"`},
			"IndexOfRefraction": map[string]any{"value": "1.5"},
			"EmissiveStrength":  map[string]any{"value": "2"},
			"Transmission":      map[string]any{"nodeId": "transmission", "port": "Out"},
			"Volume":            map[string]any{"nodeId": "volume", "port": "Out"},
			"Dispersion":        map[string]any{"nodeId": "dispersion", "port": "Out"},
		}},
		map[string]any{"alias": "gem", "type": gltfModelType, "inputs": map[string]any{
			"Name":          map[string]any{"value": `"Gem"`},
			"Mesh":          map[string]any{"nodeId": "cube", "port": "Out"},
			"Material":      map[string]any{"nodeId": "glass", "port": "Out"},
			"Gpu Instances": map[string]any{"value": `[{"position":{"x":0,"y":0,"z":0}},{"position":{"x":5,"y":0,"z":0}},{"position":{"x":10,"y":0,"z":0}}]`},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"elements": []any{map[string]any{"nodeId": "gem", "port": "Out"}}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var out polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": created.Nodes["manifest"]}, &out)

	require.Len(t, out.Materials, 1)
	glass := out.Materials[0]
	assert.Equal(t, "Glass", glass.Name)
	assert.InDelta(t, 0.9, glass.Extensions["KHR_materials_transmission"]["transmissionFactor"], 1e-9)
	assert.InDelta(t, 0.5, glass.Extensions["KHR_materials_volume"]["thicknessFactor"], 1e-9)
	assert.InDelta(t, 1.5, glass.Extensions["KHR_materials_ior"]["ior"], 1e-9)
	assert.InDelta(t, 0.36, glass.Extensions["KHR_materials_dispersion"]["dispersion"], 1e-9)
	assert.InDelta(t, 2, glass.Extensions["KHR_materials_emissive_strength"]["emissiveStrength"], 1e-9)
	assert.ElementsMatch(t, []string{
		"EXT_mesh_gpu_instancing",
		"KHR_materials_dispersion",
		"KHR_materials_emissive_strength",
		"KHR_materials_ior",
		"KHR_materials_transmission",
		"KHR_materials_volume",
	}, out.ExtensionsUsed)
	assert.Equal(t, []string{"EXT_mesh_gpu_instancing"}, out.ExtensionsRequired)

	require.Len(t, out.Models, 1)
	gem := out.Models[0]
	assert.Equal(t, "Gem", gem.Path)
	assert.Equal(t, 3, gem.Instances)
	assert.Equal(t, 12, gem.StoredTriangles)
	assert.Equal(t, 36, gem.Triangles)
	assert.Equal(t, 36, out.TotalTriangles)
	assert.Equal(t, 12, out.StoredTriangles)
	assert.Equal(t, gem.StoredVertices*3, gem.Vertices)
	assert.Equal(t, gem.StoredVertices, out.StoredVertices)
	assert.Equal(t, []string{"NORMAL", "POSITION", "TEXCOORD_0"}, gem.Attributes)
	require.Len(t, gem.AttributeRanges, 3)
	assert.Equal(t, "VEC3", gem.AttributeRanges[1].Type)
	assert.Equal(t, "float", gem.AttributeRanges[1].Component)
	assert.InDeltaSlice(t, []float64{-0.5, -0.5, -0.5}, gem.AttributeRanges[1].Min, 1e-9)
	assert.InDeltaSlice(t, []float64{0.5, 0.5, 0.5}, gem.AttributeRanges[1].Max, 1e-9)
	assert.Equal(t, "VEC2", gem.AttributeRanges[2].Type)
	require.NotNil(t, gem.Bounds)
	assert.InDelta(t, -0.5, gem.Bounds.Min.X, 1e-9, "the first instance's near face")
	assert.InDelta(t, 10.5, gem.Bounds.Max.X, 1e-9, "the last instance's far face")
	assert.Equal(t, gem.Bounds.Min, out.Bounds.Min)
	assert.Equal(t, gem.Bounds.Max, out.Bounds.Max)
	assert.Nil(t, gem.WorldPosition, "an identity transform is left out")
	assert.Nil(t, gem.WorldRotation)
	assert.Nil(t, gem.WorldScale)
	assert.Empty(t, out.Warnings)

	var wider polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": created.Nodes["manifest"], "variables": map[string]any{"R": "2"}}, &wider)
	assert.InDelta(t, 11, wider.Bounds.Max.X, 1e-9, "the override widened every instance")

	var list polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &list)
	require.Len(t, list.Variables, 1)
	assert.EqualValues(t, 1, list.Variables[0].Value, "the override was restored")

	callTool(t, session, "save_profile", map[string]any{"name": "Big", "variables": map[string]any{"R": "4"}}, &polyformmcp.SaveProfileOutput{})
	var profiled polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": created.Nodes["manifest"], "profile": "Big"}, &profiled)
	assert.InDelta(t, 12, profiled.Bounds.Max.X, 1e-9)

	var lean polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": created.Nodes["manifest"], "includeAttributes": false}, &lean)
	assert.True(t, lean.AttributeRangesOmitted)
	assert.Equal(t, []string{"NORMAL", "POSITION", "TEXCOORD_0"}, lean.Models[0].Attributes, "the names stay, only the ranges go")
	assert.Nil(t, lean.Models[0].AttributeRanges)
}

func TestDescribeManifestWarnsWhenAMaterialNeedsUVsTheMeshLacks(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "sphere", "type": sphereNodeType},
		map[string]any{"alias": "big", "type": sphereNodeType, "inputs": map[string]any{"Radius": map[string]any{"value": "5"}}},
		map[string]any{"alias": "anisotropy", "type": anisotropyExtType, "inputs": map[string]any{
			"AnisotropyStrength": map[string]any{"value": "0.8"},
		}},
		map[string]any{"alias": "volume", "type": volumeExtType, "inputs": map[string]any{
			"ThicknessFactor": map[string]any{"value": "0.5"},
		}},
		map[string]any{"alias": "brushed", "type": materialNodeType, "inputs": map[string]any{
			"Name":       map[string]any{"value": `"Brushed"`},
			"Anisotropy": map[string]any{"nodeId": "anisotropy", "port": "Out"},
			"Volume":     map[string]any{"nodeId": "volume", "port": "Out"},
		}},
		map[string]any{"alias": "small", "type": gltfModelType, "inputs": map[string]any{
			"Name":     map[string]any{"value": `"Small"`},
			"Mesh":     map[string]any{"nodeId": "sphere", "port": "Out"},
			"Material": map[string]any{"nodeId": "brushed", "port": "Out"},
		}},
		map[string]any{"alias": "large", "type": gltfModelType, "inputs": map[string]any{
			"Name":     map[string]any{"value": `"Large"`},
			"Mesh":     map[string]any{"nodeId": "big", "port": "Out"},
			"Material": map[string]any{"nodeId": "brushed", "port": "Out"},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"elements": []any{
				map[string]any{"nodeId": "small", "port": "Out"},
				map[string]any{"nodeId": "large", "port": "Out"},
			}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var out polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": created.Nodes["manifest"]}, &out)

	assert.Equal(t, []string{"POSITION"}, out.Models[0].Attributes)
	require.Len(t, out.Warnings, 3, "%v", out.Warnings)
	assert.Contains(t, out.Warnings[0], `Small: material "Brushed" uses KHR_materials_anisotropy but the mesh has no TEXCOORD_0`)
	assert.Contains(t, out.Warnings[1], `Large: material "Brushed" uses KHR_materials_anisotropy`)
	assert.Contains(t, out.Warnings[2], `volume thicknessFactor 0.5 is shared by Small (1 across) and Large (10 across)`)
}

const projectUVType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/meshops.ProjectUVNode]"

func anisotropicCubeManifest(t *testing.T, session *mcpsdk.ClientSession, flattenUVs bool) string {
	t.Helper()
	mesh := map[string]any{"nodeId": "cube", "port": "Out"}
	nodes := []any{
		map[string]any{"alias": "cube", "type": cubeNodeType},
		map[string]any{"alias": "flat", "type": projectUVType, "inputs": map[string]any{"Mesh": mesh}},
		map[string]any{"alias": "anisotropy", "type": anisotropyExtType, "inputs": map[string]any{
			"AnisotropyStrength": map[string]any{"value": "0.8"},
		}},
		map[string]any{"alias": "steel", "type": materialNodeType, "inputs": map[string]any{
			"Name":       map[string]any{"value": `"Steel"`},
			"Anisotropy": map[string]any{"nodeId": "anisotropy", "port": "Out"},
		}},
	}
	if flattenUVs {
		mesh = map[string]any{"nodeId": "flat", "port": "Out"}
	}
	nodes = append(nodes,
		map[string]any{"alias": "shell", "type": gltfModelType, "inputs": map[string]any{
			"Name":     map[string]any{"value": `"Shell"`},
			"Mesh":     mesh,
			"Material": map[string]any{"nodeId": "steel", "port": "Out"},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"elements": []any{map[string]any{"nodeId": "shell", "port": "Out"}}},
		}},
	)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": nodes}, &created)
	require.Empty(t, created.Errors)
	return created.Nodes["manifest"]
}

func TestDescribeManifestWarnsAboutZeroAreaUVsUnderAnisotropy(t *testing.T) {
	session := testSession(t)
	var out polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": anisotropicCubeManifest(t, session, true)}, &out)

	require.Len(t, out.Warnings, 1, "%v", out.Warnings)
	assert.Contains(t, out.Warnings[0], `Shell: 8 of 12 triangles have zero-area UVs, so material "Steel"`)
}

func TestDescribeManifestDoesNotWarnAboutAnisotropyOnGoodUVs(t *testing.T) {
	session := testSession(t)
	var out polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": anisotropicCubeManifest(t, session, false)}, &out)

	assert.Empty(t, out.Warnings)
}

func TestDescribeManifestRoundsToTheScenesScale(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType},
		map[string]any{"alias": "turned", "type": gltfModelType, "inputs": map[string]any{
			"Mesh":     map[string]any{"nodeId": "cube", "port": "Out"},
			"Rotation": map[string]any{"value": `{"x":0,"y":0.3826834323650898,"z":0,"w":0.9238795325112867}`},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"elements": []any{map[string]any{"nodeId": "turned", "port": "Out"}}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var out polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": created.Nodes["manifest"]}, &out)

	require.Len(t, out.Models, 1)
	assert.Equal(t, 0.3827, out.Models[0].WorldRotation[1])
	assert.Equal(t, 0.7071, out.Models[0].Bounds.Max.X)
	assert.Equal(t, 1.4142, out.Bounds.Size.X)
}

func TestDescribeManifestListsLargeScenesCompactly(t *testing.T) {
	session := testSession(t)
	nodes := []any{map[string]any{"alias": "cube", "type": cubeNodeType}}
	children := []any{}
	for i := range 41 {
		alias := fmt.Sprintf("part%d", i)
		nodes = append(nodes, map[string]any{"alias": alias, "type": gltfModelType, "inputs": map[string]any{
			"Name": map[string]any{"value": fmt.Sprintf(`"Part %d"`, i)},
			"Mesh": map[string]any{"nodeId": "cube", "port": "Out"},
		}})
		children = append(children, map[string]any{"nodeId": alias, "port": "Out"})
	}
	nodes = append(nodes,
		map[string]any{"alias": "group", "type": gltfModelType, "inputs": map[string]any{
			"Name":     map[string]any{"value": `"Board"`},
			"Children": map[string]any{"elements": children},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"elements": []any{map[string]any{"nodeId": "group", "port": "Out"}}},
		}},
	)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": nodes}, &created)
	require.Empty(t, created.Errors)
	manifest := created.Nodes["manifest"]

	var auto polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": manifest}, &auto)
	assert.Equal(t, "compact", auto.ModelDetail)
	assert.Equal(t, 42, auto.ModelCount)
	require.Len(t, auto.Models, 42)
	assert.Equal(t, 12, auto.Models[1].Triangles)
	assert.Nil(t, auto.Models[1].Bounds)
	assert.Nil(t, auto.Models[1].Attributes)
	assert.Nil(t, auto.Models[1].AttributeRanges)
	assert.Equal(t, 41*12, auto.TotalTriangles)
	require.Len(t, auto.Files, 1)
	assert.Positive(t, auto.Files[0].MeasuredBytes)

	var one polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": manifest, "path": "Board/Part 7"}, &one)
	assert.Equal(t, "full", one.ModelDetail)
	require.Len(t, one.Models, 1)
	assert.NotNil(t, one.Models[0].Bounds)
	assert.Equal(t, 41*12, one.TotalTriangles, "totals still cover the whole file")

	var none polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": manifest, "models": "none"}, &none)
	assert.Empty(t, none.Models)
	assert.Equal(t, 42, none.ModelCount)

	msg := callToolExpectingError(t, session, "describe_manifest", map[string]any{"nodeId": manifest, "path": "Board/Part 99"})
	assert.Contains(t, msg, `no model at path "Board/Part 99"`)
}

func TestDescribeManifestListsAnimationChannelsByTargetPath(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType},
		map[string]any{"alias": "mover", "type": gltfModelType, "inputs": map[string]any{
			"Name": map[string]any{"value": `"Mover"`},
			"Mesh": map[string]any{"nodeId": "cube", "port": "Out"},
		}},
		map[string]any{"alias": "frames", "type": uniformFramesType, "inputs": map[string]any{
			"Data":     map[string]any{"value": `[{"x":0,"y":0,"z":0},{"x":1,"y":0,"z":0},{"x":0,"y":0,"z":0}]`},
			"Duration": map[string]any{"value": "2"},
		}},
		map[string]any{"alias": "slide", "type": translationChannel, "inputs": map[string]any{
			"Target": map[string]any{"nodeId": "mover", "port": "Out"},
			"Frames": map[string]any{"nodeId": "frames", "port": "Out"},
		}},
		map[string]any{"alias": "anim", "type": animationNodeType, "inputs": map[string]any{
			"Name":     map[string]any{"value": `"Slide"`},
			"Channels": map[string]any{"elements": []any{map[string]any{"nodeId": "slide", "port": "Out"}}},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models":     map[string]any{"elements": []any{map[string]any{"nodeId": "mover", "port": "Out"}}},
			"Animations": map[string]any{"elements": []any{map[string]any{"nodeId": "anim", "port": "Out"}}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var out polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": created.Nodes["manifest"]}, &out)

	require.Len(t, out.Animations, 1)
	slide := out.Animations[0]
	assert.Equal(t, "Slide", slide.Name)
	assert.InDelta(t, 2, slide.Seconds, 1e-9)
	require.Len(t, slide.Channels, 1)
	assert.Equal(t, "Mover", slide.Channels[0].Target)
	assert.Equal(t, out.Models[0].Node, slide.Channels[0].Node)
	assert.Equal(t, "translation", slide.Channels[0].Path)
	assert.Equal(t, 3, slide.Channels[0].Keys)
	assert.Empty(t, out.Warnings)
}

func TestATransformOnlyParentIsNotEmptyButAMeshlessLeafIs(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType},
		map[string]any{"alias": "part", "type": gltfModelType, "inputs": map[string]any{
			"Name": map[string]any{"value": `"Part"`},
			"Mesh": map[string]any{"nodeId": "cube", "port": "Out"},
		}},
		map[string]any{"alias": "group", "type": gltfModelType, "inputs": map[string]any{
			"Name":        map[string]any{"value": `"Group"`},
			"Children":    map[string]any{"elements": []any{map[string]any{"nodeId": "part", "port": "Out"}}},
			"Translation": map[string]any{"value": `{"x":3,"y":0,"z":0}`},
		}},
		map[string]any{"alias": "gone", "type": gltfModelType, "inputs": map[string]any{
			"Name": map[string]any{"value": `"Gone"`},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"elements": []any{
				map[string]any{"nodeId": "group", "port": "Out"},
				map[string]any{"nodeId": "gone", "port": "Out"},
			}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var out polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": created.Nodes["manifest"]}, &out)

	require.Len(t, out.Models, 3)
	assert.Equal(t, "Group", out.Models[0].Path)
	assert.False(t, out.Models[0].Empty)
	assert.Equal(t, 0, out.Models[0].Triangles)
	assert.Equal(t, "Group/Part", out.Models[1].Path)
	assert.InDelta(t, 3, (out.Models[1].Bounds.Min.X+out.Models[1].Bounds.Max.X)/2, 1e-9, "the parent's transform carries into the child's bounds")
	require.NotNil(t, out.Models[1].WorldPosition)
	assert.Equal(t, 3.0, out.Models[1].WorldPosition.X)
	assert.True(t, out.Models[2].Empty)
	assert.Equal(t, 1, out.EmptyModels)
	assert.Equal(t, 1, out.Models[0].Node, "the file lists the child before its parent")
}

func TestAnInstancedGroupDrawsItsChildrenOnceAndSaysSo(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType},
		map[string]any{"alias": "body", "type": gltfModelType, "inputs": map[string]any{
			"Mesh": map[string]any{"nodeId": "cube", "port": "Out"},
		}},
		map[string]any{"alias": "group", "type": gltfModelType, "inputs": map[string]any{
			"Children":      map[string]any{"elements": []any{map[string]any{"nodeId": "body", "port": "Out"}}},
			"Gpu Instances": map[string]any{"value": `[{"position":{"x":0,"y":0,"z":0}},{"position":{"x":3,"y":0,"z":0}},{"position":{"x":6,"y":0,"z":0}}]`},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"elements": []any{map[string]any{"nodeId": "group", "port": "Out"}}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var render polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{"nodeId": created.Nodes["manifest"], "outputPath": filepath.Join(t.TempDir(), "m.png")}, &render)
	assert.Equal(t, 12, render.TriangleCount, "instancing repeats a model's own mesh, not its children")
	require.Len(t, render.NodeErrors, 1)
	assert.Contains(t, render.NodeErrors[0].Error, "has GPU instances but no mesh")

	msg := callToolExpectingError(t, session, "describe_manifest", map[string]any{"nodeId": created.Nodes["manifest"]})
	assert.Contains(t, msg, "has GPU instances but no mesh", "the export refuses it too")
}
