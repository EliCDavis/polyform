package materials

import (
	"math/rand"
	"time"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Metal struct {
	color vector3.Float64
	fuzz  float64
	r     *rand.Rand
}

func NewMetal(color vector3.Float64) Metal {
	return Metal{
		color: color,
		fuzz:  0,
		r:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func NewFuzzyMetal(color vector3.Float64, fuzz float64) Metal {
	return Metal{
		color: color,
		fuzz:  fuzz,
		r:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m Metal) Scatter(in geometry.Ray, rec *rendering.HitRecord, attenuation *vector3.Float64, scattered *geometry.Ray) bool {
	reflected := in.Direction().Normalized().Reflect(rec.Normal)
	*scattered = geometry.NewRay(rec.Point, reflected.Add(vector3.RandInUnitSphere(m.r).Scale(m.fuzz)))
	*attenuation = m.color
	return scattered.Direction().Dot(rec.Normal) > 0
}

func (m Metal) Emitted(uv vector2.Float64, pont vector3.Float64) vector3.Float64 {
	return vector3.Zero[float64]()
}
