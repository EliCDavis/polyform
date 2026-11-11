package graph

import (
	"bytes"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/variable"
)

// State is the graph as saved, with file content held beside the JSON by hash rather than inside it.
type State struct {
	saved []byte
	blobs map[string][]byte
}

// The JSON names every blob by hash, so it alone decides equality.
func (s State) Equal(other State) bool {
	return bytes.Equal(s.saved, other.saved)
}

func (s State) Size() int {
	return len(s.saved)
}

func (a *Instance) snapshotState() (State, error) {
	a.mu().RLock()
	defer a.mu().RUnlock()
	return a.snapshotStateLocked()
}

// Blob slices are shared with the variables, never copied.
func (a *Instance) snapshotStateLocked() (State, error) {
	blobs := make(map[string][]byte)
	encoder := &jbtf.Encoder{}
	saved, err := a.encodeApp(encoder, variable.NewEncoder(encoder, blobs))
	if err != nil {
		return State{}, err
	}
	return State{saved: saved, blobs: blobs}, nil
}

func (a *Instance) restoreState(state State) error {
	a.mu().Lock()
	defer a.mu().Unlock()
	return a.restoreStateLocked(state)
}

func (a *Instance) restoreStateLocked(state State) error {
	return a.applyAppSchema(state.saved, state.blobs)
}
