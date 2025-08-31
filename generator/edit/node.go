package edit

import (
	"log"
	"net/http"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/serialize"
)

const (
	nodeOutputEndpointPath = "/node/output/"
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
					graphInstance.DeleteNodeById(request.Body.NodeID)
					saver.Save()
					return EmptyResponse{}, nil
				},
			),
		},
	}
}

// type ResponseWriter[Response any] interface {
// 	Serialize(w http.ResponseWriter, response Response) (err error)
// 	ContentType(r *http.Request) ContentType
// }

func (as *Server) writeNodeOutput(w http.ResponseWriter, r *http.Request) error {
	resolved, err := getNodeOutputFromURLPath(r, nodeOutputEndpointPath, as.Graph)
	if err != nil {
		return err
	}
	entry := as.NodeOutputSerialization.Run(resolved.output)

	artifact := entry.Artifact
	w.Header().Set("Content-Type", artifact.Mime())
	return artifact.Write(w)
}

func (as *Server) NodeOutputEndpoint(w http.ResponseWriter, r *http.Request) {
	// defer func() {
	// 	if recErr := recover(); recErr != nil {
	// 		fmt.Printf("err: %s\n", recErr)
	// 		fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
	// 		// err = fmt.Errorf("panic recover: %v", recErr)
	// 	}
	// }()

	w.Header().Add("Cache-Control", "no-cache")

	// Required for sharedMemoryForWorkers to work
	w.Header().Add("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Add("Cross-Origin-Resource-Policy", "cross-origin")
	w.Header().Add("Cross-Origin-Embedder-Policy", "require-corp")

	err := as.writeNodeOutput(w, r)

	if err != nil {
		log.Print(err)
		w.Header().Set("Content-Type", string(endpoint.JsonContentType))
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONError(w, err)
	}
}

type RegisteredTypes struct {
	NodeTypes            []schema.NodeType `json:"nodeTypes"`
	SerializeOutputTypes []string          `json:"serializableOutputTypes"`
}

func nodeTypesEndpoint(graphInstance *graph.Instance, serializer *serialize.TypeSwitch[manifest.Entry]) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.JsonResponseMethod(
				func(r *http.Request) (RegisteredTypes, error) {
					b := RegisteredTypes{
						NodeTypes: graphInstance.BuildSchemaForAllNodeTypes(),
					}
					if serializer != nil {
						b.SerializeOutputTypes = serializer.Types()
					}
					return b, nil
				},
			),
		},
	}
}
