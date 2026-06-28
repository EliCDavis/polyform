package edit

import (
	"net/http"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
)

func metadataKeyFromRequestURL(url string) string {
	const marker = "/metadata/"
	idx := strings.Index(url, marker)
	if idx == -1 {
		panic("url should contain /metadata/")
	}

	metadataPath := url[idx+len(marker):]
	if len(metadataPath) > 0 && metadataPath[0] == '/' {
		metadataPath = metadataPath[1:]
	}

	if len(metadataPath) > 0 {
		metadataPath = strings.Replace(metadataPath, "/", ".", -1)
	}
	return metadataPath
}

func graphMetadataEndpointForInstance(target *graph.Instance, saver *GraphSaver) endpoint.Handler {
	type EditRequest any

	type EmptyResponse struct{}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonMethod(
				func(request endpoint.Request[EditRequest]) (EmptyResponse, error) {
					target.SetMetadata(metadataKeyFromRequestURL(request.Url), request.Body)
					saver.Save()
					return EmptyResponse{}, nil
				},
			),

			http.MethodDelete: endpoint.Func(func(r *http.Request) error {
				target.DeleteMetadata(metadataKeyFromRequestURL(r.URL.Path))
				saver.Save()
				return nil
			}),
		},
	}
}

func graphMetadataEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
	return graphMetadataEndpointForInstance(graphInstance, saver)
}
