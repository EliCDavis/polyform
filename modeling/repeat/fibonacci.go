package repeat

import (
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func FibonacciSpherePoints(samples int, offsetRadius float64) []vector3.Float64 {
	points := make([]vector3.Float64, samples)
	phi := math.Pi * (3.0 - math.Sqrt(5.0)) // golden angle in radians

	for i := 0; i < samples; i++ {
		y := 1 - (float64(i)/float64(samples-1))*2. // y goes from 1 to -1
		radius := math.Sqrt(1 - y*y)                // radius at y

		theta := phi * float64(i) // golden angle increment

		x := math.Cos(theta) * radius
		z := math.Sin(theta) * radius

		points[i] = vector3.New(x, y, z).Scale(offsetRadius)
	}

	return points
}

func FibonacciSpiral(samples int, radius float64) []trs.TRS {

	points := make([]trs.TRS, samples)
	phi := math.Pi * (3.0 - math.Sqrt(5.0)) // golden angle in radians

	for i := 0; i < samples; i++ {
		theta := phi * float64(i) // golden angle increment

		d := math.Sqrt(float64(i) / float64(samples-1))
		x := math.Cos(theta) * radius * d
		z := math.Sin(theta) * radius * d

		p := vector3.New(x, 0, z)
		points[i] = trs.New(p, quaternion.Identity(), vector3.One[float64]())
	}

	return points
}

func FibonacciSphere(samples int, radius float64) []trs.TRS {
	points := FibonacciSpherePoints(samples, radius)
	transforms := make([]trs.TRS, len(points))
	for i, p := range points {
		transforms[i] = trs.New(
			p,
			quaternion.FromTheta(0, p.Normalized()),
			vector3.One[float64](),
		)
	}

	return transforms
}

type FibonacciSphereNode = nodes.Struct[[]trs.TRS, FibonacciSphereNodeData]

type FibonacciSphereNodeData struct {
	Count  nodes.NodeOutput[int]
	Radius nodes.NodeOutput[float64]
}

func (fpnd FibonacciSphereNodeData) Process() ([]trs.TRS, error) {
	radius := 1.
	count := 10

	if fpnd.Count != nil {
		count = fpnd.Count.Value()
	}

	if fpnd.Radius != nil {
		radius = fpnd.Radius.Value()
	}

	return FibonacciSphere(count, radius), nil
}

type FibonacciSpiralNode = nodes.Struct[[]trs.TRS, FibonacciSpiralNodeData]

type FibonacciSpiralNodeData struct {
	Count  nodes.NodeOutput[int]
	Radius nodes.NodeOutput[float64]
}

func (fpnd FibonacciSpiralNodeData) Process() ([]trs.TRS, error) {
	radius := 1.
	count := 10

	if fpnd.Count != nil {
		count = fpnd.Count.Value()
	}

	if fpnd.Radius != nil {
		radius = fpnd.Radius.Value()
	}

	return FibonacciSpiral(count, radius), nil
}
