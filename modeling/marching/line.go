package marching

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

func Line(start, end vector.Vector3, radius, strength float64) Field {
	dir := end.Sub(start).Normalized()
	bounds := modeling.NewAABB(start, vector.Vector3Zero())
	bounds.EncapsulatePoint(end)
	bounds.EncapsulatePoint(end.Add(dir.MultByConstant(radius + strength)))
	bounds.EncapsulatePoint(start.Sub(dir.MultByConstant(radius + strength)))
	return Field{
		Domain: bounds,
		Float1Functions: map[string]sample.Vec3ToFloat{
			modeling.PositionAttribute: sdf.Line(start, end, radius).Scale(strength),
		},
	}
}
