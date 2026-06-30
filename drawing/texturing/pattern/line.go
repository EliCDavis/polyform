package pattern

import (
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/vector/vector2"
)

type Line[T any] struct {
	LineValue T
	Start     vector2.Float64
	End       vector2.Float64
	Width     float64
}

func (c Line[T]) Draw(tex texturing.Texture[T]) {

}
