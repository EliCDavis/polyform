package materials

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// barycentric
type Barycentric struct{}

func NewBarycentric() Barycentric {
	return Barycentric{}
}

func (l Barycentric) Scatter(in geometry.Ray, rec *rendering.HitRecord, attenuation *vector3.Float64, scattered *geometry.Ray) bool {
	scatterDir := rec.Normal
	*scattered = geometry.NewRay(rec.Point, scatterDir)
	*attenuation = rec.Float3Data["barycentric"]
	return true
}

func (l Barycentric) Emitted(uv vector2.Float64, pont vector3.Float64) vector3.Float64 {
	return vector3.Zero[float64]()
}
