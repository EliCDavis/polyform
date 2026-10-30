package edit

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

type builtOutput struct {
	identity string
	// Held so its address, which the identity names, is not reused.
	source nodes.Node
	mime   string
	data   []byte
}

type builtOutputs struct {
	mu    sync.Mutex
	built map[string]builtOutput
}

// Read under Graph.Evaluate: an output's identity can evaluate it.
func outputIdentity(port nodes.OutputPort) (string, nodes.Node) {
	return nodes.Identity(port), nodes.Source(port).Node()
}

func (b *builtOutputs) serve(w http.ResponseWriter, r *http.Request, identity string, source nodes.Node, artifact manifest.Artifact) error {
	key := r.URL.Path + "?scope=" + r.URL.Query().Get("scope")

	b.mu.Lock()
	cached, ok := b.built[key]
	b.mu.Unlock()

	if !ok || cached.identity != identity {
		buf := &bytes.Buffer{}
		if err := artifact.Write(buf); err != nil {
			return err
		}
		cached = builtOutput{identity: identity, source: source, mime: artifact.Mime(), data: buf.Bytes()}

		b.mu.Lock()
		if b.built == nil {
			b.built = make(map[string]builtOutput)
		}
		b.built[key] = cached
		b.mu.Unlock()
	}

	w.Header().Set("Content-Type", cached.mime)
	_, err := w.Write(cached.data)
	return err
}
