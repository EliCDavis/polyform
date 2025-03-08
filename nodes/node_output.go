package nodes

func TryGetOutputValue[T any](output Output[T], fallback T) T {
	if output == nil {
		return fallback
	}
	return output.Value()
}

func GetNodeOutputPort[T any](node Node, portName string) Output[T] {
	return node.Outputs()[portName].(Output[T])
}
