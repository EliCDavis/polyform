package cli

import (
	"errors"
	"fmt"
	"os"
)

type App struct {
	Commands       []*Command
	ConfigProvided func(config string) error
}

func (a *App) Run(args []string) error {
	commandMap := make(map[string]*Command)
	for _, cmd := range a.Commands {
		for _, alias := range cmd.Aliases {
			commandMap[alias] = cmd
		}
	}

	argsWithoutProg := args[1:]

	if len(argsWithoutProg) == 0 {
		return commandMap["help"].Run(nil)
	}

	firstArg := argsWithoutProg[0]
	if cmd, ok := commandMap[firstArg]; ok {
		return cmd.Run(args[2:])
	}

	if !fileExists(firstArg) {
		return fmt.Errorf("unrecognized command %s", firstArg)
	}

	err := a.ConfigProvided(firstArg)
	if err != nil {
		return err
	}

	argsWithoutGraph := argsWithoutProg[1:]
	if len(argsWithoutGraph) == 0 {
		return commandMap["help"].Run(nil)
	}

	firstArg = argsWithoutGraph[0]
	if cmd, ok := commandMap[firstArg]; ok {
		return cmd.Run(argsWithoutGraph[1:])
	}

	return fmt.Errorf("unrecognized command %s", firstArg)

}

func fileExists(fp string) bool {
	if _, err := os.Stat(fp); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}
