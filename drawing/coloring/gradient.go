package coloring

import (
	"sort"

	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector1"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type GradientKey[T any] struct {
	Time  float64
	Value T
}

type normalizedGradientKey[T any] struct {
	Time  float64
	Color T
}

func NewGradient1D(keys ...GradientKey[float64]) Gradient[float64] {
	return NewGradient(vector1.Space[float64]{}, keys...)
}

func NewGradient2D(keys ...GradientKey[vector2.Float64]) Gradient[vector2.Float64] {
	return NewGradient(vector2.Space[float64]{}, keys...)
}

func NewGradient3D(keys ...GradientKey[vector3.Float64]) Gradient[vector3.Float64] {
	return NewGradient(vector3.Space[float64]{}, keys...)
}

func NewGradient4D(keys ...GradientKey[vector4.Float64]) Gradient[vector4.Float64] {
	return NewGradient(vector4.Space[float64]{}, keys...)
}

func NewGradientColor(keys ...GradientKey[Color]) Gradient[Color] {
	return NewGradient(Space{}, keys...)
}

func NewGradient[T any](space vector.Space[T], keys ...GradientKey[T]) Gradient[T] {
	if len(keys) == 0 {
		return Gradient[T]{
			space: space,
			keys:  []normalizedGradientKey[T]{},
		}
	}

	// Create a copy to avoid modifying the original slice
	keysCopy := make([]GradientKey[T], len(keys))
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
	normalizedKeys := make([]normalizedGradientKey[T], len(keysCopy))
	for i, key := range keysCopy {
		normalizedTime := 0.0
		if timeRange > 0 {
			normalizedTime = (key.Time - minTime) / timeRange
		}
		normalizedKeys[i] = normalizedGradientKey[T]{
			Time:  normalizedTime,
			Color: key.Value,
		}
	}

	return Gradient[T]{
		keys:  normalizedKeys,
		space: space,
	}
}

type Gradient[T any] struct {
	keys  []normalizedGradientKey[T]
	space vector.Space[T]
}

func (g Gradient[T]) Sample(t float64) T {
	if len(g.keys) == 0 {
		var t T
		return t
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
	return g.space.Add(
		g.space.Scale(leftKey.Color, 1-factor),
		g.space.Scale(rightKey.Color, factor),
	)
}
