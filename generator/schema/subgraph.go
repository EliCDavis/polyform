package schema

type SubGraphDefinition struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description,omitempty"`
	Nodes       map[string]AppNodeInstance     `json:"nodes"`
	Notes       map[string]any                 `json:"notes,omitempty"`
	Variables   NestedGroup[PersistedVariable] `json:"variables,omitempty"`
	Metadata    map[string]any                 `json:"metadata,omitempty"`
}

type SubGraphInputBoundary struct {
	PortName string `json:"portName"`
	PortType string `json:"portType"`
}

type SubGraphOutputBoundary struct {
	PortName string `json:"portName"`
	PortType string `json:"portType"`
}

type RuntimeSubGraphDefinition struct {
	Name        string                            `json:"name"`
	Description string                            `json:"description,omitempty"`
	Nodes       map[string]NodeInstance           `json:"nodes"`
	Notes       map[string]any                    `json:"notes,omitempty"`
	Variables   NestedGroup[RuntimeVariable]      `json:"variables,omitempty"`
}
