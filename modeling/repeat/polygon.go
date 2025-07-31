package repeat

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func Polygon(times, sides int, radius float64) []trs.TRS {
	if sides <= 0 {
		panic(fmt.Errorf("polygon can not have %d sides", sides))
	}

	if times <= 0 {
		return nil
	}

	sideIncrement := 1.0 / float64(sides)

	angleIncrement := sideIncrement * 2.0 * math.Pi
	polygonPoints := make([]vector3.Float64, sides)
	polygonRotations := make([]quaternion.Quaternion, sides)
	for i := range sides {
		angle := angleIncrement * float64(i)
		polygonPoints[i] = vector3.New(math.Cos(angle)*radius, 0, math.Sin(angle)*radius)
		polygonRotations[i] = quaternion.FromTheta((angle+(angleIncrement/2))-(math.Pi/2), vector3.Down[float64]())
	}
	polygonPoints = append(polygonPoints, polygonPoints[0])

	transforms := make([]trs.TRS, times)
	timeIncrement := 1. / float64(times)
	currentEnd := 1
	currentEndTime := sideIncrement
	for i := range times {
		time := (float64(i) * timeIncrement) + (timeIncrement / 2) // We add a half to make things more evenly spaced
		for currentEndTime < time {
			currentEndTime += sideIncrement
			currentEnd += 1
		}

		adjusted := (time - (currentEndTime - sideIncrement)) / sideIncrement

		pos := vector3.Lerp(polygonPoints[currentEnd-1], polygonPoints[currentEnd], adjusted)
		transforms[i] = trs.New(pos, polygonRotations[currentEnd-1], vector3.One[float64]())
	}

	return transforms
}

type polygonNode struct {
	Radius nodes.Output[float64]
	Sides  nodes.Output[int]
	Times  nodes.Output[int]
}

func (r polygonNode) Out() nodes.StructOutput[[]trs.TRS] {
	out := nodes.StructOutput[[]trs.TRS]{}
	out.Set(Polygon(
		max(nodes.TryGetOutputValue(&out, r.Times, 1), 0),
		max(nodes.TryGetOutputValue(&out, r.Sides, 1), 3),
		nodes.TryGetOutputValue(&out, r.Radius, 0.),
	))
	return out
}
