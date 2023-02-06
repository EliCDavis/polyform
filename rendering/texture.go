package rendering

import (
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Texture interface {
	Value(uv vector2.Float64, p vector3.Float64) vector3.Float64
}
