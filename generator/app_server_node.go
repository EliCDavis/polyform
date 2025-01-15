package generator

import (
	"fmt"
	"net/http"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
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
					if !as.app.types.KeyRegistered(request.Body.NodeType) {
						return CreateResponse{}, fmt.Errorf("no factory registered with ID %s", request.Body.NodeType)
					}

					newNode := as.app.types.New(request.Body.NodeType)
					casted, ok := newNode.(nodes.Node)
					if !ok {
						panic(fmt.Errorf("Regiestered type did not create a node. How'd ya manage that: %s", request.Body.NodeType))
					}
					as.app.buildIDsForNode(casted)

					as.AutosaveGraph()

					return CreateResponse{
						NodeID: as.app.nodeIDs[casted],
						Data:   as.app.buildNodeInstanceSchema(casted),
					}, nil
				},
			),
			http.MethodDelete: endpoint.JsonMethod(
				func(request endpoint.Request[DeleteRequest]) (EmptyResponse, error) {
					var nodeToDelete nodes.Node

					for n, id := range as.app.nodeIDs {
						if id == request.Body.NodeID {
							nodeToDelete = n
						}
					}

					for filename, producer := range as.app.Producers {
						if as.app.nodeIDs[producer.Node()] == request.Body.NodeID {
							delete(as.app.Producers, filename)
						}
					}

					delete(as.app.nodeIDs, nodeToDelete)

					as.AutosaveGraph()

					return EmptyResponse{}, nil
				},
			),
		},
	}
}
