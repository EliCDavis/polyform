package coloring

import (
	"math"
)

/*
type Space[T any] interface {
	Distance(a, b T) float64
	Add(a, b T) T
	Sub(a, b T) T
	Scale(a T, amount float64) T
	Dot(a, b T) float64
	Length(a T) float64
	Normalized(a T) T
	Lerp(a, b T, time float64) T
}

*/

type Space struct {
}

func (c Space) Distance(a, b WebColor) float64 {
	return c.Length(c.Sub(b, a))
}

func (Space) Add(a, b WebColor) WebColor {
	return WebColor{
		R: a.R + b.R,
		G: a.G + b.G,
		B: a.B + b.B,
		A: a.A + b.A,
	}
}

func (Space) Sub(a, b WebColor) WebColor {
	return WebColor{
		R: a.R - b.R,
		G: a.G - b.G,
		B: a.B - b.B,
		A: a.A - b.A,
	}
}

func (Space) Scale(a WebColor, amount float64) WebColor {
	return WebColor{
		R: a.R * amount,
		G: a.G * amount,
		B: a.B * amount,
		A: a.A * amount,
	}
}

func (Space) Dot(a, b WebColor) float64 {
	return (a.R * b.R) + (a.G * b.G) + (a.B * b.B) + (a.A * b.A)
}

func (Space) Length(c WebColor) float64 {
	return math.Sqrt((c.R * c.R) + (c.G * c.G) + (c.B * c.B) + (c.A * c.A))
}

func (Space) Normalized(a WebColor) WebColor {
	// Shit. uh. shit
	// I'm not sure what we can do about this given we're not in floating point
	// land.

	// At the moment. I don't know when I'd used `ColorSpace` struct in vector
	// math that would require normalizaiton. If it ever comes time, we probably
	// will just change this function to do whatever it needs done.

	return a
}

func (Space) Lerp(a, b WebColor, time float64) WebColor {
	return a.Lerp(b, time)
}
