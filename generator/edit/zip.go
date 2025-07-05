package edit

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
)

func writeZip(out io.Writer, g *graph.Instance) error {
	z := zip.NewWriter(out)

	if err := graph.WriteToZip(g, z); err != nil {
		return err
	}

	return z.Close()
}

func (as *Server) zipEndpoint(w http.ResponseWriter, r *http.Request) error {

	// Their requesting a zip of the entire graph, just zip the entire thing
	if r.URL.Path == "/zip/" {
		return writeZip(w, as.Graph)
	}

	resolvedNode, err := getNodeOutputFromURLPath[manifest.Manifest](r, "/zip/", as.Graph)
	if err != nil {
		return err
	}

	if resolvedNode == nil {
		return fmt.Errorf("zip endpoint")
	}

	z := zip.NewWriter(w)
	err = graph.WriteManifestToZip(as.Graph, z, resolvedNode.node, resolvedNode.output)
	if err != nil {
		return err
	}

	return z.Close()

}

func (as *Server) ZipEndpoint(w http.ResponseWriter, r *http.Request) {
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
