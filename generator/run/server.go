package run

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

func getAvailableManifests(g *graph.Instance) ([]nodeAndOutput[manifest.Manifest], error) {
	manifestEndpoints := make([]nodeAndOutput[manifest.Manifest], 0)
	err := graph.ForeachManifestNodeOutput(g, func(nodeId string, node nodes.Node, output nodes.Output[manifest.Manifest]) error {

		name := nodeId
		portName, named := g.IsPortNamed(node, output.Name())
		if named {
			name = portName
		}

		manifestEndpoints = append(manifestEndpoints, nodeAndOutput[manifest.Manifest]{
			nodeID:     name,
			node:       node,
			outputName: output.Name(),
			output:     output,
		})
		return nil
	})

	return manifestEndpoints, err
}

func computeManifestEndpoints(nodes []nodeAndOutput[manifest.Manifest]) (endpoint.StaticResponse, error) {
	manifests := make([]AvailableManifest, len(nodes))
	for i, n := range nodes {
		manifests[i] = AvailableManifest{
			Name: n.nodeID,
			Port: n.outputName,
		}
	}
	return endpoint.StaticJson(manifests)
}

type Server struct {
	Graph *graph.Instance

	Host, Port string
	Tls        bool
	CertPath   string
	KeyPath    string

	CacheSize int

	cache         *lruCache[manifest.Manifest]
	lock          sync.RWMutex
	manifestNodes []nodeAndOutput[manifest.Manifest]
}

func (s *Server) Handler() (*http.ServeMux, error) {
	availableManifests, err := getAvailableManifests(s.Graph)
	if err != nil {
		return nil, fmt.Errorf("unable to compute graph manifests: %w", err)
	}

	manifestEndpoints, err := computeManifestEndpoints(availableManifests)
	if err != nil {
		return nil, fmt.Errorf("unable to compute static /manifests response: %w", err)
	}

	variableEndpoint, err := endpoint.StaticJson(s.Graph.SwaggerDefinition().Properties)
	if err != nil {
		return nil, fmt.Errorf("unable to compute static /profile response: %w", err)
	}

	s.manifestNodes = availableManifests
	s.cache = &lruCache[manifest.Manifest]{
		max:  s.CacheSize,
		data: make(map[string]*lruCacheEntry[manifest.Manifest]),
	}

	mux := http.NewServeMux()
	mux.Handle("/manifest/", s.manifestEndpoint())
	mux.Handle("/manifests", endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: manifestEndpoints,
		},
	})
	mux.Handle("/profile", endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: variableEndpoint,
		},
	})
	return mux, nil
}

func (s *Server) protocol() string {
	if s.Tls {
		return "https"
	}
	return "http"
}

func (s *Server) Serve() error {
	mux, err := s.Handler()
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	fmt.Printf("Serving over: %s://%s\n", s.protocol(), addr)
	if s.Tls {
		return http.ListenAndServeTLS(addr, s.CertPath, s.KeyPath, mux)
	}

	return http.ListenAndServe(addr, mux)
}
