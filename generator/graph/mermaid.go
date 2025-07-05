package graph

import (
	"fmt"
	"io"
	"strings"
)

func sanitizeMermaidName(in string) string {
	if in == "" {
		return ""
	}
	return "[" + strings.ReplaceAll(strings.ReplaceAll(in, "[", "."), "]", "") + "]"
}

func (a *Instance) WriteMermaid(out io.Writer) error {
	fmt.Fprintf(out, "---\ntitle: %s\n---\n\nflowchart LR\n", a.details.Name)

	schema := a.Schema()
	for id, n := range schema.Nodes {

		if len(n.AssignedInput) > 0 {
			fmt.Fprintf(out, "\tsubgraph %s%s\n\tdirection LR\n", id, sanitizeMermaidName(n.Name))
			fmt.Fprintf(out, "\tsubgraph %s-In[%s]\n\tdirection TB\n", id, "Input")
		} else {
			fmt.Fprintf(out, "\t%s%s\n", id, sanitizeMermaidName(n.Name))
		}

		depIndex := 0
		for name := range n.AssignedInput {
			fmt.Fprintf(out, "\t%s-%d(%s)\n", id, depIndex, sanitizeMermaidName(name))
			depIndex++
		}

		if len(n.AssignedInput) > 0 {
			fmt.Fprint(out, "\tend\n")
			fmt.Fprint(out, "\tend\n")
		}
	}

	for id, n := range schema.Nodes {
		depIndex := 0
		for _, d := range n.AssignedInput {
			fmt.Fprintf(out, "\t%s --> %s-%d\n", d.NodeId, id, depIndex)
		}
	}

	return nil
}
