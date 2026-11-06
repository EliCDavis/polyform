// Package variabletypes is the one table of variable types an app can
// create by name. It lives apart from generator/variable because the value
// types it names (coloring, trs, ...) import generator themselves.
package variabletypes

import (
	"encoding/json"
	"fmt"
	"image"
	"reflect"
	"slices"
	"strings"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/norm"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type entry struct {
	key       string
	build     func() variable.Variable
	valueType reflect.Type
}

func keyOf[T any]() string {
	return refutil.TypeResolution{StripSinglePointer: true}.Resolve(new(T))
}

func typed[T any]() entry {
	valueType := reflect.TypeFor[T]()
	return entry{
		key: keyOf[T](),
		build: func() variable.Variable {
			v := &variable.TypeVariable[T]{}
			if initial, ok := nodes.LiteralInitial(valueType); ok {
				v.SetValue(initial.(T))
			}
			return v
		},
		valueType: valueType,
	}
}

var entries = []entry{
	typed[float64](),
	typed[norm.UFloat64](),
	typed[norm.SFloat64](),
	typed[int](),
	typed[string](),
	typed[bool](),
	typed[vector2.Float64](),
	typed[vector2.Int](),
	typed[vector3.Float64](),
	typed[vector3.Int](),
	typed[vector4.Float64](),
	typed[[]float64](),
	typed[[]int](),
	typed[[]string](),
	typed[[]vector2.Float64](),
	typed[[]vector2.Int](),
	typed[[]vector3.Float64](),
	typed[[]vector3.Int](),
	typed[geometry.AABB](),
	typed[quaternion.Quaternion](),
	typed[trs.TRS](),
	typed[[]trs.TRS](),
	typed[coloring.Color](),
	typed[[]coloring.Color](),
	typed[coloring.Gradient[coloring.Color]](),
	{key: keyOf[image.Image](), build: func() variable.Variable { return &variable.ImageVariable{} }},
	{key: "file", build: func() variable.Variable { return &variable.FileVariable{} }},
}

var registered []entry

func RegisterEnum[E comparable](set enum.Set[E]) {
	key := keyOf[E]()
	for _, e := range all() {
		if strings.EqualFold(e.key, key) {
			panic(fmt.Errorf("variable type %q registered twice", key))
		}
	}
	registered = append(registered, entry{
		key:       key,
		build:     func() variable.Variable { return variable.NewEnumTypeVariable(set) },
		valueType: reflect.TypeFor[E](),
	})
}

func all() []entry {
	return append(slices.Clone(entries), registered...)
}

func KeyOfPortType(portType string) string {
	if enum.RuntimeTypeID(portType) != "" {
		return portType
	}
	outer := portType
	if bracket := strings.Index(portType, "["); bracket != -1 {
		outer = portType[:bracket]
	}
	return portType[strings.LastIndex(outer, "/")+1:]
}

func ParseDefault(key, text string) (json.RawMessage, error) {
	if enum.RuntimeTypeID(key) != "" {
		return json.Marshal(strings.TrimSpace(text))
	}
	for _, e := range all() {
		if !strings.EqualFold(e.key, key) {
			continue
		}
		if e.valueType == nil {
			return nil, fmt.Errorf("a %s has no text form for a default tag", key)
		}
		return nodes.DefaultJSON(text, e.valueType)
	}
	return nil, fmt.Errorf("unsupported variable type %q, expected one of: %s", key, strings.Join(Keys(), ", "))
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
