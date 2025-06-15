package generator

import (
	"net/http"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/variable"
)

const (
	variableInstanceEndpointPath        = "/variable/instance/"
	variableValueEndpointPath           = "/variable/value/"
	variableNameDescriptionEndpointPath = "/variable/info/"
)

func variableInstanceEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {

	type CreateVariableRequest struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	}

	type CreateVariableResponse struct {
		NodeType schema.NodeType `json:"nodeType"`
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{

			// Create a new instance of a variable
			http.MethodPost: endpoint.JsonMethod(func(request endpoint.Request[CreateVariableRequest]) (CreateVariableResponse, error) {
				variablePath := request.Url[len(variableInstanceEndpointPath):]
				variableInstance, err := variable.CreateVariable(request.Body.Type)
				if err != nil {
					return CreateVariableResponse{}, err
				}

				registeredType := graphInstance.NewVariable(variablePath, variableInstance)
				err = graphInstance.SetVariableDescription(variablePath, request.Body.Description)
				if err != nil {
					return CreateVariableResponse{}, err
				}
				saver.Save()
				return CreateVariableResponse{
					NodeType: graph.BuildNodeTypeSchema(registeredType, variableInstance.NodeReference()),
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

			http.MethodDelete: endpoint.Func(
				func(request *http.Request) error {
					variablePath := request.URL.Path[len(variableInstanceEndpointPath):]
					graphInstance.DeleteVariable(variablePath)
					saver.Save()
					return nil
				},
			),
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
					return graphInstance.VariableData(variablePath)
				},
			},
		},
	}
}

func variableInfoEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {

	type SetVariableInfoBody struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.JsonMethod(func(request endpoint.Request[SetVariableInfoBody]) (struct{}, error) {
				variablePath := request.Url[len(variableNameDescriptionEndpointPath):]
				err := graphInstance.SetVariableInfo(variablePath, request.Body.Name, request.Body.Description)
				saver.Save()
				return struct{}{}, err
			}),
		},
	}
}
