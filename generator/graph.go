package generator

import (
	"encoding/json"

	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/generator/schema"
)

type Graph struct {
	Name        string                       `json:"name,omitempty"`
	Version     string                       `json:"version,omitempty"`
	Description string                       `json:"description,omitempty"`
	WebScene    *room.WebScene               `json:"webScene"`
	Producers   map[string]schema.Producer   `json:"producers"`
	Nodes       map[string]GraphNodeInstance `json:"nodes"`
	Metadata    map[string]any               `json:"metadata,omitempty"`
}

type GraphNodeInstance struct {
	Type         string                  `json:"type"`
	Dependencies []schema.NodeDependency `json:"dependencies,omitempty"`
	Data         json.RawMessage         `json:"data,omitempty"`
}
