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
		return port.(Output[T])
	}

	msg := ""
	for port := range outputs {
		msg += port + " "
	}

	panic(fmt.Errorf("node does not contain a port named %s, only %s", portName, msg))
}
