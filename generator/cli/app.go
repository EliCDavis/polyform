package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
)

type App struct {
	Commands []*Command
	Out      io.Writer
	Err      io.Writer
}

func (a *App) Run(args []string) error {
	runState := &RunState{
		Out: a.Out,
		Err: a.Err,
	}

	if a.Out == nil {
		runState.Out = os.Stdout
	}

	if a.Err == nil {
		runState.Err = os.Stderr
	}

	commandMap := make(map[string]*Command)
	for _, cmd := range a.Commands {
		for _, alias := range cmd.Aliases {
			commandMap[alias] = cmd
		}
	}

	argsWithoutProg := args[1:]

	if len(argsWithoutProg) == 0 {
		return commandMap["help"].Run(runState)
	}

	commandName := argsWithoutProg[0]
	cmd, ok := commandMap[commandName]
	if !ok {
		return fmt.Errorf("unrecognized command %s", commandName)
	}

	runState.Args = args[2:]
	flags := flag.NewFlagSet(commandName, flag.ExitOnError)
	flags.SetOutput(runState.Err)

	registeredFlags := make(map[string]Flag)
	for _, flag := range cmd.Flags {
		flag.add(flags)
		registeredFlags[flag.name()] = flag
	}

	err := flags.Parse(runState.Args)
	if err != nil {
		return err
	}

	runState.flags = registeredFlags
	for _, flag := range cmd.Flags {
		if flag.required() && !flag.set() {
			return fmt.Errorf("flag %q is required but not set", flag.name())
		}
	}

	for _, flag := range cmd.Flags {
		if err := flag.action(); err != nil {
			return err
		}
	}

	return cmd.Run(runState)

}
