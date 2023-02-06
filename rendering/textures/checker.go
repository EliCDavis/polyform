package textures

import (
	"math"

	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type CheckerPattern struct {
	even rendering.Texture
	odd  rendering.Texture
}

func NewCheckerPattern(even, odd rendering.Texture) CheckerPattern {
	return CheckerPattern{
		even: even,
		odd:  odd,
	}
}

func NewCheckerColorPattern(even, odd vector3.Float64) CheckerPattern {
	return CheckerPattern{
		even: NewSolidColorTexture(even),
		odd:  NewSolidColorTexture(odd),
	}
}

func (ct CheckerPattern) Value(uv vector2.Float64, p vector3.Float64) vector3.Float64 {
	sines := math.Sin(10*p.X()) * math.Sin(10*p.Y()) * math.Sin(10*p.Z())
	if sines < 0 {
		return ct.odd.Value(uv, p)
	}
	return ct.even.Value(uv, p)
}
