package texturing

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/EliCDavis/vector/vector2"
)

func NewTexture[T any](width, height int) Texture[T] {
	return Texture[T]{
		width:  width,
		height: height,
		data:   make([]T, width*height),
	}
}

type Texture[T any] struct {
	width  int
	height int
	data   []T
}

func (t Texture[T]) Set(x, y int, v T) {
	t.data[x+(y*t.width)] = v
}

func (t Texture[T]) Get(x, y int) T {
	return t.data[x+(y*t.width)]
}

func (t Texture[T]) Fill(v T) {
	for i := range t.data {
		t.data[i] = v
	}
}

func (t Texture[T]) Width() int {
	return t.width
}

func (t Texture[T]) Height() int {
	return t.height
}

func (t Texture[T]) Scan(cb func(x int, y int, v T)) {
	for y := range t.height {
		yAdjust := y * t.width
		for x := range t.width {
			cb(x, y, t.data[x+yAdjust])
		}
	}
}

func (t Texture[T]) ToImage(f func(T) color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, t.width, t.height))
	for y := range t.height {
		yAdjust := y * t.width
		for x := range t.width {
			img.Set(x, y, f(t.data[x+yAdjust]))
		}
	}
	return img
}

func (t Texture[T]) Copy() Texture[T] {
	destination := make([]T, len(t.data))
	copy(destination, t.data)
	return Texture[T]{
		data:   destination,
		width:  t.width,
		height: t.height,
	}
}

func (t Texture[T]) SaveImage(fp string, conv func(T) color.Color) error {
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, t.ToImage(conv))
}

func (t Texture[T]) SearchNeighborhood(start vector2.Int, size int, cb func(x int, y int, v T)) {
	startX := start.X()
	startY := start.Y()
	for y := startY - size; y <= startY+size; y++ {
		for x := startX - size; x <= startX+size; x++ {
			cb(x, y, t.Get(x, y))
		}
	}
}
