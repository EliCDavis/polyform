package properties

import (
	"github.com/EliCDavis/polyform/examples/ply-utils/flags"
	"github.com/urfave/cli/v2"
)

var PropertiesCommand = &cli.Command{
	Name:  "property",
	Usage: "commands around interacting with properties within a ply file",
	Flags: []cli.Flag{
		flags.PlyFile,
	},
	Subcommands: []*cli.Command{
		removePropertiesCommand,
		addPropertiesCommand,
		analyzePropertiesCommand,
	},
}
