package materials

import (
	"math/rand"
	"time"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Lambertian struct {
	tex rendering.Texture
	r   *rand.Rand
}

func NewLambertian(tex rendering.Texture) *Lambertian {
	return &Lambertian{
		tex: tex,
		r:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (l Lambertian) Scatter(in geometry.Ray, rec *rendering.HitRecord, attenuation *vector3.Float64, scattered *geometry.Ray) bool {
	scatterDir := rec.Normal.Add(vector3.RandNormal(l.r))
	if scatterDir.NearZero() {
		scatterDir = rec.Normal
	}
	*scattered = geometry.NewRay(rec.Point, scatterDir)
	*attenuation = l.tex.Value(rec.UV, rec.Point)
	return true
}

func (l Lambertian) Emitted(uv vector2.Float64, pont vector3.Float64) vector3.Float64 {
	return vector3.Zero[float64]()
}
