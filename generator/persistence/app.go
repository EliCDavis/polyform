package persistence

import (
	"encoding/json"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/schema"
)

type App struct {
	Name        string                       `json:"name,omitempty"`
	Version     string                       `json:"version,omitempty"`
	Description string                       `json:"description,omitempty"`
	Authors     []Author                     `json:"authors,omitempty"`
	WebScene    *WebScene                    `json:"webScene,omitempty"`
	Producers   map[string]schema.Producer   `json:"producers"`
	Nodes       map[string]Node              `json:"nodes"`
	Metadata    map[string]any               `json:"metadata,omitempty"`
	Variables   schema.NestedGroup[Variable] `json:"variables,omitempty"`
	Profiles    map[string]Profile           `json:"profiles,omitempty"`
	Variants    map[string]VariantSet        `json:"variants,omitempty"`
	SubGraphs   map[string]SubGraph          `json:"subGraphs,omitempty"`
	Enums       map[string]Enum              `json:"enums,omitempty"`
}

type Enum struct {
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Members     []enum.Member `json:"members"`
}

type Profile struct {
	Data map[string]json.RawMessage `json:"data,omitempty"`
}

type VariantSet struct {
	Dimensions map[string]json.RawMessage `json:"dimensions,omitempty"`
}

type Node struct {
	Type          string                          `json:"type"`
	AssignedInput map[string]schema.PortReference `json:"assignedInput,omitempty"`
	Data          json.RawMessage                 `json:"data,omitempty"`
	Variable      *string                         `json:"variable,omitempty"`
}

type Author struct {
	Name        string          `json:"name"`
	ContactInfo []AuthorContact `json:"contactInfo,omitempty"`
}

type AuthorContact struct {
	Medium string `json:"medium"`
	Value  string `json:"value"`
}

type Variable struct {
	Description string          `json:"description"`
	Data        json.RawMessage `json:"data"`
}

type SubGraph struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Group       string          `json:"group,omitempty"`
	Nodes       map[string]Node `json:"nodes"`
	Notes       map[string]any  `json:"notes,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
}
