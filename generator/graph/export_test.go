package graph

import (
	"bytes"
	"fmt"
	"maps"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/persistence"
)

func CopiesMatchDefinitions(root *Instance) error {
	saved := func(graph *Graph) map[string]persistence.Node {
		return graph.savedNodes(&jbtf.Encoder{})
	}

	var check func(graph *Graph, path string) error
	check = func(graph *Graph, path string) error {
		for node, id := range graph.nodeIDs {
			placement, ok := node.(*SubgraphInstanceNode)
			if !ok {
				continue
			}
			at := path + id
			live := placement.LiveGraph()
			want, got := saved(root.subGraphs[placement.subGraphID].instance), saved(live)
			if len(want) != len(got) {
				return fmt.Errorf("%s: copy has %d nodes, its definition %d", at, len(got), len(want))
			}
			for nodeID, original := range want {
				copied, ok := got[nodeID]
				switch {
				case !ok:
					return fmt.Errorf("%s: copy is missing node %s", at, nodeID)
				case copied.Type != original.Type:
					return fmt.Errorf("%s: node %s is a %s, its original a %s", at, nodeID, copied.Type, original.Type)
				case !bytes.Equal(copied.Data, original.Data):
					return fmt.Errorf("%s: node %s holds %s, its original %s", at, nodeID, copied.Data, original.Data)
				case !maps.Equal(copied.AssignedInput, original.AssignedInput):
					return fmt.Errorf("%s: node %s reads %v, its original %v", at, nodeID, copied.AssignedInput, original.AssignedInput)
				}
			}
			if err := check(live, at+"/"); err != nil {
				return err
			}
		}
		return nil
	}

	if err := check(root.Graph, ""); err != nil {
		return err
	}
	for id, definition := range root.subGraphs {
		if err := check(definition.instance, "definition "+id+": "); err != nil {
			return err
		}
	}
	return nil
}
