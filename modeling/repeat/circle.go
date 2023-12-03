package repeat

import (
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
)

func CirclePoints(count int, radius float64) []vector3.Float64 {
	angleIncrement := (1.0 / float64(count)) * 2.0 * math.Pi
	final := make([]vector3.Float64, count)

	for i := 0; i < count; i++ {
		angle := angleIncrement * float64(i)
		final[i] = vector3.New(math.Cos(angle)*radius, 0, math.Sin(angle)*radius)
	}

	return final
}

func Circle(in modeling.Mesh, times int, radius float64) modeling.Mesh {
	angleIncrement := (1.0 / float64(times)) * 2.0 * math.Pi

	final := modeling.EmptyMesh(in.Topology())

	for i := 0; i < times; i++ {
		angle := angleIncrement * float64(i)

		pos := vector3.New(math.Cos(angle), 0, math.Sin(angle)).Scale(radius)
		rot := quaternion.FromTheta(angle-(math.Pi/2), vector3.Down[float64]())

		final = final.Append(
			in.Rotate(rot).
				Transform(
					meshops.TranslateAttribute3DTransformer{
						Amount: pos,
					},
				),
		)
	}

	return final
}
