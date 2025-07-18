package endpoint

import (
	"net/http"
)

type Handler struct {
	Methods map[string]Method
}

func (se Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method, ok := se.Methods[r.Method]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", string(method.ContentType()))
	method.Handle(w, r)
}
