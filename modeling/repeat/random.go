package repeat

import (
	"math"
	"math/rand/v2"

	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func RandomPointInSphere(radius float64) vector3.Float64 {
	// Sample random direction (normalized Gaussian)
	x := rand.NormFloat64()
	y := rand.NormFloat64()
	z := rand.NormFloat64()
	len := math.Sqrt(x*x + y*y + z*z)

	// Uniform random radius
	u := rand.Float64()
	r := radius * math.Cbrt(u)

	return vector3.New(x, y, z).DivByConstant(len).Scale(r)
}

func RandomPointsInSphere(radius float64, count int) []vector3.Float64 {
	results := make([]vector3.Float64, count)
	for i := range results {
		results[i] = RandomPointInSphere(radius)
	}
	return results
}

type RandomPointsInSphereNode struct {
	Radius nodes.Output[float64] `description:"Radius of the sphere containing the random points"`
	Points nodes.Output[int]     `description:"number of points to generate"`
}

func (g RandomPointsInSphereNode) points(recorder nodes.ExecutionRecorder) []vector3.Float64 {
	points := nodes.TryGetOutputValue(recorder, g.Points, 1)
	if points < 0 {
		recorder.CaptureError(nodes.InvalidInputError{
			Input:   g.Points,
			Message: "point count can not be negative",
		})
		return nil
	}

	return RandomPointsInSphere(
		nodes.TryGetOutputValue(recorder, g.Radius, 0.5),
		points,
	)
}

func (g RandomPointsInSphereNode) TRS(out *nodes.StructOutput[[]trs.TRS]) {
	points := g.points(out)
	result := make([]trs.TRS, len(points))
	for i, v := range points {
		result[i] = trs.Position(v)
	}
	out.Set(result)
}

func (g RandomPointsInSphereNode) Vector3(out *nodes.StructOutput[[]vector3.Float64]) {
	out.Set(g.points(out))
}
