package edit

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
)

const (
	subGraphDefinitionEndpointPath = "/subgraph/definition/"
	subGraphBoundaryEndpointPath   = "/subgraph/boundary/"
)

func subGraphDefinitionEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
	type CreateRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	type CreateResponse struct {
		NodeType schema.NodeType `json:"nodeType"`
	}

	type InfoRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	type EmptyResponse struct{}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonMethod(
				func(request endpoint.Request[CreateRequest]) (CreateResponse, error) {
					subGraphID := strings.TrimPrefix(request.Url, subGraphDefinitionEndpointPath)
					if subGraphID == "" {
						return CreateResponse{}, fmt.Errorf("sub-graph id is required")
					}

					name := request.Body.Name
					if name == "" {
						name = subGraphID
					}

					err := graphInstance.CreateSubGraph(subGraphID, name, request.Body.Description)
					if err != nil {
						return CreateResponse{}, err
					}

					typePath := subgraph.RuntimeTypePath(subGraphID)
					nodeType := graph.BuildNodeTypeSchema(typePath, graph.NewRuntimeNode(graphInstance, subGraphID))
					saver.Save()
					return CreateResponse{NodeType: nodeType}, nil
				},
			),
			http.MethodPut: endpoint.JsonMethod(
				func(request endpoint.Request[InfoRequest]) (EmptyResponse, error) {
					subGraphID := strings.TrimPrefix(request.Url, subGraphDefinitionEndpointPath)
					err := graphInstance.SetSubGraphInfo(subGraphID, request.Body.Name, request.Body.Description)
					if err != nil {
						return EmptyResponse{}, err
					}
					saver.Save()
					return EmptyResponse{}, nil
				},
			),
			http.MethodDelete: endpoint.Func(
				func(request *http.Request) error {
					subGraphID := strings.TrimPrefix(request.URL.Path, subGraphDefinitionEndpointPath)
					err := graphInstance.DeleteSubGraph(subGraphID)
					if err != nil {
						return err
					}
					saver.Save()
					return nil
				},
			),
		},
	}
}

func subGraphBoundaryEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
	type BoundaryInfoRequest struct {
		PortName string      `json:"portName"`
		Scope    graph.Scope `json:"scope"`
	}

	type BoundaryInfoResponse struct {
		NodeType schema.NodeType `json:"nodeType,omitempty"`
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonMethod(
				func(request endpoint.Request[BoundaryInfoRequest]) (BoundaryInfoResponse, error) {
					prefix := subGraphBoundaryEndpointPath
					rest := strings.TrimPrefix(request.Url, prefix)
					parts := strings.SplitN(rest, "/", 2)
					if len(parts) < 1 || parts[0] == "" {
						return BoundaryInfoResponse{}, fmt.Errorf("node id is required")
					}
					nodeID := parts[0]

					scopeInstance, err := request.Body.Scope.ResolveInstance(graphInstance)
					if err != nil {
						return BoundaryInfoResponse{}, err
					}

					err = scopeInstance.SetBoundaryNodeInfo(nodeID, request.Body.PortName)
					if err != nil {
						return BoundaryInfoResponse{}, err
					}

					subGraphID := scopeInstance.SubGraphScopeID()
					var response BoundaryInfoResponse
					if subGraphID != "" {
						typePath := subgraph.RuntimeTypePath(subGraphID)
						response.NodeType = graph.BuildNodeTypeSchema(typePath, graph.NewRuntimeNode(graphInstance, subGraphID))
					}

					saver.Save()
					return response, nil
				},
			),
		},
	}
}

func scopedNodeEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
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
					scope, err := parseSubGraphScopeFromURL(request.Url)
					if err != nil {
						return CreateResponse{}, err
					}

					scopeInstance, err := scope.ResolveInstance(graphInstance)
					if err != nil {
						return CreateResponse{}, err
					}

					node, id, err := createNodeFromRequest(scopeInstance, request.Body.NodeType, request.Body.PortType)
					if err != nil {
						return CreateResponse{}, err
					}
					saver.Save()

					return CreateResponse{
						NodeID: id,
						Data:   scopeInstance.NodeInstanceSchema(node),
					}, nil
				},
			),
			http.MethodDelete: endpoint.JsonMethod(
				func(request endpoint.Request[DeleteRequest]) (EmptyResponse, error) {
					scope, err := parseSubGraphScopeFromURL(request.Url)
					if err != nil {
						return EmptyResponse{}, err
					}

					scopeInstance, err := scope.ResolveInstance(graphInstance)
					if err != nil {
						return EmptyResponse{}, err
					}

					if !scopeInstance.HasNodeWithId(request.Body.NodeID) {
						return EmptyResponse{}, fmt.Errorf("no node exists with id %s", request.Body.NodeID)
					}

					scopeInstance.DeleteNodeById(request.Body.NodeID)
					saver.Save()
					return EmptyResponse{}, nil
				},
			),
		},
	}
}

func scopedNodeConnectionEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
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
					scope, err := parseSubGraphScopeFromURL(request.Url)
					if err != nil {
						return EmptyResponse{}, err
					}

					scopeInstance, err := scope.ResolveInstance(graphInstance)
					if err != nil {
						return EmptyResponse{}, err
					}

					scopeInstance.ConnectNodes(
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
					scope, err := parseSubGraphScopeFromURL(request.Url)
					if err != nil {
						return EmptyResponse{}, err
					}

					scopeInstance, err := scope.ResolveInstance(graphInstance)
					if err != nil {
						return EmptyResponse{}, err
					}

					if !scopeInstance.HasNodeWithId(request.Body.NodeId) {
						return EmptyResponse{}, fmt.Errorf("no node exists with id %s", request.Body.NodeId)
					}

					scopeInstance.DeleteNodeInputConnection(request.Body.NodeId, request.Body.InPortName)
					saver.Save()
					return EmptyResponse{}, nil
				},
			),
		},
	}
}

func parseSubGraphScopeFromURL(urlPath string) (graph.Scope, error) {
	rest, err := pathSuffixAfterMarker(urlPath, "/graph/subgraph/")
	if err != nil {
		return "", fmt.Errorf("invalid scoped graph url: %s", urlPath)
	}

	subGraphID := rest
	for _, suffix := range []string{"/node", "/connection", "/metadata/"} {
		if idx := strings.Index(rest, suffix); idx != -1 {
			subGraphID = rest[:idx]
			break
		}
	}

	if subGraphID == "" {
		return "", fmt.Errorf("invalid scoped graph url: %s", urlPath)
	}
	return graph.SubGraphScope(subGraphID), nil
}

func scopedGraphHandler(graphInstance *graph.Instance, saver *GraphSaver) http.Handler {
	nodeHandler := scopedNodeEndpoint(graphInstance, saver)
	connectionHandler := scopedNodeConnectionEndpoint(graphInstance, saver)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/connection") {
			connectionHandler.ServeHTTP(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/node") {
			nodeHandler.ServeHTTP(w, r)
			return
		}
		if strings.Contains(r.URL.Path, "/metadata/") {
			scope, err := parseSubGraphScopeFromURL(r.URL.Path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			scopeInstance, err := scope.ResolveInstance(graphInstance)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			graphMetadataEndpointForInstance(scopeInstance, saver).ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})
}
