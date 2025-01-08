package generator

import (
	"encoding/json"
	"net/http"

	"github.com/EliCDavis/polyform/generator/endpoint"
)

func nodeMetadataEndpoint(as *AppServer) endpoint.Handler {

	type EditRequest struct {
		NodeID   string          `json:"nodeId"`
		Metadata json.RawMessage `json:"metadata"`
	}

	type EmptyResponse struct{}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonMethod(
				func(request endpoint.Request[EditRequest]) (EmptyResponse, error) {
					as.app.nodeMetadata[request.Body.NodeID] = request.Body.Metadata
					as.AutosaveGraph()
					return EmptyResponse{}, nil
				},
			),
		},
	}
}
