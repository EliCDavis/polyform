package rendering

import (
	"math"
	"math/rand"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector3"
)

type Camera struct {
	lowerLeftCorner    vector3.Float64
	horizontal         vector3.Float64
	vertical           vector3.Float64
	origin             vector3.Float64
	u, v, w            vector3.Float64
	lensRadius         float64
	timeStart, timeEnd float64
	background         sample.Vec3ToVec3
}

func NewCamera(
	vfov, aspectRatio, aperture, focusDist float64,
	origin, lookAt, up vector3.Float64,
	timeStart, timeEnd float64,
	background sample.Vec3ToVec3,
) Camera {
	theta := vfov * (math.Pi / 180.)
	h := math.Tan(theta / 2)
	viewportHeight := 2.0 * h
	viewportWidth := aspectRatio * viewportHeight

	w := origin.Sub(lookAt).Normalized()
	u := up.Cross(w).Normalized()
	v := w.Cross(u)

	horizontal := u.Scale(viewportWidth * focusDist)
	vertical := v.Scale(viewportHeight * focusDist)

	lowerLeftCorner := origin.
		Sub(horizontal.Scale(0.5)).
		Sub(vertical.Scale(0.5)).
		Sub(w.Scale(focusDist))

	return Camera{
		lowerLeftCorner: lowerLeftCorner,
		horizontal:      horizontal,
		vertical:        vertical,
		origin:          origin,
		lensRadius:      aperture / 2,
		u:               u,
		v:               v,
		w:               w,
		timeStart:       timeStart,
		timeEnd:         timeEnd,
		background:      background,
	}
}

func (c Camera) GetRay(s, t float64) TemporalRay {
	rd := randUnitDisk().Scale(c.lensRadius)
	offset := c.u.Scale(rd.X()).Add(c.v.Scale(rd.Y()))

	dir := c.lowerLeftCorner.
		Add(c.horizontal.Scale(s)).
		Add(c.vertical.Scale(t)).
		Sub(c.origin).
		Sub(offset)

	dif := c.timeEnd - c.timeStart

	return NewTemporalRay(
		c.origin.Add(offset),
		dir,
		c.timeStart+(rand.Float64()*dif),
	)
}

func randUnitDisk() vector3.Float64 {
	for {
		p := vector3.RandRange[float64](-1, 1).SetZ(0)
		if p.LengthSquared() < 1 {
			return p
		}
	}
}
