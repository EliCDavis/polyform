package extrude_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The default strip: x across (the perimeter) becomes v, y along (the
// path) becomes u, both spanning 0..1.
func unitStrip() *primitives.StripUVs {
	return &primitives.StripUVs{
		Start: vector2.New(0., 0.5),
		End:   vector2.New(1., 0.5),
		Width: 1,
	}
}

func TestOutlineWritesNoUVsUnlessAsked(t *testing.T) {
	mesh, err := extrude.Outline{Shape: unitSquare, Path: straightUp}.Extrude()
	require.NoError(t, err)
	assert.False(t, mesh.HasFloat2Attribute(modeling.TexCoordAttribute))
}

func TestOutlineStripUVsSpanTheSweep(t *testing.T) {
	mesh, err := extrude.Outline{Shape: unitSquare, Path: straightUp, UVs: unitStrip()}.Extrude()
	require.NoError(t, err)
	require.True(t, mesh.HasFloat2Attribute(modeling.TexCoordAttribute))

	uv := mesh.Float2Attribute(modeling.TexCoordAttribute)
	positions := mesh.Float3Attribute(modeling.PositionAttribute)
	require.Equal(t, positions.Len(), uv.Len(), "every vertex needs a uv")

	var minU, maxU, minV, maxV = 1., 0., 1., 0.
	for i := 0; i < uv.Len(); i++ {
		c := uv.At(i)
		assert.GreaterOrEqual(t, c.X(), -1e-9, "u stays in range at vertex %d", i)
		assert.LessOrEqual(t, c.X(), 1+1e-9, "u stays in range at vertex %d", i)
		assert.GreaterOrEqual(t, c.Y(), -1e-9, "v stays in range at vertex %d", i)
		assert.LessOrEqual(t, c.Y(), 1+1e-9, "v stays in range at vertex %d", i)
		minU, maxU = min(minU, c.X()), max(maxU, c.X())
		minV, maxV = min(minV, c.Y()), max(maxV, c.Y())
	}
	assert.InDelta(t, 0, minU, 1e-9)
	assert.InDelta(t, 1, maxU, 1e-9)
	assert.InDelta(t, 0, minV, 1e-9)
	assert.InDelta(t, 1, maxV, 1e-9)

	// On the shell, u tracks the path: the bottom ring reads 0 and the top
	// ring 1. The caps are their own unwrap, so they are exempt - hence
	// "some vertex at this height" rather than "every vertex".
	bottom, top := false, false
	for i := 0; i < positions.Len(); i++ {
		switch {
		case positions.At(i).Y() == 0 && uv.At(i).X() == 0:
			bottom = true
		case positions.At(i).Y() == 3 && uv.At(i).X() == 1:
			top = true
		}
	}
	assert.True(t, bottom, "the ring at the path's start reads u 0")
	assert.True(t, top, "the ring at the path's end reads u 1")
}

// A profile that runs smoothly through its closing point still has to put
// perimeter 0 and perimeter 1 somewhere, or the texture mirrors back across
// the final edge.
func TestOutlineStripUVsDoNotReverseAtTheSeam(t *testing.T) {
	circle := make([]vector2.Float64, 0, 32)
	for i := 0; i < 32; i++ {
		angle := float64(i) / 32 * 2 * math.Pi
		circle = append(circle, vector2.New(math.Cos(angle), math.Sin(angle)))
	}

	smoothed := 180.
	mesh, err := extrude.Outline{
		Shape:          circle,
		Path:           straightUp,
		SmoothingAngle: &smoothed,
		UVs:            unitStrip(),
	}.Extrude()
	require.NoError(t, err)

	uv := mesh.Float2Attribute(modeling.TexCoordAttribute)
	seen := map[float64]bool{}
	for i := 0; i < uv.Len(); i++ {
		seen[uv.At(i).Y()] = true
	}
	assert.True(t, seen[0], "the seam's near side reads 0")
	assert.True(t, seen[1], "the seam's far side reads 1")
}

func TestOutlineStripUVsHonourTheStrip(t *testing.T) {
	mesh, err := extrude.Outline{
		Shape: unitSquare,
		Path:  straightUp,
		UVs:   &primitives.StripUVs{Start: vector2.New(0., 0.25), End: vector2.New(0.5, 0.25), Width: 0.5},
	}.Extrude()
	require.NoError(t, err)

	uv := mesh.Float2Attribute(modeling.TexCoordAttribute)
	for i := 0; i < uv.Len(); i++ {
		c := uv.At(i)
		assert.LessOrEqual(t, c.X(), 0.5+1e-9, "u stays inside the strip at vertex %d", i)
		assert.GreaterOrEqual(t, c.Y(), 0-1e-9, "v stays inside the strip at vertex %d", i)
		assert.LessOrEqual(t, c.Y(), 0.5+1e-9, "v stays inside the strip at vertex %d", i)
	}
}

func TestOutlineStripUVsOnAClosedPath(t *testing.T) {
	ring := []vector3.Float64{
		vector3.New(2., 0., 0.), vector3.New(0., 0., 2.),
		vector3.New(-2., 0., 0.), vector3.New(0., 0., -2.),
	}
	mesh, err := extrude.Outline{Shape: unitSquare, Path: ring, Closed: true, UVs: unitStrip()}.Extrude()
	require.NoError(t, err)

	uv := mesh.Float2Attribute(modeling.TexCoordAttribute)
	require.Equal(t, mesh.Float3Attribute(modeling.PositionAttribute).Len(), uv.Len())

	var maxU float64
	for i := 0; i < uv.Len(); i++ {
		maxU = max(maxU, uv.At(i).X())
	}
	assert.InDelta(t, 1, maxU, 1e-9, "the closing segment reaches the far end of the strip")
}
