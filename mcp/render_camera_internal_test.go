package mcp

import (
	"testing"

	"github.com/fogleman/fauxgl"
	"github.com/stretchr/testify/assert"
)

func near(a, b fauxgl.Vector) bool {
	return a.Sub(b).Length() < 1e-9
}

func TestCameraUpFollowsAzimuthStraightAboveAndBelow(t *testing.T) {
	assert.True(t, near(cameraUp(37, 0), fauxgl.V(0, 1, 0)), "eye level keeps world up")
	assert.True(t, near(cameraUp(0, 90), fauxgl.V(0, 0, -1)))
	assert.True(t, near(cameraUp(0, -90), fauxgl.V(0, 0, 1)))
	assert.True(t, near(cameraUp(90, 90), fauxgl.V(-1, 0, 0)))
	assert.False(t, near(cameraUp(0, -90), cameraUp(180, -90)), "azimuth turns the view from below")
}

func finishShader(f finish, light fauxgl.Vector) previewShader {
	phong := fauxgl.NewPhongShader(fauxgl.Identity(), light, fauxgl.V(0, 0, 5))
	phong.AmbientColor = fauxgl.Color{R: defaultLightAmbient, G: defaultLightAmbient, B: defaultLightAmbient, A: 1}
	phong.DiffuseColor = fauxgl.Color{R: 0.65, G: 0.65, B: 0.65, A: 1}
	return previewShader{PhongShader: phong, finish: f}
}

var gold = fauxgl.Color{R: 1, G: 0.77, B: 0.34, A: 1}

func TestMetalHighlightTakesItsOwnColor(t *testing.T) {
	front := fauxgl.V(0, 0, 1)
	metal := fragment(finishShader(finish{set: true, metallic: 1, roughness: 0.6}, front), gold)
	plastic := fragment(finishShader(finish{set: true, metallic: 0, roughness: 0.6}, front), gold)

	assert.Less(t, metal.B/metal.R, plastic.B/plastic.R, "a dielectric's highlight is white, a metal's is gold")
}

func TestSmoothSurfaceHasTheBrighterHighlight(t *testing.T) {
	front := fauxgl.V(0, 0, 1)
	gray := fauxgl.Color{R: 0.5, G: 0.5, B: 0.5, A: 1}
	smooth := fragment(finishShader(finish{set: true, metallic: 0, roughness: 0.15}, front), gray)
	rough := fragment(finishShader(finish{set: true, metallic: 0, roughness: 0.9}, front), gray)

	assert.Greater(t, smooth.R, rough.R)
}

func TestMetalWithNoDirectLightIsNotBlack(t *testing.T) {
	behind := fauxgl.V(0, 0, -1)
	got := fragment(finishShader(finish{set: true, metallic: 1, roughness: 0.3}, behind), gold)

	assert.Greater(t, got.R, 0.4)
	assert.Greater(t, got.R, got.B, "and still reads gold")
}
