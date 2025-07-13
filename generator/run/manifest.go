package run

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
)

type nodeAndOutput[T any] struct {
	nodeID     string
	node       nodes.Node
	outputName string
	output     nodes.Output[T]
}

func urlComponents(url, base string) []string {
	if strings.Index(url, base) != 0 {
		panic(fmt.Errorf("expected url %q to begin with %q", url, base))
	}
	components := strings.Split(url[len(base):], "/")
	if len(components) == 1 && components[0] == "" {
		return nil
	}

	return components
}

func getNodeOutputFromURLPath[T any](requestUrl string, base string, graph []nodeAndOutput[T]) (*nodeAndOutput[T], error) {
	components := urlComponents(requestUrl, base)

	if len(components) == 0 {
		return nil, fmt.Errorf("url is missing manifest node name or id")
	}

	if len(components) == 1 {
		return nil, fmt.Errorf("url is missing manifest port name")
	}

	for _, n := range graph {
		if n.nodeID != components[0] {
			continue
		}

		if n.outputName != components[1] {
			continue
		}

		return &n, nil
	}

	return nil, fmt.Errorf("%s/%s does not match any node/port combination that produces a manifest", components[0], components[1])
}

func (s *Server) manifestEndpoint() http.Handler {
	post := func(request endpoint.Request[*variable.Profile]) (CreateManifestResponse, error) {
		s.lock.Lock()
		defer s.lock.Unlock()

		response := CreateManifestResponse{}
		if request.Body != nil {
			err := s.Graph.ApplyProfile(*request.Body)
			if err != nil {
				return response, fmt.Errorf("unable to apply profile: %w", err)
			}
		}

		resolvedNode, err := getNodeOutputFromURLPath(request.Url, "/manifest/", s.manifestNodes)
		if err != nil {
			return response, err
		}

		response.Manifest = resolvedNode.output.Value()
		response.Id = s.cache.Add(response.Manifest)

		return response, nil
	}

	get := func(r *http.Request) ([]byte, error) {
		s.lock.RLock()
		defer s.lock.RUnlock()

		components := urlComponents(r.URL.Path, "/manifest/")
		if len(components) != 2 {
			return nil, fmt.Errorf("invalid url: %q", r.URL.Path)
		}
		id := components[0]
		entryPath := components[1]

		m, ok := s.cache.Get(id)
		if !ok {
			return nil, fmt.Errorf("no manifest exists with id %q", id)
		}

		entry, ok := m.Entries[entryPath]
		if !ok {
			return nil, fmt.Errorf("manifest %q contains no entry %q", id, entryPath)
		}

		out := &bytes.Buffer{}
		err := entry.Artifact.Write(out)
		return out.Bytes(), err
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.BodyResponseMethod[*variable.Profile, CreateManifestResponse]{
				ResponseWriter: endpoint.JsonResponseWriter[CreateManifestResponse]{},
				Request: endpoint.RequestReaderFunc[*variable.Profile](func(r *http.Request) (*variable.Profile, error) {
					data, err := io.ReadAll(r.Body)
					if err != nil {
						return nil, err
					}
					if len(data) == 0 {
						return nil, nil
					}
					var v *variable.Profile
					err = json.Unmarshal(data, &v)
					if err != nil {
						return nil, errors.New("unable to interpret variable profile")
					}
					return v, nil
				}),
				Handler: post,
			},
			http.MethodGet: endpoint.ResponseMethod[[]byte]{
				ResponseWriter: endpoint.BinaryResponseWriter{},
				Handler:        get,
			},
		},
	}
}
