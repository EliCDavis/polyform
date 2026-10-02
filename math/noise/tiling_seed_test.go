package noise_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/noise"
	"github.com/stretchr/testify/assert"
)

func TestTilingNoiseIsTheSameEveryTime(t *testing.T) {
	first := noise.NewTilingNoise(64, 1/16., 3)
	second := noise.NewTilingNoise(64, 1/16., 3)

	for y := range 8 {
		for x := range 8 {
			assert.Equal(t, first.Noise(x, y), second.Noise(x, y), "at %d,%d", x, y)
		}
	}
}

func TestTilingNoiseSeedChangesThePattern(t *testing.T) {
	a := noise.NewTilingNoiseWithSeed(64, 1/16., 3, 1)
	b := noise.NewTilingNoiseWithSeed(64, 1/16., 3, 2)

	differs := false
	for y := range 8 {
		for x := range 8 {
			if a.Noise(x, y) != b.Noise(x, y) {
				differs = true
			}
		}
	}
	assert.True(t, differs, "two seeds should not give the same pattern")
}

func TestTilingNoiseSameSeedSamePattern(t *testing.T) {
	a := noise.NewTilingNoiseWithSeed(64, 1/16., 3, 7)
	b := noise.NewTilingNoiseWithSeed(64, 1/16., 3, 7)
	assert.Equal(t, a.Noise(3, 5), b.Noise(3, 5))
}
