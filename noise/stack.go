package noise

import (
	"math/rand"

	"github.com/EliCDavis/vector"
)

type Sampler2D func(vector.Vector2) float64

type Stack2DEntry struct {
	Scalar    float64
	Amplitude float64
}

type Stack2D struct {
	f       Sampler2D
	entries []Stack2DEntry
}

func PerlinStack(entries []Stack2DEntry) Stack2D {
	newVals := make([]float64, 512)
	for i := 0; i < len(newVals); i++ {
		newVals[i] = rand.Float64()
	}
	return Stack2D{
		f: func(v vector.Vector2) float64 {
			return noise2D(v, fade, gradientOverValues(newVals))
		},
		entries: entries,
	}
}

func (s2d Stack2D) Value(v vector.Vector2) float64 {
	sum := 0.
	for _, entry := range s2d.entries {
		sum += s2d.f(v.MultByConstant(entry.Scalar)) * entry.Amplitude
	}
	return sum
}
