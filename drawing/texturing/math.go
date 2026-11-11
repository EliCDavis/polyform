package texturing

import (
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
)

func combine[T any](out *nodes.StructOutput[Texture[T]], textures []Texture[T], f func(values []T) T) {
	if len(textures) == 0 {
		return
	}
	if !resolutionsMatch(textures) {
		out.CaptureError(ErrMismatchDimensions)
		return
	}
	result := Empty[T](textures[0].width, textures[0].height)
	result.MutateParallel(func(x, y int, _ T) T {
		values := make([]T, len(textures))
		for i, tex := range textures {
			values[i] = tex.Get(x, y)
		}
		return f(values)
	})
	out.Set(result)
}

func mapTex[T any](out *nodes.StructOutput[Texture[T]], in nodes.Output[Texture[T]], f func(T) T) {
	if in == nil {
		return
	}
	out.Set(Map(nodes.GetOutputValue(out, in), f))
}

// ============================================================================

type SubtractFloat1Node struct {
	A nodes.Output[Texture[float64]]
	B nodes.Output[Texture[float64]]
}

func (n SubtractFloat1Node) Description() string {
	return "A - B, pixel by pixel."
}

func (n SubtractFloat1Node) Result(out *nodes.StructOutput[Texture[float64]]) {
	if n.A == nil || n.B == nil {
		return
	}
	combine(out, []Texture[float64]{nodes.GetOutputValue(out, n.A), nodes.GetOutputValue(out, n.B)}, func(v []float64) float64 {
		return v[0] - v[1]
	})
}

// ============================================================================

type MinFloat1Node struct {
	Textures []nodes.Output[Texture[float64]]
}

func (n MinFloat1Node) Description() string {
	return "The smallest value of the textures at each pixel."
}

func (n MinFloat1Node) Result(out *nodes.StructOutput[Texture[float64]]) {
	combine(out, nodes.GetOutputValues(out, n.Textures), func(v []float64) float64 {
		m := v[0]
		for _, x := range v[1:] {
			m = math.Min(m, x)
		}
		return m
	})
}

type MaxFloat1Node struct {
	Textures []nodes.Output[Texture[float64]]
}

func (n MaxFloat1Node) Description() string {
	return "The largest value of the textures at each pixel."
}

func (n MaxFloat1Node) Result(out *nodes.StructOutput[Texture[float64]]) {
	combine(out, nodes.GetOutputValues(out, n.Textures), func(v []float64) float64 {
		m := v[0]
		for _, x := range v[1:] {
			m = math.Max(m, x)
		}
		return m
	})
}

// ============================================================================

type RemapFloat1Node struct {
	Texture nodes.Output[Texture[float64]]
	InMin   nodes.Output[float64]
	InMax   nodes.Output[float64] `default:"1"`
	OutMin  nodes.Output[float64]
	OutMax  nodes.Output[float64] `default:"1"`
	Clamp   nodes.Output[bool]    `description:"Keep results between Out Min and Out Max."`
}

func (n RemapFloat1Node) Description() string {
	return "Maps every value from the In range onto the Out range, so In Min becomes Out Min and In Max becomes Out Max."
}

func (n RemapFloat1Node) Result(out *nodes.StructOutput[Texture[float64]]) {
	inMin, inMax := nodes.GetOutputValue(out, n.InMin), nodes.GetOutputValue(out, n.InMax)
	outMin, outMax := nodes.GetOutputValue(out, n.OutMin), nodes.GetOutputValue(out, n.OutMax)
	clamp := nodes.GetOutputValue(out, n.Clamp)
	span := inMax - inMin
	mapTex(out, n.Texture, func(v float64) float64 {
		t := 0.
		if span != 0 {
			t = (v - inMin) / span
		}
		if clamp {
			t = math.Min(math.Max(t, 0), 1)
		}
		return outMin + t*(outMax-outMin)
	})
}

// ============================================================================

type ClampFloat1Node struct {
	Texture nodes.Output[Texture[float64]]
	Min     nodes.Output[float64]
	Max     nodes.Output[float64] `default:"1"`
}

func (n ClampFloat1Node) Description() string {
	return "Limits every value to between Min and Max."
}

func (n ClampFloat1Node) Result(out *nodes.StructOutput[Texture[float64]]) {
	lo, hi := nodes.GetOutputValue(out, n.Min), nodes.GetOutputValue(out, n.Max)
	mapTex(out, n.Texture, func(v float64) float64 {
		return math.Min(math.Max(v, lo), hi)
	})
}

// ============================================================================

type PowerFloat1Node struct {
	Texture  nodes.Output[Texture[float64]]
	Exponent nodes.Output[float64] `default:"1"`
}

func (n PowerFloat1Node) Description() string {
	return "Raises every value to Exponent. Above 1 pulls 0-1 values toward 0, sharpening peaks; below 1 pushes them toward 1."
}

func (n PowerFloat1Node) Result(out *nodes.StructOutput[Texture[float64]]) {
	exponent := nodes.GetOutputValue(out, n.Exponent)
	mapTex(out, n.Texture, func(v float64) float64 {
		return math.Pow(v, exponent)
	})
}

// ============================================================================

type SmoothstepFloat1Node struct {
	Texture nodes.Output[Texture[float64]]
	Low     nodes.Output[float64] `description:"At or below this the result is 0."`
	High    nodes.Output[float64] `description:"At or above this the result is 1." default:"1"`
}

func (n SmoothstepFloat1Node) Description() string {
	return "A soft threshold: 0 below Low, 1 above High, an eased ramp between. Low equal to High is a hard cut."
}

func (n SmoothstepFloat1Node) Result(out *nodes.StructOutput[Texture[float64]]) {
	lo, hi := nodes.GetOutputValue(out, n.Low), nodes.GetOutputValue(out, n.High)
	mapTex(out, n.Texture, func(v float64) float64 {
		if hi <= lo {
			if v < lo {
				return 0
			}
			return 1
		}
		t := math.Min(math.Max((v-lo)/(hi-lo), 0), 1)
		return t * t * (3 - 2*t)
	})
}

// ============================================================================

func mix[T any](out *nodes.StructOutput[Texture[T]], a, b nodes.Output[Texture[T]], factor nodes.Output[Texture[float64]], amount nodes.Output[float64], lerp func(a, b T, t float64) T) {
	if a == nil || b == nil {
		return
	}
	ta, tb := nodes.GetOutputValue(out, a), nodes.GetOutputValue(out, b)
	if !resolutionsMatch([]Texture[T]{ta, tb}) {
		out.CaptureError(ErrMismatchDimensions)
		return
	}

	result := Empty[T](ta.width, ta.height)
	if factor == nil {
		t := nodes.GetOutputValue(out, amount)
		result.MutateParallel(func(x, y int, _ T) T { return lerp(ta.Get(x, y), tb.Get(x, y), t) })
		out.Set(result)
		return
	}

	tf := nodes.GetOutputValue(out, factor)
	if tf.width != ta.width || tf.height != ta.height {
		out.CaptureError(ErrMismatchDimensions)
		return
	}
	result.MutateParallel(func(x, y int, _ T) T { return lerp(ta.Get(x, y), tb.Get(x, y), tf.Get(x, y)) })
	out.Set(result)
}

type MixFloat1Node struct {
	A      nodes.Output[Texture[float64]]
	B      nodes.Output[Texture[float64]]
	Factor nodes.Output[Texture[float64]] `description:"Per-pixel blend, 0 for A and 1 for B. Replaces Amount when wired."`
	Amount nodes.Output[float64]          `description:"Blend everywhere, 0 for A and 1 for B." default:"0.5"`
}

func (n MixFloat1Node) Description() string {
	return "Blends texture A toward texture B, by one Amount or a Factor texture."
}

func (n MixFloat1Node) Result(out *nodes.StructOutput[Texture[float64]]) {
	mix(out, n.A, n.B, n.Factor, n.Amount, func(a, b, t float64) float64 { return a + (b-a)*t })
}

type MixColorNode struct {
	A      nodes.Output[Texture[coloring.Color]]
	B      nodes.Output[Texture[coloring.Color]]
	Factor nodes.Output[Texture[float64]] `description:"Per-pixel blend, 0 for A and 1 for B. Replaces Amount when wired."`
	Amount nodes.Output[float64]          `description:"Blend everywhere, 0 for A and 1 for B." default:"0.5"`
}

func (n MixColorNode) Description() string {
	return "Blends color texture A toward color texture B, by one Amount or a Factor texture."
}

func (n MixColorNode) Result(out *nodes.StructOutput[Texture[coloring.Color]]) {
	mix(out, n.A, n.B, n.Factor, n.Amount, coloring.Color.Lerp)
}
