package graph

import "slices"

// A DroppedEdge was removed because an edit left its producer carrying a
// type its input cannot take, or took the port it was wired to.
type DroppedEdge struct {
	// The subgraph the edge was in; empty for the root graph.
	Scope string

	From     string
	FromPort string
	To       string
	// "Port", or "Port.N" for an element of an array input.
	ToPort string

	Reason string
}

type dropLog struct {
	watchers int
	dropped  []DroppedEdge
}

// watchDrops records every edge dropped anywhere in the graph until the
// returned function is called, which returns them.
func (a *Graph) watchDrops() (stop func() []DroppedEdge) {
	log := &a.Root().drops
	log.watchers++
	start := len(log.dropped)

	return func() []DroppedEdge {
		seen := slices.Clone(log.dropped[start:])
		log.watchers--
		if log.watchers == 0 {
			log.dropped = nil
		}
		return seen
	}
}

func (a *Graph) recordDrop(drop DroppedEdge) {
	log := &a.Root().drops
	if log.watchers > 0 {
		log.dropped = append(log.dropped, drop)
	}
}
