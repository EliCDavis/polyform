package nodes

import "time"

// StepTiming represents the duration of a labeled sub-operation within an execution.
type StepTiming struct {
	Label    string        `json:"label,omitempty"` // Descriptive name for this step
	Duration time.Duration `json:"duration"`        // Time taken by this step
	Steps    []StepTiming  `json:"steps,omitempty"` // Detailed timing of sub-operations
}

type ExecutionReport struct {
	Errors    []string      `json:"errors,omitempty"` // Any errors that occurred during execution
	Logs      []string      `json:"logs,omitempty"`   // Log messages produced
	TotalTime time.Duration `json:"totalTime"`        // Total time taken to compute the output
	Steps     []StepTiming  `json:"steps,omitempty"`  // Detailed timing of sub-operations
}

// ObservableOutput represents an output whose execution report can be inspected.
type ObservableExecution interface {
	ExecutionReport() ExecutionReport
}
