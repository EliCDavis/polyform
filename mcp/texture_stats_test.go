package mcp

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A texture has no useful JSON form, so a build could only ever eyeball
// whether a pattern covered anything.
func TestTextureStatsReportsTheDistribution(t *testing.T) {
	tex := texturing.Empty[float64](4, 4)
	for y := range 4 {
		for x := range 4 {
			tex.Set(x, y, 0)
		}
	}
	// A quarter of the texture set high: the case "did the camo cover
	// anything" actually asks about.
	tex.Set(0, 0, 1)
	tex.Set(1, 0, 1)
	tex.Set(2, 0, 1)
	tex.Set(3, 0, 1)

	stats := textureStats(tex)
	require.NotNil(t, stats)
	assert.Equal(t, 4, stats.Width)
	assert.Equal(t, 16, stats.Pixels)

	require.Len(t, stats.Channel, 1)
	c := stats.Channel[0]
	assert.Equal(t, 0., c.Min)
	assert.Equal(t, 1., c.Max)
	assert.InDelta(t, 0.25, c.Mean, 1e-9, "a quarter of the pixels are high")
	assert.Equal(t, 12, c.Histogram[0], "and the rest sit in the bottom bucket")
	assert.Equal(t, 4, c.Histogram[len(c.Histogram)-1])
}

func TestTextureStatsOfAFlatTextureDoesNotDivideByZero(t *testing.T) {
	tex := texturing.Empty[float64](2, 2)
	tex.Fill(0.5)

	stats := textureStats(tex)
	require.NotNil(t, stats)
	c := stats.Channel[0]
	assert.Equal(t, 0.5, c.Min)
	assert.Equal(t, 0.5, c.Max)
	assert.Equal(t, 4, c.Histogram[0])
}

func TestTextureStatsIgnoresThingsThatAreNotTextures(t *testing.T) {
	assert.Nil(t, textureStats(42))
	assert.Nil(t, textureStats("not a texture"))
}
