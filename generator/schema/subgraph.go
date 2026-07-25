package schema

type SubGraphPortBoundary struct {
	PortName string `json:"portName"`
	PortType string `json:"portType"`
}

type SubGraph struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Nodes       map[string]Node   `json:"nodes"`
	Notes       map[string]any    `json:"notes,omitempty"`
}
