package predicate

import (
	"math"
	"math/rand"
	"testing"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/require"
)

func sign(v float64) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}

// A few ulps either side of a value, or the value itself.
func nudge(rng *rand.Rand, v float64) float64 {
	for i := rng.Intn(5) - 2; i != 0; {
		if i > 0 {
			v = math.Nextafter(v, math.Inf(1))
			i--
		} else {
			v = math.Nextafter(v, math.Inf(-1))
			i++
		}
	}
	return v
}

func scaled(rng *rand.Rand, v float64) float64 {
	return math.Ldexp(v, rng.Intn(200)-100)
}

func TestOrient2DAgreesWithExactArithmetic(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	cases := map[string]func() (a, b, c vector2.Float64){
		"random": func() (a, b, c vector2.Float64) {
			return vector2.New(rng.Float64(), rng.Float64()), vector2.New(rng.Float64(), rng.Float64()), vector2.New(rng.Float64(), rng.Float64())
		},
		"grid": func() (a, b, c vector2.Float64) {
			p := func() vector2.Float64 { return vector2.New(float64(rng.Intn(7)), float64(rng.Intn(7))) }
			return p(), p(), p()
		},
		"near collinear": func() (a, b, c vector2.Float64) {
			a = vector2.New(rng.Float64()*2-1, rng.Float64()*2-1)
			b = vector2.New(rng.Float64()*2-1, rng.Float64()*2-1)
			along := a.Add(b.Sub(a).Scale(rng.Float64()*3 - 1))
			return a, b, vector2.New(nudge(rng, along.X()), nudge(rng, along.Y()))
		},
		"wide exponents": func() (a, b, c vector2.Float64) {
			a = vector2.New(rng.Float64()*2-1, rng.Float64()*2-1)
			b = vector2.New(rng.Float64()*2-1, rng.Float64()*2-1)
			along := a.Add(b.Sub(a).Scale(rng.Float64()))
			k := rng.Intn(200) - 100
			return vector2.New(math.Ldexp(a.X(), k), math.Ldexp(a.Y(), k)),
				vector2.New(math.Ldexp(b.X(), k), math.Ldexp(b.Y(), k)),
				vector2.New(nudge(rng, math.Ldexp(along.X(), k)), nudge(rng, math.Ldexp(along.Y(), k)))
		},
	}

	for name, next := range cases {
		disagreed, zeros := 0, 0
		for i := 0; i < 200_000; i++ {
			a, b, c := next()
			got := orient2D(a, b, c)
			want := orient2DExact(a, b, c)
			if sign(got) != sign(want) {
				disagreed++
				if disagreed < 4 {
					t.Errorf("%s: %v %v %v: adaptive %v exact %v", name, a, b, c, got, want)
				}
			}
			if want == 0 {
				zeros++
			}
		}
		require.Zerof(t, disagreed, "%s: %d disagreements", name, disagreed)
		if name == "grid" {
			require.Positive(t, zeros, "grid points give exactly collinear triples")
		}
	}
}

func TestOrient3DAgreesWithExactArithmetic(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	v := func() vector3.Float64 { return vector3.New(rng.Float64()*2-1, rng.Float64()*2-1, rng.Float64()*2-1) }
	grid := func() vector3.Float64 { return vector3.New(float64(rng.Intn(5)), float64(rng.Intn(5)), float64(rng.Intn(5))) }
	onPlane := func(a, b, c vector3.Float64) vector3.Float64 {
		p := a.Add(b.Sub(a).Scale(rng.Float64()*3 - 1)).Add(c.Sub(a).Scale(rng.Float64()*3 - 1))
		return vector3.New(nudge(rng, p.X()), nudge(rng, p.Y()), nudge(rng, p.Z()))
	}
	cases := map[string]func() (a, b, c, d vector3.Float64){
		"random":   func() (a, b, c, d vector3.Float64) { return v(), v(), v(), v() },
		"grid":     func() (a, b, c, d vector3.Float64) { return grid(), grid(), grid(), grid() },
		"on plane": func() (a, b, c, d vector3.Float64) { a, b, c = v(), v(), v(); return a, b, c, onPlane(a, b, c) },
		"wide exponents": func() (a, b, c, d vector3.Float64) {
			a, b, c = v(), v(), v()
			d = onPlane(a, b, c)
			k := rng.Intn(200) - 100
			s := func(p vector3.Float64) vector3.Float64 {
				return vector3.New(math.Ldexp(p.X(), k), math.Ldexp(p.Y(), k), math.Ldexp(p.Z(), k))
			}
			return s(a), s(b), s(c), s(d)
		},
	}

	for name, next := range cases {
		disagreed, zeros, fellThrough := 0, 0, 0
		for i := 0; i < 200_000; i++ {
			a, b, c, d := next()
			got, ok := orient3D(a, b, c, d)
			want := orient3DExact(a, b, c, d)
			if !ok {
				fellThrough++
				continue
			}
			if sign(got) != sign(want) {
				disagreed++
				if disagreed < 4 {
					t.Errorf("%s: %v %v %v %v: adaptive %v exact %v", name, a, b, c, d, got, want)
				}
			}
			if want == 0 {
				zeros++
			}
		}
		require.Zerof(t, disagreed, "%s: %d disagreements", name, disagreed)
		t.Logf("%s: %d exact zeros settled, %d fell through to big arithmetic", name, zeros, fellThrough)
		if name == "grid" {
			require.Positive(t, zeros, "grid points settle coplanar cases without big arithmetic")
			require.Zero(t, fellThrough, "grid points have no rounded differences to fall through on")
		}
	}
}

func TestInCircleAgreesWithExactArithmetic(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	v := func() vector2.Float64 { return vector2.New(rng.Float64()*2-1, rng.Float64()*2-1) }
	grid := func() vector2.Float64 { return vector2.New(float64(rng.Intn(9)), float64(rng.Intn(9))) }
	onCircle := func() (a, b, c, d vector2.Float64) {
		center := v()
		radius := rng.Float64() + 0.5
		at := func() vector2.Float64 {
			angle := rng.Float64() * 2 * math.Pi
			p := center.Add(vector2.New(math.Cos(angle)*radius, math.Sin(angle)*radius))
			return vector2.New(nudge(rng, p.X()), nudge(rng, p.Y()))
		}
		return at(), at(), at(), at()
	}
	cases := map[string]func() (a, b, c, d vector2.Float64){
		"random":    func() (a, b, c, d vector2.Float64) { return v(), v(), v(), v() },
		"grid":      func() (a, b, c, d vector2.Float64) { return grid(), grid(), grid(), grid() },
		"on circle": onCircle,
		"wide exponents": func() (a, b, c, d vector2.Float64) {
			a, b, c, d = onCircle()
			k := rng.Intn(200) - 100
			s := func(p vector2.Float64) vector2.Float64 { return vector2.New(math.Ldexp(p.X(), k), math.Ldexp(p.Y(), k)) }
			return s(a), s(b), s(c), s(d)
		},
	}

	for name, next := range cases {
		disagreed, zeros, fellThrough := 0, 0, 0
		for i := 0; i < 200_000; i++ {
			a, b, c, d := next()
			got, ok := inCircle(a, b, c, d)
			want := inCircleExact(a, b, c, d)
			if !ok {
				fellThrough++
				continue
			}
			if sign(got) != sign(want) {
				disagreed++
				if disagreed < 4 {
					t.Errorf("%s: %v %v %v %v: adaptive %v exact %v", name, a, b, c, d, got, want)
				}
			}
			if want == 0 {
				zeros++
			}
		}
		require.Zerof(t, disagreed, "%s: %d disagreements", name, disagreed)
		t.Logf("%s: %d exact zeros settled, %d fell through to big arithmetic", name, zeros, fellThrough)
		if name == "grid" {
			require.Positive(t, zeros, "grid points settle cocircular cases without big arithmetic")
			require.Zero(t, fellThrough)
		}
	}
}

func TestExpansionArithmeticIsExact(t *testing.T) {
	rng := rand.New(rand.NewSource(4))
	for i := 0; i < 100_000; i++ {
		a, b := scaled(rng, rng.Float64()*2-1), scaled(rng, rng.Float64()*2-1)

		x, y := twoSum(a, b)
		require.Equal(t, exactArithmeticFor(a, b).add(exactArithmeticFor(a, b).float(a), exactArithmeticFor(a, b).float(b)).Cmp(
			exactArithmeticFor(x, y).add(exactArithmeticFor(x, y).float(x), exactArithmeticFor(x, y).float(y))), 0, "twoSum %v %v", a, b)

		x, y = twoProduct(a, b)
		arith := exactArithmeticFor(a, b, x, y)
		require.Equal(t, arith.mul(arith.float(a), arith.float(b)).Cmp(arith.add(arith.float(x), arith.float(y))), 0, "twoProduct %v %v", a, b)
	}
}
