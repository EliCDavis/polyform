package generator

type NodeOutput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type NodeInput struct {
	Type string `json:"type"`
}

type NodeInstanceSchema struct {
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Version      int                    `json:"version"`
	Dependencies []NodeDependencySchema `json:"dependencies"`
	Parameter    ParameterSchema        `json:"parameter,omitempty"`

	// node      nodes.Node
	parameter Parameter
}

type NodeDependencySchema struct {
	DependencyID   string `json:"dependencyID"`
	DependencyPort string `json:"dependencyPort"`
	Name           string `json:"name"`
}

type NodeTypeSchema struct {
	DisplayName string               `json:"displayName"`
	Type        string               `json:"type"`
	Path        string               `json:"path"`
	Outputs     []NodeOutput         `json:"outputs,omitempty"`
	Inputs      map[string]NodeInput `json:"inputs,omitempty"`
	Parameter   ParameterSchema      `json:"parameter,omitempty"`
}

type ProducerSchema struct {
	NodeID string `json:"nodeID"`
	Port   string `json:"port"` // Name of node out port
}

type AppSchema struct {
	Producers map[string]ProducerSchema     `json:"producers"`
	Nodes     map[string]NodeInstanceSchema `json:"nodes"`
	Types     []NodeTypeSchema              `json:"types"`
}
