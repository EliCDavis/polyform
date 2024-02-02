package generator

type NodeSchema struct {
	Name         string
	Dependencies []NodeDependencySchema
	Parameter    ParameterSchema
	Version      int
}

type NodeDependencySchema struct {
	DependencyID string
	Name         string
}

type AppSchema struct {
	Producers []string `json:"producers"`
	Nodes     map[string]NodeSchema
}
