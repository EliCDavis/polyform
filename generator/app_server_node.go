package generator

import (
	"net/http"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/schema"
)

func nodeEndpoint(as *AppServer) endpoint.Handler {
	type CreateRequest struct {
		NodeType string `json:"nodeType"`
	}

	type CreateResponse struct {
		NodeID string              `json:"nodeID"`
		Data   schema.NodeInstance `json:"data"`
	}

	type DeleteRequest struct {
		NodeID string `json:"nodeID"`
	}

	type PatchMethod struct {
	}

	type EmptyResponse struct{}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonMethod(
				func(request endpoint.Request[CreateRequest]) (CreateResponse, error) {
					node, id, err := as.app.graphInstance.CreateNode(request.Body.NodeType)
					if err != nil {
						return CreateResponse{}, err
					}
					as.AutosaveGraph()

					return CreateResponse{
						NodeID: id,
						Data:   as.app.graphInstance.NodeInstanceSchema(node),
					}, nil
				},
			),
			http.MethodDelete: endpoint.JsonMethod(
				func(request endpoint.Request[DeleteRequest]) (EmptyResponse, error) {
					as.app.graphInstance.DeleteNode(request.Body.NodeID)
					as.AutosaveGraph()
					return EmptyResponse{}, nil
				},
			),
		},
	}
}
