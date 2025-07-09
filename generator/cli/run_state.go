package cli

import (
	"fmt"
	"io"
	"time"
)

type RunState struct {
	Out   io.Writer
	Err   io.Writer
	Args  []string
	flags map[string]Flag
}

func (rs RunState) String(flagName string) string {
	return getFlagValue[string](rs.flags, flagName)
}

func (rs RunState) Bool(flagName string) bool {
	return getFlagValue[bool](rs.flags, flagName)
}

func (rs RunState) Duration(flagName string) time.Duration {
	return getFlagValue[time.Duration](rs.flags, flagName)
}

func (rs RunState) Int64(flagName string) int64 {
	return getFlagValue[int64](rs.flags, flagName)
}

func (rs RunState) Int(flagName string) int {
	return getFlagValue[int](rs.flags, flagName)
}

func getFlagValue[T any](flags map[string]Flag, flagName string) T {
	flag, ok := flags[flagName]
	if !ok {
		panic(fmt.Errorf("no flag with name %q", flagName))
	}

	v, ok := flag.value().(T)
	if !ok {
		panic(fmt.Errorf("could not cast %q to %T", flagName, v))
	}
	return v
}
