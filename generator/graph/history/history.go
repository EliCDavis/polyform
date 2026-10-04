// Package history records undoable steps against anything that can
// snapshot itself and be restored from a snapshot.
package history

import (
	"errors"
	"fmt"
)

const (
	// Bounded by bytes too: a graph carrying image data runs to megabytes.
	maxEntries = 64
	maxBytes   = 64 << 20
)

// After this the subject is in whatever state the failed step and the failed
// restore left it, which can be anything down to empty.
var ErrRollbackFailed = errors.New("the graph could not be rolled back")

type entry struct {
	label string
	state []byte
}

type History struct {
	snapshot func() ([]byte, error)
	restore  func([]byte) error

	past   []entry
	future []entry
}

func New(snapshot func() ([]byte, error), restore func([]byte) error) *History {
	return &History{snapshot: snapshot, restore: restore}
}

// Transact runs f as one undoable step. A step that fails is rolled back
// rather than left half applied.
func (h *History) Transact(label string, f func() error) (err error) {
	before, err := h.snapshot()
	if err != nil {
		return fmt.Errorf("could not snapshot the graph before %q: %w", label, err)
	}

	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}

		if err == nil {
			h.record(entry{label: label, state: before})
			return
		}
		if restoreErr := safely(func() error { return h.restore(before) }); restoreErr != nil {
			err = fmt.Errorf("%w (and %w: %v)", err, ErrRollbackFailed, restoreErr)
		}
	}()

	return f()
}

func (h *History) record(e entry) {
	// Acting after an undo abandons the branch that was undone.
	h.future = nil
	h.past = append(h.past, e)

	total := 0
	for _, kept := range h.past {
		total += len(kept.state)
	}

	drop := 0
	for drop < len(h.past)-1 && (len(h.past)-drop > maxEntries || total > maxBytes) {
		total -= len(h.past[drop].state)
		drop++
	}
	h.past = append(h.past[:0], h.past[drop:]...)
}

// A restore runs from inside Transact's recovery, where a second panic is
// not recoverable and would take the process down.
func safely(f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return f()
}

func (h *History) CanUndo() bool { return len(h.past) > 0 }

func (h *History) CanRedo() bool { return len(h.future) > 0 }

func (h *History) Undo() (string, error) {
	return h.step(&h.past, &h.future, "undo")
}

func (h *History) Redo() (string, error) {
	return h.step(&h.future, &h.past, "redo")
}

func (h *History) step(from, to *[]entry, verb string) (string, error) {
	if len(*from) == 0 {
		return "", fmt.Errorf("nothing to %s", verb)
	}
	last := (*from)[len(*from)-1]

	current, err := h.snapshot()
	if err != nil {
		return "", fmt.Errorf("could not snapshot the graph before %s: %w", verb, err)
	}
	if err := safely(func() error { return h.restore(last.state) }); err != nil {
		return "", fmt.Errorf("could not %s %q: %w", verb, last.label, err)
	}

	*from = (*from)[:len(*from)-1]
	*to = append(*to, entry{label: last.label, state: current})
	return last.label, nil
}

// Most recent step first.
type Steps struct {
	Undo []string `json:"undo"`
	Redo []string `json:"redo"`
}

func (h *History) Steps() Steps {
	labels := func(stack []entry) []string {
		out := make([]string, 0, len(stack))
		for i := len(stack) - 1; i >= 0; i-- {
			out = append(out, stack[i].label)
		}
		return out
	}
	return Steps{Undo: labels(h.past), Redo: labels(h.future)}
}
