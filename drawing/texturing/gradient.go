package texturing

import (
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type LinearGradient[T any] struct {
	Repetitions float64
	Rotation    float64
	Width       int
	Height      int
	Gradient    coloring.Gradient[T]
}

func (lg LinearGradient[T]) Texture() Texture[T] {
	tex := NewTexture[T](lg.Width, lg.Height)

	boxCenter := vector3.New(lg.Width, lg.Height, 0).
		ToFloat64().
		Scale(0.5)

	aabb := geometry.NewAABB(
		boxCenter,
		vector3.New(float64(lg.Width), float64(lg.Height), 1),
	)
	dir := vector3.New(math.Cos(lg.Rotation), math.Sin(lg.Rotation), 0.).Normalized()
	ray := geometry.NewRay(boxCenter, dir)
	intersection := ray.At(aabb.RayIntersect(ray))
	length := intersection.XY().Distance(boxCenter.XY()) * 2

	ray = geometry.NewRay(intersection, dir.Scale(-1))
	for x := range lg.Width {
		for y := range lg.Height {
			pix := vector3.New(float64(x), float64(y), 0.)
			t := ray.TimeOnRay(pix)
			tex.Set(x, y, lg.Gradient.Sample((t/length)*lg.Repetitions))
		}
	}

	return tex
}

type LinearGradientNode[T any] struct {
	Repetitions nodes.Output[float64]
	Rotation    nodes.Output[float64]
	Width       nodes.Output[int]
	Height      nodes.Output[int]
	Gradient    nodes.Output[coloring.Gradient[T]]
}

func (n LinearGradientNode[T]) LinearGradient(out *nodes.StructOutput[Texture[T]]) {
	width := nodes.TryGetOutputValue(out, n.Width, 1)
	height := nodes.TryGetOutputValue(out, n.Height, 1)

	if n.Gradient == nil {
		out.Set(NewTexture[T](width, height))
		return
	}

	lg := LinearGradient[T]{
		Repetitions: nodes.TryGetOutputValue(out, n.Repetitions, 1),
		Rotation:    nodes.TryGetOutputValue(out, n.Rotation, 0),
		Width:       width,
		Height:      height,
		Gradient:    nodes.GetOutputValue(out, n.Gradient),
	}
	out.Set(lg.Texture())
}
