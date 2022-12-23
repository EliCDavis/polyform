package repeat

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

func Circle(in modeling.Mesh, times int, radius float64) modeling.Mesh {
	angleIncrement := (1.0 / float64(times)) * 2.0 * math.Pi

	final := modeling.EmptyMesh()

	for i := 0; i < times; i++ {
		angle := angleIncrement * float64(i)

		pos := vector.NewVector3(math.Cos(angle), 0, math.Sin(angle)).MultByConstant(radius)
		rot := modeling.UnitQuaternionFromTheta(-angle, vector.Vector3Up())

		final = final.Append(in.Rotate(rot).Translate(pos))
	}

	return final
}
