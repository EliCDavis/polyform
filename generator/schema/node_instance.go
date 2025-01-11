package schema

type NodeInstance struct {
	Type         string           `json:"type"`
	Name         string           `json:"name"`
	Version      int              `json:"version"`
	Dependencies []NodeDependency `json:"dependencies"`
	Parameter    Parameter        `json:"parameter,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty"`
}
