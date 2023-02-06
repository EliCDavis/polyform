package textures

import (
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type SolidColorTexture struct {
	c vector3.Float64
}

func NewSolidColorTexture(c vector3.Float64) SolidColorTexture {
	return SolidColorTexture{c}
}

func (sct SolidColorTexture) Value(uv vector2.Float64, p vector3.Float64) vector3.Float64 {
	return sct.c
}
