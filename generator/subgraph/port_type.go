package subgraph

import (
	"reflect"
	"sync"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

var (
	portTypeMu      sync.RWMutex
	portTypeProxies = map[string]nodes.ProxyOutputBuilder{}
)

// DiscoverPortTypes scans every node registered with the factory and indexes
// the value types their output ports produce, so those types can be used as
// sub-graph boundary port types without manual registration.
func DiscoverPortTypes(factory *refutil.TypeFactory) {
	if factory == nil {
		return
	}
	for _, registeredType := range factory.Types() {
		instance := factory.New(registeredType)
		node, ok := instance.(nodes.Node)
		if !ok {
			continue
		}
		DiscoverNodePortTypes(node)
	}
}

// DiscoverNodePortTypes indexes the value types produced by a single node's
// output ports.
func DiscoverNodePortTypes(node nodes.Node) {
	for _, port := range node.Outputs() {
		builder, ok := port.(nodes.ProxyOutputBuilder)
		if !ok {
			continue
		}

		keys := portTypeKeys(port)
		if len(keys) == 0 {
			continue
		}

		portTypeMu.Lock()
		for _, key := range keys {
			portTypeProxies[key] = builder
		}
		portTypeMu.Unlock()
	}
}

// IsPortTypeKnown reports whether boundary ports of the given type can be
// exposed as strongly typed outputs.
func IsPortTypeKnown(portType string) bool {
	if portType == "" {
		return false
	}
	portTypeMu.RLock()
	defer portTypeMu.RUnlock()
	_, ok := portTypeProxies[portType]
	return ok
}

// LookupPortTypeProxy returns a builder capable of constructing a strongly
// typed output port for the given port type, if one has been discovered.
func LookupPortTypeProxy(portType string) (nodes.ProxyOutputBuilder, bool) {
	if portType == "" {
		return nil, false
	}
	portTypeMu.RLock()
	defer portTypeMu.RUnlock()
	builder, ok := portTypeProxies[portType]
	return builder, ok
}

// portTypeKeys derives the canonical type strings the port's produced value
// is known by.
func portTypeKeys(port nodes.OutputPort) []string {
	keys := make([]string, 0, 2)

	if rt, ok := outputValueReturnType(port); ok {
		resolver := refutil.TypeResolution{
			IncludePackage: true,
			IncludePointer: false,
		}
		keys = append(keys, resolver.Resolve(reflect.New(rt).Interface()))
	}

	if typed, ok := port.(nodes.Typed); ok {
		if key := typed.Type(); key != "" && (len(keys) == 0 || keys[0] != key) {
			keys = append(keys, key)
		}
	}

	return keys
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
		mt := m.Type
		if mt.NumIn() != 1 || mt.NumOut() != 1 {
			continue
		}
		return mt.Out(0), true
	}
	return nil, false
}