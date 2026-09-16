package csg

import (
	"fmt"
	"os"
	"testing"

	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func benchSolid(t testing.TB, dimensions int) *solid {
	t.Helper()
	faces, err := facesOf(benchSphere(dimensions, vector3.Zero[float64]()))
	require.NoError(t, err)
	return newSolid(faces, toleranceFor(faces, nil), 0)
}

func TestEnclosed(t *testing.T) {
	sphere := benchSolid(t, 10)

	assert.Equal(t, inside, sphere.enclosed(vector3.Zero[float64]()))
	assert.Equal(t, inside, sphere.enclosed(vector3.New(0.3, 0.2, -0.1)))
	assert.Equal(t, outside, sphere.enclosed(vector3.New(0.6, 0., 0.)))
	assert.Equal(t, outside, sphere.enclosed(vector3.New(5., 5., 5.)))

	// A whisker inside and outside a face, closer than any ray would trust.
	f := sphere.faces[0]
	at := f.barycenter()
	assert.Equal(t, inside, sphere.enclosed(at.Sub(f.normal.Scale(1e-9))))
	assert.Equal(t, outside, sphere.enclosed(at.Add(f.normal.Scale(1e-9))))
}

func BenchmarkEnclosed(b *testing.B) {
	sizes := []int{10, 92}
	if os.Getenv("BENCH_BIG") != "" {
		sizes = append(sizes, 289)
	}

	for _, dimensions := range sizes {
		sphere := benchSolid(b, dimensions)
		b.Run(fmt.Sprintf("%d faces", len(sphere.faces)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sphere.enclosed(vector3.New(0.1, 0.2, 0.3))
			}
		})
	}
}
