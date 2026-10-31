package edit

import (
	"log"
	"net/http"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/serialize"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/utils"
)

const (
	nodeOutputEndpointPath = "/node/output/"
)

// createNodeFromRequest routes node creation to the right constructor: only
// sub-graph boundary nodes carry a port type.
func createNodeFromRequest(instance *graph.Graph, nodeType, portType string) (nodes.Node, string, error) {
	if portType != "" {
		return instance.CreateBoundaryNode(nodeType, portType)
	}
	return instance.CreateNode(nodeType)
}

func nodeEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
	type CreateRequest struct {
		NodeType string `json:"nodeType"`
		PortType string `json:"portType,omitempty"`
	}

	type CreateResponse struct {
		NodeID string      `json:"nodeID"`
		Data   schema.Node `json:"data"`
	}

	type DeleteRequest struct {
		NodeID string `json:"nodeID"`
	}

	type EmptyResponse struct{}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonMethod(
				func(request endpoint.Request[CreateRequest]) (CreateResponse, error) {
					node, id, err := createNodeFromRequest(graphInstance.Graph, request.Body.NodeType, request.Body.PortType)
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
				func(request endpoint.Request[DeleteRequest]) (DroppedResponse, error) {
					dropped, err := graphInstance.DeleteNodeById(request.Body.NodeID)
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

// type ResponseWriter[Response any] interface {
// 	Serialize(w http.ResponseWriter, response Response) (err error)
// 	ContentType(r *http.Request) ContentType
// }

func (as *Server) writeNodeOutput(w http.ResponseWriter, r *http.Request) error {
	resolved, err := getNodeOutputFromURLPath(r, nodeOutputEndpointPath, as.Graph)
	if err != nil {
		return err
	}
	var entry manifest.Entry
	var identity string
	var source nodes.Node
	as.Graph.Evaluate(func() {
		entry = as.NodeOutputSerialization.Run(resolved.output)
		identity, source = outputIdentity(resolved.output)
	})
	return as.built.serve(w, r, identity, source, entry.Artifact)
}

func (as *Server) NodeOutputEndpoint(w http.ResponseWriter, r *http.Request) {
	// defer func() {
	// 	if recErr := recover(); recErr != nil {
	// 		fmt.Printf("err: %s\n", recErr)
	// 		fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
	// 		// err = fmt.Errorf("panic recover: %v", recErr)
	// 	}
	// }()

	w.Header().Add("Cache-Control", "no-store")

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
	NodeTypes            []schema.NodeType      `json:"nodeTypes"`
	SerializeOutputTypes []string               `json:"serializableOutputTypes"`
	PortTypes            []string               `json:"portTypes"`
	Enums                map[string]schema.Enum `json:"enums,omitempty"`
}

func goEnums() map[string]schema.Enum {
	out := map[string]schema.Enum{}
	for _, def := range enum.Definitions() {
		key := def.TypeName()[strings.LastIndex(def.TypeName(), "/")+1:]
		members := make([]enum.Member, 0)
		for _, name := range def.Names() {
			members = append(members, enum.Member{Name: name, Description: def.Describe(name)})
		}
		out[key] = schema.Enum{
			Name:    utils.CamelCaseToSpaceCase(key[strings.LastIndex(key, ".")+1:]),
			Members: members,
		}
	}
	return out
}

func nodeTypesEndpoint(graphInstance *graph.Instance, serializer *serialize.TypeSwitch[manifest.Entry]) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.JsonResponseMethod(
				func(r *http.Request) (RegisteredTypes, error) {
					b := RegisteredTypes{
						NodeTypes: graphInstance.BuildSchemaForAllNodeTypes(),
						PortTypes: nodes.KnownPortTypes(),
						Enums:     goEnums(),
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
