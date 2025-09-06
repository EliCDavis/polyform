package texturing

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector1"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
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

func fract(v float64) float64 { return v - math.Floor(v) }

func negativeWrap(f float64) float64 {
	if f >= 0 {
		return f
	}
	return 1 - f
}

func (t Texture[T]) UV(x, y float64) T {
	return t.Get(
		int(negativeWrap(fract(x))*float64(t.width)),
		int(negativeWrap(fract(y))*float64(t.height)),
	)
}

func (t Texture[T]) Pixels() int {
	return t.width * t.height
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

type UniformNode[T any] struct {
	Fill   nodes.Output[T]
	Width  nodes.Output[int]
	Height nodes.Output[int]
}

func (n UniformNode[T]) Texture(out *nodes.StructOutput[Texture[T]]) {
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
	Texture nodes.Output[Texture[coloring.Color]]
}

func (n ColorToImageNode) Image(out *nodes.StructOutput[image.Image]) {
	if n.Texture == nil {
		return
	}
	out.Set(nodes.GetOutputValue(out, n.Texture).ToImage(func(c coloring.Color) color.Color {
		return coloring.Color{
			R: math.Max(0, math.Min(1, c.R)),
			G: math.Max(0, math.Min(1, c.G)),
			B: math.Max(0, math.Min(1, c.B)),
			A: math.Max(0, math.Min(1, c.A)),
		}
	}))
}

// ============================================================================

type FloatToImageNode struct {
	R nodes.Output[Texture[float64]]
	G nodes.Output[Texture[float64]]
	B nodes.Output[Texture[float64]]
	A nodes.Output[Texture[float64]]

	RFill nodes.Output[float64]
	GFill nodes.Output[float64]
	BFill nodes.Output[float64]
	AFill nodes.Output[float64]
}

func (n FloatToImageNode) tex(out nodes.ExecutionRecorder) Texture[coloring.Color] {
	if n.R == nil && n.G == nil && n.B == nil && n.A == nil {
		return NewTexture[coloring.Color](0, 0)
	}

	rFill := nodes.TryGetOutputValue(out, n.RFill, 0.)
	gFill := nodes.TryGetOutputValue(out, n.GFill, 0.)
	bFill := nodes.TryGetOutputValue(out, n.BFill, 0.)
	aFill := nodes.TryGetOutputValue(out, n.AFill, 0.)

	rTex := nodes.TryGetOutputReference(out, n.R, nil)
	gTex := nodes.TryGetOutputReference(out, n.G, nil)
	bTex := nodes.TryGetOutputReference(out, n.B, nil)
	aTex := nodes.TryGetOutputReference(out, n.A, nil)

	texs := make([]Texture[float64], 0)
	if rTex != nil {
		texs = append(texs, *rTex)
	}
	if gTex != nil {
		texs = append(texs, *gTex)
	}
	if bTex != nil {
		texs = append(texs, *bTex)
	}
	if aTex != nil {
		texs = append(texs, *aTex)
	}

	for i := 1; i < len(texs); i++ {
		if texs[0].width != texs[i].width || texs[0].height != texs[i].height {
			out.CaptureError(fmt.Errorf("mismatch texture dimensions"))
			return NewTexture[coloring.Color](0, 0)
		}
	}

	tex := NewTexture[coloring.Color](texs[0].width, texs[0].height)
	for y := range texs[0].height {
		for x := range texs[0].width {
			c := coloring.Color{
				R: rFill,
				G: gFill,
				B: bFill,
				A: aFill,
			}
			if rTex != nil {
				c.R = rTex.Get(x, y)
			}

			if gTex != nil {
				c.G = gTex.Get(x, y)
			}

			if bTex != nil {
				c.B = bTex.Get(x, y)
			}

			if aTex != nil {
				c.A = aTex.Get(x, y)
			}

			tex.Set(x, y, c)
		}
	}
	return tex
}

func (n FloatToImageNode) Texture(out *nodes.StructOutput[Texture[coloring.Color]]) {
	out.Set(n.tex(out))
}

func (n FloatToImageNode) Image(out *nodes.StructOutput[image.Image]) {
	texture := n.tex(out)
	out.Set(texture.ToImage(func(c coloring.Color) color.Color {
		return c
	}))
}

// ============================================================================

type ApplyMaskNode[T any] struct {
	Texture nodes.Output[Texture[T]]
	Mask    nodes.Output[Texture[bool]]
	Fill    nodes.Output[T]
}

func (n ApplyMaskNode[T]) process(out *nodes.StructOutput[Texture[T]], keep bool) {
	if n.Texture == nil {
		return
	}

	if n.Mask == nil {
		out.Set(nodes.GetOutputValue(out, n.Texture))
		return
	}

	var fill T
	if n.Fill != nil {
		fill = nodes.GetOutputValue(out, n.Fill)
	}

	mask := nodes.GetOutputValue(out, n.Mask)
	tex := nodes.GetOutputValue(out, n.Texture)

	if mask.Height() != tex.Height() || tex.Width() != mask.Width() {
		out.CaptureError(fmt.Errorf("mask and texture dimensions do not match"))
		return
	}

	result := NewTexture[T](tex.width, tex.height)
	for y := range result.height {
		for x := range result.width {
			if mask.Get(x, y) == keep {
				result.Set(x, y, tex.Get(x, y))
			} else {
				result.Set(x, y, fill)
			}
		}
	}
	out.Set(result)
}

func (n ApplyMaskNode[T]) Kept(out *nodes.StructOutput[Texture[T]]) {
	n.process(out, true)
}

func (n ApplyMaskNode[T]) Removed(out *nodes.StructOutput[Texture[T]]) {
	n.process(out, false)
}

// ============================================================================

func addTextures[T any](textures []Texture[T], out *nodes.StructOutput[Texture[T]], space vector.Space[T]) {
	if len(textures) == 0 {
		return
	}

	if len(textures) == 1 {
		out.Set(textures[0])
		return
	}

	if !resolutionsMatch(textures) {
		out.CaptureError(fmt.Errorf("mismatch texture resolution"))
		return
	}

	result := NewTexture[T](textures[0].Width(), textures[0].Height())
	for y := range result.Height() {
		for x := range result.Width() {
			var v T
			for _, tex := range textures {
				v = space.Add(tex.Get(x, y), v)
			}
			result.Set(x, y, v)
		}
	}

	out.Set(result)
}

type AddFloat1Node struct {
	Textures []nodes.Output[Texture[float64]]
}

func (n AddFloat1Node) Result(out *nodes.StructOutput[Texture[float64]]) {
	addTextures(nodes.GetOutputValues(out, n.Textures), out, vector1.Space[float64]{})
}

type AddFloat2Node struct {
	Textures []nodes.Output[Texture[vector2.Float64]]
}

func (n AddFloat2Node) Result(out *nodes.StructOutput[Texture[vector2.Float64]]) {
	addTextures(nodes.GetOutputValues(out, n.Textures), out, vector2.Space[float64]{})
}

type AddFloat3Node struct {
	Textures []nodes.Output[Texture[vector3.Float64]]
}

func (n AddFloat3Node) Result(out *nodes.StructOutput[Texture[vector3.Float64]]) {
	addTextures(nodes.GetOutputValues(out, n.Textures), out, vector3.Space[float64]{})
}

type AddFloat4Node struct {
	Textures []nodes.Output[Texture[vector4.Float64]]
}

func (n AddFloat4Node) Result(out *nodes.StructOutput[Texture[vector4.Float64]]) {
	addTextures(nodes.GetOutputValues(out, n.Textures), out, vector4.Space[float64]{})
}

type AddColorNode struct {
	Textures []nodes.Output[Texture[coloring.Color]]
}

func (n AddColorNode) Result(out *nodes.StructOutput[Texture[coloring.Color]]) {
	addTextures(nodes.GetOutputValues(out, n.Textures), out, coloring.Space{})
}

// ============================================================================

func resolutionsMatch[T any](textures []Texture[T]) bool {
	if len(textures) < 2 {
		return true
	}
	for i := 1; i < len(textures); i++ {
		if textures[0].width != textures[i].width || textures[0].height != textures[i].height {
			return false
		}
	}
	return true
}

type MultiplyFloat1Node struct {
	Textures []nodes.Output[Texture[float64]]
}

func (n MultiplyFloat1Node) Result(out *nodes.StructOutput[Texture[float64]]) {
	textures := nodes.GetOutputValues(out, n.Textures)
	if len(textures) == 0 {
		return
	}

	if len(textures) == 1 {
		out.Set(textures[0])
		return
	}

	if !resolutionsMatch(textures) {
		out.CaptureError(fmt.Errorf("mismatch texture resolution"))
		return
	}

	result := NewTexture[float64](textures[0].Width(), textures[0].Height())
	for y := range result.Height() {
		for x := range result.Width() {
			v := textures[0].Get(x, y)
			for i := 1; i < len(textures); i++ {
				v *= textures[i].Get(x, y)
			}
			result.Set(x, y, v)
		}
	}

	out.Set(result)
}

// ============================================================================

func scaleTexture[T any](texturePort nodes.Output[Texture[T]], out *nodes.StructOutput[Texture[T]], space vector.Space[T], amount nodes.Output[float64]) {
	if texturePort == nil {
		return
	}
	texture := texturePort.Value()

	amt := nodes.TryGetOutputValue(out, amount, 1)
	result := NewTexture[T](texture.Width(), texture.Height())
	for y := range result.Height() {
		for x := range result.Width() {
			result.Set(x, y, space.Scale(texture.Get(x, y), amt))
		}
	}

	out.Set(result)
}

type ScaleFloat1Node struct {
	Texture nodes.Output[Texture[float64]]
	Scale   nodes.Output[float64]
}

func (n ScaleFloat1Node) Result(out *nodes.StructOutput[Texture[float64]]) {
	scaleTexture(n.Texture, out, vector1.Space[float64]{}, n.Scale)
}

type ScaleFloat2Node struct {
	Texture nodes.Output[Texture[vector2.Float64]]
	Scale   nodes.Output[float64]
}

func (n ScaleFloat2Node) Result(out *nodes.StructOutput[Texture[vector2.Float64]]) {
	scaleTexture(n.Texture, out, vector2.Space[float64]{}, n.Scale)
}

type ScaleFloat3Node struct {
	Texture nodes.Output[Texture[vector3.Float64]]
	Scale   nodes.Output[float64]
}

func (n ScaleFloat3Node) Result(out *nodes.StructOutput[Texture[vector3.Float64]]) {
	scaleTexture(n.Texture, out, vector3.Space[float64]{}, n.Scale)
}

type ScaleFloat4Node struct {
	Texture nodes.Output[Texture[vector4.Float64]]
	Scale   nodes.Output[float64]
}

func (n ScaleFloat4Node) Result(out *nodes.StructOutput[Texture[vector4.Float64]]) {
	scaleTexture(n.Texture, out, vector4.Space[float64]{}, n.Scale)
}

type ScaleColorNode struct {
	Texture nodes.Output[Texture[coloring.Color]]
	Scale   nodes.Output[float64]
}

func (n ScaleColorNode) Result(out *nodes.StructOutput[Texture[coloring.Color]]) {
	scaleTexture(n.Texture, out, coloring.Space{}, n.Scale)
}
