package generator

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/EliCDavis/polyform/nodes"
)

type CreateNodeRequest struct {
	NodeType string `json:"nodeType"`
}

type CreateNodeResponse struct {
	NodeID string             `json:"nodeID"`
	Data   NodeInstanceSchema `json:"data"`
}

type DeleteNodeRequest struct {
	NodeID string `json:"nodeID"`
}

func (as *AppServer) NodeEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response any

	switch r.Method {
	case "POST":
		createRequest, err := readJSON[CreateNodeRequest](r.Body)
		if err != nil {
			panic(err)
		}

		resp, err := as.nodeEncpoint_Post(createRequest)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSONError(w, err)
		} else {
			data, err := json.Marshal(resp)
			if err != nil {
				panic(err)
			}
			w.Write(data)
		}
		return

	case "DELETE":
		deleteRequest, err := readJSON[DeleteNodeRequest](r.Body)
		if err != nil {
			panic(err)
		}

		var nodeToDelete nodes.Node
		for n, id := range as.app.nodeIDs {
			if id == deleteRequest.NodeID {
				nodeToDelete = n
			}
		}

		delete(as.app.nodeIDs, nodeToDelete)

	default:
		panic(fmt.Errorf("node endpoint has not implemented HTTP method: '%s'", r.Method))
	}

	data, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	w.Write(data)
}

func (as *AppServer) nodeEncpoint_Post(req CreateNodeRequest) (resp CreateNodeResponse, err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()
	if as.app.types.KeyRegistered(req.NodeType) {
		newNode := as.app.types.New(req.NodeType)
		casted, ok := newNode.(nodes.Node)
		if !ok {
			panic(fmt.Errorf("what the fuck: %s", req.NodeType))
		}
		as.app.buildIDsForNode(casted)

		return CreateNodeResponse{
			NodeID: as.app.nodeIDs[casted],
			Data:   as.app.buildNodeInstanceSchema(casted),
		}, nil
	}

	return CreateNodeResponse{}, fmt.Errorf("no factory registered with ID %s", req.NodeType)
}
