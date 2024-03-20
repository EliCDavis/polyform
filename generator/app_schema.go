package generator

type NodeOutput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type NodeInput struct {
	Type string `json:"type"`
}

type NodeSchema struct {
	Name         string                 `json:"name"`
	Version      int                    `json:"version"`
	Dependencies []NodeDependencySchema `json:"dependencies"`
	Outputs      []NodeOutput           `json:"outputs"`
	Inputs       map[string]NodeInput   `json:"inputs"`
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
