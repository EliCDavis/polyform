package generator

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

type DeleteNodeConnectionRequest struct {
	NodeId     string `json:"nodeId"`
	InPortName string `json:"inPortName"`
}

type DeleteNodeConnectionResponse struct {
}

type CreateNodeConnectionRequest struct {
	NodeOutId   string `json:"nodeOutId"`
	OutPortName string `json:"outPortName"`
	NodeInId    string `json:"nodeInId"`
	InPortName  string `json:"inPortName"`
}

type CreateNodeConnectionResponse struct {
}

func (as *AppServer) NodeConnectionEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response any
	var err error

	switch r.Method {
	case "POST":
		createRequest, castErr := readJSON[CreateNodeConnectionRequest](r.Body)
		if castErr != nil {
			writeJSONError(w, castErr)
			return
		}
		response, err = as.nodeConnectionEndpoint_post(createRequest)

	case "DELETE":
		createRequest, castErr := readJSON[DeleteNodeConnectionRequest](r.Body)
		if castErr != nil {
			writeJSONError(w, castErr)
			return
		}
		response, err = as.nodeConnectionEndpoint_delete(createRequest)
	}

	if err != nil {
		writeJSONError(w, err)
	} else {
		data, err := json.Marshal(response)
		if err != nil {
			panic(err)
		}
		w.Write(data)
	}
}

func (as *AppServer) nodeConnectionEndpoint_post(req CreateNodeConnectionRequest) (resp CreateNodeConnectionResponse, err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()

	inNode := as.app.nodeFromID(req.NodeInId)
	outNode := as.app.nodeFromID(req.NodeOutId)

	ref := refutil.CallFuncValuesOfType(outNode, req.OutPortName)[0].(nodes.ReferencesNode)
	inNode.SetInput(
		req.InPortName,
		nodes.Output{
			NodeOutput: ref,
			Name:       req.OutPortName,
		},
	)

	return CreateNodeConnectionResponse{}, nil
}

func (as *AppServer) nodeConnectionEndpoint_delete(req DeleteNodeConnectionRequest) (resp DeleteNodeConnectionResponse, err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()

	inNode := as.app.nodeFromID(req.NodeId)

	inNode.SetInput(
		req.InPortName,
		nodes.Output{
			NodeOutput: nil,
			Name:       "",
		},
	)

	return DeleteNodeConnectionResponse{}, nil
}
