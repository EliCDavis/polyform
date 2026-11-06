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
		if spread, ok := SpreadValues[G](out); ok {
			recorder.CaptureTiming(out.Name(), time.Since(start))
			captureOutputError(recorder, out)
			results = append(results, spread...)
			continue
		}
		results = append(results, readOutput(recorder, out))
	}

	return results
}

type FailingOutput interface {
	Err() error
}

func captureOutputError(recorder ExecutionRecorder, output OutputPort) {
	if failing, ok := output.(FailingOutput); ok {
		recorder.CaptureError(failing.Err())
	}
}

func readOutput[G any](recorder ExecutionRecorder, output Output[G]) G {
	start := time.Now()
	v := output.Value()
	recorder.CaptureTiming(output.Name(), time.Since(start))
	captureOutputError(recorder, output)
	return v
}

func TryGetOutputValue[G any](recorder ExecutionRecorder, output Output[G], fallback G) G {
	if output == nil {
		return fallback
	}
	return readOutput(recorder, output)
}

func GetOutputValue[G any](recorder ExecutionRecorder, output Output[G]) G {
	return readOutput(recorder, output)
}

func TryGetOutputReference[G any](result ExecutionRecorder, output Output[G], fallback *G) *G {
	if output == nil {
		return fallback
	}
	v := readOutput(result, output)
	return &v
}

func GetOutputReference[G any](result ExecutionRecorder, output Output[G]) *G {
	v := readOutput(result, output)
	return &v
}

func GetNodeOutputPort[T any](node Node, portName string) Output[T] {
	outputs := node.Outputs()

	port, ok := outputs[portName]
	if !ok {
		for _, candidate := range outputs {
			if labeled, isLabeled := candidate.(Labeled); isLabeled && labeled.Label() == portName {
				port, ok = candidate, true
			}
		}
	}
	if ok {
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
