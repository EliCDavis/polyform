package edit

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

type resolvedNodeUrl[T any] struct {
	nodeID       string
	node         nodes.Node
	outputName   string
	output       nodes.Output[T]
	remainingUrl string
}

func getNodeOutputFromURLPath[T any](r *http.Request, base string, graph *graph.Instance) (*resolvedNodeUrl[T], error) {
	if strings.Index(r.URL.Path, base) != 0 {
		panic(fmt.Errorf("expected url to begin with manifest, instead: %q", r.URL.Path))
	}
	components := strings.Split(r.URL.Path[len(base):], "/")

	nodeID := components[0]
	node := graph.Node(nodeID)

	outputs := node.Outputs()
	outPortName := components[1]
	output, ok := outputs[outPortName]
	if !ok {
		return nil, fmt.Errorf("Node %q does not contain output %q", nodeID, outPortName)
	}

	manifestOutput, ok := output.(nodes.Output[T])
	if !ok {
		return nil, fmt.Errorf("Node %q output %q does not produce specified type", nodeID, outPortName)
	}
	return &resolvedNodeUrl[T]{
		nodeID:       nodeID,
		node:         node,
		outputName:   outPortName,
		output:       manifestOutput,
		remainingUrl: strings.Join(components[2:], "/"),
	}, nil
}

func (as *Server) writeManifest(w http.ResponseWriter, r *http.Request) error {
	resolvedNode, err := getNodeOutputFromURLPath[manifest.Manifest](r, "/manifest/", as.Graph)
	if err != nil {
		return err
	}

	manifest := resolvedNode.output.Value()

	// We're just trying to get the manifest of the node's output manifest
	if resolvedNode.remainingUrl == "" {
		w.Header().Set("Content-Type", string(endpoint.JsonContentType))
		data, err := json.Marshal(manifest)
		if err != nil {
			return err
		}
		w.Write(data)
		return nil
	}

	entry, ok := manifest.Entries[resolvedNode.remainingUrl]
	if !ok {
		return fmt.Errorf("Node %q output %q Manifest does not contain an entry %q", resolvedNode.nodeID, resolvedNode.outputName, resolvedNode.remainingUrl)
	}

	artifact := entry.Artifact
	w.Header().Set("Content-Type", artifact.Mime())
	return artifact.Write(w)
}

func (as *Server) ManifestEndpoint(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if recErr := recover(); recErr != nil {
			fmt.Printf("err: %s\n", recErr)
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			// err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()

	w.Header().Add("Cache-Control", "no-cache")

	// Required for sharedMemoryForWorkers to work
	w.Header().Add("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Add("Cross-Origin-Resource-Policy", "cross-origin")
	w.Header().Add("Cross-Origin-Embedder-Policy", "require-corp")

	if strings.Index(r.URL.Path, "/manifest/") != 0 {
		panic(fmt.Errorf("expected url to begin with manifest, instead: %q", r.URL.Path))
	}
	err := as.writeManifest(w, r)

	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONError(w, err)
	}
}
