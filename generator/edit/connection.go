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
		Insert      bool   `json:"insert,omitempty"`
	}

	type EmptyResponse struct{}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonMethod(
				func(request endpoint.Request[CreateRequest]) (EmptyResponse, error) {
					connect := graphInstance.ConnectNodes
					if request.Body.Insert {
						connect = graphInstance.InsertNodeConnection
					}
					err := connect(
						request.Body.NodeOutId,
						request.Body.OutPortName,
						request.Body.NodeInId,
						request.Body.InPortName,
					)
					if err != nil {
						return EmptyResponse{}, err
					}
					saver.Save()
					return EmptyResponse{}, nil
				},
			),
			http.MethodDelete: endpoint.JsonMethod(
				func(request endpoint.Request[DeleteRequest]) (DroppedResponse, error) {
					if !graphInstance.HasNodeWithId(request.Body.NodeId) {
						return DroppedResponse{}, fmt.Errorf("no node exists with id %s", request.Body.NodeId)
					}

					dropped, err := graphInstance.DeleteNodeInputConnection(request.Body.NodeId, request.Body.InPortName)
					if err != nil {
						return DroppedResponse{}, err
					}
					saver.Save()
					return DroppedResponse{Dropped: dropped}, nil
				},
			),
		},
	}
}

type DroppedResponse struct {
	Dropped []graph.DroppedEdge `json:"dropped,omitempty"`
}
