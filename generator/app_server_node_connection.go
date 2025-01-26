package generator

import (
	"net/http"

	"github.com/EliCDavis/polyform/generator/endpoint"
)

func nodeConnectionEndpoint(as *AppServer) endpoint.Handler {
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
					as.app.
						graphInstance.
						ConnectNodes(
							request.Body.NodeOutId,
							request.Body.OutPortName,
							request.Body.NodeInId,
							request.Body.InPortName,
						)
					as.incModelVersion()
					as.AutosaveGraph()
					return EmptyResponse{}, nil
				},
			),
			http.MethodDelete: endpoint.JsonMethod(
				func(request endpoint.Request[DeleteRequest]) (EmptyResponse, error) {
					as.app.
						graphInstance.
						DeleteNodeInputConnection(
							request.Body.NodeId,
							request.Body.InPortName,
						)
					as.incModelVersion()
					as.AutosaveGraph()
					return EmptyResponse{}, nil
				},
			),
		},
	}
}
