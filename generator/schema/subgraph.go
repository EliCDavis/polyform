package schema

type SubGraphInputBoundary struct {
	PortName string `json:"portName"`
	PortType string `json:"portType"`
}

type SubGraphOutputBoundary struct {
	PortName string `json:"portName"`
	PortType string `json:"portType"`
}

type SubGraphInstance struct {
	Nodes map[string]Node `json:"nodes"`
	Notes map[string]any  `json:"notes,omitempty"`
}
