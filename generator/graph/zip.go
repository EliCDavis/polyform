package graph

import (
	"archive/zip"
	"fmt"
	"os"
	"path"

	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

func manifestFileName(i *Instance, nodeId string, node nodes.Node, out nodes.Output[manifest.Manifest]) string {
	portName := out.Name()

	// TODO - There's gonna be an issue if anyone ever names a manifest
	// output the same name as a nodeID-port combo
	//
	// IE: Naming a manifest "Node-8-Out", could conflict with Node-8's
	// out port
	if name, named := i.IsPortNamed(node, portName); named {
		return name
	}
	return fmt.Sprintf("%s-%s", nodeId, portName)
}

func WriteManifestToZip(i *Instance, zw *zip.Writer, nodeId string, node nodes.Node, out nodes.Output[manifest.Manifest]) error {
	manifest := out.Value()
	manifestName := manifestFileName(i, nodeId, node, out)

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

	return foreachManifestNodeOutput(i, func(s string, n nodes.Node, o nodes.Output[manifest.Manifest]) error {
		return WriteManifestToZip(i, zw, s, n, o)
	})
}

func writeManifestToFolder(i *Instance, folder string, nodeId string, node nodes.Node, out nodes.Output[manifest.Manifest]) error {
	manifest := out.Value()
	manifestName := manifestFileName(i, nodeId, node, out)

	manifestFolder := path.Join(folder, manifestName)
	err := os.MkdirAll(manifestFolder, os.ModeDir)
	if err != nil {
		return err
	}

	entries := manifest.Entries
	for artifactName, entry := range entries {
		artifacePath := path.Join(manifestFolder, artifactName)

		f, err := os.Create(artifacePath)
		if err != nil {
			return fmt.Errorf("unable to create file %q: %w", artifacePath, err)
		}
		defer f.Close()

		if err = entry.Artifact.Write(f); err != nil {
			return err
		}
	}

	return nil
}

func WriteToFolder(i *Instance, folder string) error {
	return foreachManifestNodeOutput(i, func(s string, n nodes.Node, o nodes.Output[manifest.Manifest]) error {
		return writeManifestToFolder(i, folder, s, n, o)
	})
}

func foreachManifestNodeOutput(i *Instance, f func(string, nodes.Node, nodes.Output[manifest.Manifest]) error) error {
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
