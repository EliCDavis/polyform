package textures

import (
	"math"

	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type CheckerPattern struct {
	even       rendering.Texture
	odd        rendering.Texture
	tilingRate float64
}

func NewCheckerPattern(even, odd rendering.Texture) CheckerPattern {
	return CheckerPattern{
		even:       even,
		odd:        odd,
		tilingRate: 10,
	}
}

func NewCheckerPatternWithTilingRate(even, odd rendering.Texture, tilingRate float64) CheckerPattern {
	return CheckerPattern{
		even:       even,
		odd:        odd,
		tilingRate: tilingRate,
	}
}

func NewCheckerColorPattern(even, odd vector3.Float64) CheckerPattern {
	return CheckerPattern{
		even:       NewSolidColorTexture(even),
		odd:        NewSolidColorTexture(odd),
		tilingRate: 10,
	}
}

func (ct CheckerPattern) Value(uv vector2.Float64, p vector3.Float64) vector3.Float64 {
	sines := math.Sin(ct.tilingRate*p.X()) * math.Sin(ct.tilingRate*p.Y()) * math.Sin(ct.tilingRate*p.Z())
	if sines < 0 {
		return ct.odd.Value(uv, p)
	}
	return ct.even.Value(uv, p)
}
