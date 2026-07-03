package schema

type NodeTypeOutput struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type NodeTypeInput struct {
	Type        string `json:"type"`
	IsArray     bool   `json:"isArray"`
	Description string `json:"description,omitempty"`
}

type NodeType struct {
	DisplayName string                    `json:"displayName"`
	Info        string                    `json:"info"`
	Type        string                    `json:"type"`
	Path        string                    `json:"path"`
	Outputs     map[string]NodeTypeOutput `json:"outputs,omitempty"`
	Inputs      map[string]NodeTypeInput  `json:"inputs,omitempty"`
	Parameter   Parameter                 `json:"parameter,omitempty"`
}
