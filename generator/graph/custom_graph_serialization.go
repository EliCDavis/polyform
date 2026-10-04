package graph

import (
	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/nodes"
)

type CustomGraphSerialization interface {
	ToJSON(encoder *jbtf.Encoder) ([]byte, error)
	FromJSON(decoder jbtf.Decoder, body []byte) error
}

// Optional: a copy of a node takes its state from the original instead of parsing it again.
type stateCopier interface {
	CopyStateFrom(original nodes.Node) bool
}
