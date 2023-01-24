package marching

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

func Box(pos vector3.Float64, size vector3.Float64, strength float64) Field {
	domain := modeling.NewAABB(
		pos,
		size,
	)
	domain.Expand(strength)
	return Field{
		Domain: domain,
		Float1Functions: map[string]sample.Vec3ToFloat{
			modeling.PositionAttribute: sdf.Box(pos, size).Scale(strength),
		},
	}
}
