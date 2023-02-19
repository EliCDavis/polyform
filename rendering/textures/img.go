package textures

import (
	"image"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type ImageTexture struct {
	i    image.Image
	w, h int
}

func NewImage(c image.Image) ImageTexture {
	return ImageTexture{
		i: c,
		w: c.Bounds().Dx(),
		h: c.Bounds().Dy(),
	}
}

func (it ImageTexture) Value(uv vector2.Float64, p vector3.Float64) vector3.Float64 {
	r, g, b, _ := it.i.At(
		int(float64(it.w)*uv.X()),
		int(float64(it.h)*uv.Y()),
	).RGBA()

	v := vector3.New(
		float64(r>>8)/255.,
		float64(g>>8)/255.,
		float64(b>>8)/255.,
	)
	return v
}
