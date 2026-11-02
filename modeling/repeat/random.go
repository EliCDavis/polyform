package repeat

import (
	"math"
	"math/rand/v2"

	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func RandomPointInSphere(radius float64, rnd *rand.Rand) vector3.Float64 {
	// Sample random direction (normalized Gaussian)
	x := rnd.NormFloat64()
	y := rnd.NormFloat64()
	z := rnd.NormFloat64()
	len := math.Sqrt(x*x + y*y + z*z)

	// Uniform random radius
	u := rnd.Float64()
	r := radius * math.Cbrt(u)

	return vector3.New(x, y, z).DivByConstant(len).Scale(r)
}

func RandomPointsInSphere(radius float64, count int, rnd *rand.Rand) []vector3.Float64 {
	results := make([]vector3.Float64, count)
	for i := range results {
		results[i] = RandomPointInSphere(radius, rnd)
	}
	return results
}

type RandomPointsInSphereNode struct {
	Radius nodes.Output[float64] `description:"Radius of the sphere containing the random points." default:"0.5"`
	Points nodes.Output[int]     `description:"number of points to generate. Must not be negative." default:"1"`
	Seed   nodes.Output[int]     `description:"The same seed always produces the same scatter."`
}

func (g RandomPointsInSphereNode) Description() string {
	return "Scatters Points points uniformly at random inside a solid sphere of radius Radius, not just on its surface."
}

func (g RandomPointsInSphereNode) points(recorder nodes.ExecutionRecorder) []vector3.Float64 {
	points := nodes.GetOutputValue(recorder, g.Points)
	if points < 0 {
		recorder.CaptureError(nodes.InvalidInputError{
			Input:   g.Points,
			Message: "point count can not be negative",
		})
		return nil
	}

	return RandomPointsInSphere(
		nodes.GetOutputValue(recorder, g.Radius),
		points,
		chance.FromSeed(nodes.GetOutputValue(recorder, g.Seed)),
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
