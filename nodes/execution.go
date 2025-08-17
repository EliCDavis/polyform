package nodes

import "time"

// StepTiming represents the duration of a labeled sub-operation within an execution.
type StepTiming struct {
	Label    string        `json:"label,omitempty"` // Descriptive name for this step
	Duration time.Duration `json:"duration"`        // Time taken by this step
	Steps    []StepTiming  `json:"steps,omitempty"` // Detailed timing of sub-operations
}

type ExecutionReport struct {
	Errors    []string       `json:"errors,omitempty"`   // Any errors that occurred during execution
	Logs      []string       `json:"logs,omitempty"`     // Log messages produced
	TotalTime time.Duration  `json:"totalTime"`          // Total time taken to compute the output
	SelfTime  *time.Duration `json:"selfTime,omitempty"` // Time spent within the node itself, not counting waiting on other outputs
	Steps     []StepTiming   `json:"steps,omitempty"`    // Detailed timing of sub-operations, all sub operation times should result in TotalTime - SelfTime
}

// ObservableOutput represents an output whose execution report can be inspected.
type ObservableExecution interface {
	ExecutionReport() ExecutionReport
}

type ExecutionRecorder interface {
	CaptureTiming(title string, timing time.Duration)
	CaptureError(err error)
}
