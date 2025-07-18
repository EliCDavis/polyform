package graph

import (
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

func ForeachManifestNodeOutput(i *Instance, f func(nodeId string, node nodes.Node, output nodes.Output[manifest.Manifest]) error) error {
	nodeIds := i.NodeIds()

	for _, nodeId := range nodeIds {
		node := i.Node(nodeId)
		outputs := node.Outputs()
		for _, out := range outputs {
			manifestOut, ok := out.(nodes.Output[manifest.Manifest])
			if !ok {
				continue
			}
			err := f(nodeId, node, manifestOut)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
