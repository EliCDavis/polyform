package edit

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
)

const nodeLiteralEndpointPath = "/node/literal/"

func literalLocation(url string) (nodeID, port string, err error) {
	rest, err := pathSuffixAfterMarker(url, nodeLiteralEndpointPath)
	if err != nil {
		return "", "", err
	}
	nodeID, port, ok := strings.Cut(rest, "/")
	if !ok || nodeID == "" || port == "" {
		return "", "", fmt.Errorf("url should name a node and a port after %q", nodeLiteralEndpointPath)
	}
	return nodeID, port, nil
}

func nodeLiteralEndpoint(graphInstance *graph.Graph, saver *GraphSaver) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.BodyMethod[[]byte]{
				Request: endpoint.BinaryRequestReader{},
				Handler: func(request endpoint.Request[[]byte]) error {
					nodeID, port, err := literalLocation(request.Url)
					if err != nil {
						return err
					}
					if err := graphInstance.SetLiteral(nodeID, port, request.Body); err != nil {
						return err
					}
					saver.Save()
					return nil
				},
			},

			http.MethodGet: endpoint.ResponseMethod[[]byte]{
				ResponseWriter: endpoint.BinaryResponseWriter{},
				Handler: func(r *http.Request) ([]byte, error) {
					nodeID, port, err := literalLocation(r.URL.Path)
					if err != nil {
						return nil, err
					}
					return graphInstance.LiteralValue(nodeID, port)
				},
			},
		},
	}
}
