package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathLengthNode(t *testing.T) {
	lPath := []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(3., 0., 0.), vector3.New(3., 0., 4.)}
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"an L keeps its corner",
			nodetest.NewNode(geometry.PathLengthNode{Path: nodetest.NewPortValue(lPath)}),
			nodetest.AssertOutput("Length", 7.),
			nodetest.AssertOutput("Distances", []float64{0, 3, 7}),
		),
		nodetest.NewTestCase(
			"one length per path",
			nodetest.NewNode(geometry.PathLengthNode{Path: nodetest.NewPortValue([][]vector3.Float64{lPath, lPath[:2]})}),
			nodetest.AssertOutput("Length", []float64{7, 3}),
		),
		nodetest.NewTestCase(
			"a single point",
			nodetest.NewNode(geometry.PathLengthNode{Path: nodetest.NewPortValue(lPath[:1])}),
			nodetest.AssertOutput("Length", 0.),
		),
	).Run(t)
}

func TestSamplePath(t *testing.T) {
	lPath := []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(3., 0., 0.), vector3.New(3., 0., 4.)}
	cases := map[string]struct {
		path     []vector3.Float64
		distance float64
		position vector3.Float64
		tangent  vector3.Float64
		segment  int
	}{
		"first leg":            {lPath, 1, vector3.New(1., 0., 0.), vector3.New(1., 0., 0.), 0},
		"second leg":           {lPath, 5, vector3.New(3., 0., 2.), vector3.New(0., 0., 1.), 1},
		"a corner leaves":      {lPath, 3, vector3.New(3., 0., 0.), vector3.New(0., 0., 1.), 1},
		"the end":              {lPath, 7, vector3.New(3., 0., 4.), vector3.New(0., 0., 1.), 1},
		"past the end clamps":  {lPath, 9, vector3.New(3., 0., 4.), vector3.New(0., 0., 1.), 1},
		"before start clamps":  {lPath, -2, vector3.New(0., 0., 0.), vector3.New(1., 0., 0.), 0},
		"repeated point skips": {[]vector3.Float64{lPath[0], lPath[0], lPath[1]}, 1, vector3.New(1., 0., 0.), vector3.New(1., 0., 0.), 1},
		"one point":            {lPath[:1], 1, lPath[0], vector3.Float64{}, 0},
		"repeated last point":  {[]vector3.Float64{lPath[0], lPath[1], lPath[1]}, 5, vector3.New(3., 0., 0.), vector3.New(1., 0., 0.), 1},
		"one spot throughout":  {[]vector3.Float64{lPath[0], lPath[0]}, 1, lPath[0], vector3.Float64{}, 0},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sample := geometry.SamplePath(c.path, c.distance)
			assert.InDelta(t, 0, sample.Position.Distance(c.position), 1e-9)
			assert.InDelta(t, 0, sample.Tangent.Distance(c.tangent), 1e-9)
			assert.Equal(t, c.segment, sample.Segment)
		})
	}
}

func TestSamplePathNodeTakesADistancePerSample(t *testing.T) {
	lPath := []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(3., 0., 0.), vector3.New(3., 0., 4.)}
	node := &nodes.Struct[geometry.SamplePathNode]{Data: geometry.SamplePathNode{
		Path:     nodetest.NewPortValue(lPath),
		Distance: nodetest.NewPortValue([]float64{1, 5}),
	}}
	assert.Equal(t, []int{0, 1}, nodes.GetNodeOutputPort[[]int](node, "Segment").Value())

	transforms := nodes.GetNodeOutputPort[[]trs.TRS](node, "Transform").Value()
	require.Len(t, transforms, 2)
	facing := transforms[0].Rotation().Rotate(vector3.Forward[float64]())
	assert.InDelta(t, 0, facing.Distance(vector3.New(1., 0., 0.)), 1e-9, "forward turns to the direction of travel")
}
