package materials

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/vector/vector3"
)

type Metal struct {
	color vector3.Float64
	fuzz  float64
}

func NewMetal(color vector3.Float64) Metal {
	return Metal{color, 0}
}

func NewFuzzyMetal(color vector3.Float64, fuzz float64) Metal {
	return Metal{color, fuzz}
}

func (l Metal) Scatter(in geometry.Ray, rec *rendering.HitRecord, attenuation *vector3.Float64, scattered *geometry.Ray) bool {
	reflected := in.Direction().Normalized().Reflect(rec.Normal)
	*scattered = geometry.NewRay(rec.Point, reflected.Add(vector3.RandInUnitSphere().Scale(l.fuzz)))
	*attenuation = l.color
	return scattered.Direction().Dot(rec.Normal) > 0
}
