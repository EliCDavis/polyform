package graph

import (
	"flag"

	"github.com/EliCDavis/polyform/generator/schema"
)

type Parameter interface {
	DisplayName() string
	Schema() schema.Parameter
	InitializeForCLI(set *flag.FlagSet)

	SetName(name string)
	SetDescription(name string)

	ApplyMessage(msg []byte) (bool, error)
	ToMessage() []byte
}

// ============================================================================
