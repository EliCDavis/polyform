package meshops_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func nickedCube(t *testing.T) modeling.Mesh {
	t.Helper()

	whole := meshops.Weld(
		primitives.Cube{Width: 2, Height: 2, Depth: 2}.UnweldedQuads(),
		modeling.PositionAttribute, 1e-6,
	)
	requireWatertight(t, whole)

	indices := whole.Indices()
	kept := make([]int, 0, indices.Len()-3)
	for i := 3; i < indices.Len(); i++ {
		kept = append(kept, indices.At(i))
	}

	return modeling.NewTriangleMesh(kept).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: readPositions(whole),
		})
}

func readPositions(m modeling.Mesh) []vector3.Float64 {
	data := m.Float3Attribute(modeling.PositionAttribute)
	out := make([]vector3.Float64, data.Len())
	for i := range out {
		out[i] = data.At(i)
	}
	return out
}

func requireWatertight(t *testing.T, m modeling.Mesh) {
	t.Helper()
	require.Empty(t, m.BoundaryEdges())
}

func TestFillHolesClosesAndWindsWithTheSurface(t *testing.T) {
	filled, err := meshops.FillHoles(nickedCube(t))
	require.NoError(t, err)

	assert.Empty(t, filled.BoundaryLoops())
	requireWatertight(t, filled)
}

func TestFillHolesClosesEveryHoleAtOnce(t *testing.T) {
	const sides = 12
	positions := make([]vector3.Float64, 0, sides*2)
	for i := 0; i < sides; i++ {
		angle := 2 * math.Pi * float64(i) / float64(sides)
		x, z := math.Cos(angle), math.Sin(angle)
		positions = append(positions,
			vector3.New(x, -1, z),
			vector3.New(x, 1, z),
		)
	}

	indices := make([]int, 0, sides*6)
	for i := 0; i < sides; i++ {
		lowA, highA := 2*i, 2*i+1
		lowB, highB := 2*((i+1)%sides), 2*((i+1)%sides)+1
		indices = append(indices, lowA, highA, highB, lowA, highB, lowB)
	}

	tube := modeling.NewTriangleMesh(indices).
		SetFloat3Data(map[string][]vector3.Float64{modeling.PositionAttribute: positions})

	require.Len(t, tube.BoundaryLoops(), 2, "sanity: a tube is open at both ends")

	filled, err := meshops.FillHoles(tube)
	require.NoError(t, err)
	assert.Empty(t, filled.BoundaryLoops())
	requireWatertight(t, filled)

	// n points take n-2 triangles.
	assert.Equal(t, tube.PrimitiveCount()+2*(sides-2), filled.PrimitiveCount())
	assert.Equal(t, tube.AttributeLength(), filled.AttributeLength())
}

func TestFillHolesLeavesAClosedMeshAlone(t *testing.T) {
	closed := meshops.Weld(
		primitives.Cube{Width: 2, Height: 2, Depth: 2}.UnweldedQuads(),
		modeling.PositionAttribute, 1e-6,
	)

	filled, err := meshops.FillHoles(closed)
	require.NoError(t, err)
	assert.Equal(t, closed.PrimitiveCount(), filled.PrimitiveCount())
	assert.Equal(t, closed.AttributeLength(), filled.AttributeLength())
}

func TestFillHolesCarriesEveryAttribute(t *testing.T) {
	nicked := nickedCube(t)
	withUVs := nicked.SetFloat2Attribute(modeling.TexCoordAttribute,
		make([]vector2.Float64, nicked.AttributeLength()))

	filled, err := meshops.FillHoles(withUVs)
	require.NoError(t, err)

	assert.ElementsMatch(t, withUVs.Float2Attributes(), filled.Float2Attributes())
	for _, attr := range filled.Float2Attributes() {
		assert.Equal(t, filled.AttributeLength(), filled.Float2Attribute(attr).Len(), attr)
	}
	for _, attr := range filled.Float3Attributes() {
		assert.Equal(t, filled.AttributeLength(), filled.Float3Attribute(attr).Len(), attr)
	}
}

func TestFillHolesReportsHolesItCannotName(t *testing.T) {
	// Two square holes touching at one corner.
	positions := make([]vector3.Float64, 0, 25)
	for y := 0; y <= 4; y++ {
		for x := 0; x <= 4; x++ {
			positions = append(positions, vector3.New(float64(x), 0, float64(y)))
		}
	}
	at := func(x, y int) int { return y*5 + x }

	indices := make([]int, 0)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if (x == 1 && y == 1) || (x == 2 && y == 2) {
				continue
			}
			indices = append(indices,
				at(x, y), at(x+1, y), at(x+1, y+1),
				at(x, y), at(x+1, y+1), at(x, y+1))
		}
	}
	pinched := modeling.NewTriangleMesh(indices).
		SetFloat3Data(map[string][]vector3.Float64{modeling.PositionAttribute: positions})

	filled, err := meshops.FillHoles(pinched)
	require.NoError(t, err)

	assert.Empty(t, filled.BoundaryLoops(),
		"the pinched pair has no rim to return, before or after filling")
	assert.NotEmpty(t, filled.BoundaryEdges(),
		"but the mesh is still open, which is what the node has to report")
}

func TestFillHolesLeavesARimWithNoAreaOpen(t *testing.T) {
	collinear := modeling.NewTriangleMesh([]int{0, 1, 2, 2, 1, 3}).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New(0., 0., 0.),
				vector3.New(1., 0., 0.),
				vector3.New(2., 0., 0.),
				vector3.New(3., 0., 0.),
			},
		})
	require.NotEmpty(t, collinear.BoundaryLoops(), "sanity: this really does have a rim")

	filled, err := meshops.FillHoles(collinear)
	require.NoError(t, err)

	assert.Equal(t, collinear.PrimitiveCount(), filled.PrimitiveCount())
	assert.Equal(t, collinear.AttributeLength(), filled.AttributeLength())
	assert.Equal(t, collinear.BoundaryEdges(), filled.BoundaryEdges())
}

func TestFillHolesNamesTheAttributeItNeeds(t *testing.T) {
	withoutPositions := modeling.NewTriangleMesh([]int{0, 1, 2})
	require.NotEmpty(t, withoutPositions.BoundaryLoops(),
		"sanity: rims come from indices, so this gets past the early return")

	_, err := meshops.FillHoles(withoutPositions)
	assert.EqualError(t, err, "mesh is required to have the vector3 attribute: 'Position'")
}

func notchedDisc(notch vector3.Float64) modeling.Mesh {
	positions := []vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(1., 0., 0.),
		vector3.New(1., 0., 1.),
		notch,
		vector3.New(0., 0., 1.),
		vector3.New(0.5, 0., 0.5),
	}
	normals := make([]vector3.Float64, len(positions))
	for i := range normals {
		normals[i] = vector3.New(0., -1., 0.)
	}
	return modeling.NewTriangleMesh([]int{5, 0, 1, 5, 1, 2, 5, 2, 3, 5, 3, 4, 5, 4, 0}).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: positions,
			modeling.NormalAttribute:   normals,
		}).
		SetFloat2Attribute(modeling.TexCoordAttribute, make([]vector2.Float64, len(positions))).
		SetFloat1Attribute("weight", make([]float64, len(positions)))
}

func TestFillHolesFansARimTheTriangulatorWouldMerge(t *testing.T) {
	t.Run("a distinct notch triangulates across the rim", func(t *testing.T) {
		disc := notchedDisc(vector3.New(0.9, 0., 1.))
		require.Len(t, disc.BoundaryLoops(), 1, "sanity: one rim around the disc")

		filled, err := meshops.FillHoles(disc)
		require.NoError(t, err)

		requireWatertight(t, filled)
		assert.Equal(t, disc.AttributeLength(), filled.AttributeLength())
	})

	t.Run("a notch on top of its neighbour falls back to a fan", func(t *testing.T) {
		disc := notchedDisc(vector3.New(1., 0., 1.))
		require.Len(t, disc.BoundaryLoops(), 1, "sanity: one rim around the disc")

		filled, err := meshops.FillHoles(disc)
		require.NoError(t, err)

		requireWatertight(t, filled)
		middle := disc.AttributeLength()
		assert.Equal(t, middle+1, filled.AttributeLength())

		for _, attr := range filled.Float3Attributes() {
			assert.Equal(t, middle+1, filled.Float3Attribute(attr).Len(), attr)
		}
		for _, attr := range filled.Float2Attributes() {
			assert.Equal(t, middle+1, filled.Float2Attribute(attr).Len(), attr)
		}
		for _, attr := range filled.Float1Attributes() {
			assert.Equal(t, middle+1, filled.Float1Attribute(attr).Len(), attr)
		}

		// The disc faces -Y, so the sheet closing it faces the other way.
		normal := filled.Float3Attribute(modeling.NormalAttribute).At(middle)
		assert.InDelta(t, 1., normal.Y(), 1e-12)
	})
}
