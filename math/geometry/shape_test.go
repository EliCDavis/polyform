package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
)

func TestPointInShape(t *testing.T) {
	shape := geometry.Shape([]vector2.Float64{
		vector2.New(0., 0.),
		vector2.New(0., 1.),
		vector2.New(1., 1.),
		vector2.New(1., 0.),
	})

	point := vector2.New(0.5, 0.5)
	if shape.IsInside(point) == false {
		t.Error("Point should be inside shape")
	}
}

func TestGetPointtInShape(t *testing.T) {
	shape := geometry.Shape([]vector2.Float64{
		vector2.New(0., 0.),
		vector2.New(0., 1.),
		vector2.New(1., 1.),
		vector2.New(1., 0.),
	})

	point := shape.RandomPointInShape()
	if shape.IsInside(point) == false {
		t.Error("Point should be inside shape")
	}
}

func TestSplit(t *testing.T) {
	shape := geometry.Shape([]vector2.Float64{
		vector2.New(0., 0.),
		vector2.New(0., 1.),
		vector2.New(1., 1.),
		vector2.New(1., 0.),
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
	shape = geometry.Shape([]vector2.Float64{
		vector2.New(0., 0.), // l
		vector2.New(.7, .5), // r
		vector2.New(0., 1.), // l
		vector2.New(.3, 1.), // l
		vector2.New(.9, .5), // r
		vector2.New(.3, 0.), // l
	})

	left, right = shape.Split(.5)

	if len(left) != 2 {
		t.Errorf("Split should have returned 2 left shapes! Instead returned %d", len(left))
	}

	if len(right) != 1 {
		t.Errorf("Split should have returned 1 right shape! Instead returned %d", len(right))
	}
}
