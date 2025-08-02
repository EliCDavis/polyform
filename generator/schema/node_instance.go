package schema

import "github.com/EliCDavis/polyform/nodes"

type PortReference struct {
	NodeId   string `json:"id"`
	PortName string `json:"port"`
}

type NodeInstanceOutputPort struct {
	Version int                    `json:"version"`
	Report  *nodes.ExecutionReport `json:"report,omitempty"`
}

type NodeInstance struct {
	Type          string                            `json:"type"`
	Name          string                            `json:"name"`
	AssignedInput map[string]PortReference          `json:"assignedInput"`
	Output        map[string]NodeInstanceOutputPort `json:"output"`

	Parameter Parameter      `json:"parameter,omitempty"`
	Variable  any            `json:"variable,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}
