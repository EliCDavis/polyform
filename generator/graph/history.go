package graph

import (
	"fmt"

	"github.com/EliCDavis/polyform/refutil"
)

const (
	// Bounded by bytes too: a graph carrying image data runs to megabytes.
	defaultHistoryDepth = 64
	defaultHistoryBytes = 64 << 20
)

type historyEntry struct {
	label string
	state []byte
}

type history struct {
	past   []historyEntry
	future []historyEntry

	maxEntries int
	maxBytes   int
}

func (h *history) limits() (entries, bytes int) {
	entries, bytes = h.maxEntries, h.maxBytes
	if entries <= 0 {
		entries = defaultHistoryDepth
	}
	if bytes <= 0 {
		bytes = defaultHistoryBytes
	}
	return
}

func (h *history) record(e historyEntry) {
	// Acting after an undo abandons the branch that was undone.
	h.future = nil
	h.past = append(h.past, e)

	maxEntries, maxBytes := h.limits()
	total := 0
	for _, entry := range h.past {
		total += len(entry.state)
	}

	drop := 0
	for drop < len(h.past)-1 && (len(h.past)-drop > maxEntries || total > maxBytes) {
		total -= len(h.past[drop].state)
		drop++
	}
	if drop > 0 {
		h.past = append(h.past[:0], h.past[drop:]...)
	}
}

func pop(stack []historyEntry) ([]historyEntry, historyEntry, bool) {
	if len(stack) == 0 {
		return stack, historyEntry{}, false
	}
	last := len(stack) - 1
	return stack[:last], stack[last], true
}

// Transact runs f as one undoable step. A step that fails is rolled back
// rather than left half applied, and the snapshot is taken at the root, so
// a step touching a subgraph undoes as one with everything else it changed.
func (a *Instance) Transact(label string, f func() error) (err error) {
	before, err := a.EncodeToAppSchema()
	if err != nil {
		return fmt.Errorf("could not snapshot the graph before %q: %w", label, err)
	}
	beforeTypes := (&refutil.TypeFactory{}).Combine(a.typeFactory)

	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}

		if err == nil {
			a.history.record(historyEntry{label: label, state: before})
			return
		}
		// A type the failed step registered would be instantiated by the
		// reload, failing the rollback too.
		a.typeFactory = beforeTypes

		if restoreErr := a.restore(before); restoreErr != nil {
			err = fmt.Errorf("%w (and the graph could not be rolled back: %v)", err, restoreErr)
		}
	}()

	return f()
}

// A snapshot that will not load leaves the graph as the failed step left
// it, since loading clears the graph before it can know the payload is good.
func (a *Instance) restore(snapshot []byte) error {
	return safely(func() error {
		return a.ApplyAppSchema(snapshot)
	})
}

// restore runs from inside Transact's recovery, where a second panic is
// not recoverable and would take the process down.
func safely(f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return f()
}

func (a *Instance) CanUndo() bool { return len(a.history.past) > 0 }

func (a *Instance) CanRedo() bool { return len(a.history.future) > 0 }

func (a *Instance) Undo() (string, error) {
	return a.step(&a.history.past, &a.history.future, "undo")
}

func (a *Instance) Redo() (string, error) {
	return a.step(&a.history.future, &a.history.past, "redo")
}

func (a *Instance) step(from, to *[]historyEntry, verb string) (string, error) {
	remaining, entry, ok := pop(*from)
	if !ok {
		return "", fmt.Errorf("nothing to %s", verb)
	}

	current, err := a.EncodeToAppSchema()
	if err != nil {
		return "", fmt.Errorf("could not snapshot the graph before %s: %w", verb, err)
	}
	if err := a.restore(entry.state); err != nil {
		return "", fmt.Errorf("could not %s %q: %w", verb, entry.label, err)
	}

	*from = remaining
	*to = append(*to, historyEntry{label: entry.label, state: current})
	return entry.label, nil
}

// Most recent step first.
type HistoryState struct {
	Undo []string `json:"undo"`
	Redo []string `json:"redo"`
}

func (a *Instance) History() HistoryState {
	labels := func(stack []historyEntry) []string {
		out := make([]string, 0, len(stack))
		for i := len(stack) - 1; i >= 0; i-- {
			out = append(out, stack[i].label)
		}
		return out
	}
	return HistoryState{Undo: labels(a.history.past), Redo: labels(a.history.future)}
}
