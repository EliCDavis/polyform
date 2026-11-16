package triangulation_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sharpLens(length, width float64, samples int, center vector2.Float64, degrees float64) []vector2.Float64 {
	points := make([]vector2.Float64, samples)
	end := 2*math.Pi - 2*math.Pi/float64(samples)
	rad := degrees * math.Pi / 180
	for i := range points {
		th := end * float64(i) / float64(samples-1)
		s := math.Sin(th)
		p := vector2.New(length/2*(1-math.Cos(th)), -width/2*s*math.Abs(s))
		points[i] = vector2.New(
			p.X()*math.Cos(rad)-p.Y()*math.Sin(rad),
			p.X()*math.Sin(rad)+p.Y()*math.Cos(rad),
		).Add(center)
	}
	return points
}

func triangleArea(region triangulation.Region) float64 {
	total := 0.
	for _, tri := range region.Triangles {
		a, b, c := region.Points[tri[0]], region.Points[tri[1]], region.Points[tri[2]]
		total += math.Abs(b.Sub(a).X()*c.Sub(a).Y()-b.Sub(a).Y()*c.Sub(a).X()) / 2
	}
	return total
}

func coveredArea(extent float64, shapes ...geometry.Shape) float64 {
	const steps = 600
	cell := 2 * extent / steps
	inside := 0
	for i := range steps {
		for j := range steps {
			p := vector2.New(-extent+(float64(i)+0.5)*cell, -extent+(float64(j)+0.5)*cell)
			for _, s := range shapes {
				if s.IsInside(p) {
					inside++
					break
				}
			}
		}
	}
	return float64(inside) * cell * cell
}

func TestFillCoversSharpTippedOutlinesSharingATip(t *testing.T) {
	for _, length := range []float64{1, 0.1, 0.03} {
		for degrees := 5.; degrees < 180; degrees += 15 {
			a := sharpLens(length, length*0.45, 32, vector2.Zero[float64](), 90+degrees)
			b := sharpLens(length, length*0.45, 32, vector2.Zero[float64](), 90-degrees)
			want := coveredArea(length, geometry.Shape(a), geometry.Shape(b))
			region, err := triangulation.Fill(a, b)
			require.NoError(t, err)
			assert.InDelta(t, want, triangleArea(region), want*0.01, "length %v at ±%v°", length, degrees)
		}
	}
}

func TestFillCoversASharpTippedOutlineCompletely(t *testing.T) {
	for _, length := range []float64{1, 0.1, 0.03} {
		for degrees := 0.; degrees < 360; degrees += 15 {
			lens := sharpLens(length, length*0.45, 32, vector2.New(length*0.3, -length*0.2), degrees)
			want := math.Abs(geometry.Shape(lens).SignedArea())
			region, err := triangulation.Fill(lens)
			require.NoError(t, err)
			assert.InDelta(t, want, triangleArea(region), want*1e-6, "length %v at %v°", length, degrees)
		}
	}
}
