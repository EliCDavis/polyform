package extrude_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/curves"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircleRejectsWhatItCannotSweep(t *testing.T) {
	up := []vector3.Float64{vector3.New(0., 0., 0.), vector3.New(0., 1., 0.)}

	for name, circle := range map[string]extrude.Circle{
		"one point path":         {Resolution: 8, Radius: 1, Path: up[:1]},
		"two sides":              {Resolution: 2, Radius: 1, Path: up},
		"radii of another count": {Resolution: 8, Radii: []float64{1, 2, 3}, Path: up},
	} {
		t.Run(name, func(t *testing.T) {
			mesh, err := circle.Extrude()
			require.Error(t, err)
			assert.Zero(t, mesh.PrimitiveCount())
		})
	}
}

func TestCircleNodeGivesNothingForAOnePointPath(t *testing.T) {
	mesh := nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[extrude.CircleNode]{
		Data: extrude.CircleNode{
			Path: nodes.ConstOutput[[]vector3.Float64]{Val: []vector3.Float64{vector3.New(0., 0., 0.)}},
		},
	}, "Out").Value()

	assert.Zero(t, mesh.PrimitiveCount())
}

func TestCircleAlongAClosedSplineHasNoDoubledRing(t *testing.T) {
	spline := curves.CatmullRomSplineParameters{
		Points: []vector3.Float64{
			vector3.New(0., 0., 0.), vector3.New(4., 0., 0.), vector3.New(4., 0., 4.), vector3.New(0., 0., 4.),
		},
		Alpha:  0.5,
		Closed: true,
	}.Spline()

	mesh, err := extrude.CircleAlongSpline{
		CircleResolution: 8,
		Radius:           0.2,
		ClosePath:        true,
		Spline:           &spline,
		SplineResolution: 33,
		UVs:              &primitives.StripUVs{Start: vector2.New(0, 0.5), End: vector2.New(1, 0.5), Width: 1},
	}.Extrude()
	require.NoError(t, err)

	positions := mesh.Float3Attribute(modeling.PositionAttribute)
	for i := 0; i < mesh.PrimitiveCount(); i++ {
		tri := mesh.Tri(i)
		a, b, c := positions.At(tri.P1()), positions.At(tri.P2()), positions.At(tri.P3())
		require.Positivef(t, b.Sub(a).Cross(c.Sub(a)).Length(), "triangle %d has no area", i)
	}

	uvs := mesh.Float2Attribute(modeling.TexCoordAttribute)
	low, high := math.Inf(1), math.Inf(-1)
	for i := 0; i < uvs.Len(); i++ {
		low, high = min(low, uvs.At(i).X()), max(high, uvs.At(i).X())
	}
	assert.InDelta(t, 0, low, 1e-9)
	assert.InDelta(t, 1, high, 1e-9)
}

func TestLineWithoutUVsWritesFiniteOnes(t *testing.T) {
	mesh := extrude.Line([]extrude.LinePoint{
		{Point: vector3.New(0., 0., 0.), Up: vector3.Up[float64](), Width: 1},
		{Point: vector3.New(0., 0., 1.), Up: vector3.Up[float64](), Width: 1},
	})

	uvs := mesh.Float2Attribute(modeling.TexCoordAttribute)
	for i := 0; i < uvs.Len(); i++ {
		assert.Falsef(t, math.IsNaN(uvs.At(i).X()+uvs.At(i).Y()), "uv %d is NaN", i)
	}
}
