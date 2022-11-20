package noise

import "github.com/EliCDavis/vector"

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
	return Stack2D{
		f:       Perlin2D,
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
