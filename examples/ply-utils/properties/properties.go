package properties

import "github.com/urfave/cli/v2"

var PropertiesCommand = &cli.Command{
	Name:  "property",
	Usage: "commands around interacting with properties within a ply file",
	Subcommands: []*cli.Command{
		removePropertiesCommand,
		addPropertiesCommand,
	},
}
