// Package history records undoable steps against anything that can
// snapshot itself and be restored from a snapshot.
package history

import (
	"errors"
	"fmt"
	"sync"
)

const (
	// Bounded by bytes too: a graph carrying image data runs to megabytes.
	maxEntries = 64
	maxBytes   = 64 << 20
)

// After this the subject is in whatever state the failed step and the failed
// restore left it, which can be anything down to empty.
var ErrRollbackFailed = errors.New("the graph could not be rolled back")

type Snapshot[S any] interface {
	Equal(S) bool
	Size() int
}

type entry[S any] struct {
	label string
	state S
}

type History[S Snapshot[S]] struct {
	snapshot func() (S, error)
	restore  func(S) error

	// Held for a whole step: a rollback restores the state from before it, which would erase any step overlapping it.
	mu     sync.Mutex
	past   []entry[S]
	future []entry[S]
}

func New[S Snapshot[S]](snapshot func() (S, error), restore func(S) error) *History[S] {
	return &History[S]{snapshot: snapshot, restore: restore}
}

// Transact runs f as one undoable step. A step that fails is rolled back
// rather than left half applied, and one that changed nothing is not a step.
func (h *History[S]) Transact(label string, f func() error) (err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	before, err := h.snapshot()
	if err != nil {
		return fmt.Errorf("could not snapshot the graph before %q: %w", label, err)
	}

	defer func() {
		if r := recover(); r != nil {
			err = Recovered(r)
		}

		if err == nil {
			if after, snapshotErr := h.snapshot(); snapshotErr != nil || !before.Equal(after) {
				h.record(entry[S]{label: label, state: before})
			}
			return
		}
		if restoreErr := safely(func() error { return h.restore(before) }); restoreErr != nil {
			err = fmt.Errorf("%w (and %w: %v)", err, ErrRollbackFailed, restoreErr)
		}
	}()

	return f()
}

func (h *History[S]) record(e entry[S]) {
	// Acting after an undo abandons the branch that was undone.
	h.future = nil
	h.past = append(h.past, e)

	total := 0
	for _, kept := range h.past {
		total += kept.state.Size()
	}

	drop := 0
	for drop < len(h.past)-1 && (len(h.past)-drop > maxEntries || total > maxBytes) {
		total -= h.past[drop].state.Size()
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

func (h *History[S]) CanUndo() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.past) > 0
}

func (h *History[S]) CanRedo() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.future) > 0
}

func (h *History[S]) Undo() (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.step(&h.past, &h.future, "undo")
}

func (h *History[S]) Redo() (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.step(&h.future, &h.past, "redo")
}

func (h *History[S]) step(from, to *[]entry[S], verb string) (string, error) {
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
	*to = append(*to, entry[S]{label: last.label, state: current})
	return last.label, nil
}

// Most recent step first.
type Steps struct {
	Undo []string `json:"undo"`
	Redo []string `json:"redo"`
}

func (h *History[S]) Steps() Steps {
	h.mu.Lock()
	defer h.mu.Unlock()

	labels := func(stack []entry[S]) []string {
		out := make([]string, 0, len(stack))
		for i := len(stack) - 1; i >= 0; i-- {
			out = append(out, stack[i].label)
		}
		return out
	}
	return Steps{Undo: labels(h.past), Redo: labels(h.future)}
}
