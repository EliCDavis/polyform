package texturing

import (
	"image/color"
	"image/png"
	"io"
)

type Artifact[T any] struct {
	Texture    Texture[T]
	Conversion func(v T) color.Color
}

func (a Artifact[T]) Mime() string {
	return "image/png"
}

func (a Artifact[T]) Write(w io.Writer) error {
	return png.Encode(w, a.Texture.ToImage(a.Conversion))
}
