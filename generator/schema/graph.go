package schema

type Graph struct {
	Producers map[string]Producer          `json:"producers"`
	Nodes     map[string]Node              `json:"nodes"`
	Notes     map[string]any               `json:"notes"`
	Variables NestedGroup[RuntimeVariable] `json:"variables,omitempty"`
	Profiles  []string                     `json:"profiles,omitempty"`
	SubGraphs map[string]SubGraphInstance  `json:"subGraphs,omitempty"`
}
