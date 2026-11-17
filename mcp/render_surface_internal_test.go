package mcp

import (
	"image"
	"image/color"
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/fogleman/fauxgl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func upFacingFrame() []tangentFrame {
	return []tangentFrame{{tangent: fauxgl.V(1, 0, 0), bitangent: fauxgl.V(0, 1, 0)}}
}

func facingVertex() fauxgl.Vertex {
	return fauxgl.Vertex{Normal: fauxgl.V(0, 0, 1), Texture: fauxgl.V(0.5, 0.5, 0), Color: fauxgl.Color{R: 1, G: 1, B: 1, A: 1}}
}

func TestNormalMapGreenTiltsTheNormalTowardTheImageTop(t *testing.T) {
	s := surface{normal: solidTexture(color.RGBA{R: 128, G: 255, B: 128, A: 255}), normalScale: 1, tangents: upFacingFrame()}

	n := s.shadingNormal(facingVertex())

	assert.Greater(t, n.Y, 0.9)
	assert.InDelta(t, 0, n.X, 0.01)
}

func TestFlatNormalMapLeavesTheNormalAlone(t *testing.T) {
	s := surface{normal: solidTexture(color.RGBA{R: 128, G: 128, B: 255, A: 255}), normalScale: 1, tangents: upFacingFrame()}

	n := s.shadingNormal(facingVertex())

	assert.InDelta(t, 1, n.Z, 0.001)
}

func TestFaceTangentsFollowUAcrossAndTheImageTopUp(t *testing.T) {
	frame := faceTangentFrame(
		vector3.New(0., 0., 0.), vector3.New(1., 0., 0.), vector3.New(0., 1., 0.),
		vector2.New(0., 1.), vector2.New(1., 1.), vector2.New(0., 0.),
	)

	assert.Equal(t, fauxgl.V(1, 0, 0), frame.tangent)
	assert.Equal(t, fauxgl.V(0, 1, 0), frame.bitangent)
}

func TestTexturesSampleVZeroFromTheImageTop(t *testing.T) {
	im := image.NewRGBA(image.Rect(0, 0, 1, 2))
	im.Set(0, 0, color.RGBA{R: 255, A: 255})
	im.Set(0, 1, color.RGBA{B: 255, A: 255})

	top := imageTexture(&gltf.PolyformTexture{Image: im}).Sample(0.5, 0.25)

	assert.Equal(t, 1., top.R)
	assert.Equal(t, 0., top.B)
}

func TestEmissionLightsADarkSurface(t *testing.T) {
	shader := unlitShader(nil)
	shader.AmbientColor = fauxgl.Color{A: 1}
	dark := fragment(shader, fauxgl.Color{A: 1})
	shader.surface = surface{emissive: fauxgl.Color{R: 1}}

	lit := fragment(shader, fauxgl.Color{A: 1})

	assert.Equal(t, 0., dark.R)
	assert.Greater(t, lit.R, 0.5)
	assert.Equal(t, 0., lit.G)
}

func TestOcclusionMapDarkensAmbientLight(t *testing.T) {
	shader := unlitShader(nil)
	open := fragment(shader, fauxgl.Color{R: 1, G: 1, B: 1, A: 1})
	shader.surface = surface{occlusion: solidTexture(color.RGBA{R: 64, G: 64, B: 64, A: 255}), occlusionStrength: 1}

	occluded := fragment(shader, fauxgl.Color{R: 1, G: 1, B: 1, A: 1})

	assert.Less(t, occluded.R, open.R)
}

func TestOcclusionMapDarkensAMetal(t *testing.T) {
	shader := unlitShader(nil)
	shader.finish = finish{set: true, metallic: 1, roughness: 0.4}
	shader.LightDirection = fauxgl.V(1, 0, 0)
	open := fragment(shader, fauxgl.Color{R: 1, G: 1, B: 1, A: 1})
	shader.surface = surface{occlusion: solidTexture(color.RGBA{R: 64, G: 64, B: 64, A: 255}), occlusionStrength: 1}

	occluded := fragment(shader, fauxgl.Color{R: 1, G: 1, B: 1, A: 1})

	assert.Less(t, occluded.R, open.R*0.8)
}

func TestUVCheckerPaintsPartsWithUVsAndListsTheRest(t *testing.T) {
	positions := []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(1., 0., 0.), vector3.New(0., 1., 0.)}
	mapped := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(modeling.PositionAttribute, positions).
		SetFloat2Attribute(modeling.TexCoordAttribute, []vector2.Float64{vector2.New(0., 1.), vector2.New(1., 1.), vector2.New(0., 0.)})
	bare := modeling.NewTriangleMesh([]int{0, 1, 2}).SetFloat3Attribute(modeling.PositionAttribute, positions)

	scene, err := buildScene([]*gltf.PolyformModel{{Name: "Hat", Mesh: &mapped}, {Name: "Band", Mesh: &bare}}, modelRefs{}, modelRefs{}, "Node-1", 0, true)
	require.NoError(t, err)

	require.Len(t, scene.groups, 2)
	assert.Equal(t, uvCheckerTexture{}, scene.groups[0].Texture)
	assert.Nil(t, scene.groups[1].Texture)
	assert.Equal(t, []string{"uvChecker: no UVs, drawn without the checker: Band"}, scene.warnings)
}

func layersOf(t *testing.T, extensions ...gltf.MaterialExtension) layers {
	t.Helper()
	return materialLayers(&gltf.PolyformMaterial{Extensions: extensions})
}

var (
	towardCamera = fauxgl.V(0, 0, 1)
	grey         = fauxgl.Color{R: 0.5, G: 0.5, B: 0.5, A: 1}
)

func TestTransmissionSeesThroughFacingAndMirrorsAtTheEdge(t *testing.T) {
	glass := layersOf(t, gltf.PolyformTransmission{Factor: 1})

	_, facing := glass.apply(grey, grey, towardCamera, towardCamera, towardCamera)
	_, edge := glass.apply(grey, grey, fauxgl.V(1, 0, 0.05).Normalize(), towardCamera, towardCamera)

	assert.Less(t, facing, 0.5)
	assert.Greater(t, edge, facing)
	assert.True(t, glass.transmissive())
}

func TestVolumeTintsWhatTransmissionShows(t *testing.T) {
	green := layersOf(t,
		gltf.PolyformTransmission{Factor: 1},
		gltf.PolyformVolume{ThicknessFactor: 1, AttenuationColor: color.RGBA{G: 255, A: 255}},
	)

	shaded, _ := green.apply(grey, fauxgl.Color{R: 1, G: 1, B: 1, A: 1}, towardCamera, towardCamera, towardCamera)

	assert.Greater(t, shaded.G, shaded.R)
}

func TestClearcoatAddsATightHighlight(t *testing.T) {
	coat := layersOf(t, gltf.PolyformClearcoat{ClearcoatFactor: 1, ClearcoatRoughnessFactor: 0.03})

	lacquered, _ := coat.apply(grey, grey, towardCamera, towardCamera, towardCamera)
	plain, _ := layers{}.apply(grey, grey, towardCamera, towardCamera, towardCamera)

	assert.Greater(t, lacquered.R, plain.R)
}

func TestSheenGlowsMostAtGrazingAngles(t *testing.T) {
	velvet := layersOf(t, gltf.PolyformSheen{SheenColorFactor: color.RGBA{R: 255, A: 255}, SheenRoughnessFactor: 0.6})
	edgeNormal := fauxgl.V(1, 0, 0.1).Normalize()

	facing, _ := velvet.apply(grey, grey, towardCamera, towardCamera, towardCamera)
	edge, _ := velvet.apply(grey, grey, edgeNormal, towardCamera, towardCamera)

	assert.Greater(t, facing.R, grey.R)
	assert.Greater(t, edge.R, facing.R)
}

func TestSheenColorTextureVariesTheGlow(t *testing.T) {
	velvet := layersOf(t, gltf.PolyformSheen{
		SheenColorFactor:  color.RGBA{R: 255, G: 255, B: 255, A: 255},
		SheenColorTexture: &gltf.PolyformTexture{Image: image.NewUniform(color.RGBA{R: 255, A: 255})},
	})

	glow := velvet.at(fauxgl.Vertex{Texture: fauxgl.V(0.5, 0.5, 0)}).sheen

	assert.Greater(t, glow.R, 0.9)
	assert.Equal(t, 0., glow.G)
}
