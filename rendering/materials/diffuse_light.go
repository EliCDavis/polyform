package materials

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/polyform/rendering/textures"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type DiffuseLight struct {
	texture rendering.Texture
}

func NewDiffuseLight(tex rendering.Texture) DiffuseLight {
	return DiffuseLight{
		texture: tex,
	}
}

func NewDiffuseLightWithColor(color vector3.Float64) DiffuseLight {
	return DiffuseLight{
		texture: textures.NewSolidColorTexture(color),
	}
}

func (dl DiffuseLight) Scatter(in geometry.Ray, rec *rendering.HitRecord, attenuation *vector3.Float64, scattered *geometry.Ray) bool {
	return false
}

func (dl DiffuseLight) Emitted(uv vector2.Float64, point vector3.Float64) vector3.Float64 {
	return dl.texture.Value(uv, point)
}
