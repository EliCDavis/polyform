package texturing

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
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
	for y := max(startY-size, 0); y <= min(startY+size, t.height-1); y++ {
		for x := max(startX-size, 0); x <= min(startX+size, t.width-1); x++ {
			cb(x, y, t.Get(x, y))
		}
	}
}

func Convert[T, G any](in Texture[T], cb func(x int, y int, v T) G) Texture[G] {
	out := NewTexture[G](in.width, in.height)
	for y := 0; y < in.height; y++ {
		for x := 0; x < in.width; x++ {
			out.Set(x, y, cb(x, y, in.Get(x, y)))
		}
	}
	return out
}

type TextureNode[T any] struct {
	Fill   nodes.Output[T]
	Width  nodes.Output[int]
	Height nodes.Output[int]
}

func (n TextureNode[T]) Texture(out *nodes.StructOutput[Texture[T]]) {
	t := NewTexture[T](
		nodes.TryGetOutputValue(out, n.Width, 1),
		nodes.TryGetOutputValue(out, n.Height, 1),
	)

	if n.Fill != nil {
		t.Fill(nodes.GetOutputValue(out, n.Fill))
	}

	out.Set(t)
}

// ============================================================================

type CompareValueTextureNode[T vector.Number] struct {
	Texture nodes.Output[Texture[T]]
	Value   nodes.Output[T]
}

func (n CompareValueTextureNode[T]) compare(out *nodes.StructOutput[Texture[bool]], f func(in, value T) bool) {
	if n.Texture == nil {
		return
	}

	in := nodes.GetOutputValue(out, n.Texture)
	value := nodes.TryGetOutputValue(out, n.Value, 0.)
	out.Set(Convert(in, func(x, y int, v T) bool {
		return f(v, value)
	}))
}

func (n CompareValueTextureNode[T]) GreaterThan(out *nodes.StructOutput[Texture[bool]]) {
	n.compare(out, func(in, value T) bool { return in > value })
}

func (n CompareValueTextureNode[T]) LessThan(out *nodes.StructOutput[Texture[bool]]) {
	n.compare(out, func(in, value T) bool { return in < value })
}

func (n CompareValueTextureNode[T]) GreaterThanOrEqualTo(out *nodes.StructOutput[Texture[bool]]) {
	n.compare(out, func(in, value T) bool { return in >= value })
}

func (n CompareValueTextureNode[T]) LessThanOrEqualTo(out *nodes.StructOutput[Texture[bool]]) {
	n.compare(out, func(in, value T) bool { return in <= value })
}

func (n CompareValueTextureNode[T]) Equal(out *nodes.StructOutput[Texture[bool]]) {
	n.compare(out, func(in, value T) bool { return in == value })
}

// ============================================================================

type FromArrayNode[T any] struct {
	Array  nodes.Output[[]T]
	Width  nodes.Output[int]
	Height nodes.Output[int]
}

func (n FromArrayNode[T]) Texture(out *nodes.StructOutput[Texture[T]]) {
	if n.Width == nil && n.Height == nil {
		return
	}

	width := nodes.TryGetOutputValue(out, n.Width, 1)
	height := nodes.TryGetOutputValue(out, n.Height, 1)
	arr := nodes.TryGetOutputValue(out, n.Array, nil)

	if width*height == len(arr) {
		out.Set(Texture[T]{
			width:  width,
			height: height,
			data:   arr,
		})
		return
	}

	tex := Texture[T]{
		width:  width,
		height: height,
		data:   make([]T, width*height),
	}
	copy(tex.data, arr)
	out.Set(tex)
}

// ============================================================================
type SelectNode[T any] struct {
	Texture nodes.Output[Texture[T]]
}

func (n SelectNode[T]) Array(out *nodes.StructOutput[[]T]) {
	if n.Texture == nil {
		return
	}
	out.Set(nodes.GetOutputValue(out, n.Texture).data)
}

func (n SelectNode[T]) Width(out *nodes.StructOutput[int]) {
	if n.Texture == nil {
		return
	}
	out.Set(nodes.GetOutputValue(out, n.Texture).Width())
}

func (n SelectNode[T]) Height(out *nodes.StructOutput[int]) {
	if n.Texture == nil {
		return
	}
	out.Set(nodes.GetOutputValue(out, n.Texture).Height())
}

// ============================================================================

type ColorToImageNode struct {
	Texture nodes.Output[Texture[color.Color]]
}

func (n ColorToImageNode) Image(out *nodes.StructOutput[image.Image]) {
	if n.Texture == nil {
		return
	}
	out.Set(nodes.GetOutputValue(out, n.Texture).ToImage(func(c color.Color) color.Color {
		if c == nil {
			return color.Black
		}
		return c
	}))
}
