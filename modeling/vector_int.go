package modeling

import (
	"math"

	"github.com/EliCDavis/vector/vector3"
)

func Vector3ToInt(v vector3.Float64, power int) vector3.Int {
	newPower := math.Pow10(power)
	return vector3.New(
		int(math.Round(v.X()*newPower)),
		int(math.Round(v.Y()*newPower)),
		int(math.Round(v.Z()*newPower)),
	)
}
