package cli

import "io"

type Command struct {
	Name        string
	Description string
	Aliases     []string
	Run         func(state *RunState) error
}

type RunState struct {
	Out  io.Writer
	Args []string
}
