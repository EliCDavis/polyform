package nodes

import (
	"fmt"
	"time"
)

func GetOutputValues[G any](recorder ExecutionRecorder, output []Output[G]) []G {
	if len(output) == 0 {
		return make([]G, 0)
	}

	results := make([]G, 0, len(output))
	for _, out := range output {
		if out == nil {
			continue
		}
		start := time.Now()
		v := out.Value()
		recorder.CaptureTiming(out.Name(), time.Since(start))
		results = append(results, v)
	}

	return results
}

func TryGetOutputValue[G any](result ExecutionRecorder, output Output[G], fallback G) G {
	if output == nil {
		return fallback
	}
	start := time.Now()
	v := output.Value()
	result.CaptureTiming(output.Name(), time.Since(start))
	return v
}

func GetOutputValue[T, G any](result StructOutput[T], output Output[G]) G {
	start := time.Now()
	v := output.Value()
	result.CaptureTiming(output.Name(), time.Since(start))
	return v
}

func TryGetOutputReference[T, G any](result StructOutput[T], output Output[G], fallback *G) *G {
	if output == nil {
		return fallback
	}
	start := time.Now()
	v := output.Value()
	result.CaptureTiming(output.Name(), time.Since(start))
	return &v
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
