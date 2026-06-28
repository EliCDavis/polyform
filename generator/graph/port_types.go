package graph

import (
	"sort"

	"github.com/EliCDavis/polyform/generator/schema"
)

func CollectAllPortTypes(nodeTypes []schema.NodeType) []string {
	typeSet := make(map[string]struct{})

	for _, nodeType := range nodeTypes {
		for _, input := range nodeType.Inputs {
			if input.Type != "" && input.Type != "any" {
				typeSet[input.Type] = struct{}{}
			}
		}
		for _, output := range nodeType.Outputs {
			if output.Type != "" && output.Type != "any" {
				typeSet[output.Type] = struct{}{}
			}
		}
	}

	types := make([]string, 0, len(typeSet))
	for t := range typeSet {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}
