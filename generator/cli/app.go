package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
)

type App struct {
	Commands       []*Command
	Out            io.Writer
	ConfigProvided func(config string) error
}

func (a *App) Run(args []string) error {

	runState := &RunState{
		Out: a.Out,
	}

	if a.Out == nil {
		runState.Out = os.Stdout
	}

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
		runState.Args = args[2:]
		return cmd.Run(runState)
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
		runState.Args = argsWithoutGraph[1:]
		return cmd.Run(runState)
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
