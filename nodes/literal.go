package nodes

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/EliCDavis/polyform/utils"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type literalType struct {
	initial any
	parse   func(text string) (any, error)
	port    func(node Node, name string, value any) OutputPort
}

var (
	literalMu     sync.RWMutex
	literalTypes  = map[reflect.Type]literalType{}
	literalByName = map[string]reflect.Type{}
)

func RegisterLiteralType[T any](initial T) {
	registerLiteral(initial, nil)
}

func RegisterLiteralParser[T any](initial T, parse func(text string) (T, error)) {
	registerLiteral(initial, func(text string) (any, error) { return parse(text) })
}

func registerLiteral[T any](initial T, parse func(string) (any, error)) {
	t := reflect.TypeOf(initial)
	literalMu.Lock()
	defer literalMu.Unlock()
	literalTypes[t] = literalType{
		initial: initial,
		parse:   parse,
		port: func(node Node, name string, value any) OutputPort {
			return ConstOutput[T]{Ref: node, Val: value.(T), PortName: name}
		},
	}
	literalByName[typeKey(t)] = t
}

// Matches how reflect spells a type argument: import path, not package name.
func typeKey(t reflect.Type) string {
	if t.PkgPath() == "" {
		return t.Name()
	}
	return t.PkgPath() + "." + t.Name()
}

func init() {
	RegisterLiteralType(float64(0))
	RegisterLiteralType(float32(0))
	RegisterLiteralType(int(0))
	RegisterLiteralType(int32(0))
	RegisterLiteralType(int64(0))
	RegisterLiteralType(false)
	RegisterLiteralType("")
	RegisterLiteralType(vector2.Zero[float64]())
	RegisterLiteralType(vector2.Zero[int]())
	RegisterLiteralType(vector3.Zero[float64]())
	RegisterLiteralType(vector3.Zero[int]())
	RegisterLiteralType(vector4.Zero[float64]())
}

func literalOf(t reflect.Type) (literalType, bool) {
	literalMu.RLock()
	defer literalMu.RUnlock()
	lt, ok := literalTypes[t]
	return lt, ok
}

func LiteralInitial(t reflect.Type) (any, bool) {
	lt, ok := literalOf(t)
	if !ok {
		return nil, false
	}
	return lt.initial, true
}

func DefaultJSON(text string, t reflect.Type) (json.RawMessage, error) {
	text = strings.TrimSpace(text)
	if lt, ok := literalOf(t); ok && lt.parse != nil {
		return json.Marshal(text)
	}

	switch t.Kind() {
	case reflect.Float32, reflect.Float64:
		if bound, ok := namedBound(text, t.Kind(), t.Bits()); ok {
			return json.Marshal(bound)
		}
		v, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil, fmt.Errorf("default %q is not a number", text)
		}
		return json.Marshal(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if bound, ok := namedBound(text, t.Kind(), t.Bits()); ok {
			return json.Marshal(bound)
		}
		v, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("default %q is not an integer", text)
		}
		return json.Marshal(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if bound, ok := namedBound(text, t.Kind(), t.Bits()); ok {
			return json.Marshal(bound)
		}
		v, err := strconv.ParseUint(text, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("default %q is not an unsigned integer", text)
		}
		return json.Marshal(v)
	case reflect.Bool:
		if text != "true" && text != "false" {
			return nil, fmt.Errorf("default %q is not true or false", text)
		}
		return json.RawMessage(text), nil
	case reflect.String:
		return json.Marshal(text)
	}

	name := t.String()
	switch {
	case name == "coloring.Color":
		if !isHexColor(text) {
			return nil, fmt.Errorf("default %q is not a #rgb, #rgba, #rrggbb or #rrggbbaa color", text)
		}
		return json.Marshal(text)
	case strings.HasPrefix(name, "vector2.Vector["):
		return components(text, strings.HasSuffix(name, "[int]"), "x", "y")
	case strings.HasPrefix(name, "vector3.Vector["):
		return components(text, strings.HasSuffix(name, "[int]"), "x", "y", "z")
	case strings.HasPrefix(name, "vector4.Vector["), name == "quaternion.Quaternion":
		return components(text, false, "x", "y", "z", "w")
	case name == "geometry.AABB":
		center, size, ok := strings.Cut(text, ";")
		if !ok {
			return nil, fmt.Errorf("default %q needs a center and a size separated by a semicolon", text)
		}
		centerJSON, err := components(center, false, "x", "y", "z")
		if err != nil {
			return nil, err
		}
		sizeJSON, err := components(size, false, "x", "y", "z")
		if err != nil {
			return nil, err
		}
		var extents map[string]float64
		if err := json.Unmarshal(sizeJSON, &extents); err != nil {
			return nil, err
		}
		for axis := range extents {
			extents[axis] /= 2
		}
		extentsJSON, err := json.Marshal(extents)
		if err != nil {
			return nil, err
		}
		return json.RawMessage(fmt.Sprintf(`{"center":%s,"extents":%s}`, centerJSON, extentsJSON)), nil
	}
	return nil, fmt.Errorf("a %s has no text form for a default tag", name)
}

func namedBound(text string, kind reflect.Kind, bits int) (any, bool) {
	isMax := strings.EqualFold(text, "max")
	if !isMax && !strings.EqualFold(text, "min") {
		return nil, false
	}
	switch kind {
	case reflect.Float32, reflect.Float64:
		largest := math.MaxFloat64
		if bits == 32 {
			largest = math.MaxFloat32
		}
		if isMax {
			return largest, true
		}
		return -largest, true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		largest := int64(1)<<(bits-1) - 1
		if isMax {
			return largest, true
		}
		return -largest - 1, true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if isMax {
			return ^uint64(0) >> (64 - bits), true
		}
		return uint64(0), true
	}
	return nil, false
}

func components(text string, integer bool, names ...string) (json.RawMessage, error) {
	parts := strings.Split(text, ",")
	if len(parts) != len(names) {
		return nil, fmt.Errorf("default %q needs %d comma-separated components", text, len(names))
	}
	kind, bits := reflect.Float64, 64
	if integer {
		kind, bits = reflect.Int, strconv.IntSize
	}
	parsed := make(map[string]any, len(names))
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if bound, ok := namedBound(part, kind, bits); ok {
			parsed[names[i]] = bound
			continue
		}
		if integer {
			v, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("default %q: component %q is not an integer", text, part)
			}
			parsed[names[i]] = v
			continue
		}
		v, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, fmt.Errorf("default %q: component %q is not a number", text, part)
		}
		parsed[names[i]] = v
	}
	return json.Marshal(parsed)
}

func isHexColor(text string) bool {
	switch len(text) {
	case 4, 5, 7, 9:
	default:
		return false
	}
	if text[0] != '#' {
		return false
	}
	_, err := strconv.ParseUint(text[1:], 16, 64)
	return err == nil
}

func defaultValue(text string, t reflect.Type) (any, error) {
	if lt, ok := literalOf(t); ok && lt.parse != nil {
		return lt.parse(strings.TrimSpace(text))
	}
	raw, err := DefaultJSON(text, t)
	if err != nil {
		return nil, err
	}
	ptr := reflect.New(t)
	if err := json.Unmarshal(raw, ptr.Interface()); err != nil {
		return nil, fmt.Errorf("default %q: %w", text, err)
	}
	return ptr.Elem().Interface(), nil
}

// A lifted port names T only in its field type's name, so T is looked up by it.
func portElemType(field reflect.StructField) (reflect.Type, bool) {
	t := field.Type
	if t.Kind() != reflect.Interface {
		return nil, false
	}
	if m, ok := t.MethodByName("Value"); ok && m.Type.NumIn() == 0 && m.Type.NumOut() == 1 {
		return m.Type.Out(0), true
	}
	elem, ok := strings.CutPrefix(t.String(), liftedPortFieldType+"[")
	if !ok {
		return nil, false
	}
	literalMu.RLock()
	defer literalMu.RUnlock()
	rt, ok := literalByName[strings.TrimSuffix(elem, "]")]
	return rt, ok
}

func (s *Struct[T]) fillDefaults() {
	data := reflect.ValueOf(&s.Data).Elem()
	dataType := data.Type()
	filled := false
	for i := range dataType.NumField() {
		field := dataType.Field(i)
		value := data.Field(i)
		if value.Kind() != reflect.Interface || !value.IsNil() || !value.CanSet() {
			continue
		}
		elem, ok := portElemType(field)
		if !ok {
			continue
		}
		lt, ok := literalOf(elem)
		if !ok {
			continue
		}
		held := lt.initial
		if text := field.Tag.Get("default"); text != "" {
			parsed, err := defaultValue(text, elem)
			if err != nil {
				panic(fmt.Errorf("%s.%s: default tag: %w", dataType, field.Name, err))
			}
			held = parsed
		}
		port := lt.port(s, utils.CamelCaseToSpaceCase(field.Name), held)
		if !reflect.TypeOf(port).AssignableTo(field.Type) {
			continue
		}
		value.Set(reflect.ValueOf(port))
		filled = true
	}
	if filled {
		s.refreshInputVersions()
	}
}

type defaultFiller interface {
	fillDefaults()
}

func fillDefaults(node Node) {
	if filler, ok := node.(defaultFiller); ok {
		filler.fillDefaults()
	}
}
