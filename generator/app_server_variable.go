package generator

import (
	"net/http"
	"path"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/variable"
)

const (
	variableInstanceEndpointPath = "/variable/instance/"
	variableValueEndpointPath    = "/variable/value/"
)

func variableInstanceEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {

	type CreateVariableResponse struct {
		NodeType schema.NodeType `json:"nodeType"`
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{

			// Create a new instance of a variable
			http.MethodPost: endpoint.JsonMethod(func(request endpoint.Request[variable.JsonContainer]) (CreateVariableResponse, error) {
				variablePath := request.Url[len(variableInstanceEndpointPath):]
				registeredType := graphInstance.NewVariable(variablePath, request.Body.Variable)
				saver.Save()
				return CreateVariableResponse{
					NodeType: graph.BuildNodeTypeSchema(registeredType, request.Body.Variable.NodeReference()),
				}, nil
			}),

			http.MethodGet: endpoint.ResponseMethod[variable.Variable]{
				ResponseWriter: endpoint.JsonResponseWriter[variable.Variable]{},
				Handler: func(request *http.Request) (variable.Variable, error) {
					variablePath := request.URL.Path[len(variableInstanceEndpointPath):]
					graphInstance.GetVariable(variablePath)
					return graphInstance.GetVariable(variablePath), nil
				},
			},
		},
	}
}

func variableValueEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {

	updateVariable := func(variablePath string, body []byte) error {
		_, err := graphInstance.UpdateVariable(variablePath, body)
		return err
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{

			http.MethodPost: endpoint.BodyMethod[[]byte]{
				Request: endpoint.BinaryRequestReader{},
				Handler: func(request endpoint.Request[[]byte]) error {

					variablePath := request.Url[len(variableValueEndpointPath):]
					err := updateVariable(variablePath, request.Body)
					if err != nil {
						return err
					}

					saver.Save()
					return nil
				},
			},

			http.MethodGet: endpoint.ResponseMethod[[]byte]{
				ResponseWriter: endpoint.BinaryResponseWriter{},
				Handler: func(r *http.Request) ([]byte, error) {
					variablePath := r.URL.Path[len(variableValueEndpointPath):]
					n := graphInstance.VariableData(variablePath)
					return n, nil
				},
			},
		},
	}
}

func variableNameEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.ResponseMethod[string]{
				ResponseWriter: endpoint.TextResponseWriter{},
				Handler: func(r *http.Request) (string, error) {
					parameterId := path.Base(r.URL.Path)
					return graphInstance.Parameter(parameterId).DisplayName(), nil
				},
			},

			http.MethodPost: endpoint.BodyMethod[string]{
				Request: endpoint.TextRequestReader{},
				Handler: func(req endpoint.Request[string]) error {
					parameterId := path.Base(req.Url)
					graphInstance.Parameter(parameterId).SetName(req.Body)
					saver.Save()
					return nil
				},
			},
		},
	}
}

func variableDescriptionEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.BodyMethod[string]{
				Request: endpoint.TextRequestReader{},
				Handler: func(req endpoint.Request[string]) error {
					parameterId := path.Base(req.Url)
					graphInstance.Parameter(parameterId).SetDescription(req.Body)
					saver.Save()
					return nil
				},
			},
		},
	}
}
