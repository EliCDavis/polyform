package extrude_test

import (
	"testing"

	"github.com/EliCDavis/mesh/extrude"
	"github.com/EliCDavis/vector"
	"github.com/stretchr/testify/assert"
)

func TestSimplePolygonSpike(t *testing.T) {
	extrusionPoints := []extrude.ExtrusionPoint{
		{
			Point:       vector.Vector3Zero(),
			Thickness:   1,
			UvThickness: 1,
			UvPoint:     vector.NewVector2(0.5, 0),
		},
		{
			Point:       vector.Vector3Up(),
			Thickness:   0,
			UvThickness: 1,
			UvPoint:     vector.NewVector2(0.5, 1),
		},
	}
	// ACT ====================================================================
	m := extrude.Polygon(3, extrusionPoints)
	view := m.View()

	// ASSERT =================================================================
	assert.Len(t, view.Normals, 8)
	if assert.Len(t, view.Vertices, 8) {
		assert.Equal(t, vector.Vector3Up(), view.Vertices[7])
	}
}
