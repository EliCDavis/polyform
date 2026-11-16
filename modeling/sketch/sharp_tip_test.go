package sketch_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sharpLens(length, width float64, samples int) []vector2.Float64 {
	points := make([]vector2.Float64, samples)
	end := 2*math.Pi - 2*math.Pi/float64(samples)
	for i := range points {
		th := end * float64(i) / float64(samples-1)
		s := math.Sin(th)
		points[i] = vector2.New(length/2*(1-math.Cos(th)), -width/2*s*math.Abs(s))
	}
	return points
}

func TestASharpTippedOutlineKeepsItsWholeAreaThroughBooleans(t *testing.T) {
	lens := sharpLens(0.2, 0.06, 32)
	want := math.Abs(geometry.Shape(lens).SignedArea())

	filled, err := sketch.Fill([]geometry.Shape{lens}, nil)
	require.NoError(t, err)
	assert.InDelta(t, want, filled.Size(), want*1e-6, "fill")

	frame, err := sketch.Fill([]geometry.Shape{{
		vector2.New(-1., -1.), vector2.New(1., -1.), vector2.New(1., 1.), vector2.New(-1., 1.),
	}}, nil)
	require.NoError(t, err)

	inside, err := sketch.Intersect(filled, frame)
	require.NoError(t, err)
	assert.InDelta(t, want, inside.Size(), want*1e-6, "intersect")

	cut, err := sketch.Subtract(frame, inside)
	require.NoError(t, err)
	assert.InDelta(t, frame.Size()-want, cut.Size(), want*1e-6, "subtract")
}

func TestSharpTippedOutlinesKeepTheirAreaAtAnyRotation(t *testing.T) {
	lens := sharpLens(0.1, 0.045, 32)
	want := math.Abs(geometry.Shape(lens).SignedArea())

	for degrees := 0.; degrees < 360; degrees += 7 {
		moved := sketch.Transform(lens, vector2.New(0.03, -0.02), degrees, vector2.One[float64]())
		filled, err := sketch.Fill([]geometry.Shape{moved}, nil)
		require.NoError(t, err)
		if !assert.InDelta(t, want, filled.Size(), want*1e-4, "fill at %v°", degrees) {
			continue
		}
		frame, _ := sketch.Fill([]geometry.Shape{{
			vector2.New(-0.2, -0.2), vector2.New(0.2, -0.2), vector2.New(0.2, 0.2), vector2.New(-0.2, 0.2),
		}}, nil)
		inside, err := sketch.Intersect(filled, frame)
		require.NoError(t, err)
		assert.InDelta(t, want, inside.Size(), want*1e-4, "intersect at %v°", degrees)
	}
}
