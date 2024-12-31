package generator

import (
	"net/http"

	"github.com/EliCDavis/polyform/generator/endpoint"
)

func graphEndpoint(as *AppServer) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.ResponseMethod[[]byte]{
				ResponseWriter: endpoint.BinaryResponseWriter{},
				Handler: func(r *http.Request) ([]byte, error) {
					return as.app.Graph(), nil
				},
			},

			http.MethodPost: endpoint.BodyMethod[[]byte]{
				Request: endpoint.BinaryRequestReader{},
				Handler: func(request endpoint.Request[[]byte]) error {
					return as.app.ApplyGraph(request.Body)
				},
			},
		},
	}
}
