package coloring

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

// RGBToHSV converts red/green/blue in 0-1 into hue in degrees (0-360) plus
// saturation and value in 0-1.
func RGBToHSV(r, g, b float64) (h, s, v float64) {
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	v = max
	if max > 0 {
		s = delta / max
	}

	if delta == 0 {
		return 0, s, v
	}

	switch max {
	case r:
		h = 60 * math.Mod((g-b)/delta, 6)
	case g:
		h = 60 * (((b - r) / delta) + 2)
	default:
		h = 60 * (((r - g) / delta) + 4)
	}

	if h < 0 {
		h += 360
	}
	return h, s, v
}

// HSVToRGB converts hue in degrees plus saturation and value in 0-1 into
// red/green/blue in 0-1.
func HSVToRGB(h, s, v float64) (r, g, b float64) {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c

	var rp, gp, bp float64
	switch {
	case h < 60:
		rp, gp, bp = c, x, 0
	case h < 120:
		rp, gp, bp = x, c, 0
	case h < 180:
		rp, gp, bp = 0, c, x
	case h < 240:
		rp, gp, bp = 0, x, c
	case h < 300:
		rp, gp, bp = x, 0, c
	default:
		rp, gp, bp = c, 0, x
	}

	return rp + m, gp + m, bp + m
}

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}

// ============================================================================

type AdjustHSVNode struct {
	In         nodes.Output[Color]   `description:"Color to adjust."`
	Hue        nodes.Output[float64] `description:"Degrees to rotate the hue by, positive or negative. Wraps around the color wheel. Defaults to 0."`
	Saturation nodes.Output[float64] `description:"Multiplier on saturation - below 1 washes the color out toward grey, above 1 makes it more vivid. Defaults to 1."`
	Value      nodes.Output[float64] `description:"Multiplier on brightness - below 1 darkens, above 1 lightens. Defaults to 1."`
}

func (n AdjustHSVNode) Description() string {
	return "Shifts a color's hue and scales its saturation and brightness. Lighten, darken, tint, or desaturate a color."
}

func (n AdjustHSVNode) Out(out *nodes.StructOutput[Color]) {
	if n.In == nil {
		out.Set(Color{R: 0, G: 0, B: 0, A: 1})
		return
	}

	in := nodes.GetOutputValue(out, n.In)
	h, s, v := RGBToHSV(in.R, in.G, in.B)

	h += nodes.TryGetOutputValue(out, n.Hue, 0)
	s = clamp01(s * nodes.TryGetOutputValue(out, n.Saturation, 1))
	v = clamp01(v * nodes.TryGetOutputValue(out, n.Value, 1))

	r, g, b := HSVToRGB(h, s, v)
	out.Set(Color{R: r, G: g, B: b, A: in.A})
}

// ============================================================================

type BrightnessNode struct {
	In     nodes.Output[Color]   `description:"Color to lighten or darken."`
	Amount nodes.Output[float64] `description:"Multiplier on each color channel - 0.5 is half as bright, 2 is twice. Defaults to 1."`
}

func (n BrightnessNode) Description() string {
	return "Lightens or darkens a color by scaling its red, green and blue channels."
}

func (n BrightnessNode) Out(out *nodes.StructOutput[Color]) {
	if n.In == nil {
		out.Set(Color{R: 0, G: 0, B: 0, A: 1})
		return
	}

	in := nodes.GetOutputValue(out, n.In)
	amount := nodes.TryGetOutputValue(out, n.Amount, 1)

	out.Set(Color{
		R: clamp01(in.R * amount),
		G: clamp01(in.G * amount),
		B: clamp01(in.B * amount),
		A: in.A,
	})
}

// ============================================================================

type MultiplyNode struct {
	A nodes.Output[Color] `description:"First color."`
	B nodes.Output[Color] `description:"Second color, multiplied into the first channel by channel."`
}

func (n MultiplyNode) Description() string {
	return "Multiplies two colors channel by channel, tinting one by the other."
}

func (n MultiplyNode) Out(out *nodes.StructOutput[Color]) {
	a := nodes.TryGetOutputValue(out, n.A, Color{R: 1, G: 1, B: 1, A: 1})
	b := nodes.TryGetOutputValue(out, n.B, Color{R: 1, G: 1, B: 1, A: 1})

	out.Set(Color{
		R: a.R * b.R,
		G: a.G * b.G,
		B: a.B * b.B,
		A: a.A * b.A,
	})
}

// ============================================================================

type FromHSVNode struct {
	Hue        nodes.Output[float64] `description:"Hue in degrees, 0-360. Defaults to 0."`
	Saturation nodes.Output[float64] `description:"Saturation, 0 (grey) to 1 (fully vivid). Defaults to 1."`
	Value      nodes.Output[float64] `description:"Brightness, 0 (black) to 1. Defaults to 1."`
}

func (n FromHSVNode) Description() string {
	return "Builds a color from hue, saturation and value."
}

func (n FromHSVNode) Out(out *nodes.StructOutput[Color]) {
	r, g, b := HSVToRGB(
		nodes.TryGetOutputValue(out, n.Hue, 0),
		clamp01(nodes.TryGetOutputValue(out, n.Saturation, 1)),
		clamp01(nodes.TryGetOutputValue(out, n.Value, 1)),
	)
	out.Set(Color{R: r, G: g, B: b, A: 1})
}

// ============================================================================

type ToHSVNode struct {
	In nodes.Output[Color] `description:"Color to break apart."`
}

func (n ToHSVNode) Description() string {
	return "Splits a color into hue, saturation and value."
}

func (n ToHSVNode) hsv(out *nodes.StructOutput[float64]) (h, s, v float64) {
	if n.In == nil {
		return 0, 0, 0
	}
	in := nodes.GetOutputValue(out, n.In)
	return RGBToHSV(in.R, in.G, in.B)
}

func (n ToHSVNode) Hue(out *nodes.StructOutput[float64]) {
	h, _, _ := n.hsv(out)
	out.Set(h)
}

func (n ToHSVNode) Saturation(out *nodes.StructOutput[float64]) {
	_, s, _ := n.hsv(out)
	out.Set(s)
}

func (n ToHSVNode) Value(out *nodes.StructOutput[float64]) {
	_, _, v := n.hsv(out)
	out.Set(v)
}

// ============================================================================

// FromVectorNode is the inverse of ToVectorNode. Without it a color could be
// taken apart into numbers and operated on, but never reassembled, so any
// color computed with vector math was a dead end.
type FromVectorNode struct {
	Vector3 nodes.Output[vector3.Float64] `description:"RGB as x/y/z in 0-1. Alpha becomes 1."`
	Vector4 nodes.Output[vector4.Float64] `description:"RGBA as x/y/z/w in 0-1. Takes priority over Vector3 when both are wired."`
}

func (n FromVectorNode) Description() string {
	return "Builds a color from a vector's components. Inverse of To Vector."
}

func (n FromVectorNode) Out(out *nodes.StructOutput[Color]) {
	if n.Vector4 != nil {
		v := nodes.GetOutputValue(out, n.Vector4)
		out.Set(Color{R: clamp01(v.X()), G: clamp01(v.Y()), B: clamp01(v.Z()), A: clamp01(v.W())})
		return
	}

	if n.Vector3 != nil {
		v := nodes.GetOutputValue(out, n.Vector3)
		out.Set(Color{R: clamp01(v.X()), G: clamp01(v.Y()), B: clamp01(v.Z()), A: 1})
		return
	}

	out.Set(Color{R: 0, G: 0, B: 0, A: 1})
}

// ============================================================================

type FromVectorArrayNode struct {
	Vector3 nodes.Output[[]vector3.Float64] `description:"One RGB triple per color, components in 0-1."`
}

func (n FromVectorArrayNode) Description() string {
	return "Builds an array of colors from an array of vectors. Inverse of To Vector Array."
}

func (n FromVectorArrayNode) Out(out *nodes.StructOutput[[]Color]) {
	in := nodes.TryGetOutputValue(out, n.Vector3, nil)
	arr := make([]Color, len(in))
	for i, v := range in {
		arr[i] = Color{R: clamp01(v.X()), G: clamp01(v.Y()), B: clamp01(v.Z()), A: 1}
	}
	out.Set(arr)
}
