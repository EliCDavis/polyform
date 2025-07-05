package graph

import (
	"archive/zip"
	"fmt"
	"path"

	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

func WriteManifestToZip(i *Instance, zw *zip.Writer, node nodes.Node, out nodes.Output[manifest.Manifest]) error {
	manifest := out.Value()
	manifestName := fmt.Sprintf("%s-%s", i.nodeIDs[node], out.Name())

	// TODO - There's gonna be an issue if anyone ever names a manifest
	// output the same name as a nodeID-port combo
	//
	// IE: Naming a manifest "Node-8-Out", could conflict with Node-8's
	// out port
	if name, named := i.IsPortNamed(node, out.Name()); named {
		manifestName = name
	}

	entries := manifest.Entries
	for artifactName, entry := range entries {
		f, err := zw.Create(path.Join(manifestName, artifactName))
		if err != nil {
			return err
		}

		if err = entry.Artifact.Write(f); err != nil {
			return err
		}
	}

	return nil
}

func WriteToZip(i *Instance, zw *zip.Writer) error {
	if zw == nil {
		panic("can't write to nil zip writer")
	}

	nodeIds := i.NodeIds()

	for _, nodeId := range nodeIds {
		node := i.Node(nodeId)
		outputs := node.Outputs()
		for _, out := range outputs {
			manifestOut, ok := out.(nodes.Output[manifest.Manifest])
			if !ok {
				continue
			}
			WriteManifestToZip(i, zw, node, manifestOut)
		}
	}

	return nil
}
