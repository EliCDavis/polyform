package noise

import (
	"math"

	vec "github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

const (
	tau   = 6.28318530718
	eps   = 1e-9
	root2 = 1.41421356237
)

// ---------- small vector helpers (component-wise) ----------

func x(v vec.Float64) float64 { return v.X() }
func y(v vec.Float64) float64 { return v.Y() }

func mul2(a, b vec.Float64) vec.Float64          { return vec.New(x(a)*x(b), y(a)*y(b)) }
func mul2s(a vec.Float64, s float64) vec.Float64 { return vec.New(x(a)*s, y(a)*s) }

func fract(v float64) float64          { return v - math.Floor(v) }
func fract2(a vec.Float64) vec.Float64 { return vec.New(fract(x(a)), fract(y(a))) }

func mod2(a, b vec.Float64) vec.Float64 {
	return vec.New(math.Mod(x(a), x(b)), math.Mod(y(a), y(b)))
}

func len2(a vec.Float64) float64 { return math.Hypot(x(a), y(a)) }

func mix(a, b, t float64) float64 { return a*(1.0-t) + b*t }

// Quintic fade used by Perlin/simplex
func fade2(f vec.Float64) vec.Float64 {
	part := f.Scale(6.0).Sub(vec.New(15.0, 15.0))
	part = mul2(f, part)
	part = part.Add(vec.New(10.0, 10.0))

	return mul2(mul2(mul2(f, f), f), part)
}

// ---------- hash / rand helpers (GLSL-style) ----------
// Common 2D-to-N hash (not cryptographic), tuned to look like shader hashes.
func hash2(v vec.Float64, a, b, k float64) float64 {
	return fract(math.Sin(x(v)*a+y(v)*b) * k)
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
func ValueNoise2D(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := fract2(coord)

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
		mix(p00, p10, x(t)),
		mix(p01, p11, x(t)),
		y(t),
	)
}

func FBM2DValue(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := range octaves {
		noise := ValueNoise2D(mul2(coord, sz), sz, offset, seed+float64(i))
		for range folds {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = mul2s(sz, 2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

// ---------- Perlin Noise ----------
func PerlinNoise2D(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := fract2(coord)

	a00 := randf(mod2(o, size))*tau + offset*tau
	a01 := randf(mod2(o.Add(vec.New(0.0, 1.0)), size))*tau + offset*tau
	a10 := randf(mod2(o.Add(vec.New(1.0, 0.0)), size))*tau + offset*tau
	a11 := randf(mod2(o.Add(vec.New(1.0, 1.0)), size))*tau + offset*tau

	v00 := vec.New(math.Cos(a00), math.Sin(a00))
	v01 := vec.New(math.Cos(a01), math.Sin(a01))
	v10 := vec.New(math.Cos(a10), math.Sin(a10))
	v11 := vec.New(math.Cos(a11), math.Sin(a11))

	p00 := v00.Dot(f)
	p01 := v01.Dot(f.Sub(vec.New(0.0, 1.0)))
	p10 := v10.Dot(f.Sub(vec.New(1.0, 0.0)))
	p11 := v11.Dot(f.Sub(vec.New(1.0, 1.0)))

	t := fade2(f)

	return 0.5 + mix(
		mix(p00, p10, x(t)),
		mix(p01, p11, x(t)),
		y(t),
	)
}

func FBM2DPerlin(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := PerlinNoise2D(mul2(coord, sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = mul2s(sz, 2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

func PerlinAbsNoise2D(coord, size vec.Float64, offset, seed float64) float64 {
	return math.Abs(2.0*PerlinNoise2D(coord, size, offset, seed) - 1.0)
}

func FBM2DPerlinAbs(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := PerlinAbsNoise2D(mul2(coord, sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = mul2s(sz, 2.0)
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
	u := permute(permute(x(p))+y(p))*0.0243902439 + rot // /41
	u = fract(u+seed) * tau
	return vec.New(math.Cos(u), math.Sin(u))
}

// ---------- Simplex Noise (tiling-aware as in shader) ----------
func SimplexNoise2D(coord, size vec.Float64, offset, seed float64) float64 {
	// Make it tile by doubling
	coord = mul2s(coord, 2.0)
	coord = coord.Add(rand2(vec.New(seed, 1.0-seed)).Add(size))
	size = mul2s(size, 2.0)
	coord = coord.Add(vec.New(0.0, 0.001)) // epsilon shift

	uv := vec.New(x(coord)+y(coord)*0.5, y(coord))
	i0 := uv.Floor()
	f0 := fract2(uv)

	var i1 vec.Float64
	if x(f0) > y(f0) {
		i1 = vec.New(1.0, 0.0)
	} else {
		i1 = vec.New(0.0, 1.0)
	}

	p0 := vec.New(x(i0)-y(i0)*0.5, y(i0))
	p1 := vec.New(x(p0)+x(i1)-y(i1)*0.5, y(p0)+y(i1))
	p2 := vec.New(x(p0)+0.5, y(p0)+1.0)

	// i1sum := add2(i0, i1)
	// i2 := add2(i0, vec.New(1.0, 1.0))

	d0 := coord.Sub(p0)
	d1 := coord.Sub(p1)
	d2 := coord.Sub(p2)

	xw := [3]float64{math.Mod(x(p0), x(size)), math.Mod(x(p1), x(size)), math.Mod(x(p2), x(size))}
	yw := [3]float64{math.Mod(y(p0), y(size)), math.Mod(y(p1), y(size)), math.Mod(y(p2), y(size))}
	iuw := [3]float64{xw[0] + 0.5*yw[0], xw[1] + 0.5*yw[1], xw[2] + 0.5*yw[2]}
	ivw := [3]float64{yw[0], yw[1], yw[2]}

	g0 := rgrad2(vec.New(iuw[0], ivw[0]), offset, seed)
	g1 := rgrad2(vec.New(iuw[1], ivw[1]), offset, seed)
	g2 := rgrad2(vec.New(iuw[2], ivw[2]), offset, seed)

	w0 := g0.Dot(d0)
	w1 := g1.Dot(d1)
	w2 := g2.Dot(d2)

	t0 := 0.8 - (x(d0)*x(d0) + y(d0)*y(d0))
	t1 := 0.8 - (x(d1)*x(d1) + y(d1)*y(d1))
	t2 := 0.8 - (x(d2)*x(d2) + y(d2)*y(d2))

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

func FBM2DSimplex(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := range octaves {
		noise := SimplexNoise2D(mul2(coord, sz), sz, offset, seed+float64(i))
		for range folds {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = mul2s(sz, 2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

// ---------- Cellular / Worley variants ----------
func CellularNoise2D(coord, size vec.Float64, offset, seed float64) float64 {
	_ = fract2(coord) // f in shader; not needed explicitly aside from diff computation
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := fract2(coord)

	minDist := 2.0
	for ix := -1; ix <= 1; ix++ {
		for iy := -1; iy <= 1; iy++ {
			neighbor := vec.New(float64(ix), float64(iy))
			node := rand2(mod2(o.Add(neighbor), size))
			node = vec.New(0.5, 0.5).Add(mul2s(vec.New(math.Sin(offset*tau+tau*x(node)), math.Sin(offset*tau+tau*y(node))), 0.25))
			diff := neighbor.Add(node).Sub(f)
			dist := len2(diff)
			if dist < minDist {
				minDist = dist
			}
		}
	}
	return minDist
}

func FBM2DCellular(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := CellularNoise2D(mul2(coord, sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = mul2s(sz, 2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

func Cellular2Noise2D(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := fract2(coord)

	min1, min2 := 2.0, 2.0
	for ix := -1; ix <= 1; ix++ {
		for iy := -1; iy <= 1; iy++ {
			neighbor := vec.New(float64(ix), float64(iy))
			node := rand2(mod2(o.Add(neighbor), size))
			node = vec.New(0.5, 0.5).Add(mul2s(vec.New(math.Sin(offset*tau+tau*x(node)), math.Sin(offset*tau+tau*y(node))), 0.25))
			diff := neighbor.Add(node).Sub(f)
			dist := len2(diff)
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

func FBM2DCellular2(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := Cellular2Noise2D(mul2(coord, sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = mul2s(sz, 2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

func Cellular3Noise2D(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := fract2(coord)

	minDist := 2.0
	for ix := -1; ix <= 1; ix++ {
		for iy := -1; iy <= 1; iy++ {
			neighbor := vec.New(float64(ix), float64(iy))
			node := rand2(mod2(o.Add(neighbor), size))
			node = vec.New(0.5, 0.5).Add(mul2s(vec.New(math.Sin(offset*tau+tau*x(node)), math.Sin(offset*tau+tau*y(node))), 0.25))
			diff := neighbor.Add(node).Sub(f)
			// Manhattan distance
			dist := math.Abs(x(diff)) + math.Abs(y(diff))
			if dist < minDist {
				minDist = dist
			}
		}
	}
	return minDist
}

func FBM2DCellular3(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := Cellular3Noise2D(mul2(coord, sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = mul2s(sz, 2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

func Cellular4Noise2D(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed)).Add(size))
	f := fract2(coord)

	min1, min2 := 2.0, 2.0
	for ix := -1; ix <= 1; ix++ {
		for iy := -1; iy <= 1; iy++ {
			neighbor := vec.New(float64(ix), float64(iy))
			node := rand2(mod2(o.Add(neighbor), size))
			node = vec.New(0.5, 0.5).Add(mul2s(vec.New(math.Sin(offset*tau+tau*x(node)), math.Sin(offset*tau+tau*y(node))), 0.25))
			diff := neighbor.Add(node).Sub(f)
			dist := math.Abs(x(diff)) + math.Abs(y(diff)) // Manhattan
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

func FBM2DCellular4(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := Cellular4Noise2D(mul2(coord, sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = mul2s(sz, 2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

func Cellular5Noise2D(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := fract2(coord)

	minDist := 2.0
	for ix := -1; ix <= 1; ix++ {
		for iy := -1; iy <= 1; iy++ {
			neighbor := vec.New(float64(ix), float64(iy))
			node := rand2(mod2(o.Add(neighbor), size))
			node = vec.New(0.5, 0.5).Add(mul2s(vec.New(math.Sin(offset*tau+tau*x(node)), math.Sin(offset*tau+tau*y(node))), 0.5))
			diff := neighbor.Add(node).Sub(f)
			// Chebyshev distance
			ax := math.Abs(x(diff))
			ay := math.Abs(y(diff))
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

func FBM2DCellular5(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := Cellular5Noise2D(mul2(coord, sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = mul2s(sz, 2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

func Cellular6Noise2D(coord, size vec.Float64, offset, seed float64) float64 {
	o := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := fract2(coord)

	min1, min2 := 2.0, 2.0
	for ix := -1; ix <= 1; ix++ {
		for iy := -1; iy <= 1; iy++ {
			neighbor := vec.New(float64(ix), float64(iy))
			node := rand2(mod2(o.Add(neighbor), size))
			node = vec.New(0.5, 0.5).Add(mul2s(vec.New(math.Sin(offset*tau+tau*x(node)), math.Sin(offset*tau+tau*y(node))), 0.25))
			diff := neighbor.Add(node).Sub(f)
			ax := math.Abs(x(diff))
			ay := math.Abs(y(diff))
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

func FBM2DCellular6(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := 0; i < octaves; i++ {
		noise := Cellular6Noise2D(mul2(coord, sz), sz, offset, seed+float64(i))
		for f := 0; f < folds; f++ {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = mul2s(sz, 2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}

// ---------- Voronoise (Inigo Quilez MIT, adapted) ----------
func VoronoiseNoise2D(coord, size vec.Float64, offset, seed float64) float64 {
	i := coord.Floor().Add(rand2(vec.New(seed, 1.0-seed))).Add(size)
	f := fract2(coord)

	aX, aY := 0.0, 0.0

	for yy := -2; yy <= 2; yy++ {
		for xx := -2; xx <= 2; xx++ {
			g := vec.New(float64(xx), float64(yy))
			o := rand3(mod2(i.Add(g), size).Add(vec.New(seed, seed)))
			ox := o.X()
			oy := o.Y()
			ox += 0.25 * math.Sin(offset*tau+tau*ox)
			oy += 0.25 * math.Sin(offset*tau+tau*oy)
			d := g.Sub(f).Sub(vec.New(-ox, -oy))                 // g - f + o.xy
			w := math.Pow(1.0-math.Min(1.0, len2(d)/root2), 1.0) // smoothstep approx
			aX += o.Z() * w
			aY += w
		}
	}
	return aX / max(aY, eps)
}

func FBM2DVoronoise(coord, size vec.Float64, folds, octaves int, persistence, offset, seed float64) float64 {
	norm, value, scale := 0.0, 0.0, 1.0
	sz := size
	for i := range octaves {
		noise := VoronoiseNoise2D(mul2(coord, sz), sz, offset, seed+float64(i))
		for range folds {
			noise = math.Abs(2.0*noise - 1.0)
		}
		value += noise * scale
		norm += scale
		sz = mul2s(sz, 2.0)
		scale *= persistence
	}
	return value / max(norm, eps)
}
