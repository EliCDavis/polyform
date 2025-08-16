package coloring

import (
	"image/color"
	"sort"
)

type GradientKey struct {
	Time  float64
	Color color.Color
}

type normalizedGradientKey struct {
	Time  float64
	Color color.Color
}

func NewGradient(keys ...GradientKey) Gradient {
	if len(keys) == 0 {
		return Gradient{keys: []normalizedGradientKey{}}
	}

	// Create a copy to avoid modifying the original slice
	keysCopy := make([]GradientKey, len(keys))
	copy(keysCopy, keys)

	// Sort keys by time
	sort.Slice(keysCopy, func(i, j int) bool {
		return keysCopy[i].Time < keysCopy[j].Time
	})

	// Find min and max times for normalization
	minTime := keysCopy[0].Time
	maxTime := keysCopy[len(keysCopy)-1].Time
	timeRange := maxTime - minTime

	// Create normalized keys
	normalizedKeys := make([]normalizedGradientKey, len(keysCopy))
	for i, key := range keysCopy {
		normalizedTime := 0.0
		if timeRange > 0 {
			normalizedTime = (key.Time - minTime) / timeRange
		}
		normalizedKeys[i] = normalizedGradientKey{
			Time:  normalizedTime,
			Color: key.Color,
		}
	}

	return Gradient{keys: normalizedKeys}
}

type Gradient struct {
	keys []normalizedGradientKey
}

func (g Gradient) Sample(t float64) color.Color {
	if len(g.keys) == 0 {
		return color.RGBA{0, 0, 0, 255} // Default to black
	}

	if len(g.keys) == 1 {
		return g.keys[0].Color
	}

	if t <= g.keys[0].Time {
		return g.keys[0].Color
	}

	// If t is at or after the last key
	if t >= g.keys[len(g.keys)-1].Time {
		return g.keys[len(g.keys)-1].Color
	}

	// Find corresponding left and right gradient keys straddling time
	leftIdx := 0
	rightIdx := len(g.keys) - 1

	// Find the two keys that straddle the time t
	for i := 0; i < len(g.keys)-1; i++ {
		if t >= g.keys[i].Time && t <= g.keys[i+1].Time {
			leftIdx = i
			rightIdx = i + 1
			break
		}
	}

	leftKey := g.keys[leftIdx]
	rightKey := g.keys[rightIdx]

	// Calculate interpolation factor
	timeRange := rightKey.Time - leftKey.Time
	if timeRange == 0 {
		return leftKey.Color
	}
	factor := (t - leftKey.Time) / timeRange

	// Interpolate between the two colors
	return interpolateColors(leftKey.Color, rightKey.Color, factor)
}

// Helper function to interpolate between two colors
func interpolateColors(c1, c2 color.Color, factor float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	// Convert to 8-bit values for easier arithmetic
	r1, g1, b1, a1 = r1>>8, g1>>8, b1>>8, a1>>8
	r2, g2, b2, a2 = r2>>8, g2>>8, b2>>8, a2>>8

	// Linear interpolation
	r := uint8(float64(r1)*(1-factor) + float64(r2)*factor)
	g := uint8(float64(g1)*(1-factor) + float64(g2)*factor)
	b := uint8(float64(b1)*(1-factor) + float64(b2)*factor)
	a := uint8(float64(a1)*(1-factor) + float64(a2)*factor)

	return color.RGBA{r, g, b, a}
}
