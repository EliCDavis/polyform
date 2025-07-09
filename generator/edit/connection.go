package edit

import (
	"fmt"
	"net/http"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
)

func nodeConnectionEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
	type DeleteRequest struct {
		NodeId     string `json:"nodeId"`
		InPortName string `json:"inPortName"`
	}

	type CreateRequest struct {
		NodeOutId   string `json:"nodeOutId"`
		OutPortName string `json:"outPortName"`
		NodeInId    string `json:"nodeInId"`
		InPortName  string `json:"inPortName"`
	}

	type EmptyResponse struct{}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonMethod(
				func(request endpoint.Request[CreateRequest]) (EmptyResponse, error) {
					graphInstance.
						ConnectNodes(
							request.Body.NodeOutId,
							request.Body.OutPortName,
							request.Body.NodeInId,
							request.Body.InPortName,
						)
					saver.Save()
					return EmptyResponse{}, nil
				},
			),
			http.MethodDelete: endpoint.JsonMethod(
				func(request endpoint.Request[DeleteRequest]) (EmptyResponse, error) {
					if !graphInstance.HasNodeWithId(request.Body.NodeId) {
						return EmptyResponse{}, fmt.Errorf("no node exists with id %s", request.Body.NodeId)
					}

					graphInstance.
						DeleteNodeInputConnection(
							request.Body.NodeId,
							request.Body.InPortName,
						)
					saver.Save()
					return EmptyResponse{}, nil
				},
			),
		},
	}
}
