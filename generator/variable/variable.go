package variable

import (
	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
)

type Variable interface {
	NodeReference() nodes.Node

	Info() Info
	setInfo(i Info) error

	ApplyMessage(msg []byte) (bool, error)
	ToMessage() []byte

	runtimeSchema() schema.RuntimeVariable

	currentValue() any
	currentVersion() int

	toPersistantJSON(encoder *jbtf.Encoder) ([]byte, error)
	fromPersistantJSON(decoder jbtf.Decoder, body []byte) error
	// fromPersistantJSON(decoder jbtf.Decoder, body []byte) error
	// Schema() schema.Parameter
	// InitializeForCLI(set *flag.FlagSet)
}

type Reference interface {
	Reference() Variable
}
