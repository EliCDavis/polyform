package edit

import (
	"net/http"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
)

func graphMetadataEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {

	urlToMetadataKey := func(url string) string {
		// We're making the assumption the url starts like this,
		// so assert it.
		if strings.Index(url, "/graph/metadata") != 0 {
			panic("url should begin with /graph/metadata")
		}

		metadataPath := url[len("/graph/metadata"):]

		if metadataPath[0] == '/' {
			metadataPath = metadataPath[1:]
		}

		if len(metadataPath) > 0 {
			metadataPath = strings.Replace(metadataPath, "/", ".", -1)
		}
		return metadataPath
	}

	type EditRequest any

	type EmptyResponse struct{}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonMethod(
				func(request endpoint.Request[EditRequest]) (EmptyResponse, error) {
					graphInstance.SetMetadata(urlToMetadataKey(request.Url), request.Body)
					saver.Save()
					return EmptyResponse{}, nil
				},
			),

			http.MethodDelete: endpoint.Func(func(r *http.Request) error {
				graphInstance.DeleteMetadata(urlToMetadataKey(r.URL.Path))
				saver.Save()
				return nil
			}),
		},
	}
}
