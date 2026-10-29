package edit

import (
	"net/http"
	"strings"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/persistence"
)

func exampleGraphEndpoint(as *Server) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.BodyMethod[string]{
				Request: endpoint.TextRequestReader{},
				Handler: func(request endpoint.Request[string]) error {
					as.showNewGraphPopup = false
					data, err := loadExample(request.Body)
					if err != nil {
						return err
					}
					return as.Graph.ApplyAppSchema(data)
				},
			},
		},
	}
}

func newGraphEndpoint(editServer *Server) endpoint.Handler {
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
					editServer.showNewGraphPopup = false
					editServer.Graph.Reset()
					editServer.Graph.SetDetails(graph.Details{
						Name:        clean(request.Body.Name, "New Graph"),
						Description: clean(request.Body.Description, ""),
						Version:     clean(request.Body.Version, "v0.0.1"),
						Authors:     []persistence.Author{{Name: clean(request.Body.Author, "")}},
					})
					return nil
				},
			},
		},
	}
}

type graphDetails struct {
	Name        string               `json:"name"`
	Version     string               `json:"version"`
	Description string               `json:"description"`
	Authors     []persistence.Author `json:"authors"`
}

func cleanAuthors(authors []persistence.Author) []persistence.Author {
	var cleaned []persistence.Author
	for _, author := range authors {
		name := strings.TrimSpace(author.Name)
		if name == "" {
			continue
		}
		var contacts []persistence.AuthorContact
		for _, contact := range author.ContactInfo {
			medium, value := strings.TrimSpace(contact.Medium), strings.TrimSpace(contact.Value)
			if medium == "" && value == "" {
				continue
			}
			contacts = append(contacts, persistence.AuthorContact{Medium: medium, Value: value})
		}
		cleaned = append(cleaned, persistence.Author{Name: name, ContactInfo: contacts})
	}
	return cleaned
}

func graphDetailsEndpoint(as *Server, saver *GraphSaver) endpoint.Handler {
	read := func() graphDetails {
		details := as.Graph.Details()
		return graphDetails{
			Name:        details.Name,
			Version:     details.Version,
			Description: details.Description,
			Authors:     details.Authors,
		}
	}
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.JsonResponseMethod(func(r *http.Request) (graphDetails, error) {
				return read(), nil
			}),
			http.MethodPost: endpoint.JsonMethod(func(request endpoint.Request[graphDetails]) (graphDetails, error) {
				as.Graph.SetDetails(graph.Details{
					Name:        strings.TrimSpace(request.Body.Name),
					Version:     strings.TrimSpace(request.Body.Version),
					Description: strings.TrimSpace(request.Body.Description),
					Authors:     cleanAuthors(request.Body.Authors),
				})
				saver.Save()
				return read(), nil
			}),
		},
	}
}

func graphEndpoint(as *Server) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.ResponseMethod[[]byte]{
				ResponseWriter: endpoint.BinaryResponseWriter{},
				Handler: func(r *http.Request) ([]byte, error) {
					return as.Graph.EncodeToAppSchema()
				},
			},

			http.MethodPost: endpoint.BodyMethod[[]byte]{
				Request: endpoint.BinaryRequestReader{},
				Handler: func(request endpoint.Request[[]byte]) error {
					as.showNewGraphPopup = false
					return as.Graph.ApplyAppSchema(request.Body)
				},
			},
		},
	}
}
