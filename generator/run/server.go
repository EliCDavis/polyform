package run

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
)

type Server struct {
	Graph *graph.Instance

	Host, Port string
	Tls        bool
	CertPath   string
	KeyPath    string

	CacheSize int

	cache *lruCache[manifest.Manifest]
	lock  sync.RWMutex
}

func (s *Server) Handler() (*http.ServeMux, error) {
	s.cache = &lruCache[manifest.Manifest]{
		max:  s.CacheSize,
		data: make(map[string]*lruCacheEntry[manifest.Manifest]),
	}

	mux := http.NewServeMux()

	mux.Handle("/manifest/", s.manifestEndpoint())

	return mux, nil
}

func (as *Server) Serve() error {
	mux, err := as.Handler()
	if err != nil {
		return err
	}

	protocol := "http"
	if as.Tls {
		protocol = "https"
	}

	addr := fmt.Sprintf("%s:%s", as.Host, as.Port)

	fmt.Printf("Serving over: %s://%s\n", protocol, addr)
	if as.Tls {
		return http.ListenAndServeTLS(addr, as.CertPath, as.KeyPath, mux)
	}

	return http.ListenAndServe(addr, mux)
}
