package schema

import (
	"encoding/json"
)

type App struct {
	Name        string                     `json:"name,omitempty"`
	Version     string                     `json:"version,omitempty"`
	Description string                     `json:"description,omitempty"`
	Authors     []Author                   `json:"authors,omitempty"`
	WebScene    *WebScene                  `json:"webScene,omitempty"`
	Producers   map[string]Producer        `json:"producers"`
	Nodes       map[string]AppNodeInstance `json:"nodes"`
	Metadata    map[string]any             `json:"metadata,omitempty"`
}

type AppNodeInstance struct {
	Type         string           `json:"type"`
	Dependencies []NodeDependency `json:"dependencies,omitempty"`
	Data         json.RawMessage  `json:"data,omitempty"`
}
