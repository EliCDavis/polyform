package schema

type NodeOutput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type NodeInput struct {
	Type    string `json:"type"`
	IsArray bool   `json:"isArray"`
}

type NodeType struct {
	DisplayName string               `json:"displayName"`
	Info        string               `json:"info"`
	Type        string               `json:"type"`
	Path        string               `json:"path"`
	Outputs     []NodeOutput         `json:"outputs,omitempty"`
	Inputs      map[string]NodeInput `json:"inputs,omitempty"`
	Parameter   Parameter            `json:"parameter,omitempty"`
}
