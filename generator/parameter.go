package generator

import (
	"flag"

	"github.com/EliCDavis/polyform/formats/swagger"
)

type Parameter interface {
	DisplayName() string
	Schema() ParameterSchema
	InitializeForCLI(set *flag.FlagSet)

	ApplyMessage(msg []byte) (bool, error)
	ToMessage() []byte
}

type SwaggerParameter interface {
	Parameter

	SwaggerProperty() swagger.Property
}
