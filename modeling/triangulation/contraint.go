package triangulation

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
)

type Constraint struct {
	shape  geometry.Shape
	keepIn bool
}

func NewConstraint(shape []vector2.Float64) Constraint {
	if len(shape) < 3 {
		panic("constraint must contain 3 or more points to form a shape")
	}

	return Constraint{
		shape:  shape,
		keepIn: true,
	}
}

func (c Constraint) contains(points ...vector2.Float64) int {
	count := 0
	for _, p := range points {
		if c.shape.IsInside(p) {
			count += 1
		}
	}
	return count
}
