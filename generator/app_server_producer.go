package generator

import (
	"fmt"
	"net/http"
	"path"

	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func producerNameEndpoint(as *AppServer) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.BodyMethod[string]{
				Request: endpoint.TextRequestReader{},
				Handler: func(req endpoint.Request[string]) error {
					producerId := path.Base(req.Url)

					producerNode := as.app.FindNodeByID(producerId)

					if producerNode == nil {
						panic(fmt.Errorf("no node exists with id %q", producerId))
					}

					// TODO: We need to allow users to specify which output port
					// that is the actuall artifact. can't rely on "Out"
					outPortVals := refutil.CallFuncValuesOfType(producerNode, "Out")
					ref := outPortVals[0].(nodes.NodeOutput[artifact.Artifact])
					if ref == nil {
						panic(fmt.Errorf("Couldn't find Out port on Node: %s", producerId))
					}

					// We need to check and remove previous references...
					for filename, producer := range as.app.Producers {

						if as.app.nodeIDs[producer.Node()] != producerId {
							continue
						}

						// TODO: This changes once we allow multiple output
						// port artifact. Need to specify port instead of "Out"
						if producer.Port() != "Out" {
							continue
						}

						delete(as.app.Producers, filename)
					}

					as.app.Producers[req.Body] = ref

					as.incModelVersion()
					as.AutosaveGraph()

					return nil
				},
			},
		},
	}
}
