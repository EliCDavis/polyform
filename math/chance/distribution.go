package chance

import (
	"errors"
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type Distribution[T any] func(next Stream) T

// Shape is the distribution that draws f(u) for one uniform u in [0, 1).
// With f non-decreasing, f(u) is the value a share u of draws fall below.
func Shape(f func(u float64) float64) Distribution[float64] {
	return func(next Stream) float64 { return f(next()) }
}

func Constant(value float64) Distribution[float64] {
	return func(Stream) float64 { return value }
}

func Uniform(min, max float64) Distribution[float64] {
	return Shape(func(u float64) float64 { return min + (max-min)*u })
}

func normalCDF(z float64) float64 {
	return 0.5 * math.Erfc(-z/math.Sqrt2)
}

func normalQuantile(mean, deviation, clamp float64) func(u float64) float64 {
	lo, hi := 1e-12, 1-1e-12
	if clamp > 0 {
		lo, hi = normalCDF(-clamp), normalCDF(clamp)
	}
	return func(u float64) float64 {
		u = lo + (hi-lo)*u
		return mean + deviation*math.Sqrt2*math.Erfinv(2*u-1)
	}
}

// Normal clamps by remapping u into the central share, so values beyond
// Clamp deviations never occur but nothing piles up at the edges. Clamp <= 0
// leaves the tails open.
func Normal(mean, deviation, clamp float64) Distribution[float64] {
	return Shape(normalQuantile(mean, deviation, clamp))
}

func LogNormal(median, spread float64) Distribution[float64] {
	z := normalQuantile(0, 1, 0)
	return Shape(func(u float64) float64 { return median * math.Exp(spread*z(u)) })
}

func Exponential(median float64) Distribution[float64] {
	return Shape(func(u float64) float64 {
		return -median * math.Log(1-min(u, 1-1e-12)) / math.Ln2
	})
}

func Triangular(min, mode, max float64) Distribution[float64] {
	span := max - min
	if span <= 0 {
		return Constant(min)
	}
	split := (mode - min) / span
	return Shape(func(u float64) float64 {
		if u < split {
			return min + math.Sqrt(u*span*(mode-min))
		}
		return max - math.Sqrt((1-u)*span*(max-mode))
	})
}

func Power(min, max, exponent float64) Distribution[float64] {
	return Shape(func(u float64) float64 { return min + (max-min)*math.Pow(u, exponent) })
}

func discreteQuantile(values, weights []float64) (func(u float64) float64, error) {
	if len(values) == 0 {
		return nil, errors.New("no values to pick from")
	}
	if weights == nil {
		weights = make([]float64, len(values))
		for i := range weights {
			weights[i] = 1
		}
	}
	if len(weights) != len(values) {
		return nil, fmt.Errorf("%d weight(s) for %d value(s)", len(weights), len(values))
	}
	cumulative := make([]float64, len(weights))
	total := 0.
	for i, w := range weights {
		if w < 0 {
			return nil, fmt.Errorf("weight %d is negative (%v)", i, w)
		}
		total += w
		cumulative[i] = total
	}
	if total <= 0 {
		return nil, errors.New("weights must hold at least one value above 0")
	}
	return func(u float64) float64 {
		target := u * total
		for i, c := range cumulative {
			if target < c {
				return values[i]
			}
		}
		for i := len(cumulative) - 1; i >= 0; i-- {
			if weights[i] > 0 {
				return values[i]
			}
		}
		return values[len(values)-1]
	}, nil
}

// Discrete picks Values[i] with probability Weights[i] / sum. Nil weights
// pick evenly.
func Discrete(values, weights []float64) (Distribution[float64], error) {
	q, err := discreteQuantile(values, weights)
	if err != nil {
		return nil, err
	}
	return Shape(q), nil
}

func Ramp(gradient coloring.Gradient[float64]) Distribution[float64] {
	return Shape(gradient.Sample)
}

func Compose2D(x, y Distribution[float64]) Distribution[vector2.Float64] {
	return func(next Stream) vector2.Float64 {
		return vector2.New(x(next), y(next))
	}
}

func Compose3D(x, y, z Distribution[float64]) Distribution[vector3.Float64] {
	return func(next Stream) vector3.Float64 {
		return vector3.New(x(next), y(next), z(next))
	}
}

func Compose4D(x, y, z, w Distribution[float64]) Distribution[vector4.Float64] {
	return func(next Stream) vector4.Float64 {
		return vector4.New(x(next), y(next), z(next), w(next))
	}
}

func Uniform2D(min, max vector2.Float64) Distribution[vector2.Float64] {
	return Compose2D(Uniform(min.X(), max.X()), Uniform(min.Y(), max.Y()))
}

func Uniform4D(min, max vector4.Float64) Distribution[vector4.Float64] {
	return Compose4D(Uniform(min.X(), max.X()), Uniform(min.Y(), max.Y()), Uniform(min.Z(), max.Z()), Uniform(min.W(), max.W()))
}

func Uniform3D(min, max vector3.Float64) Distribution[vector3.Float64] {
	return Compose3D(Uniform(min.X(), max.X()), Uniform(min.Y(), max.Y()), Uniform(min.Z(), max.Z()))
}

func InsideSphere(radius float64) Distribution[vector3.Float64] {
	return func(next Stream) vector3.Float64 {
		z := 2*next() - 1
		ring := math.Sqrt(max(0, 1-z*z))
		phi := 2 * math.Pi * next()
		r := radius * math.Cbrt(next())
		return vector3.New(ring*math.Cos(phi), ring*math.Sin(phi), z).Scale(r)
	}
}

// Direction is a unit vector within Cone Angle degrees of Axis.
func Direction(axis vector3.Float64, coneAngle float64) Distribution[vector3.Float64] {
	rotation := quaternion.RotationTo(vector3.Forward[float64](), axis.Normalized())
	cosCone := math.Cos(coneAngle * math.Pi / 180)
	return func(next Stream) vector3.Float64 {
		cosTheta := 1 - next()*(1-cosCone)
		sinTheta := math.Sqrt(max(0, 1-cosTheta*cosTheta))
		phi := 2 * math.Pi * next()
		local := vector3.New(sinTheta*math.Cos(phi), sinTheta*math.Sin(phi), cosTheta)
		return rotation.Rotate(local)
	}
}
