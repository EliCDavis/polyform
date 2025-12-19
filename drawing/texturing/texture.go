package texturing

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"runtime"
	"sync"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector1"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

func Empty[T any](width, height int) Texture[T] {
	return Texture[T]{
		width:  width,
		height: height,
		data:   make([]T, width*height),
	}
}

func FromArray[T any](arr []T, width, height int) Texture[T] {
	if len(arr) < width*height {
		panic(fmt.Errorf("can't create texture from array, array length %d less than provided dimensions %dx%d", len(arr), width, height))
	}
	return Texture[T]{
		width:  width,
		height: height,
		data:   arr,
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

func (t Texture[T]) Mutate(cb func(x int, y int, v T) T) {
	for y := range t.height {
		yAdjust := y * t.width
		for x := range t.width {
			t.data[x+yAdjust] = cb(x, y, t.data[x+yAdjust])
		}
	}
}

func (t Texture[T]) scanWorker(start, end int, cb func(x int, y int, v T), wg *sync.WaitGroup) {
	defer wg.Done()
	for i := start; i < end; i++ {
		x := i % t.width
		y := i / t.width
		cb(x, y, t.data[i])
	}
}

func (t Texture[T]) mutateWorker(start, end int, cb func(x int, y int, v T) T, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := start; i < end; i++ {
		x := i % t.width
		y := i / t.width
		t.data[i] = cb(x, y, t.data[i])
	}
}

func (t Texture[T]) ScanParallelAcross(workers int, cb func(x int, y int, v T)) {
	if workers <= 1 || len(t.data) <= 1 {
		t.Scan(cb)
		return
	}

	total := len(t.data)
	base := total / workers
	extra := total % workers // remainder to distribute

	wg := &sync.WaitGroup{}
	start := 0
	for i := range workers {

		// This happens whenever we have more workers than pixels
		if start == total {
			break
		}

		size := base
		if i < extra {
			size++
		}
		end := start + size
		wg.Add(1)
		go t.scanWorker(start, end, cb, wg)
		start = end
	}
	wg.Wait()
}

func (t Texture[T]) MutateParallelAcross(workers int, cb func(x int, y int, v T) T) {
	if workers <= 1 || len(t.data) <= 1 {
		t.Mutate(cb)
		return
	}

	total := len(t.data)
	base := total / workers
	extra := total % workers // remainder to distribute

	wg := &sync.WaitGroup{}
	start := 0
	for i := range workers {

		// This happens whenever we have more workers than pixels
		if start == total {
			break
		}

		size := base
		if i < extra {
			size++
		}
		end := start + size
		wg.Add(1)
		go t.mutateWorker(start, end, cb, wg)
		start = end
	}
	wg.Wait()
}

func (t Texture[T]) ScanParallel(cb func(x int, y int, v T)) {
	t.ScanParallelAcross(runtime.NumCPU(), cb)
}

func (t Texture[T]) MutateParallel(cb func(x int, y int, v T) T) {
	t.MutateParallelAcross(runtime.NumCPU(), cb)
}

func (t Texture[T]) ToImage(f func(T) color.Color) *image.RGBA {
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
	destination := make([]T, t.width*t.height)
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
	out := Empty[G](in.width, in.height)
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
	t := Empty[T](
		nodes.TryGetOutputValue(out, n.Width, 1),
		nodes.TryGetOutputValue(out, n.Height, 1),
	)

	if n.Fill != nil {
		t.Fill(nodes.GetOutputValue(out, n.Fill))
	}

	out.Set(t)
}

// ============================================================================

type CompareValueNode[T vector.Number] struct {
	Texture nodes.Output[Texture[T]]
	Value   nodes.Output[T]
}

func (n CompareValueNode[T]) compareMask(out *nodes.StructOutput[Texture[bool]], f func(in, value T) bool) {
	if n.Texture == nil {
		return
	}

	in := nodes.GetOutputValue(out, n.Texture)
	value := nodes.TryGetOutputValue(out, n.Value, 0.)
	out.Set(Convert(in, func(x, y int, v T) bool {
		return f(v, value)
	}))
}

func (n CompareValueNode[T]) compare(out *nodes.StructOutput[Texture[T]], f func(in, value T) bool) {
	if n.Texture == nil {
		return
	}

	in := nodes.GetOutputValue(out, n.Texture)
	value := nodes.TryGetOutputValue(out, n.Value, 0.)
	out.Set(Convert(in, func(x, y int, v T) T {
		var t T
		if f(v, value) {
			t = v
		}
		return t
	}))
}

func (n CompareValueNode[T]) GreaterThanMask(out *nodes.StructOutput[Texture[bool]]) {
	n.compareMask(out, func(in, value T) bool { return in > value })
}

func (n CompareValueNode[T]) LessThanMask(out *nodes.StructOutput[Texture[bool]]) {
	n.compareMask(out, func(in, value T) bool { return in < value })
}

func (n CompareValueNode[T]) GreaterThanOrEqualToMask(out *nodes.StructOutput[Texture[bool]]) {
	n.compareMask(out, func(in, value T) bool { return in >= value })
}

func (n CompareValueNode[T]) LessThanOrEqualToMask(out *nodes.StructOutput[Texture[bool]]) {
	n.compareMask(out, func(in, value T) bool { return in <= value })
}

func (n CompareValueNode[T]) EqualMask(out *nodes.StructOutput[Texture[bool]]) {
	n.compareMask(out, func(in, value T) bool { return in == value })
}

func (n CompareValueNode[T]) GreaterThan(out *nodes.StructOutput[Texture[T]]) {
	n.compare(out, func(in, value T) bool { return in > value })
}

func (n CompareValueNode[T]) LessThan(out *nodes.StructOutput[Texture[T]]) {
	n.compare(out, func(in, value T) bool { return in < value })
}

func (n CompareValueNode[T]) GreaterThanOrEqualTo(out *nodes.StructOutput[Texture[T]]) {
	n.compare(out, func(in, value T) bool { return in >= value })
}

func (n CompareValueNode[T]) LessThanOrEqualTo(out *nodes.StructOutput[Texture[T]]) {
	n.compare(out, func(in, value T) bool { return in <= value })
}

func (n CompareValueNode[T]) Equal(out *nodes.StructOutput[Texture[T]]) {
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

func selectArray[T any](out *nodes.StructOutput[[]T], texture nodes.Output[Texture[T]]) {
	if texture == nil {
		return
	}
	out.Set(nodes.GetOutputValue(out, texture).data)
}

func selectWidth[T any](out *nodes.StructOutput[int], texture nodes.Output[Texture[T]]) {
	if texture == nil {
		return
	}
	out.Set(nodes.GetOutputValue(out, texture).Width())
}

func selectHeight[T any](out *nodes.StructOutput[int], texture nodes.Output[Texture[T]]) {
	if texture == nil {
		return
	}
	out.Set(nodes.GetOutputValue(out, texture).Height())
}

type SelectNode[T any] struct{ Texture nodes.Output[Texture[T]] }

func (n SelectNode[T]) Array(out *nodes.StructOutput[[]T])  { selectArray(out, n.Texture) }
func (n SelectNode[T]) Width(out *nodes.StructOutput[int])  { selectWidth(out, n.Texture) }
func (n SelectNode[T]) Height(out *nodes.StructOutput[int]) { selectHeight(out, n.Texture) }

type SelectColorNode struct {
	Texture nodes.Output[Texture[coloring.Color]]
}

func (n SelectColorNode) Width(out *nodes.StructOutput[int])  { selectWidth(out, n.Texture) }
func (n SelectColorNode) Height(out *nodes.StructOutput[int]) { selectHeight(out, n.Texture) }
func (n SelectColorNode) Array(out *nodes.StructOutput[[]coloring.Color]) {
	selectArray(out, n.Texture)
}

func (n SelectColorNode) R(out *nodes.StructOutput[Texture[float64]]) {
	if n.Texture == nil {
		return
	}
	tex := nodes.GetOutputValue(out, n.Texture)
	out.Set(Convert(tex, func(x, y int, c coloring.Color) float64 {
		return c.R
	}))
}

func (n SelectColorNode) G(out *nodes.StructOutput[Texture[float64]]) {
	if n.Texture == nil {
		return
	}
	tex := nodes.GetOutputValue(out, n.Texture)
	out.Set(Convert(tex, func(x, y int, c coloring.Color) float64 {
		return c.G
	}))
}

func (n SelectColorNode) B(out *nodes.StructOutput[Texture[float64]]) {
	if n.Texture == nil {
		return
	}
	tex := nodes.GetOutputValue(out, n.Texture)
	out.Set(Convert(tex, func(x, y int, c coloring.Color) float64 {
		return c.B
	}))
}

func (n SelectColorNode) A(out *nodes.StructOutput[Texture[float64]]) {
	if n.Texture == nil {
		return
	}
	tex := nodes.GetOutputValue(out, n.Texture)
	out.Set(Convert(tex, func(x, y int, c coloring.Color) float64 {
		return c.A
	}))
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
		return Empty[coloring.Color](0, 0)
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
			out.CaptureError(ErrMismatchDimensions)
			return Empty[coloring.Color](0, 0)
		}
	}

	tex := Empty[coloring.Color](texs[0].width, texs[0].height)
	tex.MutateParallel(func(x, y int, v coloring.Color) coloring.Color {
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

		return c
	})

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

	result := Empty[T](tex.width, tex.height)
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
		out.CaptureError(ErrMismatchDimensions)
		return
	}

	result := Empty[T](textures[0].Width(), textures[0].Height())
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

// ============================================================================

func scaleTextureUniform[T any](
	texturePort nodes.Output[Texture[T]],
	out *nodes.StructOutput[Texture[T]],
	space vector.Space[T],
	amount nodes.Output[float64],
) {
	if texturePort == nil {
		return
	}
	texture := nodes.GetOutputValue(out, texturePort)

	amt := nodes.TryGetOutputValue(out, amount, 1)
	if amt == 1 {
		out.Set(texture)
		return
	}

	out.Set(ScaleUniform(texture, amt, space))
}

func ScaleUniform[T any](tex Texture[T], amount float64, space vector.Space[T]) Texture[T] {
	result := Empty[T](tex.Width(), tex.Height())
	for y := range result.Height() {
		for x := range result.Width() {
			result.Set(x, y, space.Scale(tex.Get(x, y), amount))
		}
	}
	return result
}

func scaleTexture[T any](
	texturePort nodes.Output[Texture[T]],
	out *nodes.StructOutput[Texture[T]],
	space vector.Space[T],
	amount nodes.Output[Texture[float64]],
) {
	if texturePort == nil {
		return
	}
	texture := nodes.GetOutputValue(out, texturePort)
	if texture.width == 0 || texture.height == 0 {
		return
	}

	if amount == nil {
		out.Set(texture)
		return
	}

	amt := nodes.GetOutputValue(out, amount)
	if texture.width != amt.width || texture.height != amt.height {
		out.CaptureError(ErrMismatchDimensions)
		return
	}

	result := Empty[T](texture.Width(), texture.Height())
	for y := range result.Height() {
		for x := range result.Width() {
			result.Set(x, y, space.Scale(texture.Get(x, y), amt.Get(x, y)))
		}
	}

	out.Set(result)
}

type ScaleFloat1UniformNode struct {
	Texture nodes.Output[Texture[float64]]
	Scale   nodes.Output[float64]
}

func (n ScaleFloat1UniformNode) Result(out *nodes.StructOutput[Texture[float64]]) {
	scaleTextureUniform(n.Texture, out, vector1.Space[float64]{}, n.Scale)
}

type ScaleFloat2UniformNode struct {
	Texture nodes.Output[Texture[vector2.Float64]]
	Scale   nodes.Output[float64]
}

func (n ScaleFloat2UniformNode) Result(out *nodes.StructOutput[Texture[vector2.Float64]]) {
	scaleTextureUniform(n.Texture, out, vector2.Space[float64]{}, n.Scale)
}

type ScaleFloat3UniformNode struct {
	Texture nodes.Output[Texture[vector3.Float64]]
	Scale   nodes.Output[float64]
}

func (n ScaleFloat3UniformNode) Result(out *nodes.StructOutput[Texture[vector3.Float64]]) {
	scaleTextureUniform(n.Texture, out, vector3.Space[float64]{}, n.Scale)
}

type ScaleFloat4UniformNode struct {
	Texture nodes.Output[Texture[vector4.Float64]]
	Scale   nodes.Output[float64]
}

func (n ScaleFloat4UniformNode) Result(out *nodes.StructOutput[Texture[vector4.Float64]]) {
	scaleTextureUniform(n.Texture, out, vector4.Space[float64]{}, n.Scale)
}

type ScaleColorUniformNode struct {
	Texture nodes.Output[Texture[coloring.Color]]
	Scale   nodes.Output[float64]
}

func (n ScaleColorUniformNode) Result(out *nodes.StructOutput[Texture[coloring.Color]]) {
	scaleTextureUniform(n.Texture, out, coloring.Space{}, n.Scale)
}

type ScaleColorNode struct {
	Texture nodes.Output[Texture[coloring.Color]]
	Scale   nodes.Output[Texture[float64]]
}

func (n ScaleColorNode) Result(out *nodes.StructOutput[Texture[coloring.Color]]) {
	scaleTexture(n.Texture, out, coloring.Space{}, n.Scale)
}
