package edit

import (
	"fmt"
	"net/http"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

func undoVerb(method string) string {
	switch method {
	case http.MethodDelete:
		return "Delete"
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return "Update"
	}
	return "Change"
}

// A request that answers with a failure is rolled back, not kept as a step.
func undoable(g *graph.Instance, noun string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		recorder := &statusRecorder{ResponseWriter: w}
		_ = g.History().Transact(undoVerb(r.Method)+" "+noun, func() error {
			next.ServeHTTP(recorder, r)
			if recorder.status >= http.StatusBadRequest {
				return fmt.Errorf("%s %s answered %d", r.Method, r.URL.Path, recorder.status)
			}
			return nil
		})
	})
}

type historyResponse struct {
	Undo []string `json:"undo"`
	Redo []string `json:"redo"`

	Applied string `json:"applied,omitempty"`
}

func historyEndpoint(g *graph.Instance, saver *GraphSaver) endpoint.Handler {
	state := func(applied string) historyResponse {
		h := g.History().Steps()
		return historyResponse{Undo: h.Undo, Redo: h.Redo, Applied: applied}
	}

	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.ResponseMethod[historyResponse]{
				ResponseWriter: endpoint.JsonResponseWriter[historyResponse]{},
				Handler: func(r *http.Request) (historyResponse, error) {
					return state(""), nil
				},
			},
		},
	}
}

func undoEndpoint(g *graph.Instance, saver *GraphSaver) endpoint.Handler {
	return historyStepEndpoint(g, saver, g.History().Undo)
}

func redoEndpoint(g *graph.Instance, saver *GraphSaver) endpoint.Handler {
	return historyStepEndpoint(g, saver, g.History().Redo)
}

func historyStepEndpoint(g *graph.Instance, saver *GraphSaver, walk func() (string, error)) endpoint.Handler {
	return endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodPost: endpoint.ResponseMethod[historyResponse]{
				ResponseWriter: endpoint.JsonResponseWriter[historyResponse]{},
				Handler: func(r *http.Request) (historyResponse, error) {
					applied, err := walk()
					if err != nil {
						return historyResponse{}, err
					}
					saver.Save()
					h := g.History().Steps()
					return historyResponse{Undo: h.Undo, Redo: h.Redo, Applied: applied}, nil
				},
			},
		},
	}
}
