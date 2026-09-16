package csg

import (
	"math"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// The axis a face faces most squarely. Flattening by deleting that coordinate
// keeps every remaining bit of the input, where projecting onto a basis built
// from the normal rounds all of them.
func dominantAxis(normal vector3.Float64) int {
	x, y, z := math.Abs(normal.X()), math.Abs(normal.Y()), math.Abs(normal.Z())
	switch {
	case x >= y && x >= z:
		return 0
	case y >= z:
		return 1
	}
	return 2
}

// dropAxis flattens by deleting one coordinate. The remaining pair is taken
// in cyclic order so the projection's handedness only depends on the sign of
// the deleted component, and no arithmetic touches the numbers that survive.
func dropAxis(point vector3.Float64, axis int) vector2.Float64 {
	switch axis {
	case 0:
		return vector2.New(point.Y(), point.Z())
	case 1:
		return vector2.New(point.Z(), point.X())
	}
	return vector2.New(point.X(), point.Y())
}

// liftOntoPlane recovers the deleted coordinate from the face's plane. Only
// points the triangulator invented need it; everything else still has the
// coordinates it arrived with.
//
// The axis deleted was the one the face faces most squarely, so the divisor
// here is the largest component of the normal and never small.
func liftOntoPlane(flat vector2.Float64, axis int, f face) vector3.Float64 {
	normal := f.normal
	offset := normal.Dot(f.verts[0])

	switch axis {
	case 0:
		y, z := flat.X(), flat.Y()
		return vector3.New((offset-normal.Y()*y-normal.Z()*z)/normal.X(), y, z)
	case 1:
		z, x := flat.X(), flat.Y()
		return vector3.New(x, (offset-normal.Z()*z-normal.X()*x)/normal.Y(), z)
	}
	x, y := flat.X(), flat.Y()
	return vector3.New(x, y, (offset-normal.X()*x-normal.Y()*y)/normal.Z())
}
