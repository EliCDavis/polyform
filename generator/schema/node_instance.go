package schema

type PortReference struct {
	NodeId   string `json:"id"`
	PortName string `json:"port"`
}

type NodeOutputPort struct {
	Version int `json:"version"`

	// Sent only when it differs from the node type's declaration: a lifted output becomes an array once one is wired in.
	Type string `json:"type,omitempty"`
}

type Node struct {
	Type          string                    `json:"type"`
	Name          string                    `json:"name"`
	AssignedInput map[string]PortReference  `json:"assignedInput"`
	Output        map[string]NodeOutputPort `json:"output"`

	Parameter Parameter      `json:"parameter,omitempty"`
	Variable  any            `json:"variable,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`

	// What each of this instance's type variables settled on, empty for a
	// node without dynamic ports or when nothing has been connected to yet.
	DynamicTypes map[string]string `json:"dynamicTypes,omitempty"`

	SubGraphInputBoundary  *SubGraphPortBoundary `json:"subGraphInputBoundary,omitempty"`
	SubGraphOutputBoundary *SubGraphPortBoundary `json:"subGraphOutputBoundary,omitempty"`
	SubGraphId             string                `json:"subGraphId,omitempty"`
}
