package schema

import "github.com/EliCDavis/polyform/nodes"

type GraphExecutionReport struct {
	Nodes map[string]NodeExecutionReport `json:"nodes"`
}

type NodeExecutionReport struct {
	Output map[string]nodes.ExecutionReport `json:"output"`
}
