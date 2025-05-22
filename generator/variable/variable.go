package variable

import (
	"github.com/EliCDavis/polyform/nodes"
)

type Variable interface {
	Name() string
	NodeReference() nodes.Node
	// Schema() schema.Parameter
	// InitializeForCLI(set *flag.FlagSet)

	SetName(name string)
	SetDescription(description string)

	ApplyMessage(msg []byte) (bool, error)
	ToMessage() []byte
}
