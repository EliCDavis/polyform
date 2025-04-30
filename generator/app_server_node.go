package generator

import (
	"net/http"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
)

func nodeEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
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

	type EmptyResponse struct{}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonMethod(
				func(request endpoint.Request[CreateRequest]) (CreateResponse, error) {
					node, id, err := graphInstance.CreateNode(request.Body.NodeType)
					if err != nil {
						return CreateResponse{}, err
					}
					saver.Save()

					return CreateResponse{
						NodeID: id,
						Data:   graphInstance.NodeInstanceSchema(node),
					}, nil
				},
			),
			http.MethodDelete: endpoint.JsonMethod(
				func(request endpoint.Request[DeleteRequest]) (EmptyResponse, error) {
					graphInstance.DeleteNode(request.Body.NodeID)
					saver.Save()
					return EmptyResponse{}, nil
				},
			),
		},
	}
}
