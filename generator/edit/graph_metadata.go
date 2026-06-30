package edit

import (
	"net/http"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
)

func metadataKeyFromRequestURL(url string) string {
	metadataPath, err := pathSuffixAfterMarker(url, "/metadata/")
	if err != nil {
		panic(err.Error())
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
