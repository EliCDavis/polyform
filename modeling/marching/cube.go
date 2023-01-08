package marching

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

func Box(pos vector.Vector3, size vector.Vector3, strength float64) Field {
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
