package mcp

import (
	"math"
	"testing"

	"github.com/fogleman/fauxgl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func f64(v float64) *float64 { return &v }

// domeShading is the spread of diffuse terms across a shallow dome facing
// the camera - how much of the form a render can actually show. A light
// at the camera lights every one of these normals almost equally.
func domeShading(light RenderLight) float64 {
	toCamera := fauxgl.V(0, 0, 1)
	dir := lightDirection(toCamera, fauxgl.V(0, 1, 0), light)

	lo, hi := math.Inf(1), math.Inf(-1)
	for i := range 9 {
		// Normals tilting +/- 20 degrees off the view axis, the range a
		// shallow dome or a rolled arm actually spans.
		angle := (-20 + 5*float64(i)) * math.Pi / 180
		n := fauxgl.V(math.Sin(angle), 0, math.Cos(angle))
		d := math.Max(n.Dot(dir), 0)
		lo, hi = math.Min(lo, d), math.Max(hi, d)
	}
	return hi - lo
}

func TestDefaultLightShowsShallowFormAHeadlightCannot(t *testing.T) {
	headlight := domeShading(RenderLight{Azimuth: f64(0), Elevation: f64(0)})
	def := domeShading(RenderLight{})

	assert.Less(t, headlight, 0.07, "a light at the camera barely varies across a shallow dome")
	assert.Greater(t, def, 0.2, "the default key light spreads those same normals over a readable range")
	assert.Greater(t, def, headlight*3)
}

func TestRakingLightSpreadsShallowFormFurtherStill(t *testing.T) {
	assert.Greater(t,
		domeShading(RenderLight{Azimuth: f64(70)}),
		domeShading(RenderLight{}),
		"a wide azimuth rakes across relief the three-quarter key still softens")
}

func TestLightDirectionIsCameraRelative(t *testing.T) {
	up := fauxgl.V(0, 1, 0)

	// The same angles against two different camera positions put the
	// light in the same place relative to each camera, not in the world.
	front := lightDirection(fauxgl.V(0, 0, 1), up, RenderLight{Azimuth: f64(90), Elevation: f64(0)})
	side := lightDirection(fauxgl.V(1, 0, 0), up, RenderLight{Azimuth: f64(90), Elevation: f64(0)})

	assert.InDelta(t, 1, front.X, 1e-9, "from the front camera, +90 is world +X")
	assert.InDelta(t, 0, front.Z, 1e-9)
	assert.InDelta(t, -1, side.Z, 1e-9, "from the +X camera, +90 is world -Z")
	assert.InDelta(t, 0, side.X, 1e-9)
}

func TestLightElevationRaisesTheKey(t *testing.T) {
	straight := lightDirection(fauxgl.V(0, 0, 1), fauxgl.V(0, 1, 0), RenderLight{Azimuth: f64(0), Elevation: f64(0)})
	raised := lightDirection(fauxgl.V(0, 0, 1), fauxgl.V(0, 1, 0), RenderLight{Azimuth: f64(0), Elevation: f64(90)})

	assert.InDelta(t, 0, straight.Y, 1e-9)
	assert.InDelta(t, 1, raised.Y, 1e-9)
}

func TestLightDirectionSurvivesATopDownCamera(t *testing.T) {
	dir := lightDirection(fauxgl.V(0, 1, 0), fauxgl.V(0, 1, 0), RenderLight{})
	require.False(t, math.IsNaN(dir.X) || math.IsNaN(dir.Y) || math.IsNaN(dir.Z),
		"a camera straight overhead makes the up-cross-toCamera basis degenerate")
	assert.InDelta(t, 1, dir.Length(), 1e-9)
}

func TestAmbientDefaultsAndClamps(t *testing.T) {
	assert.Equal(t, defaultLightAmbient, RenderLight{}.ambient())
	assert.Equal(t, 0.1, RenderLight{Ambient: f64(0.1)}.ambient())
	assert.Equal(t, 0., RenderLight{Ambient: f64(-1)}.ambient())
	assert.Equal(t, 1., RenderLight{Ambient: f64(5)}.ambient())
}
