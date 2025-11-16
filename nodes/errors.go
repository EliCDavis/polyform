package nodes

import "fmt"

type InvalidInputError struct {
	Input   OutputPort
	Message string
}

func (nie InvalidInputError) Error() string {
	return fmt.Sprintf("invalid input %q: %s", nie.Input.Name(), nie.Message)
}
