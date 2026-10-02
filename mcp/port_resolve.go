package mcp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
)

// normalizePortKey collapses a port name down to a form tolerant of the
// one detail agents most often get wrong when guessing a port name instead
// of confirming it via get_node_types: port names are the CamelCase-to-
// space-case form of a Go struct field (see utils.CamelCaseToSpaceCase),
// which inserts a space before every internal capital letter and every
// digit — e.g. the Go field ColorTexture is the port "Color Texture", and
// Radius2 is "Radius 2", not the raw field name.
func normalizePortKey(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", ""))
}

// resolvePortName returns the key in ports matching requested, tolerating
// whitespace/case differences from the real port name. An exact match is
// always preferred and returned immediately. If there's no exact match,
// it falls back to a normalized comparison; if exactly one port matches,
// its real name is returned. If zero or more than one port matches after
// normalizing, requested is returned unchanged, so the existing "no such
// port" error from the graph layer still applies rather than silently
// guessing at an ambiguous or genuinely-wrong name.
func resolvePortName[T any](requested string, ports map[string]T) string {
	if _, ok := ports[requested]; ok {
		return requested
	}

	target := normalizePortKey(requested)
	match := ""
	for name := range ports {
		if normalizePortKey(name) == target {
			if match != "" {
				return requested // ambiguous after normalizing - don't guess
			}
			match = name
		}
	}
	if match == "" {
		return requested
	}
	return match
}

// resolveOutputPort is resolvePortName plus a unique-prefix fallback
// ("Float" for HalfNode's "Float 64"), since the math package names the
// same output differently from node to node. A wrong name with no unique
// prefix still passes through so the graph's own error lists the real
// ports.
func resolveOutputPort[T any](requested string, ports map[string]T) string {
	resolved := resolvePortName(requested, ports)
	if _, ok := ports[resolved]; ok {
		return resolved
	}

	target := normalizePortKey(requested)
	match := ""
	for name := range ports {
		if strings.HasPrefix(normalizePortKey(name), target) {
			if match != "" {
				match = ""
				break
			}
			match = name
		}
	}
	if match != "" {
		return match
	}

	// "Out", "Float 64", "Result" all mean "the node's main output"; when
	// the caller asked for one such name and the node has exactly one
	// output with another, that is the port they meant.
	if genericOutputNames[target] {
		match = ""
		for name := range ports {
			if !genericOutputNames[normalizePortKey(name)] {
				continue
			}
			if match != "" {
				return requested
			}
			match = name
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

// resolveInputPortNameWithIndex is resolveInputPortName for a port
// spelling that may carry a trailing array-element index (e.g. "Items.0",
// per disconnect's own port syntax) - the ".N" suffix is preserved as-is,
// only the port name itself is resolved.
func resolveInputPortNameWithIndex(node nodes.Node, requested string) string {
	base, suffix, hasIndex := strings.Cut(requested, ".")
	if !hasIndex {
		return resolveInputPortName(node, requested)
	}
	if _, err := strconv.Atoi(suffix); err != nil {
		// Not actually a ".N" index (e.g. the port name itself contains a
		// literal dot) - resolve the whole thing as one name.
		return resolveInputPortName(node, requested)
	}
	return resolveInputPortName(node, base) + "." + suffix
}

// resolveInputPortByFeeder turns "Port.@nodeId" into "Port.N" where N is
// the element currently fed by nodeId, so an element can be swapped
// without first reading the array's compacted order. Any other spelling
// is returned through resolveInputPortNameWithIndex unchanged.
func resolveInputPortByFeeder(inst *graph.Instance, node nodes.Node, requested string) (string, error) {
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
		return "", fmt.Errorf("port %q is not an array, so %q has no element to address", resolved, requested)
	}
	for i, upstream := range array.Value() {
		if upstream != nil && inst.NodeId(upstream.Node()) == suffix {
			return fmt.Sprintf("%s.%d", resolved, i), nil
		}
	}
	return "", fmt.Errorf("no element of %q is fed by node %q", resolved, suffix)
}
