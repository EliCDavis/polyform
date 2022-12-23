package repeat

import (
	"math"

	"github.com/EliCDavis/vector"
)

func Point(times int, radius float64) []vector.Vector3 {
	angleIncrement := (1.0 / float64(times)) * 2.0 * math.Pi

	final := make([]vector.Vector3, times)

	for i := 0; i < times; i++ {
		angle := angleIncrement * float64(i)

		final[i] = vector.NewVector3(math.Cos(angle), 0, math.Sin(angle)).MultByConstant(radius)

	}

	return final
}
