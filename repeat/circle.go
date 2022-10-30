package repeat

import (
	"math"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/vector"
)

func Circle(in mesh.Mesh, times int, radius float64) mesh.Mesh {
	angleIncrement := (1.0 / float64(times)) * 2.0 * math.Pi

	final := mesh.EmptyMesh()

	for i := 0; i < times; i++ {
		angle := angleIncrement * float64(i)

		pos := vector.NewVector3(math.Cos(angle)*radius, 0, math.Sin(angle)*radius)
		rot := mesh.UnitQuaternionFromTheta(angle, vector.Vector3Up())

		final = final.Append(in.Rotate(rot).Translate(pos))
	}

	return final
}
