package marching

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

func Sphere(pos vector3.Float64, radius, strength float64) Field {
	domainRadius := strength * radius * 2
	return Field{
		Domain: geometry.NewAABB(
			pos,
			vector3.New(domainRadius, domainRadius, domainRadius),
		),
		Float1Functions: map[string]sample.Vec3ToFloat{
			modeling.PositionAttribute: sdf.Sphere(pos, radius).Scale(strength),
		},
	}
}
