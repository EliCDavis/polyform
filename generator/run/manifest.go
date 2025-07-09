package run

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
)

type resolvedNodeUrl[T any] struct {
	nodeID       string
	node         nodes.Node
	outputName   string
	output       nodes.Output[T]
	remainingUrl string
}

func urlComponents(url, base string) []string {
	if strings.Index(url, base) != 0 {
		panic(fmt.Errorf("expected url %q to begin with %q", url, base))
	}
	return strings.Split(url[len(base):], "/")
}

func getNodeOutputFromURLPath[T any](requestUrl string, base string, graph *graph.Instance) (*resolvedNodeUrl[T], error) {
	components := urlComponents(requestUrl, base)

	if len(components) == 0 {
		return nil, fmt.Errorf("url is missing manifest node name or id")
	}

	if len(components) == 1 {
		return nil, fmt.Errorf("url is missing manifest port name")
	}

	nodeID := components[0]
	if !graph.HasNodeWithId(nodeID) {
		return nil, fmt.Errorf("no node exists with id %s", nodeID)
	}
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

		resolvedNode, err := getNodeOutputFromURLPath[manifest.Manifest](request.Url, "/manifest/", s.Graph)
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
			return nil, fmt.Errorf("unable to parse url: %q", r.URL.Path)
		}
		id := components[0]
		entryPath := components[1]

		m, ok := s.cache.Get(id)
		if !ok {
			return nil, fmt.Errorf("contain no manifest with id %q", id)
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
					return v, json.Unmarshal(data, &v)
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
