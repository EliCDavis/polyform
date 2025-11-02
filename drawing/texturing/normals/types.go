package normals

import (
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/vector/vector3"
)

type NormalMap = texturing.Texture[vector3.Float64]
type HeightMap = texturing.Texture[float64]
