package nodes

import "fmt"

type InvalidInputError struct {
	Input   OutputPort
	Message string
}

func (nie InvalidInputError) Error() string {
	return fmt.Sprintf("invalid input %q: %s", nie.Input.Name(), nie.Message)
}

type NilInputError struct {
	Input OutputPort
}

func (nie NilInputError) Error() string {
	return fmt.Sprintf("invalid input %q: can not be nil", nie.Input.Name())
}

type UnsetInputError struct {
	Input OutputPort
}

func (nie UnsetInputError) Error() string {
	return fmt.Sprintf("invalid input %q: is not set", nie.Input.Name())
}
