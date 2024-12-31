package generator

import (
	"fmt"
	"net/http"
	"path"

	"github.com/EliCDavis/polyform/generator/endpoint"
)

func profileEndpoint(as *AppServer) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{

			http.MethodPost: endpoint.BodyMethod[[]byte]{
				Request: endpoint.BinaryRequestReader{},
				Handler: func(request endpoint.Request[[]byte]) error {
					as.producerLock.Lock()
					defer as.producerLock.Unlock()

					profileID := path.Base(request.Url)
					_, err := as.ApplyMessage(profileID, request.Body)
					if err != nil {
						return err
					}
					as.incModelVersion()
					return nil
				},
			},

			http.MethodGet: endpoint.ResponseMethod[[]byte]{
				ResponseWriter: endpoint.BinaryResponseWriter{},
				Handler: func(r *http.Request) ([]byte, error) {
					as.producerLock.Lock()
					defer as.producerLock.Unlock()

					profileID := path.Base(r.URL.Path)
					n, ok := as.app.Schema().Nodes[profileID]
					if !ok {
						return nil, fmt.Errorf("no node registered with ID: '%s'", profileID)
					}
					return n.parameter.ToMessage(), nil
				},
			},
		},
	}
}
