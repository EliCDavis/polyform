package generator

import (
	"net/http"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
)

func nodeMetadataEndpoint(as *AppServer) endpoint.Handler {

	type EditRequest struct {
		NodeID   string         `json:"nodeId"`
		Metadata map[string]any `json:"metadata"`
	}

	type EmptyResponse struct{}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonMethod(
				func(request endpoint.Request[EditRequest]) (EmptyResponse, error) {

					// We're making the assumption the url starts like this,
					// so assert it.
					if strings.Index(request.Url, "/node/metadata") != 0 {
						panic("url should begin with /node/metadata")
					}

					metadataPath := request.Url[len("/node/metadata"):]

					if metadataPath[0] == '/' {
						metadataPath = metadataPath[1:]
					}

					if len(metadataPath) > 0 {
						metadataPath = "." + strings.Replace(metadataPath, "/", ".", -1)
					}

					as.app.nodeMetadata.Set(request.Body.NodeID+metadataPath, request.Body.Metadata)
					as.AutosaveGraph()
					return EmptyResponse{}, nil
				},
			),
		},
	}
}
