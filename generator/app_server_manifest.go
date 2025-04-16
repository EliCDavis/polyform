package generator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

func (as *AppServer) writeManifest(w http.ResponseWriter, components []string) error {
	nodeID := components[0]
	node := as.app.graphInstance.Node(nodeID)

	outputs := node.Outputs()
	outPortName := components[1]
	output, ok := outputs[outPortName]
	if !ok {
		return fmt.Errorf("Node %q does not contain output %q", nodeID, outPortName)
	}

	manifestOutput, ok := output.(nodes.Output[manifest.Manifest])
	if !ok {
		return fmt.Errorf("Node %q output %q does not produce a Manifest file", nodeID, outPortName)
	}
	manifest := manifestOutput.Value()

	// We're just trying to get the manifest of the node's output manifest
	if len(components) == 2 {
		w.Header().Set("Content-Type", string(endpoint.JsonContentType))
		data, err := json.Marshal(manifest)
		if err != nil {
			return err
		}
		w.Write(data)
		return nil
	}

	artifactPath := strings.Join(components[2:], "/")
	entry, ok := manifest.Entries[artifactPath]
	if !ok {
		return fmt.Errorf("Node %q output %q Manifest does not contain an entry %q", nodeID, outPortName, artifactPath)
	}

	artifact := entry.Artifact
	w.Header().Set("Content-Type", artifact.Mime())
	return artifact.Write(w)
}

func (as *AppServer) ManifestEndpoint(w http.ResponseWriter, r *http.Request) {
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
	components := strings.Split(r.URL.Path[len("/manifest/"):], "/")
	err := as.writeManifest(w, components)

	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONError(w, err)
	}
}
