package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
)

func TestConvexHullDropsInteriorAndEdgePoints(t *testing.T) {
	hull := geometry.ConvexHull([]vector2.Float64{
		vector2.New(2., 2.),
		vector2.New(0., 0.),
		vector2.New(1., 1.),
		vector2.New(2., 0.),
		vector2.New(1., 0.),
		vector2.New(0., 2.),
		vector2.New(0., 0.),
	})

	assert.Equal(t, []vector2.Float64{
		vector2.New(0., 0.), vector2.New(2., 0.), vector2.New(2., 2.), vector2.New(0., 2.),
	}, hull)
}

func TestConvexHullOfCollinearPointsIsTheirEnds(t *testing.T) {
	hull := geometry.ConvexHull([]vector2.Float64{vector2.New(1., 1.), vector2.New(3., 3.), vector2.New(0., 0.), vector2.New(2., 2.)})
	assert.Equal(t, []vector2.Float64{vector2.New(0., 0.), vector2.New(3., 3.)}, hull)
}

func TestConvexHullOfFewerThanThreePointsIsThePoints(t *testing.T) {
	assert.Empty(t, geometry.ConvexHull(nil))
	assert.Equal(t, []vector2.Float64{vector2.New(1., 2.)}, geometry.ConvexHull([]vector2.Float64{vector2.New(1., 2.), vector2.New(1., 2.)}))
}

func TestConvexHullNodeWrapsEachPointSetOfAList(t *testing.T) {
	hulls := nodes.GetNodeOutputPort[[][]vector2.Float64](&nodes.Struct[geometry.ConvexHullNode]{Data: geometry.ConvexHullNode{
		Points: nodes.ConstOutput[[][]vector2.Float64]{Val: [][]vector2.Float64{
			{vector2.New(0., 0.), vector2.New(1., 0.), vector2.New(0., 1.), vector2.New(0.2, 0.2)},
			{vector2.New(0., 0.), vector2.New(4., 0.), vector2.New(4., 4.), vector2.New(0., 4.), vector2.New(2., 2.)},
		}},
	}}, "Out").Value()

	assert.Len(t, hulls, 2)
	assert.Len(t, hulls[0], 3)
	assert.InDelta(t, 16., geometry.Shape(hulls[1]).SignedArea(), 1e-12)
}
