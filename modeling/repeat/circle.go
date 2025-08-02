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

func Circle(times int, radius float64) []trs.TRS {
	angleIncrement := (1.0 / float64(times)) * 2.0 * math.Pi

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
	Radius nodes.Output[float64]
	Times  nodes.Output[int]
}

func (r CircleNode) Out(out *nodes.StructOutput[[]trs.TRS]) {
	out.Set(Circle(
		max(nodes.TryGetOutputValue(out, r.Times, 1), 0),
		nodes.TryGetOutputValue(out, r.Radius, 0.),
	))
}
