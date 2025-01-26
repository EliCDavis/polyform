package generator

import (
	"net/http"
	"path"

	"github.com/EliCDavis/polyform/generator/endpoint"
)

func producerNameEndpoint(as *AppServer) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.BodyMethod[string]{
				Request: endpoint.TextRequestReader{},
				Handler: func(req endpoint.Request[string]) error {
					producerId := path.Base(req.Url)
					as.app.graphInstance.SetNodeAsProducer(producerId, req.Body)
					as.incModelVersion()
					as.AutosaveGraph()

					return nil
				},
			},
		},
	}
}
