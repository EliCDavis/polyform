package sdf

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

const (
	raycastEpsilon  = 1e-5
	raycastMaxSteps = 256
	normalEpsilon   = 1e-4
)

func Raycast(field sample.Vec3ToFloat, origin, direction vector3.Float64, maxDistance float64) (float64, bool) {
	if field == nil || direction.Length() == 0 {
		return 0, false
	}
	dir := direction.Normalized()
	inside := field(origin) < 0

	travelled := 0.
	for range raycastMaxSteps {
		step := field(origin.Add(dir.Scale(travelled)))
		if inside {
			step = -step
		}
		if step < raycastEpsilon {
			return travelled, true
		}
		travelled += step
		if travelled > maxDistance {
			return 0, false
		}
	}
	return 0, false
}

func Normal(field sample.Vec3ToFloat, point vector3.Float64) vector3.Float64 {
	if field == nil {
		return vector3.Zero[float64]()
	}

	along := func(axis vector3.Float64) float64 {
		offset := axis.Scale(normalEpsilon)
		return field(point.Add(offset)) - field(point.Sub(offset))
	}

	gradient := vector3.New(
		along(vector3.Right[float64]()),
		along(vector3.Up[float64]()),
		along(vector3.Forward[float64]()),
	)
	if gradient.Length() == 0 {
		return vector3.Zero[float64]()
	}
	return gradient.Normalized()
}

type RaycastNode struct {
	Field       nodes.Output[sample.Vec3ToFloat]  `description:"The field to march against."`
	Origin      nodes.LiftedPort[vector3.Float64] `description:"Where the ray starts. Inside the shape is fine; the march then runs outward to the wall."`
	Direction   nodes.LiftedPort[vector3.Float64] `description:"Which way to travel, any length. Defaults to +Y."`
	MaxDistance nodes.Output[float64]             `description:"How far to travel before giving up. Defaults to 10."`
}

func (n RaycastNode) Description() string {
	return "Finds where a ray meets an SDF surface, and which way that surface faces."
}

func (n RaycastNode) Keywords() []string {
	return []string{"surface", "attach"}
}

type raycastHit struct {
	point    vector3.Float64
	normal   vector3.Float64
	distance float64
	hit      bool
}

func (n RaycastNode) march(field sample.Vec3ToFloat, maxDistance float64) func(vector3.Float64, vector3.Float64) raycastHit {
	return func(origin, direction vector3.Float64) raycastHit {
		if direction.Length() == 0 {
			return raycastHit{}
		}
		dir := direction.Normalized()
		distance, ok := Raycast(field, origin, dir, maxDistance)
		if !ok {
			return raycastHit{}
		}
		point := origin.Add(dir.Scale(distance))
		return raycastHit{
			point:    point,
			normal:   Normal(field, point),
			distance: distance,
			hit:      true,
		}
	}
}

func (n RaycastNode) Point(out *nodes.Lifted[vector3.Float64]) {
	if n.Field == nil {
		return
	}
	field := nodes.GetOutputValue(out, n.Field)
	march := n.march(field, nodes.TryGetOutputValue(out, n.MaxDistance, 10))
	nodes.Zip2(out, n.Origin, nodes.LiftedOr(n.Direction, vector3.Up[float64]()),
		func(origin, direction vector3.Float64) vector3.Float64 {
			return march(origin, direction).point
		})
}

func (n RaycastNode) Normal(out *nodes.Lifted[vector3.Float64]) {
	if n.Field == nil {
		return
	}
	field := nodes.GetOutputValue(out, n.Field)
	march := n.march(field, nodes.TryGetOutputValue(out, n.MaxDistance, 10))
	nodes.Zip2(out, n.Origin, nodes.LiftedOr(n.Direction, vector3.Up[float64]()),
		func(origin, direction vector3.Float64) vector3.Float64 {
			return march(origin, direction).normal
		})
}

func (n RaycastNode) Distance(out *nodes.Lifted[float64]) {
	if n.Field == nil {
		return
	}
	field := nodes.GetOutputValue(out, n.Field)
	march := n.march(field, nodes.TryGetOutputValue(out, n.MaxDistance, 10))
	nodes.Zip2(out, n.Origin, nodes.LiftedOr(n.Direction, vector3.Up[float64]()),
		func(origin, direction vector3.Float64) float64 {
			return march(origin, direction).distance
		})
}

func (n RaycastNode) Hit(out *nodes.Lifted[bool]) {
	if n.Field == nil {
		return
	}
	field := nodes.GetOutputValue(out, n.Field)
	march := n.march(field, nodes.TryGetOutputValue(out, n.MaxDistance, 10))
	nodes.Zip2(out, n.Origin, nodes.LiftedOr(n.Direction, vector3.Up[float64]()),
		func(origin, direction vector3.Float64) bool {
			return march(origin, direction).hit
		})
}
