package generator

import (
	"net/http"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
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
					inNode := as.app.nodeFromID(request.Body.NodeInId)
					outNode := as.app.nodeFromID(request.Body.NodeOutId)
					outPortVals := refutil.CallFuncValuesOfType(outNode, request.Body.OutPortName)

					ref := outPortVals[0].(nodes.NodeOutputReference)
					inNode.SetInput(
						request.Body.InPortName,
						nodes.Output{
							NodeOutput: ref,
						},
					)
					as.incModelVersion()

					return EmptyResponse{}, nil
				},
			),
			http.MethodDelete: endpoint.JsonMethod(
				func(request endpoint.Request[DeleteRequest]) (EmptyResponse, error) {
					inNode := as.app.nodeFromID(request.Body.NodeId)

					inNode.SetInput(
						request.Body.InPortName,
						nodes.Output{
							NodeOutput: nil,
						},
					)
					as.incModelVersion()

					return EmptyResponse{}, nil
				},
			),
		},
	}
}
