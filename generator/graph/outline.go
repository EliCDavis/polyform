package graph

import (
	"fmt"
	"io"
)

func WriteOutline(graph *Instance, out io.Writer) error {
	fmt.Fprintf(out, "# %s\n\n", graph.GetName())
	if graph.GetVersion() != "" {
		fmt.Fprintf(out, "%s\n\n", graph.GetVersion())
	}
	fmt.Fprintf(out, "%s\n\n", graph.GetDescription())

	fmt.Fprintf(out, "## Variables\n\n")

	variables := graph.Schema().Variables.Variables
	if len(variables) == 0 {
		fmt.Fprintf(out, "(none)\n")
	}

	for name, variable := range variables {
		fmt.Fprintf(out, "* %s: %s", name, variable.Type)
		if variable.Description != "" {
			fmt.Fprintf(out, " - %s", variable.Description)
		}
		fmt.Fprintf(out, "\n  * Value: %v\n", variable.Value)
	}

	fmt.Fprintf(out, "\n## Profiles\n\n")

	profiles := graph.Profiles()
	if len(profiles) == 0 {
		fmt.Fprintf(out, "(none)\n\n")
	}

	for i, profile := range profiles {
		fmt.Fprintf(out, "%d. %s\n", i+1, profile)
	}
	return nil
}
