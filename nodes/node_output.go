package nodes

import "fmt"

func TryGetOutputValue[T any](output Output[T], fallback T) T {
	if output == nil {
		return fallback
	}
	return output.Value()
}

func GetNodeOutputPort[T any](node Node, portName string) Output[T] {
	outputs := node.Outputs()

	if port, ok := outputs[portName]; ok {
		if cast, ok := port.(Output[T]); ok {
			return cast
		}
		var t T
		panic(fmt.Errorf("node port %q is not type %T", portName, t))
	}

	msg := ""
	for port := range outputs {
		msg += port + " "
	}

	panic(fmt.Errorf("node does not contain a port named %q, only %s", portName, msg))
}
