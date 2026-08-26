package sdf

import (
	"errors"
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

// ColorField is a color at every point in space, the color-channel
// equivalent of sample.Vec3ToFloat.
type ColorField func(vector3.Float64) coloring.Color

// ConstantColor builds a ColorField that returns the same color
// everywhere - the common case for a single primitive.
func ConstantColor(c coloring.Color) ColorField {
	return func(vector3.Float64) coloring.Color {
		return c
	}
}

// ColoredField pairs a distance field with a color field over the same
// space, so a union combinator can blend both consistently using the same
// "how far into the transition zone am I" weight for each.
type ColoredField struct {
	Distance sample.Vec3ToFloat
	Color    ColorField
}

// SmoothUnionColored combines two or more ColoredFields the same way
// SmoothUnion combines their bare distance fields, and blends color using
// the identical blend weight the distance union computes - so color
// transitions over exactly the same physical region the geometry does,
// instead of snapping at a hard edge or drifting independently of it.
//
// For 3+ fields, only the two nearest at a given point are blended - the
// same nearest-two rule SmoothUnion's own N-ary case uses - so a point
// deep inside one field is unaffected by a color on the far side of the
// model, no matter how many other fields are in the union.
func SmoothUnionColored(radius float64, fields ...ColoredField) ColoredField {
	if len(fields) == 0 {
		panic("no fields to union")
	}
	if len(fields) == 1 {
		return fields[0]
	}

	distances := make([]sample.Vec3ToFloat, len(fields))
	for i, f := range fields {
		distances[i] = f.Distance
	}
	distanceField := SmoothUnion(radius, distances...)

	colorField := func(v vector3.Float64) coloring.Color {
		min1, min2 := math.Inf(1), math.Inf(1)
		idx1, idx2 := 0, 0
		for i, f := range fields {
			d := f.Distance(v)
			if d < min1 {
				min1, min2 = d, min1
				idx2 = idx1
				idx1 = i
			} else if d < min2 {
				min2 = d
				idx2 = i
			}
		}

		colorA := fields[idx1].Color(v)
		if radius <= 0 {
			return colorA // hard union: the nearer field's color wins outright
		}

		colorB := fields[idx2].Color(v)
		h := math.Max(radius-math.Abs(min1-min2), 0) / radius
		// h=0 far from the boundary (idx1 clearly nearer) -> pure colorA.
		// h=1 exactly on the boundary (min1==min2) -> an even 50/50 blend.
		return colorA.Lerp(colorB, h*0.5)
	}

	return ColoredField{Distance: distanceField, Color: colorField}
}

// UnionColored is the hard-union equivalent of SmoothUnionColored: at
// each point, whichever field is nearest wins outright - both its
// distance and its color - with a crisp seam and no blending, the same
// relationship Union has to SmoothUnion for plain fields.
func UnionColored(fields ...ColoredField) ColoredField {
	if len(fields) == 0 {
		panic("no fields to union")
	}
	if len(fields) == 1 {
		return fields[0]
	}

	distances := make([]sample.Vec3ToFloat, len(fields))
	for i, f := range fields {
		distances[i] = f.Distance
	}
	distanceField := Union(distances...)

	colorField := func(v vector3.Float64) coloring.Color {
		minDist := math.Inf(1)
		winner := 0
		for i, f := range fields {
			d := f.Distance(v)
			if d < minDist {
				minDist = d
				winner = i
			}
		}
		return fields[winner].Color(v)
	}

	return ColoredField{Distance: distanceField, Color: colorField}
}

type WithColorNode struct {
	Field nodes.Output[sample.Vec3ToFloat] `description:"The field to attach a color to."`
	Color nodes.Output[coloring.Color]     `description:"The color given to every point in Field."`
}

func (n WithColorNode) Description() string {
	return "Pairs an existing field with a constant color, so it can be combined with other colored fields via SmoothUnionColoredNode."
}

func (n WithColorNode) Out(out *nodes.StructOutput[ColoredField]) {
	field := nodes.TryGetOutputValue(out, n.Field, nullField)
	color := nodes.TryGetOutputValue(out, n.Color, coloring.Color{A: 1})
	out.Set(ColoredField{
		Distance: field,
		Color:    ConstantColor(color),
	})
}

type SmoothUnionColoredNode struct {
	Fields []nodes.Output[ColoredField] `description:"The colored fields to combine."`
	Radius nodes.Output[float64]        `description:"Width of the blend region in world units, same meaning as SmoothUnionNode's Radius. Zero or less is a hard union: color follows whichever field is nearer outright, with no blending. Defaults to 0.1."`
}

func (n SmoothUnionColoredNode) Description() string {
	return "Combines two or more colored fields, blending both their shapes and their colors smoothly where they meet, using the same blend weight for each."
}

func (n SmoothUnionColoredNode) Union(out *nodes.StructOutput[ColoredField]) {
	fields := nodes.GetOutputValues(out, n.Fields)
	if len(fields) == 0 {
		out.CaptureError(errors.New("No fields provided to union"))
		return
	}
	out.Set(SmoothUnionColored(nodes.TryGetOutputValue(out, n.Radius, .1), fields...))
}

type UnionColoredNode struct {
	Fields []nodes.Output[ColoredField] `description:"The colored fields to combine."`
}

func (n UnionColoredNode) Description() string {
	return "Boolean union of two or more colored fields - at each point, whichever field is nearest wins outright, both its shape and its color, with a crisp seam and no blending where they meet."
}

func (n UnionColoredNode) Union(out *nodes.StructOutput[ColoredField]) {
	fields := nodes.GetOutputValues(out, n.Fields)
	if len(fields) == 0 {
		out.CaptureError(errors.New("No fields provided to union"))
		return
	}
	out.Set(UnionColored(fields...))
}

type ColoredFieldDistanceNode struct {
	Field nodes.Output[ColoredField] `description:"The colored field to extract the plain distance field from."`
}

func (n ColoredFieldDistanceNode) Description() string {
	return "Extracts the plain distance field from a colored field, for wiring into MarchNode - a colored field's own color only becomes usable after marching, via ApplyColorFieldNode."
}

func (n ColoredFieldDistanceNode) Out(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	field := nodes.TryGetOutputValue(out, n.Field, ColoredField{})
	out.Set(field.Distance)
}
