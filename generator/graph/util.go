package graph

import (
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

func ForeachManifestNodeOutput(i *Instance, f func(nodeId string, node nodes.Node, output nodes.Output[manifest.Manifest]) error) error {
	for nodeId, node := range i.nodesByID {
		for _, out := range node.Outputs() {
			if manifestOut, ok := out.(nodes.Output[manifest.Manifest]); ok {
				if err := f(nodeId, node, manifestOut); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
