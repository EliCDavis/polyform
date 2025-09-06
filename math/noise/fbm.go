package noise

import (
	"math"

	vec "github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

const (
	tau = math.Pi * 2
	eps = 1e-9
)

func fract(v float64) float64 { return v - math.Floor(v) }

func mod2(a, b vec.Float64) vec.Float64 {
	return vec.New(math.Mod(a.X(), b.X()), math.Mod(a.Y(), b.Y()))
}

func mix(a, b, t float64) float64 { return a*(1.0-t) + b*t }

// Quintic fade used by Perlin/simplex
func fade2(f vec.Float64) vec.Float64 {
	part := f.Scale(6.0).Sub(vec.New(15.0, 15.0))
	part = f.MultByVector(part)
	part = part.Add(vec.New(10.0, 10.0))

	return f.MultByVector(f).MultByVector(f).MultByVector(part)
}

// ---------- hash / rand helpers (GLSL-style) ----------
// Common 2D-to-N hash (not cryptographic), tuned to look like shader hashes.
func hash2(v vec.Float64, a, b, k float64) float64 {
	return fract(math.Sin(v.X()*a+v.Y()*b) * k)
}

func randf(v vec.Float64) float64 {
	return hash2(v, 12.9898, 78.233, 43758.5453123)
}

func rand2(v vec.Float64) vec.Float64 {
	// Use different coefficients to decorrelate components
	rx := hash2(v, 127.1, 311.7, 43758.5453123)
	ry := hash2(v, 269.5, 183.3, 43758.5453123)
	return vec.New(rx, ry)
}

func rand3(v vec.Float64) vector3.Float64 {
	rx := hash2(v, 127.1, 311.7, 43758.5453123)
	ry := hash2(v, 269.5, 183.3, 43758.5453123)
	rz := hash2(v, 419.2, 371.9, 43758.5453123)
	return vector3.New(rx, ry, rz)
}

// ---------- Value Noise ----------
func value(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := coord.Fract()

	p00 := randf(mod2(o, size))
	p01 := randf(mod2(o.Add(vec.New(0.0, 1.0)), size))
	p10 := randf(mod2(o.Add(vec.New(1.0, 0.0)), size))
	p11 := randf(mod2(o.Add(vec.New(1.0, 1.0)), size))

	p00 = math.Sin((p00+offset)*tau)/2.0 + 0.5
	p01 = math.Sin((p01+offset)*tau)/2.0 + 0.5
	p10 = math.Sin((p10+offset)*tau)/2.0 + 0.5
	p11 = math.Sin((p11+offset)*tau)/2.0 + 0.5

	t := fade2(f)

	return mix(
		mix(p00, p10, t.X()),
		mix(p01, p11, t.X()),
		t.Y(),
	)
}

func Value(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, val, scale := 0.0, 0.0, 1.0
	sz := size
	for i := range octaves {
		noise := value(coord.MultByVector(sz), sz, offset, seed+float64(i))
		for range folds {
			noise = math.Abs(2.0*noise - 1.0)
		}
		val += noise * scale
		norm += scale
		sz = sz.Scale(2.0)
		scale *= persistence
	}
	return val / max(norm, eps)
}

// ---------- Perlin Noise ----------
func perlin(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)

	a00 := randf(mod2(o, size))*tau + offset*tau
	a01 := randf(mod2(o.Add(vec.New(0.0, 1.0)), size))*tau + offset*tau
	a10 := randf(mod2(o.Add(vec.New(1.0, 0.0)), size))*tau + offset*tau
	a11 := randf(mod2(o.Add(vec.New(1.0, 1.0)), size))*tau + offset*tau

	v00 := vec.New(math.Cos(a00), math.Sin(a00))
	v01 := vec.New(math.Cos(a01), math.Sin(a01))
	v10 := vec.New(math.Cos(a10), math.Sin(a10))
	v11 := vec.New(math.Cos(a11), math.Sin(a11))

	f := coord.Fract()
	p00 := v00.Dot(f)
	p01 := v01.Dot(f.Sub(vec.New(0.0, 1.0)))
	p10 := v10.Dot(f.Sub(vec.New(1.0, 0.0)))
	p11 := v11.Dot(f.Sub(vec.New(1.0, 1.0)))

	t := fade2(f)

	return 0.5 + mix(
		mix(p00, p10, t.X()),
		mix(p01, p11, t.X()),
		t.Y(),
	)
}

func Perlin(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := range octaves {
		noise := perlin(coord.MultByVector(sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = sz.Scale(2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

func perlinAbs(coord, size vec.Float64, offset, seed float64) float64 {
	return math.Abs(2.0*perlin(coord, size, offset, seed) - 1.0)
}

func PerlinAbs(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := perlinAbs(coord.MultByVector(sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = sz.Scale(2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

// ---------- Helpers for Simplex ----------
func mod289(x float64) float64 {
	return x - math.Floor(x*(1.0/289.0))*289.0
}

func permute(x float64) float64 {
	return mod289(((x * 34.0) + 1.0) * x)
}

func rgrad2(p vec.Float64, rot, seed float64) vec.Float64 {
	u := permute(permute(p.X())+p.Y())*0.0243902439 + rot // /41
	u = fract(u+seed) * tau
	return vec.New(math.Cos(u), math.Sin(u))
}

// ---------- simplex Noise (tiling-aware as in shader) ----------
func simplex(coord, size vec.Float64, offset, seed float64) float64 {
	// Make it tile by doubling
	coord = coord.Scale(2.0)
	coord = coord.Add(rand2(vec.New(seed, 1.0-seed)).Add(size))
	size = size.Scale(2.0)
	coord = coord.Add(vec.New(0.0, 0.001)) // epsilon shift

	uv := vec.New(coord.X()+coord.Y()*0.5, coord.Y())
	i0 := uv.Floor()
	f0 := uv.Fract()

	var i1 vec.Float64
	if f0.X() > f0.Y() {
		i1 = vec.New(1.0, 0.0)
	} else {
		i1 = vec.New(0.0, 1.0)
	}

	p0 := vec.New(i0.X()-i0.Y()*0.5, i0.Y())
	p1 := vec.New(p0.X()+i1.X()-i1.Y()*0.5, p0.Y()+i1.Y())
	p2 := vec.New(p0.X()+0.5, p0.Y()+1.0)

	// i1sum := add2(i0, i1)
	// i2 := add2(i0, vec.New(1.0, 1.0))

	d0 := coord.Sub(p0)
	d1 := coord.Sub(p1)
	d2 := coord.Sub(p2)

	xw := [3]float64{math.Mod(p0.X(), size.X()), math.Mod(p1.X(), size.X()), math.Mod(p2.X(), size.X())}
	yw := [3]float64{math.Mod(p0.Y(), size.Y()), math.Mod(p1.Y(), size.Y()), math.Mod(p2.Y(), size.Y())}
	iuw := [3]float64{xw[0] + 0.5*yw[0], xw[1] + 0.5*yw[1], xw[2] + 0.5*yw[2]}
	ivw := [3]float64{yw[0], yw[1], yw[2]}

	g0 := rgrad2(vec.New(iuw[0], ivw[0]), offset, seed)
	g1 := rgrad2(vec.New(iuw[1], ivw[1]), offset, seed)
	g2 := rgrad2(vec.New(iuw[2], ivw[2]), offset, seed)

	w0 := g0.Dot(d0)
	w1 := g1.Dot(d1)
	w2 := g2.Dot(d2)

	t0 := 0.8 - (d0.X()*d0.X() + d0.Y()*d0.Y())
	t1 := 0.8 - (d1.X()*d1.X() + d1.Y()*d1.Y())
	t2 := 0.8 - (d2.X()*d2.X() + d2.Y()*d2.Y())

	if t0 < 0 {
		t0 = 0
	}
	if t1 < 0 {
		t1 = 0
	}
	if t2 < 0 {
		t2 = 0
	}

	t0 *= t0
	t0 *= t0
	t1 *= t1
	t1 *= t1
	t2 *= t2
	t2 *= t2

	n := t0*w0 + t1*w1 + t2*w2
	return 0.5 + 5.5*n
}

func Simplex(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := range octaves {
		noise := simplex(coord.MultByVector(sz), sz, offset, seed+float64(i))
		for range folds {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = sz.Scale(2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

// ---------- Cellular / Worley variants ----------
func cellular(coord, size vec.Float64, offset, seed float64) float64 {
	// _ = coord.Fract() // f in shader; not needed explicitly aside from diff computation
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := coord.Fract()

	minDist := 2.0
	for ix := -1; ix <= 1; ix++ {
		for iy := -1; iy <= 1; iy++ {
			neighbor := vec.New(float64(ix), float64(iy))
			node := rand2(mod2(o.Add(neighbor), size))
			node = vec.New(0.5, 0.5).Add(vec.New(math.Sin(offset*tau+tau*node.X()), math.Sin(offset*tau+tau*node.Y())).Scale(0.25))
			diff := neighbor.Add(node).Sub(f)
			dist := diff.Length()
			if dist < minDist {
				minDist = dist
			}
		}
	}
	return minDist
}

func Cellular(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := range octaves {
		noise := cellular(coord.MultByVector(sz), sz, offset, seed+float64(i))
		for range folds {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = sz.Scale(2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

func cellular2(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := coord.Fract()

	ot := offset * tau

	min1, min2 := 2.0, 2.0
	for ix := -1; ix <= 1; ix++ {
		for iy := -1; iy <= 1; iy++ {
			neighbor := vec.New(float64(ix), float64(iy))
			node := rand2(mod2(o.Add(neighbor), size))
			node = vec.New(0.5, 0.5).Add(vec.New(math.Sin(ot+tau*node.X()), math.Sin(ot+tau*node.Y())).Scale(0.25))
			diff := neighbor.Add(node).Sub(f)
			dist := diff.Length()
			if dist < min1 {
				min2 = min1
				min1 = dist
			} else if dist < min2 {
				min2 = dist
			}
		}
	}
	return min2 - min1
}

func Cellular2(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := cellular2(coord.MultByVector(sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = sz.Scale(2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

func cellular3(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := coord.Fract()

	minDist := 2.0
	for ix := -1; ix <= 1; ix++ {
		for iy := -1; iy <= 1; iy++ {
			neighbor := vec.New(float64(ix), float64(iy))
			node := rand2(mod2(o.Add(neighbor), size))
			node = vec.New(0.5, 0.5).Add(vec.New(math.Sin(offset*tau+tau*node.X()), math.Sin(offset*tau+tau*node.Y())).Scale(0.25))
			diff := neighbor.Add(node).Sub(f)
			// Manhattan distance
			dist := math.Abs(diff.X()) + math.Abs(diff.Y())
			if dist < minDist {
				minDist = dist
			}
		}
	}
	return minDist
}

func Cellular3(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := cellular3(coord.MultByVector(sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = sz.Scale(2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

func cellular4(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed)).Add(size))
	f := coord.Fract()

	min1, min2 := 2.0, 2.0
	for ix := -1; ix <= 1; ix++ {
		for iy := -1; iy <= 1; iy++ {
			neighbor := vec.New(float64(ix), float64(iy))
			node := rand2(mod2(o.Add(neighbor), size))
			node = vec.New(0.5, 0.5).Add(vec.New(math.Sin(offset*tau+tau*node.X()), math.Sin(offset*tau+tau*node.Y())).Scale(0.25))
			diff := neighbor.Add(node).Sub(f)
			dist := math.Abs(diff.X()) + math.Abs(diff.Y()) // Manhattan
			if dist < min1 {
				min2 = min1
				min1 = dist
			} else if dist < min2 {
				min2 = dist
			}
		}
	}
	return min2 - min1
}

func Cellular4(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := cellular4(coord.MultByVector(sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = sz.Scale(2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

func cellular5(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := coord.Fract()

	minDist := 2.0
	for ix := -1; ix <= 1; ix++ {
		for iy := -1; iy <= 1; iy++ {
			neighbor := vec.New(float64(ix), float64(iy))
			node := rand2(mod2(o.Add(neighbor), size))
			node = vec.New(0.5, 0.5).Add(vec.New(math.Sin(offset*tau+tau*node.X()), math.Sin(offset*tau+tau*node.Y())).Scale(0.5))
			diff := neighbor.Add(node).Sub(f)
			// Chebyshev distance
			ax := math.Abs(diff.X())
			ay := math.Abs(diff.Y())
			dist := ax
			if ay > dist {
				dist = ay
			}
			if dist < minDist {
				minDist = dist
			}
		}
	}
	return minDist
}

func Cellular5(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := cellular5(coord.MultByVector(sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = sz.Scale(2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

func cellular6(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := coord.Fract()

	min1, min2 := 2.0, 2.0
	for ix := -1; ix <= 1; ix++ {
		for iy := -1; iy <= 1; iy++ {
			neighbor := vec.New(float64(ix), float64(iy))
			node := rand2(mod2(o.Add(neighbor), size))
			node = vec.New(0.5, 0.5).Add(vec.New(math.Sin(offset*tau+tau*node.X()), math.Sin(offset*tau+tau*node.Y())).Scale(0.25))
			diff := neighbor.Add(node).Sub(f)
			ax := math.Abs(diff.X())
			ay := math.Abs(diff.Y())
			dist := ax
			if ay > dist {
				dist = ay
			} // Chebyshev
			if dist < min1 {
				min2 = min1
				min1 = dist
			} else if dist < min2 {
				min2 = dist
			}
		}
	}
	return min2 - min1
}

func Cellular6(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := cellular6(coord.MultByVector(sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = sz.Scale(2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

// ---------- Voronoise (Inigo Quilez MIT, adapted) ----------
func voronoise(coord, size vec.Float64, offset, seed float64) float64 {
	i := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := coord.Fract()

	aX, aY := 0.0, 0.0

	for yy := -2; yy <= 2; yy++ {
		for xx := -2; xx <= 2; xx++ {
			g := vec.New(float64(xx), float64(yy))
			o := rand3(mod2(i.Add(g), size).Add(vec.New(seed, seed)))
			ox := o.X()
			oy := o.Y()
			ox += 0.25 * math.Sin(offset*tau+tau*ox)
			oy += 0.25 * math.Sin(offset*tau+tau*oy)
			d := g.Sub(f).Sub(vec.New(-ox, -oy))                         // g - f + o.xy
			w := math.Pow(1.0-math.Min(1.0, d.Length()/math.Sqrt2), 1.0) // smoothstep approx
			aX += o.Z() * w
			aY += w
		}
	}
	return aX / max(aY, eps)
}

func Voronoise(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := range octaves {
		noise := voronoise(coord.MultByVector(sz), sz, offset, seed+float64(i))
		for range folds {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = sz.Scale(2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}
