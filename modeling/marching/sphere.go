package marching

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

func Sphere(pos vector.Vector3, radius, strength float64) Field {
	domainRadius := strength * radius * 2
	return Field{
		Domain: modeling.NewAABB(
			pos,
			vector.NewVector3(domainRadius, domainRadius, domainRadius),
		),
		Float1Functions: map[string]sample.Vec3ToFloat{
			modeling.PositionAttribute: sdf.Sphere(pos, radius).Scale(strength),
		},
	}
}
