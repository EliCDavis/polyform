package triangulation_test

import (
	"math"
	"math/rand"
	"slices"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func voronoiSquare(size float64) []vector2.Float64 {
	return []vector2.Float64{
		vector2.New(0., 0.), vector2.New(size, 0.), vector2.New(size, size), vector2.New(0., size),
	}
}

func requireVoronoiPartition(t *testing.T, sites, domain []vector2.Float64, cells [][]vector2.Float64) {
	t.Helper()
	require.Len(t, cells, len(sites))

	total := 0.
	for i, cell := range cells {
		area := geometry.Shape(cell).SignedArea()
		require.GreaterOrEqualf(t, area, 0., "cell %d winds clockwise", i)
		total += area

		for _, corner := range cell {
			own := corner.Distance(sites[i])
			for j, other := range sites {
				require.GreaterOrEqualf(t, corner.Distance(other), own-1e-9,
					"corner %v of cell %d is closer to site %d", corner, i, j)
			}
		}
	}
	assert.InDelta(t, math.Abs(geometry.Shape(domain).SignedArea()), total, 1e-9)
}

func TestVoronoiCellsPartitionTheDomain(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	sites := make([]vector2.Float64, 200)
	for i := range sites {
		sites[i] = vector2.New(rng.Float64()*10, rng.Float64()*10)
	}
	domain := voronoiSquare(10)

	cells, err := triangulation.Voronoi(sites, domain)
	require.NoError(t, err)
	requireVoronoiPartition(t, sites, domain, cells)
}

func TestVoronoiNeighboursShareCornersExactly(t *testing.T) {
	rng := rand.New(rand.NewSource(100))
	sites := make([]vector2.Float64, 100)
	for i := range sites {
		sites[i] = vector2.New(rng.Float64(), rng.Float64())
	}

	cells, err := triangulation.Voronoi(sites, nil)
	require.NoError(t, err)

	for i, cell := range cells {
		for _, corner := range cell {
			for j, other := range cells {
				if i == j {
					continue
				}
				for _, o := range other {
					if d := corner.Distance(o); d != 0 {
						require.Greaterf(t, d, 1e-6, "cell %d corner %v nearly matches cell %d corner %v", i, corner, j, o)
					}
				}
			}
		}
	}
}

func TestVoronoiOfOneSiteIsTheDomain(t *testing.T) {
	cells, err := triangulation.Voronoi([]vector2.Float64{vector2.New(1., 1.)}, voronoiSquare(4))
	require.NoError(t, err)
	require.Len(t, cells, 1)
	assert.InDelta(t, 16., geometry.Shape(cells[0]).SignedArea(), 1e-9)
}

func TestVoronoiOfCollinearSitesIsStrips(t *testing.T) {
	sites := []vector2.Float64{vector2.New(1., 2.), vector2.New(3., 2.), vector2.New(2., 2.), vector2.New(0.5, 2.)}
	domain := voronoiSquare(4)

	cells, err := triangulation.Voronoi(sites, domain)
	require.NoError(t, err)
	requireVoronoiPartition(t, sites, domain, cells)
	assert.InDelta(t, 4., geometry.Shape(cells[2]).SignedArea(), 1e-9)
}

func TestVoronoiAcceptsAClockwiseDomain(t *testing.T) {
	sites := []vector2.Float64{vector2.New(1., 1.), vector2.New(3., 3.), vector2.New(1., 3.)}
	domain := voronoiSquare(4)
	slices.Reverse(domain)

	cells, err := triangulation.Voronoi(sites, domain)
	require.NoError(t, err)
	requireVoronoiPartition(t, sites, domain, cells)
}

func TestVoronoiRejectsAConcaveDomain(t *testing.T) {
	domain := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(4., 0.), vector2.New(2., 1.), vector2.New(4., 4.), vector2.New(0., 4.),
	}
	_, err := triangulation.Voronoi([]vector2.Float64{vector2.New(1., 1.)}, domain)
	assert.Error(t, err)
}

func TestVoronoiDefaultsToACircleAroundTheSites(t *testing.T) {
	sites := []vector2.Float64{vector2.New(0., 0.), vector2.New(4., 0.), vector2.New(0., 3.), vector2.New(4., 3.)}

	cells, err := triangulation.Voronoi(sites, nil)
	require.NoError(t, err)

	radius := 5 * 0.75
	for i, cell := range cells {
		assert.NotEmptyf(t, cell, "cell %d", i)
		for _, corner := range cell {
			assert.LessOrEqual(t, corner.Distance(vector2.New(2., 1.5)), radius+1e-9)
		}
	}

	total := 0.
	for _, cell := range cells {
		total += geometry.Shape(cell).SignedArea()
	}
	assert.InDelta(t, 64*math.Sin(2*math.Pi/64)/2*radius*radius, total, 1e-9)
}

func TestVoronoiNodeGivesOneCellPerPoint(t *testing.T) {
	cells := nodes.GetNodeOutputPort[[][]vector2.Float64](&nodes.Struct[triangulation.VoronoiNode]{Data: triangulation.VoronoiNode{
		Points: nodes.ConstOutput[[]vector2.Float64]{Val: []vector2.Float64{vector2.New(1., 1.), vector2.New(3., 1.)}},
		Domain: nodes.ConstOutput[[]vector2.Float64]{Val: voronoiSquare(4)},
	}}, "Out").Value()

	require.Len(t, cells, 2)
	assert.InDelta(t, 8., geometry.Shape(cells[0]).SignedArea(), 1e-9)
	assert.InDelta(t, 8., geometry.Shape(cells[1]).SignedArea(), 1e-9)
}
