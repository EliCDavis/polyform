// Package variabletypes is the one table of variable types an app can
// create by name. It lives apart from generator/variable because the value
// types it names (coloring, trs, ...) import generator themselves.
package variabletypes

import (
	"fmt"
	"slices"
	"strings"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type entry struct {
	key   string
	build func() variable.Variable
}

func typed[T any]() func() variable.Variable {
	return func() variable.Variable { return &variable.TypeVariable[T]{} }
}

func typedWith[T any](initial T) func() variable.Variable {
	return func() variable.Variable {
		v := &variable.TypeVariable[T]{}
		v.SetValue(initial)
		return v
	}
}

// Keys are matched case-insensitively; they are the type names refutil
// resolves for each value type, which is also what the editor sends.
var entries = []entry{
	{"float64", typed[float64]()},
	{"int", typed[int]()},
	{"string", typed[string]()},
	{"bool", typed[bool]()},
	{"vector2.Vector[float64]", typed[vector2.Float64]()},
	{"vector2.Vector[int]", typed[vector2.Int]()},
	{"vector3.Vector[float64]", typed[vector3.Float64]()},
	{"vector3.Vector[int]", typed[vector3.Int]()},
	{"vector4.Vector[float64]", typed[vector4.Float64]()},
	{"[]float64", typed[[]float64]()},
	{"[]int", typed[[]int]()},
	{"[]string", typed[[]string]()},
	{"[]vector2.Vector[float64]", typed[[]vector2.Float64]()},
	{"[]vector2.Vector[int]", typed[[]vector2.Int]()},
	{"[]vector3.Vector[float64]", typed[[]vector3.Float64]()},
	{"[]vector3.Vector[int]", typed[[]vector3.Int]()},
	{"geometry.AABB", typed[geometry.AABB]()},
	{"quaternion.Quaternion", typedWith(quaternion.Identity())},
	{"trs.TRS", typedWith(trs.Identity())},
	{"[]trs.TRS", typed[[]trs.TRS]()},
	{"coloring.Color", typed[coloring.Color]()},
	{"[]coloring.Color", typed[[]coloring.Color]()},
	{"coloring.Gradient[github.com/EliCDavis/polyform/drawing/coloring.Color]", typedWith(coloring.NewGradientColor(
		coloring.GradientKey[coloring.Color]{Time: 0, Value: coloring.Black()},
		coloring.GradientKey[coloring.Color]{Time: 1, Value: coloring.White()},
	))},
	{"image.Image", func() variable.Variable { return &variable.ImageVariable{} }},
	{"file", func() variable.Variable { return &variable.FileVariable{} }},
}

var registered []entry

func RegisterEnum[E comparable](set enum.Set[E]) {
	key := refutil.TypeResolution{}.Resolve(*new(E))
	for _, e := range all() {
		if strings.EqualFold(e.key, key) {
			panic(fmt.Errorf("variable type %q registered twice", key))
		}
	}
	registered = append(registered, entry{key, func() variable.Variable { return variable.NewEnumTypeVariable(set) }})
}

func all() []entry {
	return append(slices.Clone(entries), registered...)
}

func Keys() []string {
	all := all()
	keys := make([]string, len(all))
	for i, e := range all {
		keys[i] = e.key
	}
	return keys
}

func New(typeKey string) (variable.Variable, error) {
	for _, e := range all() {
		if strings.EqualFold(e.key, typeKey) {
			return e.build(), nil
		}
	}
	return nil, fmt.Errorf("unsupported variable type %q, expected one of: %s", typeKey, strings.Join(Keys(), ", "))
}
