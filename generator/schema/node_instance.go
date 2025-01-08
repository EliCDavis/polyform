package schema

import "encoding/json"

type NodeInstance struct {
	Type         string           `json:"type"`
	Name         string           `json:"name"`
	Version      int              `json:"version"`
	Dependencies []NodeDependency `json:"dependencies"`
	Parameter    Parameter        `json:"parameter,omitempty"`

	Metadata json.RawMessage `json:"metadata,omitempty"`
}
