package primitives_test

import (
	"fmt"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func degenerateTriangles(m modeling.Mesh) int {
	indices := m.Indices()
	positions := m.Float3Attribute(modeling.PositionAttribute)

	count := 0
	for i := 0; i+2 < indices.Len(); i += 3 {
		a := positions.At(indices.At(i))
		b := positions.At(indices.At(i + 1))
		c := positions.At(indices.At(i + 2))
		if b.Sub(a).Cross(c.Sub(a)).Length() == 0 {
			count++
		}
	}
	return count
}

// The index slice was built with make([]int, n) rather than a capacity, so
// every subdivided quad carried n leading zeros: one whole triangle of
// (0, 0, 0) per real triangle, doubling the count.
func TestSubdividedQuadsCarryNoEmptyTriangles(t *testing.T) {
	for _, subdivisions := range []int{1, 2, 3, 7, 10} {
		t.Run(fmt.Sprintf("%dx%d", subdivisions, subdivisions), func(t *testing.T) {
			quad := primitives.Quad{
				Width: 1, Depth: 1,
				Rows: subdivisions, Columns: subdivisions,
			}.ToMesh()

			require.Equal(t, subdivisions*subdivisions*2, quad.Indices().Len()/3,
				"two triangles per cell and nothing else")
			assert.Zero(t, degenerateTriangles(quad))
		})
	}
}

func TestSubdividedCubesCarryNoEmptyTriangles(t *testing.T) {
	for _, subdivisions := range []int{1, 2, 10} {
		t.Run(fmt.Sprintf("dimensions %d", subdivisions), func(t *testing.T) {
			cube := primitives.Cube{
				Width: 1, Height: 1, Depth: 1, Dimensions: subdivisions,
			}.UnweldedQuads()

			require.Equal(t, subdivisions*subdivisions*12, cube.Indices().Len()/3)
			assert.Zero(t, degenerateTriangles(cube))

			sphere := primitives.QuadSphere(0.5, primitives.Cube{
				Width: 1, Height: 1, Depth: 1, Dimensions: subdivisions,
			}, false, false)

			require.Equal(t, subdivisions*subdivisions*12, sphere.Indices().Len()/3)
			assert.Zero(t, degenerateTriangles(sphere))
		})
	}
}

func edgeUseCounts(m modeling.Mesh) map[int]int {
	indices := m.Indices()

	edges := map[[2]int]int{}
	for i := 0; i+2 < indices.Len(); i += 3 {
		corners := [3]int{indices.At(i), indices.At(i + 1), indices.At(i + 2)}
		for k := 0; k < 3; k++ {
			a, b := corners[k], corners[(k+1)%3]
			if a > b {
				a, b = b, a
			}
			edges[[2]int{a, b}]++
		}
	}

	counts := map[int]int{}
	for _, shared := range edges {
		counts[shared]++
	}
	return counts
}

func TestWeldedCubesSubdivide(t *testing.T) {
	for _, subdivisions := range []int{2, 3, 10} {
		t.Run(fmt.Sprintf("dimensions %d", subdivisions), func(t *testing.T) {
			cube := primitives.Cube{
				Width: 1, Height: 1, Depth: 1, Dimensions: subdivisions,
			}.Welded()

			positions := cube.Float3Attribute(modeling.PositionAttribute)
			assert.Equal(t, 6*subdivisions*subdivisions+2, positions.Len())
			assert.Equal(t, 12*subdivisions*subdivisions, cube.Indices().Len()/3)
			assert.Zero(t, degenerateTriangles(cube))

			// Shared by index rather than by position, which is what says the
			// corners were actually welded and not merely coincident.
			assert.Equal(t, map[int]int{2: 18 * subdivisions * subdivisions},
				edgeUseCounts(cube))
		})
	}
}

func TestWeldedCubesFaceOutward(t *testing.T) {
	cube := primitives.Cube{Width: 1, Height: 1, Depth: 1, Dimensions: 4}.Welded()

	indices := cube.Indices()
	positions := cube.Float3Attribute(modeling.PositionAttribute)
	for i := 0; i+2 < indices.Len(); i += 3 {
		a := positions.At(indices.At(i))
		b := positions.At(indices.At(i + 1))
		c := positions.At(indices.At(i + 2))

		outward := a.Add(b).Add(c).Scale(1. / 3.)
		require.Positivef(t, b.Sub(a).Cross(c.Sub(a)).Dot(outward),
			"triangle %d faces inward", i/3)
	}
}

func TestUnsubdividedWeldedCubesAreUnchanged(t *testing.T) {
	for _, subdivisions := range []int{0, 1} {
		cube := primitives.Cube{
			Width: 1, Height: 1, Depth: 1, Dimensions: subdivisions,
		}.Welded()

		assert.Equal(t, 8, cube.Float3Attribute(modeling.PositionAttribute).Len())
		assert.Equal(t, 12, cube.Indices().Len()/3)
	}
}

func TestWeldedQuadSpheresHonourResolution(t *testing.T) {
	for _, subdivisions := range []int{2, 10} {
		t.Run(fmt.Sprintf("dimensions %d", subdivisions), func(t *testing.T) {
			sphere := primitives.QuadSphere(0.5, primitives.Cube{
				Width: 1, Height: 1, Depth: 1, Dimensions: subdivisions,
			}, true, false)

			positions := sphere.Float3Attribute(modeling.PositionAttribute)
			require.Equal(t, 6*subdivisions*subdivisions+2, positions.Len())
			assert.Equal(t, 12*subdivisions*subdivisions, sphere.Indices().Len()/3)

			for i := 0; i < positions.Len(); i++ {
				assert.InDelta(t, 0.5, positions.At(i).Length(), 1e-12)
			}
		})
	}
}
