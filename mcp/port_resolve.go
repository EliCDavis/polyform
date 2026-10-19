package mcp

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
)

// Port names are the space-cased form of a Go field (ColorTexture is
// "Color Texture"), which agents often guess wrong; this tolerates that.
func normalizePortKey(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", ""))
}

// What a port answers to: its key, and its label when it has one.
func portNames(key string, port any) []string {
	if labeled, ok := port.(nodes.Labeled); ok && labeled.Label() != key {
		return []string{key, labeled.Label()}
	}
	return []string{key}
}

// resolvePortName returns the key of the port requested names, by key or
// label, exactly or up to spacing and case. An ambiguous or unknown name
// comes back unchanged, so the graph's own error names the real ports.
func resolvePortName[T any](requested string, ports map[string]T) string {
	if _, ok := ports[requested]; ok {
		return requested
	}
	for key, port := range ports {
		if portNames(key, port)[len(portNames(key, port))-1] == requested {
			return key
		}
	}

	target := normalizePortKey(requested)
	match := ""
	for key, port := range ports {
		for _, name := range portNames(key, port) {
			if normalizePortKey(name) != target {
				continue
			}
			if match != "" && match != key {
				return requested
			}
			match = key
		}
	}
	if match == "" {
		return requested
	}
	return match
}

// resolveOutputPort is resolvePortName plus a unique-prefix fallback
// ("Float" for HalfNode's "Float 64").
func resolveOutputPort[T any](requested string, ports map[string]T) string {
	resolved := resolvePortName(requested, ports)
	if _, ok := ports[resolved]; ok {
		return resolved
	}

	target := normalizePortKey(requested)
	match := ""
	for key, port := range ports {
		for _, name := range portNames(key, port) {
			if !strings.HasPrefix(normalizePortKey(name), target) {
				continue
			}
			if match != "" && match != key {
				match = ""
				goto generic
			}
			match = key
		}
	}
	if match != "" {
		return match
	}

generic:
	// "Out", "Float 64", "Result" all mean the node's main output.
	if genericOutputNames[target] {
		match = ""
		for key, port := range ports {
			name := portNames(key, port)
			if !genericOutputNames[normalizePortKey(name[len(name)-1])] {
				continue
			}
			if match != "" {
				return requested
			}
			match = key
		}
		if match != "" {
			return match
		}
	}
	return requested
}

var genericOutputNames = map[string]bool{
	"out": true, "float": true, "float64": true, "result": true, "value": true,
}

func resolveInputPortName(node nodes.Node, requested string) string {
	if node == nil {
		return requested
	}
	return resolvePortName(requested, node.Inputs())
}

func resolveOutputPortName(node nodes.Node, requested string) string {
	if node == nil {
		return requested
	}
	return resolveOutputPort(requested, node.Outputs())
}

// A trailing ".N" element index is kept; only the port name is resolved.
func resolveInputPortNameWithIndex(node nodes.Node, requested string) string {
	base, index, hasIndex := splitPortIndex(node, requested)
	if !hasIndex {
		return resolveInputPortName(node, requested)
	}
	return resolveInputPortName(node, base) + "." + strconv.Itoa(index)
}

// "Port.@nodeId" names the element nodeId currently feeds, so an element
// can be swapped without first reading the array's order.
func resolveInputPortByFeeder(inst *graph.Graph, node nodes.Node, requested string) (string, error) {
	base, suffix, found := strings.Cut(requested, ".@")
	if !found || node == nil {
		return resolveInputPortNameWithIndex(node, requested), nil
	}
	resolved := resolveInputPortName(node, base)
	port, ok := node.Inputs()[resolved]
	if !ok {
		return "", fmt.Errorf("no input port %q", base)
	}
	array, isArray := port.(nodes.ArrayValueInputPort)
	if !isArray {
		return "", fmt.Errorf("port %q is not an array, so %q has no element to address", portLabel(port), requested)
	}
	for i, upstream := range array.Value() {
		if upstream != nil && inst.NodeId(upstream.Node()) == suffix {
			return fmt.Sprintf("%s.%d", resolved, i), nil
		}
	}
	return "", fmt.Errorf("no element of %q is fed by node %q", portLabel(port), suffix)
}

func portLabel(port nodes.Port) string {
	if labeled, ok := port.(nodes.Labeled); ok {
		return labeled.Label()
	}
	return port.Name()
}

func sortedOutputNames(outputs map[string]nodes.OutputPort) []string {
	return sortedLabels(outputs)
}

func sortedOutputKeys(outputs map[string]nodes.OutputPort) []string {
	return slices.Sorted(maps.Keys(outputs))
}

func sortedPortNames(ports map[string]nodes.InputPort) []string {
	return sortedLabels(ports)
}

func sortedLabels[P nodes.Port](ports map[string]P) []string {
	names := make([]string, 0, len(ports))
	for _, port := range ports {
		names = append(names, portLabel(port))
	}
	slices.Sort(names)
	return names
}
