package schema

type Producer struct {
	NodeID string `json:"nodeID"`
	Port   string `json:"port"` // Name of node out port
}
