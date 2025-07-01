package schema

type GraphInstance struct {
	Producers map[string]Producer          `json:"producers"`
	Nodes     map[string]NodeInstance      `json:"nodes"`
	Notes     map[string]any               `json:"notes"`
	Variables NestedGroup[RuntimeVariable] `json:"variables,omitempty"`
	Profiles  []string                     `json:"profiles,omitempty"`
}
