package extrude_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestSimplePolygonSpike(t *testing.T) {
	extrusionPoints := []extrude.ExtrusionPoint{
		{
			Point:     vector3.Zero[float64](),
			Thickness: 1,
			UV: &extrude.ExtrusionPointUV{
				Thickness: 1,
				Point:     vector2.New(0.5, 0.),
			},
		},
		{
			Point:     vector3.Up[float64](),
			Thickness: 0,
			UV: &extrude.ExtrusionPointUV{
				Thickness: 1,
				Point:     vector2.New(0.5, 1.),
			},
		},
	}
	// ACT ====================================================================
	m := extrude.Polygon(3, extrusionPoints)

	// ASSERT =================================================================
	assert.Equal(t, 8, m.Float3Attribute(modeling.NormalAttribute).Len())
	if assert.Equal(t, 8, m.Float3Attribute(modeling.PositionAttribute).Len()) {
		assert.Equal(t, vector3.Up[float64](), m.Float3Attribute(modeling.PositionAttribute).At(7))
	}
}
