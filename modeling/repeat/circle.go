package repeat

import (
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func CirclePoints(count int, radius float64) []vector3.Float64 {
	angleIncrement := (1.0 / float64(count)) * 2.0 * math.Pi
	final := make([]vector3.Float64, count)

	for i := range count {
		angle := angleIncrement * float64(i)
		final[i] = vector3.New(math.Cos(angle)*radius, 0, math.Sin(angle)*radius)
	}

	return final
}

func Circle(times int, radius, revolutions float64) []trs.TRS {
	angleIncrement := (1.0 / float64(times)) * 2.0 * math.Pi * revolutions

	transforms := make([]trs.TRS, times)

	for i := range times {
		angle := angleIncrement * float64(i)

		pos := vector3.New(math.Cos(angle), 0, math.Sin(angle)).Scale(radius)
		rot := quaternion.FromTheta(angle-(math.Pi/2), vector3.Down[float64]())

		transforms[i] = trs.New(pos, rot, vector3.One[float64]())
	}

	return transforms
}

type CircleNode struct {
	Radius      nodes.Output[float64] `description:"Distance from the center to each point. Defaults to 0.5."`
	Revolutions nodes.Output[float64] `description:"How many full loops the points sweep through before reaching Times copies. 1 spaces them evenly around a single circle; other values spiral them around more than once. Defaults to 1."`
	Times       nodes.Output[int]     `description:"How many copies to place evenly around the circle. Defaults to 1."`
}

func (r CircleNode) Description() string {
	return "Produces Times transforms evenly spaced around a circle in the XZ plane, each rotated to face outward/tangent to the circle."
}

func (r CircleNode) Out(out *nodes.StructOutput[[]trs.TRS]) {
	out.Set(Circle(
		max(nodes.TryGetOutputValue(out, r.Times, 1), 0),
		nodes.TryGetOutputValue(out, r.Radius, 0.5),
		nodes.TryGetOutputValue(out, r.Revolutions, 1.),
	))
}
