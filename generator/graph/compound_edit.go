package graph

import (
	"fmt"
	"slices"
	"strings"
)

// While a compound edit is open, an edit anywhere in the graph is only
// noted; settling and refreshing placements happen once, when it closes.
type compoundEdit struct {
	depth  int
	edited []*Graph
}

// Batch runs f as one compound edit. Edges that no longer fit when f is
// done are dropped and returned as an error.
func (a *Instance) Batch(f func() error) (err error) {
	a.mu().Lock()
	end := a.beginCompoundEdit()
	a.mu().Unlock()

	defer func() {
		a.mu().Lock()
		defer a.mu().Unlock()
		if endErr := end(); endErr != nil && err == nil {
			err = endErr
		}
	}()
	return f()
}

func (a *Graph) beginCompoundEdit() (end func() error) {
	root := a.Root()
	root.compound.depth++
	return func() error {
		root.compound.depth--
		if root.compound.depth > 0 {
			return nil
		}
		edited := root.compound.edited
		root.compound.edited = nil
		return root.closeCompoundEdit(edited)
	}
}

// deferredToCompoundEdit reports whether a compound edit is open, noting a
// as edited if so.
func (a *Graph) deferredToCompoundEdit() bool {
	root := a.Root()
	if root.compound.depth == 0 {
		return false
	}
	if !slices.Contains(root.compound.edited, a) {
		root.compound.edited = append(root.compound.edited, a)
	}
	return true
}

func (root *Instance) closeCompoundEdit(edited []*Graph) error {
	stop := root.watchDrops()
	err := func() error {
		for _, graph := range edited {
			if _, err := graph.settleUnder(dropConflicts); err != nil {
				return err
			}
		}
		for _, graph := range edited {
			if err := graph.refreshPlacers(dropConflicts); err != nil {
				return err
			}
		}
		return nil
	}()
	dropped := stop()

	if err != nil {
		return err
	}
	if len(dropped) == 0 {
		return nil
	}
	reasons := make([]string, len(dropped))
	for i, drop := range dropped {
		reasons[i] = drop.Reason
	}
	return fmt.Errorf("the edit left edges whose types no longer fit, and they were removed: %s", strings.Join(reasons, "; "))
}
