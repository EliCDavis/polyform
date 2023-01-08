package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector"
)

func TestPointInShape(t *testing.T) {
	shape := geometry.Shape([]vector.Vector2{
		vector.NewVector2(0, 0),
		vector.NewVector2(0, 1),
		vector.NewVector2(1, 1),
		vector.NewVector2(1, 0),
	})

	point := vector.NewVector2(0.5, 0.5)
	if shape.IsInside(point) == false {
		t.Error("Point should be inside shape")
	}
}

func TestGetPointtInShape(t *testing.T) {
	shape := geometry.Shape([]vector.Vector2{
		vector.NewVector2(0, 0),
		vector.NewVector2(0, 1),
		vector.NewVector2(1, 1),
		vector.NewVector2(1, 0),
	})

	point := shape.RandomPointInShape()
	if shape.IsInside(point) == false {
		t.Error("Point should be inside shape")
	}
}

func TestSplit(t *testing.T) {
	shape := geometry.Shape([]vector.Vector2{
		vector.NewVector2(0, 0),
		vector.NewVector2(0, 1),
		vector.NewVector2(1, 1),
		vector.NewVector2(1, 0),
	})

	left, right := shape.Split(.5)

	if len(left) != 1 {
		t.Errorf("Split should have returned 1 left shape! Instead returned %d", len(left))
	}

	if len(right) != 1 {
		t.Errorf("Split should have returned 1 left shape! Instead returned %d", len(right))
	}

	left, right = shape.Split(1.1)

	if len(left) != 1 {
		t.Errorf("Split should have returned the same shape! Instead returned %d new shapes", len(left))
	}

	if len(right) != 0 {
		t.Errorf("Split should have returned nothing! Instead returned %d new shapes", len(right))
	}

	left, right = shape.Split(-1)

	if len(right) != 1 {
		t.Errorf("Split should have returned the same shape! Instead returned %d new shapes", len(left))
	}

	if len(left) != 0 {
		t.Errorf("Split should have returned nothing! Instead returned %d new shapes", len(right))
	}

	// l 2 points, r 1 point
	shape = geometry.Shape([]vector.Vector2{
		vector.NewVector2(0, 0),   // l
		vector.NewVector2(.7, .5), // r
		vector.NewVector2(0, 1),   // l
		vector.NewVector2(.3, 1),  // l
		vector.NewVector2(.9, .5), // r
		vector.NewVector2(.3, 0),  // l
	})

	left, right = shape.Split(.5)

	if len(left) != 2 {
		t.Errorf("Split should have returned 2 left shapes! Instead returned %d", len(left))
	}

	if len(right) != 1 {
		t.Errorf("Split should have returned 1 right shape! Instead returned %d", len(right))
	}
}
