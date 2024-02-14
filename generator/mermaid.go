package generator

import (
	"fmt"
	"io"
)

func WriteMermaid(a App, out io.Writer) error {

	schema := a.Schema()

	fmt.Fprintf(out, "---\ntitle: %s\n---\n\nflowchart LR\n", a.Name)

	for id, n := range schema.Nodes {

		if len(n.Dependencies) > 0 {
			fmt.Fprintf(out, "\tsubgraph %s[%s]\n\tdirection LR\n", id, n.Name)
			fmt.Fprintf(out, "\tsubgraph %s-In[%s]\n\tdirection TB\n", id, "Input")
		} else {
			fmt.Fprintf(out, "\t%s[%s]\n", id, n.Name)
		}

		for depIndex, dep := range n.Dependencies {
			fmt.Fprintf(out, "\t%s-%d([%s])\n", id, depIndex, dep.Name)
		}

		if len(n.Dependencies) > 0 {
			fmt.Fprint(out, "\tend\n")
			fmt.Fprint(out, "\tend\n")
		}
	}

	for id, n := range schema.Nodes {
		for depIndex, d := range n.Dependencies {
			fmt.Fprintf(out, "\t%s --> %s-%d\n", d.DependencyID, id, depIndex)
		}
	}

	return nil
}
