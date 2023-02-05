package rendering

import (
	"github.com/EliCDavis/vector/vector3"
)

type Texture interface {
	Value(u, v float64) vector3.Float64
}

type SolidColorTexture struct {
	c vector3.Float64
}

func (sct SolidColorTexture) Value(u, v float64) vector3.Float64 {
	return sct.c
}
