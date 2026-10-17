package history

import (
	"fmt"
	"runtime/debug"
)

type PanicError struct {
	Value any
	Stack string
}

// Recovered must be called from the deferred function that recovered value,
// while the panicking frames are still on the stack.
func Recovered(value any) *PanicError {
	return &PanicError{Value: value, Stack: string(debug.Stack())}
}

func (e *PanicError) Error() string {
	if err, ok := e.Value.(error); ok {
		return err.Error()
	}
	return fmt.Sprint(e.Value)
}

func (e *PanicError) Unwrap() error {
	err, _ := e.Value.(error)
	return err
}
