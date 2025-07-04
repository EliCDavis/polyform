package generator

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"runtime/debug"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

func writeManifestToZip(graph *graph.Instance, zw *zip.Writer, node nodes.Node, out nodes.Output[manifest.Manifest]) error {
	manifest := out.Value()
	manifestName := fmt.Sprintf("%s-%s", graph.NodeId(node), out.Name())

	// TODO - There's gonna be an issue if anyone ever names a manifest
	// output the same name as a nodeID-port combo
	//
	// IE: Naming a manifest "Node-8-Out", could conflict with Node-8's
	// out port
	if name, named := graph.IsPortNamed(node, out.Name()); named {
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

func writeGraphManifestsToZip(graph *graph.Instance, zw *zip.Writer) error {
	if graph == nil {
		panic("can't zip nil graph")
	}

	if zw == nil {
		panic("can't write to nil zip writer")
	}

	nodeIds := graph.NodeIds()

	for _, nodeId := range nodeIds {
		node := graph.Node(nodeId)
		outputs := node.Outputs()
		for _, out := range outputs {
			manifestOut, ok := out.(nodes.Output[manifest.Manifest])
			if !ok {
				continue
			}
			writeManifestToZip(graph, zw, node, manifestOut)
		}
	}

	return nil
}

func writeZip(out io.Writer, graph *graph.Instance) error {
	z := zip.NewWriter(out)

	if err := writeGraphManifestsToZip(graph, z); err != nil {
		return err
	}

	return z.Close()
}

func (as *EditServer) zipEndpoint(w http.ResponseWriter, r *http.Request) error {

	// Their requesting a zip of the entire graph, just zip the entire thing
	if r.URL.Path == "/zip/" {
		return writeZip(w, as.app.Graph)
	}

	resolvedNode, err := getNodeOutputFromURLPath[manifest.Manifest](r, "/zip/", as.app.Graph)
	if err != nil {
		return err
	}

	if resolvedNode == nil {
		return fmt.Errorf("zip endpoint")
	}

	z := zip.NewWriter(w)
	err = writeManifestToZip(as.app.Graph, z, resolvedNode.node, resolvedNode.output)
	if err != nil {
		return err
	}

	return z.Close()

}

func (as *EditServer) ZipEndpoint(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if recErr := recover(); recErr != nil {
			fmt.Printf("err: %s\n", recErr)
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			w.WriteHeader(http.StatusInternalServerError)
			writeJSONError(w, recErr.(error))
			// err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()

	err := as.zipEndpoint(w, r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONError(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/zip")
}
