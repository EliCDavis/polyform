package generator

import (
	"net/http"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/schema"
)

func exampleGraphEndpoint(app *App, as *AppServer) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.BodyMethod[string]{
				Request: endpoint.TextRequestReader{},
				Handler: func(request endpoint.Request[string]) error {
					as.showNewGraphPopup = false
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

func newGraphEndpoint(app *App, as *AppServer) endpoint.Handler {
	type NewGraph struct {
		Name        string `json:"name"`
		Author      string `json:"author"`
		Description string `json:"description"`
		Version     string `json:"version"`
	}

	clean := func(in, fallback string) string {
		cleaned := strings.TrimSpace(in)
		if cleaned == "" {
			return fallback
		}
		return cleaned
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.BodyMethod[NewGraph]{
				Request: endpoint.JsonRequestReader[NewGraph]{},
				Handler: func(request endpoint.Request[NewGraph]) error {
					as.showNewGraphPopup = false
					app.Name = clean(request.Body.Name, "New Graph")
					cleanedAuthor := clean(request.Body.Author, "")
					if cleanedAuthor != "" {
						app.Authors = []schema.Author{{Name: cleanedAuthor}}
					}
					app.Description = clean(request.Body.Description, "")
					app.Version = clean(request.Body.Version, "v0.0.1")
					app.graphInstance.Reset()
					return nil
				},
			},
		},
	}
}

func graphEndpoint(app *App, as *AppServer) endpoint.Handler {
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
					as.showNewGraphPopup = false
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
