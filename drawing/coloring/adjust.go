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
	In         nodes.LiftedPort[Color]   `description:"Color to adjust."`
	Hue        nodes.LiftedPort[float64] `description:"Degrees to rotate the hue by, positive or negative. Wraps around the color wheel. Defaults to 0."`
	Saturation nodes.LiftedPort[float64] `description:"Multiplier on saturation - below 1 washes the color out toward grey, above 1 makes it more vivid. Defaults to 1."`
	Value      nodes.LiftedPort[float64] `description:"Multiplier on brightness - below 1 darkens, above 1 lightens. Defaults to 1."`
}

func (n AdjustHSVNode) Description() string {
	return "Shifts a color's hue and scales its saturation and brightness. Lighten, darken, tint, or desaturate a color."
}

func (n AdjustHSVNode) Out(out *nodes.Lifted[Color]) {
	nodes.Zip4(
		out,
		nodes.LiftedOr(n.In, Color{A: 1}),
		n.Hue,
		nodes.LiftedOr(n.Saturation, 1.),
		nodes.LiftedOr(n.Value, 1.),
		func(in Color, hue, saturation, value float64) Color {
			h, s, v := RGBToHSV(in.R, in.G, in.B)
			r, g, b := HSVToRGB(h+hue, clamp01(s*saturation), clamp01(v*value))
			return Color{R: r, G: g, B: b, A: in.A}
		},
	)
}

// ============================================================================

type BrightnessNode struct {
	In     nodes.LiftedPort[Color]   `description:"Color to lighten or darken."`
	Amount nodes.LiftedPort[float64] `description:"Multiplier on each color channel - 0.5 is half as bright, 2 is twice. Defaults to 1."`
}

func (n BrightnessNode) Description() string {
	return "Lightens or darkens a color by scaling its red, green and blue channels."
}

func (n BrightnessNode) Out(out *nodes.Lifted[Color]) {
	nodes.Zip2(out, nodes.LiftedOr(n.In, Color{A: 1}), nodes.LiftedOr(n.Amount, 1.), func(in Color, amount float64) Color {
		return Color{
			R: clamp01(in.R * amount),
			G: clamp01(in.G * amount),
			B: clamp01(in.B * amount),
			A: in.A,
		}
	})
}

// ============================================================================

type MultiplyNode struct {
	A nodes.LiftedPort[Color] `description:"First color."`
	B nodes.LiftedPort[Color] `description:"Second color, multiplied into the first channel by channel."`
}

func (n MultiplyNode) Description() string {
	return "Multiplies two colors channel by channel, tinting one by the other."
}

func (n MultiplyNode) Out(out *nodes.Lifted[Color]) {
	white := Color{R: 1, G: 1, B: 1, A: 1}
	nodes.Zip2(out, nodes.LiftedOr(n.A, white), nodes.LiftedOr(n.B, white), func(a, b Color) Color {
		return Color{
			R: a.R * b.R,
			G: a.G * b.G,
			B: a.B * b.B,
			A: a.A * b.A,
		}
	})
}

// ============================================================================

type FromHSVNode struct {
	Hue        nodes.LiftedPort[float64] `description:"Hue in degrees, 0-360. Defaults to 0."`
	Saturation nodes.LiftedPort[float64] `description:"Saturation, 0 (grey) to 1 (fully vivid). Defaults to 1."`
	Value      nodes.LiftedPort[float64] `description:"Brightness, 0 (black) to 1. Defaults to 1."`
}

func (n FromHSVNode) Description() string {
	return "Builds a color from hue, saturation and value."
}

func (n FromHSVNode) Out(out *nodes.Lifted[Color]) {
	nodes.Zip3(out, n.Hue, nodes.LiftedOr(n.Saturation, 1.), nodes.LiftedOr(n.Value, 1.), func(hue, saturation, value float64) Color {
		r, g, b := HSVToRGB(hue, clamp01(saturation), clamp01(value))
		return Color{R: r, G: g, B: b, A: 1}
	})
}

// ============================================================================

type ToHSVNode struct {
	In nodes.LiftedPort[Color] `description:"Color to break apart."`
}

func (n ToHSVNode) Description() string {
	return "Splits a color into hue, saturation and value."
}

func (n ToHSVNode) channel(out *nodes.Lifted[float64], pick func(h, s, v float64) float64) {
	nodes.Zip1(out, nodes.LiftedOr(n.In, Color{A: 1}), func(in Color) float64 {
		return pick(RGBToHSV(in.R, in.G, in.B))
	})
}

func (n ToHSVNode) Hue(out *nodes.Lifted[float64]) {
	n.channel(out, func(h, _, _ float64) float64 { return h })
}

func (n ToHSVNode) Saturation(out *nodes.Lifted[float64]) {
	n.channel(out, func(_, s, _ float64) float64 { return s })
}

func (n ToHSVNode) Value(out *nodes.Lifted[float64]) {
	n.channel(out, func(_, _, v float64) float64 { return v })
}

// ============================================================================

type FromVectorNode struct {
	Vector3 nodes.LiftedPort[vector3.Float64] `description:"RGB as x/y/z in 0-1. Alpha becomes 1."`
	Vector4 nodes.LiftedPort[vector4.Float64] `description:"RGBA as x/y/z/w in 0-1. Takes priority over Vector3 when both are wired."`
}

func (n FromVectorNode) Description() string {
	return "Builds a color from a vector's components. Inverse of To Vector."
}

func (n FromVectorNode) Out(out *nodes.Lifted[Color]) {
	if n.Vector4 != nil {
		nodes.Zip1(out, n.Vector4, func(v vector4.Float64) Color {
			return Color{R: clamp01(v.X()), G: clamp01(v.Y()), B: clamp01(v.Z()), A: clamp01(v.W())}
		})
		return
	}

	nodes.Zip1(out, nodes.LiftedOr(n.Vector3, vector3.Zero[float64]()), func(v vector3.Float64) Color {
		return Color{R: clamp01(v.X()), G: clamp01(v.Y()), B: clamp01(v.Z()), A: 1}
	})
}
