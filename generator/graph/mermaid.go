package graph

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
)

func sanitizeMermaidName(in string) string {
	if in == "" {
		return ""
	}
	return "[" + strings.ReplaceAll(strings.ReplaceAll(in, "[", "."), "]", "") + "]"
}

func WriteMermaid(a *Instance, out io.Writer) error {
	fmt.Fprintf(out, "---\ntitle: %s\n---\n\nflowchart LR\n", a.details.Name)

	schema := a.Schema()
	for id, n := range schema.Nodes {

		if len(n.AssignedInput) > 0 {
			fmt.Fprintf(out, "\tsubgraph %s%s\n\tdirection LR\n", id, sanitizeMermaidName(n.Name))
			fmt.Fprintf(out, "\tsubgraph %s-In[%s]\n\tdirection TB\n", id, "Input")
		} else {
			fmt.Fprintf(out, "\t%s%s\n", id, sanitizeMermaidName(n.Name))
		}

		for i, name := range slices.Sorted(maps.Keys(n.AssignedInput)) {
			fmt.Fprintf(out, "\t%s-%d(%s)\n", id, i, sanitizeMermaidName(name))
		}

		if len(n.AssignedInput) > 0 {
			fmt.Fprint(out, "\tend\n")
			fmt.Fprint(out, "\tend\n")
		}
	}

	for id, n := range schema.Nodes {
		for i, name := range slices.Sorted(maps.Keys(n.AssignedInput)) {
			fmt.Fprintf(out, "\t%s --> %s-%d\n", n.AssignedInput[name].NodeId, id, i)
		}
	}

	return nil
}
