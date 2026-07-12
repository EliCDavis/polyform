package schema

type PortReference struct {
	NodeId   string `json:"id"`
	PortName string `json:"port"`
}

type NodeOutputPort struct {
	Version int `json:"version"`
}

type Node struct {
	Type          string                    `json:"type"`
	Name          string                    `json:"name"`
	AssignedInput map[string]PortReference  `json:"assignedInput"`
	Output        map[string]NodeOutputPort `json:"output"`

	Parameter Parameter      `json:"parameter,omitempty"`
	Variable  any            `json:"variable,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`

	SubGraphInputBoundary  *SubGraphPortBoundary `json:"subGraphInputBoundary,omitempty"`
	SubGraphOutputBoundary *SubGraphPortBoundary `json:"subGraphOutputBoundary,omitempty"`
	SubGraphId             string                `json:"subGraphId,omitempty"`
}
