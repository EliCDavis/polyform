package generator

import (
	"net/http"
	"path"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/variable"
)

const variableInstanceEndpointPath = "/variable/instance/"

func variableInstanceEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {

	// updateParameter := func(parameterId string, body []byte) error {
	// 	_, err := graphInstance.UpdateParameter(parameterId, body)
	// 	return err
	// }

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{

			// Create a new instance of a variable
			http.MethodPost: endpoint.BodyMethod[variable.JsonContainer]{
				Request: endpoint.JsonRequestReader[variable.JsonContainer]{},
				Handler: func(request endpoint.Request[variable.JsonContainer]) error {
					variablePath := request.Url[len(variableInstanceEndpointPath):]
					graphInstance.NewVariable(variablePath, request.Body.Variable)
					saver.Save()
					return nil
				},
			},

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

	updateParameter := func(parameterId string, body []byte) error {
		_, err := graphInstance.UpdateParameter(parameterId, body)
		return err
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{

			http.MethodPost: endpoint.BodyMethod[[]byte]{
				Request: endpoint.BinaryRequestReader{},
				Handler: func(request endpoint.Request[[]byte]) error {

					parameterId := path.Base(request.Url)
					err := updateParameter(parameterId, request.Body)
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
					parameterId := path.Base(r.URL.Path)
					n := graphInstance.ParameterData(parameterId)
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
