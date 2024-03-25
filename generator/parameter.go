package generator

import (
	"flag"
)

type Parameter interface {
	DisplayName() string
	Schema() ParameterSchema
	initializeForCLI(set *flag.FlagSet)
	ApplyMessage(msg []byte) (bool, error)
}
