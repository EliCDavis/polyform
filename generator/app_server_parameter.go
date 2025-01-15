package generator

import (
	"net/http"
	"path"

	"github.com/EliCDavis/polyform/generator/endpoint"
)

func parameterValueEndpoint(as *AppServer) endpoint.Handler {

	updateParameter := func(parameterId string, body []byte) error {
		as.producerLock.Lock()
		defer as.producerLock.Unlock()
		_, err := as.UpdateParameter(parameterId, body)
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

					as.AutosaveGraph()

					as.incModelVersion()
					return nil
				},
			},

			http.MethodGet: endpoint.ResponseMethod[[]byte]{
				ResponseWriter: endpoint.BinaryResponseWriter{},
				Handler: func(r *http.Request) ([]byte, error) {
					as.producerLock.Lock()
					defer as.producerLock.Unlock()

					parameterId := path.Base(r.URL.Path)
					n := as.app.GetParameter(parameterId)
					return n.ToMessage(), nil
				},
			},
		},
	}
}

func parameterNameEndpoint(as *AppServer) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.ResponseMethod[string]{
				ResponseWriter: endpoint.TextResponseWriter{},
				Handler: func(r *http.Request) (string, error) {
					parameterId := path.Base(r.URL.Path)
					return as.app.GetParameter(parameterId).DisplayName(), nil
				},
			},

			http.MethodPost: endpoint.BodyMethod[string]{
				Request: endpoint.TextRequestReader{},
				Handler: func(req endpoint.Request[string]) error {
					parameterId := path.Base(req.Url)
					as.app.GetParameter(parameterId).SetName(req.Body)
					as.AutosaveGraph()
					return nil
				},
			},
		},
	}
}
