package graph

import (
	"fmt"
	"slices"
	"strings"

	"github.com/EliCDavis/polyform/nodes"
)

// While a compound edit is open, an edit anywhere in the graph is only
// noted; settling and refreshing placements happen once, when it closes.
type compoundEdit struct {
	depth  int
	edited []*Graph
	added  map[addedInput]bool
}

type addedInput struct {
	graph    *Graph
	consumer nodes.Node
	input    string
}

func (c *compoundEdit) noteAdded(graph *Graph, e edge) {
	if c.added == nil {
		c.added = make(map[addedInput]bool)
	}
	c.added[addedInput{graph, e.consumer, e.input}] = true
}

func (c *compoundEdit) hintFor(edited []*Graph, drop DroppedEdge) string {
	for _, graph := range edited {
		if graph.SubGraphScopeID() != drop.Scope {
			continue
		}
		e, err := graph.edgeBetween(drop.From, drop.FromPort, drop.To, drop.ToPort)
		if err == nil && c.added[addedInput{graph, e.consumer, e.input}] {
			return edgeHint(e)
		}
	}
	return ""
}

// Batch runs f as one compound edit; edges left unfit are dropped and returned as *UnfitEdges.
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
		defer func() { root.compound.added = nil }()
		return root.closeCompoundEdit(edited)
	}
}

func (a *Graph) inCompoundEdit() bool {
	return a.Root().compound.depth > 0
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
	var dropped []DroppedEdge
	for _, finish := range []func(*Graph, conflictPolicy) ([]DroppedEdge, error){(*Graph).settle, (*Graph).refreshPlacers} {
		for _, graph := range edited {
			unfit, err := finish(graph, dropConflicts)
			if err != nil {
				return err
			}
			dropped = append(dropped, unfit...)
		}
	}
	if len(dropped) == 0 {
		return nil
	}
	for i := range dropped {
		dropped[i].Reason += root.compound.hintFor(edited, dropped[i])
	}
	return &UnfitEdges{Edges: dropped}
}

// UnfitEdges refuses a compound edit; a caller that keeps the edit has had Edges removed.
type UnfitEdges struct {
	Edges []DroppedEdge
}

func (u *UnfitEdges) Error() string {
	reasons := make([]string, len(u.Edges))
	for i, drop := range u.Edges {
		reasons[i] = drop.Reason
	}
	return fmt.Sprintf("refused, edges would no longer fit: %s", strings.Join(reasons, "; "))
}
