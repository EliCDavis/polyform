package generator

import (
	"net/http"
	"path"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
)

func producerNameEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.BodyMethod[string]{
				Request: endpoint.TextRequestReader{},
				Handler: func(req endpoint.Request[string]) error {
					producerId := path.Base(req.Url)
					graphInstance.SetNodeAsProducer(producerId, req.Body)
					saver.Save()

					return nil
				},
			},
		},
	}
}
