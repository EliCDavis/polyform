package generator

type NodeOutput struct {
	Name string `json:"name"`
}

type NodeSchema struct {
	Name         string                 `json:"name"`
	Version      int                    `json:"version"`
	Dependencies []NodeDependencySchema `json:"dependencies"`
	Outputs      []NodeOutput           `json:"outputs"`
	Parameter    ParameterSchema        `json:"parameter,omitempty"`

	// node      nodes.Node
	parameter Parameter
}

type NodeDependencySchema struct {
	DependencyID string `json:"dependencyID"`
	Name         string `json:"name"`
}

type AppSchema struct {
	Producers []string              `json:"producers"`
	Nodes     map[string]NodeSchema `json:"nodes"`
}
