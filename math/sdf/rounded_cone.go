package sdf

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func dot2(v vector3.Float64) float64 {
	return v.Dot(v)
}

func sign(f float64) float64 {
	if f > 0 {
		return 1
	}

	if f < 0 {
		return -1
	}
	return 0
}

// https://iquilezles.org/articles/distfunctions/
// https://www.shadertoy.com/view/tdXGWr
// Round Cone - exact
func RoundedCone(a, b vector3.Float64, r1, r2 float64) sample.Vec3ToFloat {

	// sampling independent computations (only depend on shape)
	ba := b.Sub(a)
	l2 := ba.Dot(ba)
	rr := r1 - r2
	rrr := rr * rr
	signRRR := sign(rr) * rrr
	a2 := l2 - rrr
	il2 := 1.0 / l2

	return func(v vector3.Float64) float64 {
		// sampling dependant computations
		pa := v.Sub(a)
		y := pa.Dot(ba)
		z := y - l2
		x2 := dot2(pa.Scale(l2).Sub(ba.Scale(y)))
		y2 := y * y * l2
		z2 := z * z * l2

		// single square root!
		k := signRRR * x2
		if sign(z)*a2*z2 > k {
			return math.Sqrt(x2+z2)*il2 - r2
		}
		if sign(y)*a2*y2 < k {
			return math.Sqrt(x2+y2)*il2 - r1
		}
		return (math.Sqrt(x2*a2*il2)+y*rr)*il2 - r1
	}
}

type RoundedConeNode struct {
	A       nodes.Output[vector3.Float64]
	B       nodes.Output[vector3.Float64]
	Radius1 nodes.Output[float64]
	Radius2 nodes.Output[float64]
}

func (cn RoundedConeNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	out.Set(RoundedCone(
		nodes.TryGetOutputValue(out, cn.A, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, cn.B, vector3.Up[float64]()),
		nodes.TryGetOutputValue(out, cn.Radius1, 1),
		nodes.TryGetOutputValue(out, cn.Radius2, .1),
	))
}
