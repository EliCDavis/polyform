package generator

import (
	"encoding/json"
	"flag"
)

type Parameter interface {
	DisplayName() string
	Schema() ParameterSchema
	initializeForCLI(set *flag.FlagSet)
	ApplyJsonMessage(msg json.RawMessage) error
	Reset()
}
