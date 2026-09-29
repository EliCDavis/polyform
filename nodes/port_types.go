package nodes

import (
	"reflect"
	"sort"
	"sync"

	"github.com/EliCDavis/polyform/refutil"
)

var (
	portTypeMu      sync.RWMutex
	portTypeSamples = map[string]OutputPort{}
)

// A sample port of each type is kept because the port is itself a factory
// for more ports of that type.
func DiscoverPortTypes(factory *refutil.TypeFactory) {
	if factory == nil {
		return
	}
	for _, registered := range factory.Types() {
		node, ok := factory.New(registered).(Node)
		if !ok {
			continue
		}
		DiscoverNodePortTypes(node)
	}
}

func DiscoverNodePortTypes(node Node) {
	for _, port := range node.Outputs() {
		if _, ok := port.(ProxyOutputBuilder); !ok {
			continue
		}
		for _, key := range portTypeKeys(port) {
			rememberPortType(key, port)
		}
	}
}

func rememberPortType(typeName string, port OutputPort) {
	if typeName == "" || port == nil {
		return
	}
	if _, ok := port.(ProxyOutputBuilder); !ok {
		return
	}

	portTypeMu.Lock()
	defer portTypeMu.Unlock()

	// A sample that can also build the array form of its type is worth more
	// than one that can't, so it displaces a poorer sample seen earlier.
	if existing, exists := portTypeSamples[typeName]; exists {
		_, had := existing.(DynamicArrayOutputBuilder)
		_, has := port.(DynamicArrayOutputBuilder)
		if had || !has {
			return
		}
	}
	portTypeSamples[typeName] = port
}

func lookupPortType(typeName string) (OutputPort, bool) {
	if typeName == "" {
		return nil, false
	}
	portTypeMu.RLock()
	defer portTypeMu.RUnlock()
	port, ok := portTypeSamples[typeName]
	return port, ok
}

// LookupPortTypeProxy returns a factory for a pass-through port of the given
// value type, if any registered node produces that type.
func LookupPortTypeProxy(typeName string) (ProxyOutputBuilder, bool) {
	port, ok := lookupPortType(typeName)
	if !ok {
		return nil, false
	}
	builder, ok := port.(ProxyOutputBuilder)
	return builder, ok
}

// LookupPortTypeDynamic returns a factory for a runtime-typed port of the
// given value type.
func LookupPortTypeDynamic(typeName string) (DynamicOutputBuilder, bool) {
	port, ok := lookupPortType(typeName)
	if !ok {
		return nil, false
	}
	builder, ok := port.(DynamicOutputBuilder)
	return builder, ok
}

// LookupPortTypeDynamicArray returns a factory for a port carrying an array
// of the given value type.
func LookupPortTypeDynamicArray(typeName string) (DynamicArrayOutputBuilder, bool) {
	port, ok := lookupPortType(typeName)
	if !ok {
		return nil, false
	}
	builder, ok := port.(DynamicArrayOutputBuilder)
	return builder, ok
}

func IsPortTypeKnown(typeName string) bool {
	_, ok := lookupPortType(typeName)
	return ok
}

func KnownPortTypes() []string {
	portTypeMu.RLock()
	defer portTypeMu.RUnlock()
	types := make([]string, 0, len(portTypeSamples))
	for name := range portTypeSamples {
		types = append(types, name)
	}
	sort.Strings(types)
	return types
}

// Value()'s reflected return type is canonical; a port reporting a different
// Type() is indexed under that too, so either spelling finds it.
func portTypeKeys(port OutputPort) []string {
	keys := make([]string, 0, 2)

	if rt, ok := outputValueReturnType(port); ok {
		resolver := refutil.TypeResolution{IncludePackage: true, IncludePointer: false}
		keys = append(keys, resolver.Resolve(reflect.New(rt).Interface()))
	}

	if typed, ok := port.(Typed); ok {
		if key := typed.Type(); key != "" && (len(keys) == 0 || keys[0] != key) {
			keys = append(keys, key)
		}
	}

	return keys
}

// Falls back to Value()'s return type, because a port need not implement
// Typed and a dynamic port wired to one still has to learn its type.
func portValueTypeName(port OutputPort) string {
	if typed, ok := port.(Typed); ok {
		if name := typed.Type(); name != "" {
			return name
		}
	}
	rt, ok := outputValueReturnType(port)
	if !ok {
		return ""
	}
	resolver := refutil.TypeResolution{IncludePackage: true, IncludePointer: false}
	return resolver.Resolve(reflect.New(rt).Interface())
}

func outputValueReturnType(port any) (reflect.Type, bool) {
	t := reflect.TypeOf(port)
	if t == nil {
		return nil, false
	}
	for i := range t.NumMethod() {
		m := t.Method(i)
		if m.Name != "Value" {
			continue
		}
		if m.Type.NumIn() != 1 || m.Type.NumOut() != 1 {
			continue
		}
		return m.Type.Out(0), true
	}
	return nil, false
}
