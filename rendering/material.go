package rendering

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Material interface {
	Scatter(in geometry.Ray, rec *HitRecord, attenuation *vector3.Float64, scattered *geometry.Ray) bool
	Emitted(uv vector2.Float64, pont vector3.Float64) vector3.Float64
}
