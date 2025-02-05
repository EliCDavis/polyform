package generator

import (
	"net/http"

	"github.com/EliCDavis/polyform/generator/endpoint"
)

func exampleGraphEndpoint(app *App) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.BodyMethod[string]{
				Request: endpoint.TextRequestReader{},
				Handler: func(request endpoint.Request[string]) error {
					err := app.ApplySchema(loadExample(request.Body))
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	}
}

func graphEndpoint(app *App) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.ResponseMethod[[]byte]{
				ResponseWriter: endpoint.BinaryResponseWriter{},
				Handler: func(r *http.Request) ([]byte, error) {
					return app.Schema(), nil
				},
			},

			http.MethodPost: endpoint.BodyMethod[[]byte]{
				Request: endpoint.BinaryRequestReader{},
				Handler: func(request endpoint.Request[[]byte]) error {
					err := app.ApplySchema(request.Body)
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	}
}
