package edit

import (
	"net/http"
	"path"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
)

func parameterValueEndpoint(graphInstance *graph.Graph, saver *GraphSaver) endpoint.Handler {

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
					return graphInstance.ParameterData(parameterId)
				},
			},
		},
	}
}

func parameterNameEndpoint(graphInstance *graph.Graph, saver *GraphSaver) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.ResponseMethod[string]{
				ResponseWriter: endpoint.TextResponseWriter{},
				Handler: func(r *http.Request) (string, error) {
					parameterId := path.Base(r.URL.Path)
					parameter, err := graphInstance.Parameter(parameterId)
					if err != nil {
						return "", err
					}
					return parameter.DisplayName(), nil
				},
			},

			http.MethodPost: endpoint.BodyMethod[string]{
				Request: endpoint.TextRequestReader{},
				Handler: func(req endpoint.Request[string]) error {
					parameterId := path.Base(req.Url)
					parameter, err := graphInstance.Parameter(parameterId)
					if err != nil {
						return err
					}
					parameter.SetName(req.Body)
					saver.Save()
					return nil
				},
			},
		},
	}
}

func parameterDescriptionEndpoint(graphInstance *graph.Graph, saver *GraphSaver) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.BodyMethod[string]{
				Request: endpoint.TextRequestReader{},
				Handler: func(req endpoint.Request[string]) error {
					parameterId := path.Base(req.Url)
					parameter, err := graphInstance.Parameter(parameterId)
					if err != nil {
						return err
					}
					parameter.SetDescription(req.Body)
					saver.Save()
					return nil
				},
			},
		},
	}
}
