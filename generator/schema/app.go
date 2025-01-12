package schema

type App struct {
	Producers map[string]Producer     `json:"producers"`
	Nodes     map[string]NodeInstance `json:"nodes"`
	Types     []NodeType              `json:"types"`
	Notes     map[string]any          `json:"notes"`
}
