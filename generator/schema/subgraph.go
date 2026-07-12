package schema

type SubGraphPortBoundary struct {
	PortName string `json:"portName"`
	PortType string `json:"portType"`
}

type SubGraph struct {
	Nodes map[string]Node `json:"nodes"`
	Notes map[string]any  `json:"notes,omitempty"`
}
