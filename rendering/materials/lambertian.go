package materials

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/vector/vector3"
)

type Lambertian struct {
	albedo vector3.Float64
}

func NewLambertian(color vector3.Float64) *Lambertian {
	return &Lambertian{color}
}

func (l Lambertian) Scatter(in geometry.Ray, rec *rendering.HitRecord, attenuation *vector3.Float64, scattered *geometry.Ray) bool {
	scatterDir := rec.Normal.Add(vector3.RandNormal())
	if scatterDir.NearZero() {
		scatterDir = rec.Normal
	}
	*scattered = geometry.NewRay(rec.Point, scatterDir)
	*attenuation = l.albedo
	return true
}
