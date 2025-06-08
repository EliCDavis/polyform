package variable

import (
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
)

type Variable interface {
	NodeReference() nodes.Node

	Info() Info
	setInfo(i Info) error

	ApplyMessage(msg []byte) (bool, error)
	ToMessage() []byte

	schema() schema.RuntimeVariable
	// Schema() schema.Parameter
	// InitializeForCLI(set *flag.FlagSet)
}

type Reference interface {
	Reference() Variable
}
