package generator

import (
	"flag"

	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator/schema"
)

type Parameter interface {
	DisplayName() string
	Schema() schema.Parameter
	InitializeForCLI(set *flag.FlagSet)

	ApplyMessage(msg []byte) (bool, error)
	ToMessage() []byte
}

type SwaggerParameter interface {
	Parameter

	SwaggerProperty() swagger.Property
}
