package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLineStripsFromPoints2D(t *testing.T) {
	a, b, c := vector2.New(0., 0.), vector2.New(1., 0.), vector2.New(1., 1.)

	assert.Nil(t, geometry.LineStripsFromPoints2D([]vector2.Float64{a}, true))
	assert.Equal(t, []geometry.Line2D{geometry.NewLine2D(a, b), geometry.NewLine2D(b, c)},
		geometry.LineStripsFromPoints2D([]vector2.Float64{a, b, c}, false))
	assert.Equal(t, []geometry.Line2D{geometry.NewLine2D(a, b), geometry.NewLine2D(b, c), geometry.NewLine2D(c, a)},
		geometry.LineStripsFromPoints2D([]vector2.Float64{a, b, c}, true))
	assert.Equal(t, []geometry.Line2D{geometry.NewLine2D(a, b)},
		geometry.LineStripsFromPoints2D([]vector2.Float64{a, b}, true),
		"two points close to the same segment, not a doubled one")
}

func TestLinesFromPoints2DPairsPointsAndDropsAnOddOneOut(t *testing.T) {
	a, b, c := vector2.New(0., 0.), vector2.New(1., 0.), vector2.New(1., 1.)

	assert.Equal(t, []geometry.Line2D{geometry.NewLine2D(a, b)},
		geometry.LinesFromPoints2D([]vector2.Float64{a, b, c}))
}

func TestLine2DPromotesOntoEachAxisPlane(t *testing.T) {
	line := geometry.NewLine2D(vector2.New(1., 2.), vector2.New(3., 4.))

	assert.Equal(t, geometry.NewLine3D(vector3.New(5., 1., 2.), vector3.New(5., 3., 4.)), line.PlaneX(5))
	assert.Equal(t, geometry.NewLine3D(vector3.New(1., 5., 2.), vector3.New(3., 5., 4.)), line.PlaneY(5))
	assert.Equal(t, geometry.NewLine3D(vector3.New(1., 2., 5.), vector3.New(3., 4., 5.)), line.PlaneZ(5))
}

func TestPromotedLinesLandOnTheAreaExtrudedFromTheSamePoints(t *testing.T) {
	outline := []vector2.Float64{vector2.New(0., 0.), vector2.New(2., 0.), vector2.New(2., 1.), vector2.New(0., 1.)}
	area, err := sketch.Fill([]geometry.Shape{outline}, nil)
	require.NoError(t, err)
	lines := geometry.LineStripsFromPoints2D(outline, true)

	planes := map[string]struct {
		extrude func(sketch.Area, float64) (modeling.Mesh, error)
		promote func(geometry.Line2D, float64) geometry.Line3D
	}{
		"X": {extrude.AreaX, geometry.Line2D.PlaneX},
		"Y": {extrude.AreaY, geometry.Line2D.PlaneY},
		"Z": {extrude.AreaZ, geometry.Line2D.PlaneZ},
	}
	for name, plane := range planes {
		t.Run(name, func(t *testing.T) {
			face, err := plane.extrude(area, 0)
			require.NoError(t, err)
			positions := face.Float3Attribute(modeling.PositionAttribute)
			vertices := make([]vector3.Float64, positions.Len())
			for i := range vertices {
				vertices[i] = positions.At(i)
			}

			for _, line := range lines {
				promoted := plane.promote(line, 0)
				assert.Contains(t, vertices, promoted.GetStartPoint())
				assert.Contains(t, vertices, promoted.GetEndPoint())
			}
		})
	}
}

func TestLinesFromPoints2DNodeTakesManyPaths(t *testing.T) {
	a, b, c := vector2.New(0., 0.), vector2.New(1., 0.), vector2.New(1., 1.)
	paths := [][]vector2.Float64{{a, b, c}, {c, b}}

	nodetest.NewTestCase(
		"closed strips per path",
		nodetest.NewNode(geometry.LinesFromPoints2DNode{
			Points: nodes.ConstOutput[[][]vector2.Float64]{Val: paths},
			Closed: nodes.ConstOutput[bool]{Val: true},
		}),
		nodetest.AssertOutput("Line Strips", [][]geometry.Line2D{
			geometry.LineStripsFromPoints2D(paths[0], true),
			geometry.LineStripsFromPoints2D(paths[1], true),
		}),
	).Run(t)
}

func TestLines2DTo3DNodeLiftsEachListOntoItsOwnOffset(t *testing.T) {
	line := geometry.NewLine2D(vector2.New(1., 2.), vector2.New(3., 4.))

	nodetest.NewTestCase(
		"one offset per list",
		nodetest.NewNode(geometry.Lines2DTo3DNode{
			Lines:  nodes.ConstOutput[[][]geometry.Line2D]{Val: [][]geometry.Line2D{{line}, {line}}},
			Offset: nodes.ConstOutput[[]float64]{Val: []float64{0, 1}},
		}),
		nodetest.AssertOutput("Z", [][]geometry.Line3D{{line.PlaneZ(0)}, {line.PlaneZ(1)}}),
		nodetest.AssertOutput("Y", [][]geometry.Line3D{{line.PlaneY(0)}, {line.PlaneY(1)}}),
	).Run(t)
}
