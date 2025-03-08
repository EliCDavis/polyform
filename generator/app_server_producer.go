package generator

import (
	"net/http"
	"path"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
)

func producerNameEndpoint(graphInstance *graph.Instance, saver *GraphSaver) endpoint.Handler {

	type SetProducer struct {
		NodePort string
		Producer string
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.BodyMethod[SetProducer]{
				Request: endpoint.JsonRequestReader[SetProducer]{},
				Handler: func(req endpoint.Request[SetProducer]) error {
					nodeId := path.Base(req.Url)
					graphInstance.SetNodeAsProducer(nodeId, req.Body.NodePort, req.Body.Producer)
					saver.Save()

					return nil
				},
			},
		},
	}
}
