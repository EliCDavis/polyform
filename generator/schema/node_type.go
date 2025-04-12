package schema

type NodeOutput struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type NodeInput struct {
	Type        string `json:"type"`
	IsArray     bool   `json:"isArray"`
	Description string `json:"description,omitempty"`
}

type NodeType struct {
	DisplayName string                `json:"displayName"`
	Info        string                `json:"info"`
	Type        string                `json:"type"`
	Path        string                `json:"path"`
	Outputs     map[string]NodeOutput `json:"outputs,omitempty"`
	Inputs      map[string]NodeInput  `json:"inputs,omitempty"`
	Parameter   Parameter             `json:"parameter,omitempty"`
}
