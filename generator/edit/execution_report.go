package edit

import (
	"net/http"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/schema"
)

func executionReportEndpoint(as *Server) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.JsonResponseMethod(func(r *http.Request) (schema.GraphExecutionReport, error) {
				return as.Graph.ExecutionReport(), nil
			}),
		},
	}
}
