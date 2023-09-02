package noise

import (
	"math/rand"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector2"
)

type Stack2DEntry struct {
	Scalar    float64
	Amplitude float64
}

type Stack2D struct {
	f       sample.Vec2ToFloat
	entries []Stack2DEntry
}

func PerlinStack(entries ...Stack2DEntry) Stack2D {
	newVals := make([]float64, 512)
	for i := 0; i < len(newVals); i++ {
		newVals[i] = rand.Float64()
	}
	return Stack2D{
		f: func(v vector2.Float64) float64 {
			return Noise2D(v, QuinticInterpolation, gradientOverValues2D(newVals))
		},
		entries: entries,
	}
}

func (s2d Stack2D) Value(v vector2.Float64) float64 {
	sum := 0.
	for _, entry := range s2d.entries {
		sum += s2d.f(v.Scale(entry.Scalar)) * entry.Amplitude
	}
	return sum
}
