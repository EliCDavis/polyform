package nodes

import "sync/atomic"

// generation counts every mutation that can change an output port's version.
// Struct memoizes against it, so Version() stops re-walking to the leaves.
var generation atomic.Uint64

func Touch() {
	generation.Add(1)
}

func Generation() uint64 {
	return generation.Load()
}
